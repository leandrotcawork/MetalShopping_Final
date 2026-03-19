"""Shopping Price worker scaffold (Level 1).

Worker responsibility:
- ingest shopping run snapshots from an input JSON
- write to Postgres tables owned by shopping read surface
- keep idempotent upserts (safe to re-run)

This worker does not expose HTTP endpoints.
Go server_core reads from Postgres and exposes API.
"""

from __future__ import annotations

import json
import os
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

import psycopg


@dataclass(frozen=True)
class RunItem:
    product_id: str
    seller_name: str
    channel: str
    observed_price: float
    currency_code: str
    observed_at: str


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
                seller_name=str(item["seller_name"]),
                channel=str(item["channel"]),
                observed_price=float(item["observed_price"]),
                currency_code=str(item["currency_code"]).upper(),
                observed_at=str(item["observed_at"]),
            )
        )
    return parsed


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
        cur.execute("SELECT set_config('app.current_tenant_id', %s, true)", (tenant_id,))

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
            run_item_id = str(uuid4())
            cur.execute(
                """
                INSERT INTO shopping_price_run_items (
                  run_item_id, tenant_id, run_id, product_id, seller_name, channel,
                  observed_price, currency_code, observed_at
                )
                VALUES (
                  %s, current_tenant_id(), %s, %s, %s, %s,
                  %s, %s, %s
                )
                ON CONFLICT (run_item_id) DO NOTHING
                """,
                (
                    run_item_id,
                    run_id,
                    item.product_id,
                    item.seller_name,
                    item.channel,
                    item.observed_price,
                    item.currency_code,
                    item.observed_at,
                ),
            )

            snapshot_id = f"{tenant_id}:{item.product_id}"
            cur.execute(
                """
                INSERT INTO shopping_price_latest_snapshot (
                  snapshot_id, tenant_id, product_id, run_id, seller_name, channel,
                  observed_price, currency_code, observed_at
                )
                VALUES (
                  %s, current_tenant_id(), %s, %s, %s, %s,
                  %s, %s, %s
                )
                ON CONFLICT (tenant_id, product_id) DO UPDATE SET
                  run_id = EXCLUDED.run_id,
                  seller_name = EXCLUDED.seller_name,
                  channel = EXCLUDED.channel,
                  observed_price = EXCLUDED.observed_price,
                  currency_code = EXCLUDED.currency_code,
                  observed_at = EXCLUDED.observed_at,
                  updated_at = NOW()
                """,
                (
                    snapshot_id,
                    item.product_id,
                    run_id,
                    item.seller_name,
                    item.channel,
                    item.observed_price,
                    item.currency_code,
                    item.observed_at,
                ),
            )

        cur.execute("COMMIT")
    return len(items)


def main() -> int:
    database_url = os.getenv("MS_DATABASE_URL", "").strip()
    input_path = os.getenv("MS_SHOPPING_INPUT_PATH", "").strip()

    if database_url == "" or input_path == "":
        print(
            "Missing required env vars: MS_DATABASE_URL and MS_SHOPPING_INPUT_PATH",
            file=sys.stderr,
        )
        return 2

    payload = load_input(input_path)

    tenant_id = str(payload["tenant_id"])
    run_id = str(payload.get("run_id") or uuid4())
    run_status = str(payload.get("run_status") or "completed")
    started_at = str(payload.get("started_at") or utc_now_iso())
    finished_at = payload.get("finished_at")
    notes = str(payload.get("notes") or "")
    items = parse_items(payload.get("items", []))

    log("shopping_worker_start", tenant_id=tenant_id, run_id=run_id, items=len(items))
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
        log("shopping_worker_error", tenant_id=tenant_id, run_id=run_id, error=str(exc))
        return 1

    log("shopping_worker_end", tenant_id=tenant_id, run_id=run_id, rows_written=rows_written)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
