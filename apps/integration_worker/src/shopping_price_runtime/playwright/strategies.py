from __future__ import annotations

import re
import time
import urllib.parse
from typing import Any

from ..models import RuntimeObservation, SupplierRuntimeConfig
from ..models import SupplierSignal
from ..shared.parsing import safe_float, safe_int, safe_str


def execute_playwright_runtime(
    *,
    config: SupplierRuntimeConfig,
    strategy: str,
    product_id: str,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
) -> RuntimeObservation:
    start_url = safe_str(config.config_json.get("startUrl"), "").strip()
    search_url = safe_str(config.config_json.get("searchUrl"), "").strip()
    seller_default = safe_str(config.config_json.get("sellerName"), config.supplier_code.lower())

    if strategy == "playwright.mock.v1":
        if not start_url.startswith("mock://"):
            return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "PLAYWRIGHT", None, "playwright.mock.v1 requires startUrl=mock://", strategy, lookup_term)
        offset = safe_float(config.config_json.get("mockPriceOffset"), 0.0)
        return RuntimeObservation("OK", safe_float(base_price + offset, base_price), seller_default, "PLAYWRIGHT_MOCK", 200, f"playwright_mock_runtime strategy={strategy}", strategy, lookup_term)

    if strategy == "playwright.pdp_first.v1":
        if start_url.startswith("mock://") or search_url.startswith("mock://"):
            offset = safe_float(config.config_json.get("mockPriceOffset"), 0.0)
            return RuntimeObservation("OK", safe_float(base_price + offset, base_price), seller_default, "PLAYWRIGHT_PDP_MOCK", 200, f"playwright_pdp_mock strategy={strategy}", strategy, lookup_term)
        return _execute_playwright_pdp_first_non_mock(
            config=config,
            strategy=strategy,
            product_id=product_id,
            lookup_term=lookup_term,
            base_price=base_price,
            signal=signal,
            seller_default=seller_default,
        )

    return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "PLAYWRIGHT", None, f"unsupported_playwright_strategy:{strategy}", strategy, lookup_term)


def _execute_playwright_pdp_first_non_mock(
    *,
    config: SupplierRuntimeConfig,
    strategy: str,
    product_id: str,
    lookup_term: str,
    base_price: float,
    signal: SupplierSignal | None,
    seller_default: str,
) -> RuntimeObservation:
    try:
        from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
        from playwright.sync_api import sync_playwright
    except Exception as exc:
        return RuntimeObservation(
            "ERROR",
            safe_float(base_price, base_price),
            seller_default,
            "PLAYWRIGHT",
            None,
            f"playwright_not_installed:{str(exc)[:220]}",
            strategy,
            lookup_term,
        )

    timeout_seconds = safe_int(config.config_json.get("timeoutSeconds"), 30, 1, 120)
    max_retries = safe_int(config.config_json.get("maxRetries"), 2, 1, 8)
    headless = bool(config.config_json.get("headless", True))
    wait_until = safe_str(config.config_json.get("waitUntil"), "domcontentloaded")
    locale = safe_str(config.config_json.get("locale"), "pt-BR")
    fallback_enabled = bool(config.config_json.get("fallbackSearchEnabled", False))
    if signal is not None and not signal.allow_url_discovery and safe_str(signal.product_url, "").strip() == "":
        fallback_enabled = False
    price_regex_raw = safe_str(
        config.config_json.get("priceRegex"),
        r"(\d{1,3}(?:\.\d{3})*,\d{2}|\d+(?:\.\d{2})?)",
    )
    try:
        price_regex = re.compile(price_regex_raw, flags=re.IGNORECASE | re.MULTILINE)
    except Exception:
        price_regex = re.compile(r"(\d{1,3}(?:\.\d{3})*,\d{2}|\d+(?:\.\d{2})?)", flags=re.IGNORECASE | re.MULTILINE)

    selectors = config.config_json.get("pdpSelectors")
    if not isinstance(selectors, dict):
        return RuntimeObservation(
            "ERROR",
            safe_float(base_price, base_price),
            seller_default,
            "PLAYWRIGHT",
            None,
            "playwright.pdp_first.v1 requires pdpSelectors object",
            strategy,
            lookup_term,
        )

    candidate_urls = _build_candidate_urls(
        config=config,
        product_id=product_id,
        lookup_term=lookup_term,
        signal=signal,
        fallback_enabled=fallback_enabled,
    )
    if len(candidate_urls) == 0:
        return RuntimeObservation(
            "ERROR",
            safe_float(base_price, base_price),
            seller_default,
            "PLAYWRIGHT",
            None,
            "playwright_runtime_url_missing",
            strategy,
            lookup_term,
        )

    last_observation: RuntimeObservation | None = None

    for attempt in range(1, max_retries + 1):
        for url in candidate_urls:
            with sync_playwright() as playwright:
                browser = playwright.chromium.launch(headless=headless)
                context = browser.new_context(locale=locale)
                page = context.new_page()
                response_status: int | None = None
                try:
                    response = page.goto(url, timeout=timeout_seconds * 1000, wait_until=wait_until)
                    if response is not None:
                        response_status = int(response.status)
                    html_text = page.content()
                    block_reason = _detect_block_reason(html_text, page.title())
                    if block_reason != "":
                        last_observation = RuntimeObservation(
                            "ERROR",
                            safe_float(base_price, base_price),
                            seller_default,
                            "PLAYWRIGHT",
                            response_status,
                            f"playwright_blocked:{block_reason}:attempt={attempt}",
                            strategy,
                            lookup_term,
                        )
                        continue

                    price_text = _selector_text(page, safe_str(selectors.get("price"), ""), timeout_seconds)
                    seller_text = _selector_text(page, safe_str(selectors.get("seller"), ""), timeout_seconds)
                    channel_text = _selector_text(page, safe_str(selectors.get("channel"), ""), timeout_seconds)

                    if price_text == "":
                        match = price_regex.search(html_text)
                        if match is not None:
                            price_text = safe_str(match.group(1), "")

                    observed_price = _price_from_text(price_text, 0.0)
                    if observed_price <= 0:
                        last_observation = RuntimeObservation(
                            "NOT_FOUND",
                            safe_float(base_price, base_price),
                            safe_str(seller_text, seller_default),
                            safe_str(channel_text, "PLAYWRIGHT").upper(),
                            response_status,
                            f"playwright_price_not_found:attempt={attempt}",
                            strategy,
                            lookup_term,
                        )
                        continue

                    return RuntimeObservation(
                        "OK",
                        safe_float(observed_price, base_price),
                        safe_str(seller_text, seller_default),
                        safe_str(channel_text, "PLAYWRIGHT").upper(),
                        response_status,
                        f"playwright_pdp_runtime:attempt={attempt}",
                        strategy,
                        lookup_term,
                    )
                except PlaywrightTimeoutError:
                    last_observation = RuntimeObservation(
                        "ERROR",
                        safe_float(base_price, base_price),
                        seller_default,
                        "PLAYWRIGHT",
                        response_status,
                        f"playwright_timeout:attempt={attempt}",
                        strategy,
                        lookup_term,
                    )
                except Exception as exc:
                    last_observation = RuntimeObservation(
                        "ERROR",
                        safe_float(base_price, base_price),
                        seller_default,
                        "PLAYWRIGHT",
                        response_status,
                        f"playwright_error:{str(exc)[:220]}:attempt={attempt}",
                        strategy,
                        lookup_term,
                    )
                finally:
                    try:
                        context.close()
                    except Exception:
                        pass
                    try:
                        browser.close()
                    except Exception:
                        pass
        if attempt < max_retries:
            time.sleep(min(1.0, 0.2 * attempt))

    if last_observation is not None:
        return last_observation
    return RuntimeObservation(
        "ERROR",
        safe_float(base_price, base_price),
        seller_default,
        "PLAYWRIGHT",
        None,
        "playwright_runtime_failed_without_observation",
        strategy,
        lookup_term,
    )


