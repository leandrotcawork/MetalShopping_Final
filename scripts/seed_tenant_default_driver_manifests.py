from __future__ import annotations

import json
import os
import uuid
from dataclasses import dataclass
from typing import Any

import psycopg


@dataclass(frozen=True)
class ManifestSeed:
    supplier_code: str
    family: str
    config_json: dict[str, Any]


def _seller_name(supplier_code: str) -> str:
    return f"{supplier_code.strip().lower()}_marketplace"


def _manifest_id(supplier_code: str) -> str:
    return f"man_{supplier_code.strip().lower()}_{uuid.uuid4().hex[:12]}"


def _seeds() -> list[ManifestSeed]:
    # Values extracted from the legacy drivers:
    # - C:\Users\...\Documents\Nova pasta\MetalShopping\drivers\*_driver.py
    #
    # Keep this list small and explicit: these are the operational URLs/templates that
    # must be present for tenant_default to run real catalog smokes.
    sha_vtex_product_search_v3 = "31d3fa494df1fc41efef6d16dd96a96e6911b8aed7a037868699a1f3f4d365de"

    return [
        ManifestSeed(
            supplier_code="DEXCO",
            family="http",
            config_json={
                "strategy": "http.vtex_persisted_query.v1",
                "timeoutSeconds": 12,
                "maxRetries": 2,
                "sellerName": _seller_name("DEXCO"),
                "baseUrl": "https://www.lojadexco.com.br/deca/_v/segment/graphql/v1",
                "operationName": "productSearchV3",
                "sha256Hash": sha_vtex_product_search_v3,
                # Optional knobs (kept aligned with legacy intent).
                "skusFilter": "FIRST_AVAILABLE",
                "toN": 11,
                "includeVariant": False,
                "allowFallbackFirstProduct": True,
                "requireAvailableOffer": False,
                "preferredSellerName": "Loja Dexco",
            },
        ),
        ManifestSeed(
            supplier_code="TELHA_NORTE",
            family="http",
            config_json={
                "strategy": "http.vtex_persisted_query.v1",
                "timeoutSeconds": 12,
                "maxRetries": 2,
                "sellerName": _seller_name("TELHA_NORTE"),
                "baseUrl": "https://www.telhanorte.com.br/_v/segment/graphql/v1",
                "operationName": "productSearchV3",
                "sha256Hash": sha_vtex_product_search_v3,
                "skusFilter": "ALL_AVAILABLE",
                "toN": 39,
                "includeVariant": True,
                "allowFallbackFirstProduct": False,
                "requireAvailableOffer": True,
            },
        ),
        ManifestSeed(
            supplier_code="LEROY",
            family="http",
            config_json={
                "strategy": "http.leroy_search_sellers.v1",
                "timeoutSeconds": 12,
                "maxRetries": 2,
                "sellerName": _seller_name("LEROY"),
                "searchUrlTemplate": "https://www.leroymerlin.com.br/search?term={term}",
                "sellersUrlTemplate": "https://www.leroymerlin.com.br/api/v3/products/{product_id}/sellers",
                "region": "uberlandia",
                "sellerPickStrategy": "selected",
            },
        ),
        ManifestSeed(
            supplier_code="ABC",
            family="http",
            config_json={
                "strategy": "http.html_dom_first_card.v1",
                "timeoutSeconds": 12,
                "maxRetries": 2,
                "sellerName": _seller_name("ABC"),
                "searchUrlTemplate": "https://www.abcdaconstrucao.com.br/busca?busca={term}",
                "cardRootHint": "spots-list",
                # Optional hints to stabilize extraction (legacy uses #spots-list).
                "cardItemHint": "spot__content",
                "titleHint": "spot__content-title",
                "priceHint": "spot__content-price",
                "calculatedPriceHint": "precoCalculado",
                "pricePriority": "calculated_first",
            },
        ),
        ManifestSeed(
            supplier_code="CONDEC",
            family="http",
            config_json={
                "strategy": "http.html_search.v1",
                "timeoutSeconds": 12,
                "maxRetries": 2,
                "sellerName": _seller_name("CONDEC"),
                # Legacy CONDEC driver uses a query param palavra_busca (reference-first).
                "searchUrlTemplate": "https://www.condec.com.br/loja/busca.php?palavra_busca={term}",
            },
        ),
        ManifestSeed(
            supplier_code="OBRA_FACIL",
            family="playwright",
            config_json={
                "strategy": "playwright.pdp_first.v1",
                "timeoutSeconds": 30,
                "maxRetries": 2,
                "tabs": 7,
                "sellerName": _seller_name("OBRA_FACIL"),
                # Minimum to avoid playwright_runtime_url_missing. PDP URLs are expected to
                # come from persisted signals (shopping_supplier_product_signals) when available.
                "startUrl": "https://lojaobrafacil.com.br/",
                "waitUntil": "commit",
                "headless": True,
                "fallbackSearchEnabled": False,
                "pdpSelectors": {
                    # From legacy OBRA_FACIL driver (kept intentionally simple).
                    "price": "div.col-des div.price-box p span",
                    "seller": "",
                },
            },
        ),
    ]


