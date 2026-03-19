"""Shopping Price worker (ADR-0018 Phase 1).

Worker responsibility:
- queue mode: claim shopping run requests from Postgres and process lifecycle
- legacy mode: ingest shopping run snapshots from an input JSON
- write read-model rows consumed by server_core APIs
- keep idempotent upserts to avoid duplicate run items on retry
"""

from __future__ import annotations

import json
import os
import socket
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from hashlib import sha1
from typing import Any
from uuid import uuid4

import psycopg


@dataclass(frozen=True)
class RunItem:
    product_id: str
    supplier_code: str
    item_status: str
    seller_name: str
    channel: str
    observed_price: float
    currency_code: str
    observed_at: str
    product_url: str | None = None
    http_status: int | None = None
    elapsed_s: float | None = None
    chosen_seller_json: dict[str, Any] | None = None
    notes: str | None = None


@dataclass(frozen=True)
class RunRequest:
    run_request_id: str
    tenant_id: str
    input_mode: str
    input_payload: dict[str, Any]
    requested_by: str


def utc_now_iso() -> str:
    return datetime.now(tz=timezone.utc).isoformat()


def log(event: str, **fields: Any) -> None:
    payload = {"event": event, "ts": utc_now_iso(), **fields}
    print(json.dumps(payload, ensure_ascii=True))


def load_input(path: str) -> dict[str, Any]:
    with open(path, "r", encoding="utf-8") as fp:
        payload = json.load(fp)
    if not isinstance(payload, dict):
        raise ValueError("input payload must be a JSON object")
    return payload


def parse_items(raw_items: Any) -> list[RunItem]:
    if not isinstance(raw_items, list):
        raise ValueError("items must be an array")
    parsed: list[RunItem] = []
    for item in raw_items:
        if not isinstance(item, dict):
            raise ValueError("each item must be an object")
        parsed.append(
            RunItem(
                product_id=str(item["product_id"]),
                supplier_code=str(item.get("supplier_code") or "DEFAULT").strip().upper(),
                item_status=str(item.get("item_status") or "OK").strip().upper(),
                seller_name=str(item["seller_name"]),
                channel=str(item["channel"]),
                observed_price=float(item["observed_price"]),
                currency_code=str(item["currency_code"]).upper(),
                observed_at=str(item["observed_at"]),
                product_url=str(item["product_url"]).strip() if "product_url" in item and item["product_url"] is not None else None,
                http_status=int(item["http_status"]) if "http_status" in item and item["http_status"] is not None else None,
                elapsed_s=float(item["elapsed_s"]) if "elapsed_s" in item and item["elapsed_s"] is not None else None,
                chosen_seller_json=item.get("chosen_seller_json") if isinstance(item.get("chosen_seller_json"), dict) else None,
                notes=str(item["notes"]).strip() if "notes" in item and item["notes"] is not None else None,
            )
        )
    return parsed


def parse_catalog_product_ids(payload: dict[str, Any]) -> list[str]:
    value = payload.get("catalogProductIds")
    if value is None:
        return []
    if not isinstance(value, list):
        raise ValueError("catalogProductIds must be an array when provided")
    parsed: list[str] = []
    for entry in value:
        text = str(entry).strip()
        if text != "":
            parsed.append(text)
    return parsed


def parse_xlsx_file_path(payload: dict[str, Any]) -> str:
    value = payload.get("xlsxFilePath")
    if value is None:
        return ""
    return str(value).strip()


def parse_supplier_codes(payload: dict[str, Any]) -> list[str]:
    value = payload.get("supplierCodes")
    if value is None:
        return []
    if not isinstance(value, list):
        raise ValueError("supplierCodes must be an array when provided")
    parsed: list[str] = []
    for entry in value:
        code = str(entry).strip().upper()
        if code != "":
            parsed.append(code)
    return parsed


def deterministic_run_item_id(
    tenant_id: str,
    run_id: str,
    product_id: str,
    supplier_code: str,
) -> str:
    material = f"{tenant_id}|{run_id}|{product_id}|{supplier_code}"
    digest = sha1(material.encode("utf-8")).hexdigest()
    return f"{digest[0:8]}-{digest[8:12]}-{digest[12:16]}-{digest[16:20]}-{digest[20:32]}"


