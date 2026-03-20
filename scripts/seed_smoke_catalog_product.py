from __future__ import annotations

import argparse
import os
import uuid

import psycopg


def _new_id(prefix: str) -> str:
    return f"{prefix}_{uuid.uuid4().hex[:12]}"


def seed_catalog_product(
    *,
    database_url: str,
    tenant_id: str,
    product_id: str,
    sku: str,
    name: str,
    reference: str,
    ean: str,
) -> None:
    reference = str(reference or "").strip()
    ean = str(ean or "").strip()

    if reference == "" and ean == "":
        raise ValueError("Provide at least one identifier: reference or ean.")

    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))

            cur.execute(
                """
                INSERT INTO catalog_products (
                  product_id, tenant_id, sku, name, status, created_at, updated_at
                ) VALUES (
                  %s, current_tenant_id(), %s, %s, 'active', NOW(), NOW()
                ) ON CONFLICT (product_id) DO NOTHING
                """,
                (product_id, sku, name),
            )

            if reference:
                cur.execute(
                    """
                    INSERT INTO catalog_product_identifiers (
                      product_identifier_id, product_id, tenant_id,
                      identifier_type, identifier_value, source_system, is_primary,
                      created_at, updated_at
                    ) VALUES (
                      %s, %s, current_tenant_id(),
                      'reference', %s, 'smoke', TRUE,
                      NOW(), NOW()
                    ) ON CONFLICT DO NOTHING
                    """,
                    (_new_id("pid_ref"), product_id, reference),
                )

            if ean:
                cur.execute(
                    """
                    INSERT INTO catalog_product_identifiers (
                      product_identifier_id, product_id, tenant_id,
                      identifier_type, identifier_value, source_system, is_primary,
                      created_at, updated_at
                    ) VALUES (
                      %s, %s, current_tenant_id(),
                      'ean', %s, 'smoke', FALSE,
                      NOW(), NOW()
                    ) ON CONFLICT DO NOTHING
                    """,
                    (_new_id("pid_ean"), product_id, ean),
                )

            cur.execute("COMMIT")


def main() -> None:
    parser = argparse.ArgumentParser(description="Seed a smoke catalog product with reference/ean identifiers.")
    parser.add_argument("--database-url", default=os.getenv("MS_DATABASE_URL", "").strip())
    parser.add_argument("--tenant-id", default=os.getenv("MS_TENANT_ID", "").strip() or "tenant_default")
    parser.add_argument("--product-id", required=True)
    parser.add_argument("--sku", default="")
    parser.add_argument("--name", default="")
    parser.add_argument("--reference", default="")
    parser.add_argument("--ean", default="")
    args = parser.parse_args()

    database_url = str(args.database_url or "").strip()
    if database_url == "":
        raise SystemExit("Missing --database-url (or MS_DATABASE_URL).")

    product_id = str(args.product_id or "").strip()
    if product_id == "":
        raise SystemExit("Missing --product-id.")

    sku = str(args.sku or "").strip() or f"SMOKE_{product_id}"
    name = str(args.name or "").strip() or f"Smoke {product_id}"

    seed_catalog_product(
        database_url=database_url,
        tenant_id=str(args.tenant_id),
        product_id=product_id,
        sku=sku,
        name=name,
        reference=str(args.reference or ""),
        ean=str(args.ean or ""),
    )
    print(f"Seeded catalog product {product_id}.")


if __name__ == "__main__":
    main()

