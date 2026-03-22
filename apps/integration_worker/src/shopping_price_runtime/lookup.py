from __future__ import annotations

from .models import LookupInputs, SupplierRuntimeConfig, SupplierSignal


def select_lookup_term(
    *,
    config: SupplierRuntimeConfig,
    inputs: LookupInputs,
    signal: SupplierSignal | None,
) -> str:
    ref = str(inputs.product_reference or "").strip()
    ean = str(inputs.product_ean or "").strip()

    if signal is not None:
        # Only honor stored lookup_mode when a stable/manual signal exists.
        if signal.manual_override or (signal.product_url is not None and signal.product_url.strip() != ""):
            if signal.lookup_mode == "EAN" and ean:
                return ean
            if signal.lookup_mode == "REFERENCE" and ref:
                return ref

    if config.lookup_policy.upper() == "EAN_FIRST":
        if ean:
            return ean
        if ref:
            return ref
    else:
        if ref:
            return ref
        if ean:
            return ean

    return str(inputs.product_id or "").strip()
