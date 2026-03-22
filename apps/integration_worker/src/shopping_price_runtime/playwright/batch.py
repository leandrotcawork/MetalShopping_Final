from __future__ import annotations

import asyncio
import re
import time
from dataclasses import dataclass
from typing import Any

from ..lookup import select_lookup_term
from ..models import LookupInputs, RuntimeObservation, SupplierRuntimeConfig, SupplierSignal
from ..shared.parsing import safe_float, safe_int, safe_str
from .strategies import _build_candidate_urls, _detect_block_reason, _price_from_text


@dataclass(frozen=True)
class BatchItem:
    index: int
    product_id: str
    lookup_term: str
    base_price: float
    signal: SupplierSignal | None


@dataclass(frozen=True)
class BatchResult:
    index: int
    observation: RuntimeObservation
    elapsed_s: float


async def _selector_text(page: Any, selector: str, timeout_ms: int) -> str:
    selector = selector.strip()
    if selector == "":
        return ""
    try:
        loc = page.locator(selector).first
        value = await loc.text_content(timeout=timeout_ms)
        return safe_str(value, "")
    except Exception:
        return ""


def _with_perf(note: str, nav_s: float, sel_s: float, total_s: float, sel_timeout_ms: int) -> str:
    suffix = f" t_nav_s={nav_s:.3f} t_sel_s={sel_s:.3f} t_total_s={total_s:.3f} t_sel_ms={sel_timeout_ms}"
    if note:
        return f"{note} |{suffix}"
    return suffix.strip()


async def _execute_single(
    *,
    page: Any,
    config: SupplierRuntimeConfig,
    item: BatchItem,
    strategy: str,
    seller_default: str,
    price_regex: re.Pattern[str],
    selectors: dict[str, Any],
    timeout_seconds: int,
    max_retries: int,
    wait_until: str,
) -> BatchResult:
    fallback_enabled = bool(config.config_json.get("fallbackSearchEnabled", False))
    if item.signal is not None and not item.signal.allow_url_discovery and safe_str(item.signal.product_url, "").strip() == "":
        fallback_enabled = False

    candidate_urls = _build_candidate_urls(
        config=config,
        product_id=item.product_id,
        lookup_term=item.lookup_term,
        signal=item.signal,
        fallback_enabled=fallback_enabled,
    )
    if len(candidate_urls) == 0:
        observation = RuntimeObservation(
            "ERROR",
            safe_float(item.base_price, item.base_price),
            seller_default,
            "PLAYWRIGHT",
            None,
            "playwright_runtime_url_missing",
            strategy,
            item.lookup_term,
        )
        return BatchResult(item.index, observation, 0.0)

    last_observation: RuntimeObservation | None = None
    selector_timeout_ms = int(max(200, min(1500, timeout_seconds * 1000 * 0.1)))

    for attempt in range(1, max_retries + 1):
        for url in candidate_urls:
            total_start = time.perf_counter()
            nav_s = 0.0
            sel_s = 0.0
            response_status: int | None = None
            try:
                nav_start = time.perf_counter()
                response = await page.goto(url, timeout=timeout_seconds * 1000, wait_until=wait_until)
                if response is not None:
                    response_status = int(response.status)
                html_text = await page.content()
                nav_s = time.perf_counter() - nav_start

                block_reason = _detect_block_reason(html_text, await page.title())
                if block_reason != "":
                    last_observation = RuntimeObservation(
                        "ERROR",
                        safe_float(item.base_price, item.base_price),
                        seller_default,
                        "PLAYWRIGHT",
                        response_status,
                        _with_perf(
                            f"playwright_blocked:{block_reason}:attempt={attempt}",
                            nav_s,
                            sel_s,
                            time.perf_counter() - total_start,
                            selector_timeout_ms,
                        ),
                        strategy,
                        item.lookup_term,
                    )
                    continue

                sel_start = time.perf_counter()
                price_text = ""
                match = price_regex.search(html_text)
                if match is not None:
                    price_text = safe_str(match.group(1), "")
                if price_text == "":
                    price_text = await _selector_text(
                        page,
                        safe_str(selectors.get("price"), ""),
                        selector_timeout_ms,
                    )
                seller_text = await _selector_text(
                    page,
                    safe_str(selectors.get("seller"), ""),
                    selector_timeout_ms,
                )
                channel_text = await _selector_text(
                    page,
                    safe_str(selectors.get("channel"), ""),
                    selector_timeout_ms,
                )
                sel_s = time.perf_counter() - sel_start

                observed_price = _price_from_text(price_text, 0.0)
                if observed_price <= 0:
                    last_observation = RuntimeObservation(
                        "NOT_FOUND",
                        safe_float(item.base_price, item.base_price),
                        safe_str(seller_text, seller_default),
                        safe_str(channel_text, "PLAYWRIGHT").upper(),
                        response_status,
                        _with_perf(
                            f"playwright_price_not_found:attempt={attempt}",
                            nav_s,
                            sel_s,
                            time.perf_counter() - total_start,
                            selector_timeout_ms,
                        ),
                        strategy,
                        item.lookup_term,
                    )
                    continue

                observation = RuntimeObservation(
                    "OK",
                    safe_float(observed_price, item.base_price),
                    safe_str(seller_text, seller_default),
                    safe_str(channel_text, "PLAYWRIGHT").upper(),
                    response_status,
                    _with_perf(
                        f"playwright_pdp_runtime:attempt={attempt}",
                        nav_s,
                        sel_s,
                        time.perf_counter() - total_start,
                        selector_timeout_ms,
                    ),
                    strategy,
                    item.lookup_term,
                )
                return BatchResult(item.index, observation, round(time.perf_counter() - total_start, 3))
            except Exception as exc:
                last_observation = RuntimeObservation(
                    "ERROR",
                    safe_float(item.base_price, item.base_price),
                    seller_default,
                    "PLAYWRIGHT",
                    response_status,
                    _with_perf(
                        f"playwright_error:{str(exc)[:220]}:attempt={attempt}",
                        nav_s,
                        sel_s,
                        time.perf_counter() - total_start,
                        selector_timeout_ms,
                    ),
                    strategy,
                    item.lookup_term,
                )
        if attempt < max_retries:
            await asyncio.sleep(min(1.0, 0.2 * attempt))

    if last_observation is None:
        last_observation = RuntimeObservation(
            "ERROR",
            safe_float(item.base_price, item.base_price),
            seller_default,
            "PLAYWRIGHT",
            None,
            "playwright_runtime_failed_without_observation",
            strategy,
            item.lookup_term,
        )
    return BatchResult(item.index, last_observation, 0.0)