def upsert_run(
    conn: psycopg.Connection,
    tenant_id: str,
    run_id: str,
    run_status: str,
    started_at: str,
    finished_at: str | None,
    notes: str,
    items: list[RunItem],
) -> int:
    processed_items = len(items)
    total_items = len(items)

    with conn.cursor() as cur:
        cur.execute("BEGIN")
        # RLS uses current_tenant_id() -> current_setting('app.tenant_id')
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))

        cur.execute(
            """
            INSERT INTO shopping_price_runs (
              run_id, tenant_id, run_status, started_at, finished_at, processed_items, total_items, notes
            )
            VALUES (%s, current_tenant_id(), %s, %s, %s, %s, %s, %s)
            ON CONFLICT (run_id) DO UPDATE SET
              run_status = EXCLUDED.run_status,
              finished_at = EXCLUDED.finished_at,
              processed_items = EXCLUDED.processed_items,
              total_items = EXCLUDED.total_items,
              notes = EXCLUDED.notes,
              updated_at = NOW()
            """,
            (run_id, run_status, started_at, finished_at, processed_items, total_items, notes),
        )

        for item in items:
            run_item_id = deterministic_run_item_id(
                tenant_id=tenant_id,
                run_id=run_id,
                product_id=item.product_id,
                supplier_code=item.supplier_code,
            )
            chosen_seller_json = item.chosen_seller_json or {}
            cur.execute(
                """
                INSERT INTO shopping_price_run_items (
                  run_item_id, tenant_id, run_id, product_id, seller_name, channel,
                  supplier_code, item_status, observed_price, currency_code, observed_at,
                  product_url, http_status, elapsed_s, chosen_seller_json, notes
                )
                VALUES (
                  %s, current_tenant_id(), %s, %s, %s, %s,
                  %s, %s, %s, %s, %s,
                  %s, %s, %s, %s::jsonb, %s
                )
                ON CONFLICT (tenant_id, run_id, product_id, supplier_code) DO UPDATE SET
                  seller_name = EXCLUDED.seller_name,
                  channel = EXCLUDED.channel,
                  item_status = EXCLUDED.item_status,
                  observed_price = EXCLUDED.observed_price,
                  currency_code = EXCLUDED.currency_code,
                  observed_at = EXCLUDED.observed_at,
                  product_url = EXCLUDED.product_url,
                  http_status = EXCLUDED.http_status,
                  elapsed_s = EXCLUDED.elapsed_s,
                  chosen_seller_json = EXCLUDED.chosen_seller_json,
                  notes = EXCLUDED.notes,
                  updated_at = NOW()
                """,
                (
                    run_item_id,
                    run_id,
                    item.product_id,
                    item.seller_name,
                    item.channel,
                    item.supplier_code,
                    item.item_status,
                    item.observed_price,
                    item.currency_code,
                    item.observed_at,
                    item.product_url,
                    item.http_status,
                    item.elapsed_s,
                    json.dumps(chosen_seller_json),
                    item.notes,
                ),
            )

            snapshot_id = f"{tenant_id}:{item.product_id}:{item.supplier_code}"
            cur.execute(
                """
                INSERT INTO shopping_price_latest_snapshot (
                  snapshot_id, tenant_id, product_id, run_id, seller_name, channel,
                  supplier_code, item_status, observed_price, currency_code, observed_at,
                  product_url, http_status, elapsed_s, chosen_seller_json, notes
                )
                VALUES (
                  %s, current_tenant_id(), %s, %s, %s, %s,
                  %s, %s, %s, %s, %s,
                  %s, %s, %s, %s::jsonb, %s
                )
                ON CONFLICT (tenant_id, product_id, supplier_code) DO UPDATE SET
                  run_id = EXCLUDED.run_id,
                  seller_name = EXCLUDED.seller_name,
                  channel = EXCLUDED.channel,
                  item_status = EXCLUDED.item_status,
                  observed_price = EXCLUDED.observed_price,
                  currency_code = EXCLUDED.currency_code,
                  observed_at = EXCLUDED.observed_at,
                  product_url = EXCLUDED.product_url,
                  http_status = EXCLUDED.http_status,
                  elapsed_s = EXCLUDED.elapsed_s,
                  chosen_seller_json = EXCLUDED.chosen_seller_json,
                  notes = EXCLUDED.notes,
                  updated_at = NOW()
                """,
                (
                    snapshot_id,
                    item.product_id,
                    run_id,
                    item.seller_name,
                    item.channel,
                    item.supplier_code,
                    item.item_status,
                    item.observed_price,
                    item.currency_code,
                    item.observed_at,
                    item.product_url,
                    item.http_status,
                    item.elapsed_s,
                    json.dumps(chosen_seller_json),
                    item.notes,
                ),
            )

        cur.execute("COMMIT")
    return len(items)


