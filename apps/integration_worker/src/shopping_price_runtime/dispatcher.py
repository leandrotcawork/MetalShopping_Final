from __future__ import annotations

from .http.strategies import execute_http_runtime
from .lookup import select_lookup_term
from .models import LookupInputs, RuntimeObservation, SupplierRuntimeConfig, SupplierSignal
from .playwright.strategies import execute_playwright_runtime
from .shared.parsing import safe_float, safe_str


def execute(
    *,
    config: SupplierRuntimeConfig,
    inputs: LookupInputs,
    base_price: float,
    signal: SupplierSignal | None,
) -> RuntimeObservation:
    family = config.family
    strategy = safe_str(config.config_json.get("strategy"), "").lower()
    if strategy == "":
        strategy = f"{family}.mock.v1"

    lookup_term = select_lookup_term(config=config, inputs=inputs, signal=signal)

    if family == "http":
        return execute_http_runtime(
            config=config,
            strategy=strategy,
            product_id=inputs.product_id,
            lookup_term=lookup_term,
            base_price=base_price,
            signal=signal,
        )
    if family == "playwright":
        return execute_playwright_runtime(
            config=config,
            strategy=strategy,
            lookup_term=lookup_term,
            base_price=base_price,
        )
    return RuntimeObservation(
        "ERROR",
        safe_float(base_price, base_price),
        config.supplier_code.lower(),
        config.execution_kind,
        None,
        f"unsupported_driver_family:{family}",
        strategy,
        lookup_term,
    )