async def _run_batch_async(
    *,
    config: SupplierRuntimeConfig,
    items: list[BatchItem],
    tabs: int,
) -> list[BatchResult]:
    if not items:
        return []

    try:
        from playwright.async_api import async_playwright
    except Exception as exc:
        seller_default = safe_str(config.config_json.get("sellerName"), config.supplier_code.lower())
        return [
            BatchResult(
                item.index,
                RuntimeObservation(
                    "ERROR",
                    safe_float(item.base_price, item.base_price),
                    seller_default,
                    "PLAYWRIGHT",
                    None,
                    f"playwright_not_installed:{str(exc)[:220]}",
                    safe_str(config.config_json.get("strategy"), ""),
                    item.lookup_term,
                ),
                0.0,
            )
            for item in items
        ]

    timeout_seconds = safe_int(config.config_json.get("timeoutSeconds"), 30, 1, 120)
    max_retries = safe_int(config.config_json.get("maxRetries"), 2, 1, 8)
    headless = bool(config.config_json.get("headless", True))
    wait_until = safe_str(config.config_json.get("waitUntil"), "domcontentloaded")
    locale = safe_str(config.config_json.get("locale"), "pt-BR")
    seller_default = safe_str(config.config_json.get("sellerName"), config.supplier_code.lower())
    strategy = safe_str(config.config_json.get("strategy"), "").lower() or "playwright.pdp_first.v1"

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
        return [
            BatchResult(
                item.index,
                RuntimeObservation(
                    "ERROR",
                    safe_float(item.base_price, item.base_price),
                    seller_default,
                    "PLAYWRIGHT",
                    None,
                    "playwright.pdp_first.v1 requires pdpSelectors object",
                    strategy,
                    item.lookup_term,
                ),
                0.0,
            )
            for item in items
        ]

    tabs = max(1, min(int(tabs), len(items)))
    chunks: list[list[BatchItem]] = []
    base = len(items) // tabs
    rem = len(items) % tabs
    cursor = 0
    for i in range(tabs):
        take = base + (1 if i < rem else 0)
        chunk = items[cursor : cursor + take]
        cursor += take
        chunks.append(chunk)

    async with async_playwright() as pw:
        browser = await pw.chromium.launch(headless=headless)
        results: list[BatchResult] = []

        async def _worker(chunk: list[BatchItem]) -> None:
            if not chunk:
                return
            context = await browser.new_context(locale=locale)
            page = await context.new_page()
            try:
                for item in chunk:
                    result = await _execute_single(
                        page=page,
                        config=config,
                        item=item,
                        strategy=strategy,
                        seller_default=seller_default,
                        price_regex=price_regex,
                        selectors=selectors,
                        timeout_seconds=timeout_seconds,
                        max_retries=max_retries,
                        wait_until=wait_until,
                    )
                    results.append(result)
            finally:
                try:
                    await page.close()
                except Exception:
                    pass
                try:
                    await context.close()
                except Exception:
                    pass

        await asyncio.gather(*[_worker(chunk) for chunk in chunks if chunk])
        await browser.close()

    return results


def execute_playwright_pdp_first_batch(
    *,
    config: SupplierRuntimeConfig,
    items: list[BatchItem],
    tabs: int,
) -> list[BatchResult]:
    return asyncio.run(_run_batch_async(config=config, items=items, tabs=tabs))


def build_batch_items(
    *,
    config: SupplierRuntimeConfig,
    inputs: list[tuple[int, LookupInputs, float, SupplierSignal | None]],
) -> list[BatchItem]:
    out: list[BatchItem] = []
    for index, lookup_inputs, base_price, signal in inputs:
        lookup_term = select_lookup_term(config=config, inputs=lookup_inputs, signal=signal)
        out.append(
            BatchItem(
                index=index,
                product_id=lookup_inputs.product_id,
                lookup_term=lookup_term,
                base_price=base_price,
                signal=signal,
            )
        )
    return out