def claim_next_request(
    conn: psycopg.Connection,
    tenant_id: str,
    worker_id: str,
) -> RunRequest | None:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT run_request_id, tenant_id, input_mode, input_payload_json, requested_by
            FROM shopping_price_run_requests
            WHERE tenant_id = current_tenant_id()
              AND request_status = 'queued'
            ORDER BY requested_at ASC
            FOR UPDATE SKIP LOCKED
            LIMIT 1
            """
        )
        row = cur.fetchone()
        if row is None:
            cur.execute("COMMIT")
            return None

        run_request_id = str(row[0])
        cur.execute(
            """
            UPDATE shopping_price_run_requests
            SET request_status = 'claimed',
                claimed_at = NOW(),
                worker_id = %s,
                updated_at = NOW()
            WHERE run_request_id = %s
            """,
            (worker_id, run_request_id),
        )
        cur.execute("COMMIT")

    payload = row[3]
    if not isinstance(payload, dict):
        raise ValueError("input_payload_json must be a JSON object")
    return RunRequest(
        run_request_id=run_request_id,
        tenant_id=str(row[1]),
        input_mode=str(row[2]),
        input_payload=payload,
        requested_by=str(row[4]),
    )


def mark_request_running(
    conn: psycopg.Connection,
    tenant_id: str,
    run_request_id: str,
    run_id: str,
) -> None:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            UPDATE shopping_price_run_requests
            SET request_status = 'running',
                started_at = NOW(),
                run_id = %s,
                updated_at = NOW()
            WHERE tenant_id = current_tenant_id()
              AND run_request_id = %s
            """,
            (run_id, run_request_id),
        )
        cur.execute("COMMIT")


def mark_request_completed(
    conn: psycopg.Connection,
    tenant_id: str,
    run_request_id: str,
) -> None:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            UPDATE shopping_price_run_requests
            SET request_status = 'completed',
                finished_at = NOW(),
                updated_at = NOW()
            WHERE tenant_id = current_tenant_id()
              AND run_request_id = %s
            """,
            (run_request_id,),
        )
        cur.execute("COMMIT")


def mark_request_failed(
    conn: psycopg.Connection,
    tenant_id: str,
    run_request_id: str,
    error_message: str,
) -> None:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            UPDATE shopping_price_run_requests
            SET request_status = 'failed',
                finished_at = NOW(),
                error_message = %s,
                updated_at = NOW()
            WHERE tenant_id = current_tenant_id()
              AND run_request_id = %s
            """,
            (error_message[:500], run_request_id),
        )
        cur.execute("COMMIT")


