from __future__ import annotations

import base64
import json
import re
from dataclasses import dataclass
from html.parser import HTMLParser
from typing import Any

from ..models import SupplierRuntimeConfig


def safe_float(value: Any, default: float) -> float:
    try:
        parsed = float(value)
    except (TypeError, ValueError):
        return default
    if parsed < 0:
        return 0.0
    return round(parsed, 2)


def safe_int(value: Any, default: int, minimum: int, maximum: int) -> int:
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        return default
    if parsed < minimum:
        return minimum
    if parsed > maximum:
        return maximum
    return parsed


def safe_str(value: Any, default: str) -> str:
    text = str(value).strip() if value is not None else ""
    return text if text != "" else default


@dataclass(frozen=True)
class HtmlSearchCard:
    title: str
    full_price: float | None
    sale_price: float | None

    @property
    def effective_price(self) -> float | None:
        if self.sale_price is not None:
            return self.sale_price
        return self.full_price


@dataclass(frozen=True)
class HtmlSearchCardProfile:
    profile_name: str
    card_tag: str
    card_class_tokens: tuple[str, ...]
    title_tag: str
    full_price_class_tokens: tuple[str, ...]
    sale_price_class_tokens: tuple[str, ...]
    minimum_price: float = 0.0
    fallback_to_regex: bool = True


def _normalize_html_text(text: str) -> str:
    return re.sub(r"\s+", " ", str(text or "")).strip()


def _normalize_class_tokens(value: Any) -> tuple[str, ...]:
    if isinstance(value, (list, tuple)):
        tokens = [safe_str(item, "").lower() for item in value]
    else:
        tokens = [token.strip().lower() for token in safe_str(value, "").split()]
    return tuple(token for token in tokens if token)


def _matches_required_classes(class_text: str, required_tokens: tuple[str, ...]) -> bool:
    if len(required_tokens) == 0:
        return False
    available = {token for token in class_text.lower().split() if token}
    return all(token in available for token in required_tokens)


class _HtmlSearchCardParser(HTMLParser):
    def __init__(self, profile: HtmlSearchCardProfile) -> None:
        super().__init__(convert_charrefs=True)
        self.profile = profile
        self.cards: list[HtmlSearchCard] = []
        self._inside_card = False
        self._card_depth = 0
        self._capture_target: str | None = None
        self._capture_buffer: list[str] = []
        self._current: dict[str, str] = {"title": "", "full": "", "sale": ""}

    def _flush_capture(self) -> None:
        if self._capture_target is None:
            return
        value = _normalize_html_text(" ".join(self._capture_buffer))
        if value != "":
            previous = self._current.get(self._capture_target, "")
            self._current[self._capture_target] = f"{previous} {value}".strip() if previous else value
        self._capture_target = None
        self._capture_buffer = []

    def _finalize_card(self) -> None:
        self._flush_capture()
        title = _normalize_html_text(self._current.get("title", ""))
        full_text = _normalize_html_text(self._current.get("full", ""))
        sale_text = _normalize_html_text(self._current.get("sale", ""))
        self.cards.append(
            HtmlSearchCard(
                title=title,
                full_price=_parse_html_card_price(full_text),
                sale_price=_parse_html_card_price(sale_text),
            )
        )
        self._inside_card = False
        self._card_depth = 0
        self._capture_target = None
        self._capture_buffer = []
        self._current = {"title": "", "full": "", "sale": ""}

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        tag_name = safe_str(tag, "").lower()
        attrs_map = {safe_str(key, "").lower(): safe_str(value, "") for key, value in attrs}
        class_text = attrs_map.get("class", "")

        if tag_name == self.profile.card_tag:
            if not self._inside_card and _matches_required_classes(class_text, self.profile.card_class_tokens):
                self._inside_card = True
                self._card_depth = 1
                self._capture_target = None
                self._capture_buffer = []
                self._current = {"title": "", "full": "", "sale": ""}
                return
            if self._inside_card:
                self._card_depth += 1

        if not self._inside_card:
            return

        if tag_name == self.profile.title_tag:
            self._flush_capture()
            self._capture_target = "title"
            self._capture_buffer = []
            return

        if tag_name == "span":
            if _matches_required_classes(class_text, self.profile.sale_price_class_tokens):
                self._flush_capture()
                self._capture_target = "sale"
                self._capture_buffer = []
                return
            if _matches_required_classes(class_text, self.profile.full_price_class_tokens):
                self._flush_capture()
                self._capture_target = "full"
                self._capture_buffer = []

    def handle_data(self, data: str) -> None:
        if self._inside_card and self._capture_target is not None:
            self._capture_buffer.append(str(data or ""))

    def handle_endtag(self, tag: str) -> None:
        tag_name = safe_str(tag, "").lower()
        if self._inside_card and self._capture_target == "title" and tag_name == self.profile.title_tag:
            self._flush_capture()
        elif self._inside_card and self._capture_target in {"full", "sale"} and tag_name == "span":
            self._flush_capture()

        if self._inside_card and tag_name == self.profile.card_tag:
            self._card_depth -= 1
            if self._card_depth <= 0:
                self._finalize_card()