def seed_manifests(*, database_url: str, tenant_id: str) -> None:
    seeds = _seeds()
    if not seeds:
        raise RuntimeError("No manifest seeds defined.")

    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))

            # Operational lookup behavior lives in suppliers_directory (not in the manifest).
            # ABC changed behavior and is now more reliable on reference-only searches.
            cur.execute(
                """
                UPDATE suppliers_directory
                SET lookup_policy = 'REFERENCE_FIRST', updated_at = NOW()
                WHERE tenant_id = current_tenant_id()
                  AND supplier_code = 'ABC'
                """,
            )

            for seed in seeds:
                supplier = seed.supplier_code.strip().upper()

                # Ensure only one active manifest exists per supplier.
                cur.execute(
                    """
                    UPDATE supplier_driver_manifests
                    SET is_active = FALSE, updated_at = NOW()
                    WHERE tenant_id = current_tenant_id()
                      AND supplier_code = %s
                      AND is_active = TRUE
                    """,
                    (supplier,),
                )

                cur.execute(
                    """
                    SELECT COALESCE(MAX(version_number), 0)
                    FROM supplier_driver_manifests
                    WHERE tenant_id = current_tenant_id()
                      AND supplier_code = %s
                    """,
                    (supplier,),
                )
                (max_version,) = cur.fetchone() or (0,)
                next_version = int(max_version or 0) + 1

                cur.execute(
                    """
                    INSERT INTO supplier_driver_manifests (
                      manifest_id,
                      tenant_id,
                      supplier_code,
                      version_number,
                      family,
                      config_json,
                      validation_status,
                      validation_errors_json,
                      is_active,
                      created_by,
                      created_at,
                      updated_at
                    ) VALUES (
                      %s,
                      current_tenant_id(),
                      %s,
                      %s,
                      %s,
                      %s::jsonb,
                      'valid',
                      '[]'::jsonb,
                      TRUE,
                      'seed',
                      NOW(),
                      NOW()
                    )
                    """,
                    (
                        _manifest_id(supplier),
                        supplier,
                        next_version,
                        seed.family,
                        json.dumps(seed.config_json, ensure_ascii=True),
                    ),
                )

            cur.execute("COMMIT")

    print(f"Seeded {len(seeds)} manifests for tenant_id={tenant_id}.")


def main() -> None:
    database_url = os.getenv("MS_DATABASE_URL", "").strip()
    tenant_id = os.getenv("MS_TENANT_ID", "").strip() or "tenant_default"
    if not database_url:
        raise SystemExit("Missing MS_DATABASE_URL in environment.")
    seed_manifests(database_url=database_url, tenant_id=tenant_id)


if __name__ == "__main__":
    main()