def query_catalog_items(
    conn: psycopg.Connection,
    tenant_id: str,
    product_ids: list[str],
    supplier_codes: list[str],
) -> list[RunItem]:
    if len(product_ids) == 0:
        return []
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              p.product_id,
              COALESCE(pr.price_amount, 0)::double precision AS observed_price,
              COALESCE(pr.currency_code, 'BRL') AS currency_code
            FROM catalog_products p
            LEFT JOIN pricing_product_prices pr
              ON pr.tenant_id = current_tenant_id()
             AND pr.product_id = p.product_id
             AND pr.effective_to IS NULL
            WHERE p.tenant_id = current_tenant_id()
              AND p.product_id = ANY(%s::text[])
            ORDER BY p.product_id
            """,
            (product_ids,),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    effective_suppliers = supplier_codes if len(supplier_codes) > 0 else ["DEFAULT"]
    now_iso = utc_now_iso()
    items: list[RunItem] = []
    for row in rows:
        product_id = str(row[0])
        observed_price = float(row[1])
        currency_code = str(row[2]).upper()
        for supplier_code in effective_suppliers:
            items.append(
                RunItem(
                    product_id=product_id,
                    supplier_code=supplier_code,
                    item_status="OK",
                    seller_name="catalog_reference",
                    channel="CATALOG",
                    observed_price=observed_price,
                    currency_code=currency_code,
                    observed_at=now_iso,
                )
            )
    return items


def query_xlsx_fallback_items(
    conn: psycopg.Connection,
    tenant_id: str,
    limit: int,
    supplier_codes: list[str],
) -> list[RunItem]:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              p.product_id,
              COALESCE(pr.price_amount, 0)::double precision AS observed_price,
              COALESCE(pr.currency_code, 'BRL') AS currency_code
            FROM catalog_products p
            LEFT JOIN pricing_product_prices pr
              ON pr.tenant_id = current_tenant_id()
             AND pr.product_id = p.product_id
             AND pr.effective_to IS NULL
            WHERE p.tenant_id = current_tenant_id()
            ORDER BY p.updated_at DESC
            LIMIT %s
            """,
            (limit,),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    effective_suppliers = supplier_codes if len(supplier_codes) > 0 else ["DEFAULT"]
    now_iso = utc_now_iso()
    items: list[RunItem] = []
    for row in rows:
        product_id = str(row[0])
        observed_price = float(row[1])
        currency_code = str(row[2]).upper()
        for supplier_code in effective_suppliers:
            items.append(
                RunItem(
                    product_id=product_id,
                    supplier_code=supplier_code,
                    item_status="OK",
                    seller_name="xlsx_reference",
                    channel="XLSX",
                    observed_price=observed_price,
                    currency_code=currency_code,
                    observed_at=now_iso,
                )
            )
    return items


def process_claimed_request(
    conn: psycopg.Connection,
    run_request: RunRequest,
    xlsx_fallback_limit: int,
) -> tuple[str, int]:
    run_id = str(uuid4())
    mark_request_running(conn, run_request.tenant_id, run_request.run_request_id, run_id)

    notes = f"requested_by={run_request.requested_by}; input_mode={run_request.input_mode}"
    supplier_codes = parse_supplier_codes(run_request.input_payload)
    if run_request.input_mode == "catalog":
        product_ids = parse_catalog_product_ids(run_request.input_payload)
        if len(product_ids) == 0:
            raise ValueError("catalog mode requires at least one catalogProductIds entry")
        items = query_catalog_items(conn, run_request.tenant_id, product_ids, supplier_codes)
    elif run_request.input_mode == "xlsx":
        xlsx_file_path = parse_xlsx_file_path(run_request.input_payload)
        items = query_xlsx_fallback_items(conn, run_request.tenant_id, xlsx_fallback_limit, supplier_codes)
        notes = f"{notes}; xlsx_path={xlsx_file_path or 'not_provided'}; mode=xlsx_fallback"
    else:
        raise ValueError(f"unsupported input_mode: {run_request.input_mode}")

    upsert_run(
        conn=conn,
        tenant_id=run_request.tenant_id,
        run_id=run_id,
        run_status="completed",
        started_at=utc_now_iso(),
        finished_at=utc_now_iso(),
        notes=notes,
        items=items,
    )
    mark_request_completed(conn, run_request.tenant_id, run_request.run_request_id)
    return run_id, len(items)


