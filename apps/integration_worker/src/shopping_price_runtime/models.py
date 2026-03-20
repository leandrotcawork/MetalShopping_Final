from __future__ import annotations

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class SupplierSignal:
    product_url: str | None
    lookup_mode: str
    manual_override: bool
    # URL discovery (search fallback) should be decided by orchestration using durable DB lifecycle fields.
    # Strategies should treat this as a simple capability flag.
    allow_url_discovery: bool = True


@dataclass(frozen=True)
class SupplierRuntimeConfig:
    supplier_code: str
    execution_kind: str
    lookup_policy: str
    family: str
    config_json: dict[str, Any]


@dataclass(frozen=True)
class LookupInputs:
    product_id: str
    product_reference: str
    product_ean: str


@dataclass(frozen=True)
class RuntimeObservation:
    item_status: str
    observed_price: float
    seller_name: str
    channel: str
    http_status: int | None
    notes: str | None
    strategy: str
    lookup_term: str