def _parse_html_card_price(text: str) -> float | None:
    token = safe_str(text, "")
    if token == "":
        return None
    token = re.sub(r"[^0-9,.\-]", "", token)
    if token == "":
        return None
    if "," in token and "." in token:
        token = token.replace(".", "").replace(",", ".")
    elif "," in token:
        token = token.replace(",", ".")
    try:
        return float(token)
    except (TypeError, ValueError):
        return None


def extract_html_search_cards(
    html_text: str,
    *,
    profile: HtmlSearchCardProfile,
    top_n: int,
) -> list[HtmlSearchCard]:
    parser = _HtmlSearchCardParser(profile)
    try:
        parser.feed(str(html_text or ""))
        parser.close()
    except Exception:
        return []

    unique_cards: list[HtmlSearchCard] = []
    seen: set[tuple[str, float | None, float | None]] = set()
    for card in parser.cards:
        key = (
            safe_str(card.title, "").lower(),
            round(card.full_price, 4) if card.full_price is not None else None,
            round(card.sale_price, 4) if card.sale_price is not None else None,
        )
        if key in seen:
            continue
        seen.add(key)
        unique_cards.append(card)

    limit = max(0, int(top_n or 0))
    if limit <= 0:
        return unique_cards
    return unique_cards[:limit]


def b64json(obj: dict[str, Any]) -> str:
    raw = json.dumps(obj, separators=(",", ":")).encode("utf-8")
    return base64.b64encode(raw).decode("utf-8")


def json_path_get(payload: Any, path: str) -> Any:
    if path.strip() == "":
        return None
    current = payload
    for token in path.split("."):
        key = token.strip()
        if key == "":
            return None
        if isinstance(current, dict):
            if key not in current:
                return None
            current = current[key]
            continue
        if isinstance(current, list):
            try:
                index = int(key)
            except ValueError:
                return None
            if index < 0 or index >= len(current):
                return None
            current = current[index]
            continue
        return None
    return current


def decode_price_text(value: Any, fallback: float) -> float:
    if isinstance(value, (float, int)):
        return safe_float(value, fallback)
    text = safe_str(value, "")
    if text == "":
        return safe_float(fallback, fallback)
    normalized = text.replace("R$", "").replace(" ", "")
    if "," in normalized and "." in normalized:
        normalized = normalized.replace(".", "").replace(",", ".")
    elif "," in normalized:
        normalized = normalized.replace(",", ".")
    filtered = "".join(ch for ch in normalized if ch.isdigit() or ch in {".", "-"})
    return safe_float(filtered, fallback)


def extract_by_paths(
    body_json: dict[str, Any],
    config: SupplierRuntimeConfig,
    base_price: float,
    seller_default: str,
    channel_default: str,
) -> tuple[float, str, str]:
    price_path = safe_str(config.config_json.get("pricePath"), "")
    seller_path = safe_str(config.config_json.get("sellerPath"), "")
    channel_path = safe_str(config.config_json.get("channelPath"), "")

    raw_price = json_path_get(body_json, price_path) if price_path != "" else None
    if raw_price is None:
        raw_price = body_json.get("price", body_json.get("observed_price", body_json.get("amount")))
    observed_price = decode_price_text(raw_price, base_price)

    raw_seller = json_path_get(body_json, seller_path) if seller_path != "" else None
    if raw_seller is None:
        raw_seller = body_json.get("seller", body_json.get("seller_name"))
    seller_name = safe_str(raw_seller, seller_default)

    raw_channel = json_path_get(body_json, channel_path) if channel_path != "" else None
    if raw_channel is None:
        raw_channel = body_json.get("channel")
    channel = safe_str(raw_channel, channel_default).upper()

    return observed_price, seller_name, channel
