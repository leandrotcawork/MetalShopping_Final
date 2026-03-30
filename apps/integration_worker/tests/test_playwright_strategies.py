from __future__ import annotations

import unittest

from shopping_price_runtime.models import SupplierRuntimeConfig, SupplierSignal
from shopping_price_runtime.playwright.strategies import (
    _is_static_start_url_without_signal,
    execute_playwright_runtime,
)


class PlaywrightStrategiesTest(unittest.TestCase):
    def test_pdp_first_without_signal_and_without_fallback_returns_not_found(self) -> None:
        config = SupplierRuntimeConfig(
            supplier_code="OBRA_FACIL",
            execution_kind="runtime",
            lookup_policy="REFERENCE_FIRST",
            family="playwright",
            config_json={
                "strategy": "playwright.pdp_first.v1",
                "sellerName": "obra facil",
                "startUrl": "https://lojaobrafacil.com.br/",
                "fallbackSearchEnabled": False,
                "pdpSelectors": {"price": "div.price-box span.price"},
            },
        )

        observation = execute_playwright_runtime(
            config=config,
            strategy="playwright.pdp_first.v1",
            product_id="prd_123",
            lookup_term="7891461490294",
            base_price=3720.55,
            signal=None,
        )

        self.assertEqual("NOT_FOUND", observation.item_status)
        self.assertEqual("playwright_product_url_missing_no_fallback", observation.notes)
        self.assertEqual(3720.55, observation.observed_price)

    def test_static_start_url_is_allowed_when_signal_has_product_url(self) -> None:
        config = SupplierRuntimeConfig(
            supplier_code="OBRA_FACIL",
            execution_kind="runtime",
            lookup_policy="REFERENCE_FIRST",
            family="playwright",
            config_json={"startUrl": "https://lojaobrafacil.com.br/"},
        )
        signal = SupplierSignal(
            product_url="https://lojaobrafacil.com.br/produto-x",
            lookup_mode="REFERENCE",
            manual_override=False,
            allow_url_discovery=False,
        )

        result = _is_static_start_url_without_signal(
            config=config,
            signal=signal,
            fallback_enabled=False,
        )

        self.assertFalse(result)

    def test_start_url_template_is_not_treated_as_static_homepage(self) -> None:
        config = SupplierRuntimeConfig(
            supplier_code="OBRA_FACIL",
            execution_kind="runtime",
            lookup_policy="REFERENCE_FIRST",
            family="playwright",
            config_json={"startUrl": "https://example.com/search?q={term}"},
        )

        result = _is_static_start_url_without_signal(
            config=config,
            signal=None,
            fallback_enabled=False,
        )

        self.assertFalse(result)


if __name__ == "__main__":
    unittest.main()
