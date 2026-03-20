from __future__ import annotations

from ..models import RuntimeObservation, SupplierRuntimeConfig
from ..shared.parsing import safe_float, safe_str


def execute_playwright_runtime(
    *,
    config: SupplierRuntimeConfig,
    strategy: str,
    lookup_term: str,
    base_price: float,
) -> RuntimeObservation:
    start_url = safe_str(config.config_json.get("startUrl"), "")
    search_url = safe_str(config.config_json.get("searchUrl"), "")
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
        return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "PLAYWRIGHT", None, "playwright_runtime_not_enabled", strategy, lookup_term)

    return RuntimeObservation("ERROR", safe_float(base_price, base_price), seller_default, "PLAYWRIGHT", None, f"unsupported_playwright_strategy:{strategy}", strategy, lookup_term)
