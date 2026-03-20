from __future__ import annotations

from .models import SupplierRuntimeConfig
from .shared.parsing import safe_str


def build_runtime_headers(config: SupplierRuntimeConfig) -> dict[str, str]:
    headers = {"User-Agent": "metalshopping-shopping-worker/1.0"}
    raw_headers = config.config_json.get("headers")
    if isinstance(raw_headers, dict):
        for key, value in raw_headers.items():
            header_key = safe_str(key, "")
            header_value = safe_str(value, "")
            if header_key != "" and header_value != "":
                headers[header_key] = header_value
    return headers
