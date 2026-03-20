from __future__ import annotations

import base64
import json
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
