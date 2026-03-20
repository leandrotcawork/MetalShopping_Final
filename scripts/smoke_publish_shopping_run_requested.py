from __future__ import annotations

import argparse
import json
import os
import uuid
from typing import Any

import psycopg


def _new_uuid() -> str:
    return str(uuid.uuid4())


def publish(
    *,
    database_url: str,
    tenant_id: str,
    supplier_code: str,
    input_mode: str,
    catalog_product_ids: list[str],
) -> dict[str, Any]:
    run_request_id = _new_uuid()
    event_id = _new_uuid()

    payload = {
        "inputMode": input_mode,
        "catalogProductIds": catalog_product_ids if input_mode == "catalog" else [],
        "xlsxFilePath": "",
        "supplierCodes": [supplier_code],
    }

    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
            cur.execute(
                """
                INSERT INTO shopping_price_run_requests (
                  run_request_id, tenant_id, request_status, input_mode, input_payload_json, requested_by, requested_at
                ) VALUES (
                  %s, current_tenant_id(), 'queued', %s, %s::jsonb, 'smoke', NOW()
                )
                """,
                (run_request_id, input_mode, json.dumps(payload, ensure_ascii=True)),
            )
            cur.execute("COMMIT")

        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
            event_payload = {
                "run_request_id": run_request_id,
                "tenant_id": tenant_id,
                "input_mode": input_mode,
            }
            idempotency_key = f"shopping_run_requested:{run_request_id}"
            cur.execute(
                """
                INSERT INTO outbox_events (
                  event_id, aggregate_type, aggregate_id, event_name, event_version,
                  tenant_id, trace_id, idempotency_key, payload_json, status, available_at, published_at
                ) VALUES (
                  %s, 'shopping_run_request', %s, 'shopping.run_requested', 'v1',
                  %s, 'smoke', %s, %s::jsonb, 'published', NOW(), NOW()
                ) ON CONFLICT (idempotency_key) DO NOTHING
                """,
                (
                    event_id,
                    run_request_id,
                    tenant_id,
                    idempotency_key,
                    json.dumps(event_payload, ensure_ascii=True),
                ),
            )
            cur.execute("COMMIT")

    return {"run_request_id": run_request_id, "event_id": event_id, "status": "published"}


def main() -> None:
    parser = argparse.ArgumentParser(description="Publish shopping.run_requested without seeding supplier/manifests.")
    parser.add_argument("--database-url", default=os.getenv("MS_DATABASE_URL", "").strip())
    parser.add_argument("--tenant-id", default=os.getenv("MS_TENANT_ID", "").strip() or "tenant_default")
    parser.add_argument("--supplier-code", required=True)
    parser.add_argument("--input-mode", default=os.getenv("MS_INPUT_MODE", "").strip().lower() or "catalog")
    parser.add_argument("--catalog-product-ids", default=os.getenv("MS_CATALOG_PRODUCT_IDS", "").strip())
    args = parser.parse_args()

    database_url = str(args.database_url or "").strip()
    if database_url == "":
        raise SystemExit("Missing --database-url (or MS_DATABASE_URL).")

    input_mode = str(args.input_mode or "").strip().lower()
    if input_mode not in {"catalog", "xlsx"}:
        raise SystemExit("input_mode must be catalog or xlsx")

    supplier_code = str(args.supplier_code or "").strip().upper()
    if supplier_code == "":
        raise SystemExit("Missing --supplier-code.")

    catalog_ids: list[str] = []
    if input_mode == "catalog":
        raw = str(args.catalog_product_ids or "").strip()
        if raw == "":
            raise SystemExit("Missing --catalog-product-ids (or MS_CATALOG_PRODUCT_IDS) when input_mode=catalog.")
        for part in raw.split(","):
            part = part.strip()
            if part:
                catalog_ids.append(part)

    out = publish(
        database_url=database_url,
        tenant_id=str(args.tenant_id),
        supplier_code=supplier_code,
        input_mode=input_mode,
        catalog_product_ids=catalog_ids,
    )
    print(json.dumps(out, ensure_ascii=True))


if __name__ == "__main__":
    main()