def run_queue_once(
    conn: psycopg.Connection,
    tenant_id: str,
    worker_id: str,
    xlsx_fallback_limit: int,
) -> bool:
    run_request = claim_next_request(conn, tenant_id, worker_id)
    if run_request is None:
        log("shopping_queue_idle", tenant_id=tenant_id, worker_id=worker_id)
        return False

    log(
        "shopping_queue_claimed",
        tenant_id=run_request.tenant_id,
        worker_id=worker_id,
        run_request_id=run_request.run_request_id,
        input_mode=run_request.input_mode,
    )
    try:
        run_id, rows_written = process_claimed_request(conn, run_request, xlsx_fallback_limit)
    except Exception as exc:
        mark_request_failed(conn, run_request.tenant_id, run_request.run_request_id, str(exc))
        log(
            "shopping_queue_failed",
            tenant_id=run_request.tenant_id,
            worker_id=worker_id,
            run_request_id=run_request.run_request_id,
            error=str(exc),
        )
        return True

    log(
        "shopping_queue_completed",
        tenant_id=run_request.tenant_id,
        worker_id=worker_id,
        run_request_id=run_request.run_request_id,
        run_id=run_id,
        rows_written=rows_written,
    )
    return True


def main() -> int:
    database_url = os.getenv("MS_DATABASE_URL", "").strip()
    input_path = os.getenv("MS_SHOPPING_INPUT_PATH", "").strip()
    tenant_id = os.getenv("MS_TENANT_ID", "").strip()
    worker_id = os.getenv("MS_WORKER_ID", "").strip() or f"{socket.gethostname()}-{os.getpid()}"
    xlsx_fallback_limit_raw = os.getenv("MS_SHOPPING_XLSX_FALLBACK_LIMIT", "").strip() or "50"
    max_queue_claims_raw = os.getenv("MS_SHOPPING_MAX_QUEUE_CLAIMS", "").strip() or "1"

    if database_url == "":
        print("Missing required env var: MS_DATABASE_URL", file=sys.stderr)
        return 2
    try:
        xlsx_fallback_limit = max(1, int(xlsx_fallback_limit_raw))
        max_queue_claims = max(1, int(max_queue_claims_raw))
    except ValueError:
        print("MS_SHOPPING_XLSX_FALLBACK_LIMIT and MS_SHOPPING_MAX_QUEUE_CLAIMS must be integers", file=sys.stderr)
        return 2

    if input_path != "":
        payload = load_input(input_path)

        tenant_id = str(payload["tenant_id"])
        run_id = str(payload.get("run_id") or uuid4())
        run_status = str(payload.get("run_status") or "completed")
        started_at = str(payload.get("started_at") or utc_now_iso())
        finished_at = payload.get("finished_at")
        notes = str(payload.get("notes") or "")
        items = parse_items(payload.get("items", []))

        log("shopping_worker_start", tenant_id=tenant_id, run_id=run_id, items=len(items), mode="legacy_input")
        try:
            with psycopg.connect(database_url) as conn:
                rows_written = upsert_run(
                    conn=conn,
                    tenant_id=tenant_id,
                    run_id=run_id,
                    run_status=run_status,
                    started_at=started_at,
                    finished_at=finished_at,
                    notes=notes,
                    items=items,
                )
        except Exception as exc:  # pragma: no cover - worker boundary log
            log("shopping_worker_error", tenant_id=tenant_id, run_id=run_id, error=str(exc), mode="legacy_input")
            return 1

        log("shopping_worker_end", tenant_id=tenant_id, run_id=run_id, rows_written=rows_written, mode="legacy_input")
        return 0

    if tenant_id == "":
        print("Missing required env var for queue mode: MS_TENANT_ID", file=sys.stderr)
        return 2

    log(
        "shopping_worker_start",
        tenant_id=tenant_id,
        worker_id=worker_id,
        mode="queue",
        max_queue_claims=max_queue_claims,
    )
    handled = 0
    try:
        with psycopg.connect(database_url) as conn:
            for _ in range(max_queue_claims):
                did_handle = run_queue_once(
                    conn=conn,
                    tenant_id=tenant_id,
                    worker_id=worker_id,
                    xlsx_fallback_limit=xlsx_fallback_limit,
                )
                if not did_handle:
                    break
                handled += 1
    except Exception as exc:  # pragma: no cover - worker boundary log
        log(
            "shopping_worker_error",
            tenant_id=tenant_id,
            worker_id=worker_id,
            mode="queue",
            error=str(exc),
        )
        return 1

    log(
        "shopping_worker_end",
        tenant_id=tenant_id,
        worker_id=worker_id,
        mode="queue",
        handled=handled,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