def _build_candidate_urls(
    *,
    config: SupplierRuntimeConfig,
    product_id: str,
    lookup_term: str,
    signal: SupplierSignal | None,
    fallback_enabled: bool,
) -> list[str]:
    urls: list[str] = []
    if signal is not None:
        product_url = safe_str(signal.product_url, "")
        if product_url != "":
            urls.append(product_url)

    start_url = safe_str(config.config_json.get("startUrl"), "")
    if start_url != "":
        rendered = _render_url_template(start_url, config.supplier_code, product_id, lookup_term, signal)
        if rendered != "":
            urls.append(rendered)

    search_url = safe_str(config.config_json.get("searchUrl"), "")
    if fallback_enabled and search_url != "":
        rendered = _render_url_template(search_url, config.supplier_code, product_id, lookup_term, signal)
        if rendered != "":
            urls.append(rendered)

    deduped: list[str] = []
    seen = set()
    for url in urls:
        key = url.strip()
        if key == "" or key in seen:
            continue
        seen.add(key)
        deduped.append(key)
    return deduped


def _render_url_template(
    template: str,
    supplier_code: str,
    product_id: str,
    lookup_term: str,
    signal: SupplierSignal | None,
) -> str:
    lookup_mode = signal.lookup_mode if signal is not None else "REFERENCE"
    return (
        str(template)
        .replace("{supplier_code}", urllib.parse.quote(supplier_code, safe=""))
        .replace("{product_id}", urllib.parse.quote(product_id, safe=""))
        .replace("{term}", urllib.parse.quote(lookup_term, safe=""))
        .replace("{lookup_mode}", urllib.parse.quote(lookup_mode, safe=""))
    )


def _selector_text(page: Any, selector: str, timeout_seconds: int) -> str:
    selector = selector.strip()
    if selector == "":
        return ""
    try:
        loc = page.locator(selector).first
        value = loc.text_content(timeout=timeout_seconds * 1000)
        return safe_str(value, "")
    except Exception:
        return ""


def _detect_block_reason(html_text: str, title: str) -> str:
    haystack = f"{title}\n{html_text}".lower()
    markers = {
        "cloudflare": "cloudflare",
        "error 1015": "cloudflare_1015",
        "captcha": "captcha",
        "access denied": "access_denied",
        "temporarily blocked": "temporarily_blocked",
    }
    for marker, code in markers.items():
        if marker in haystack:
            return code
    return ""


def _price_from_text(value: str, fallback: float) -> float:
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
