"""Shopping Price worker (ADR-0018/ADR-0025).

Worker responsibility:
- queue mode: claim shopping run requests from Postgres and process lifecycle
- event mode: consume published shopping.run_requested events and process targeted requests
- legacy mode: ingest shopping run snapshots from an input JSON
- write read-model rows consumed by server_core APIs
- keep idempotent upserts to avoid duplicate run items on retry
"""

from __future__ import annotations

import json
import os
import socket
import sys
import time
import threading
from concurrent.futures import Future, ThreadPoolExecutor, as_completed
from dataclasses import dataclass
from datetime import datetime, timezone
from hashlib import sha1
from typing import Any, Callable
from uuid import uuid4

import psycopg
from src.shopping_price_runtime.dispatcher import execute as runtime_execute
from src.shopping_price_runtime.playwright.batch import (
    build_batch_items,
    execute_playwright_pdp_first_batch,
)
from src.shopping_price_runtime.models import LookupInputs, RuntimeObservation, SupplierRuntimeConfig, SupplierSignal


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


@dataclass(frozen=True)
class PublishedRunRequestedEvent:
    event_id: str
    tenant_id: str
    run_request_id: str


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
            log(
                "shopping_run_item",
                tenant_id=tenant_id,
                run_id=run_id,
                product_id=item.product_id,
                supplier_code=item.supplier_code,
                item_status=item.item_status,
                channel=item.channel,
                observed_price=item.observed_price,
                currency_code=item.currency_code,
                http_status=item.http_status,
                elapsed_s=item.elapsed_s,
                product_url=item.product_url,
                notes=item.notes,
            )
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

            inferred_lookup_mode = infer_lookup_mode(item)
            cur.execute(
                """
                INSERT INTO shopping_supplier_product_signals (
                  tenant_id, product_id, supplier_code, product_url, url_status, lookup_mode, lookup_mode_source,
                  manual_override, last_checked_at, last_success_at, last_http_status, last_error_message,
                  next_discovery_at, not_found_count,
                  created_by, created_at, updated_at
                )
                VALUES (
                  current_tenant_id(), %s, %s, %s::text,
                  CASE
                    WHEN %s::int IN (404, 410) THEN 'INVALID'
                    WHEN %s::text IS NULL OR %s::text = '' THEN 'STALE'
                    ELSE 'ACTIVE'
                  END,
                  %s, 'INFERRED',
                  FALSE, (%s)::timestamptz,
                  CASE WHEN %s::text IS NULL OR %s::text = '' THEN NULL ELSE (%s)::timestamptz END,
                  %s, %s,
                  CASE WHEN %s::text = 'NOT_FOUND' THEN NOW() + INTERVAL '30 days' ELSE NULL END,
                  CASE WHEN %s::text = 'NOT_FOUND' THEN 1 ELSE 0 END,
                  'shopping_worker', NOW(), NOW()
                )
                ON CONFLICT (tenant_id, product_id, supplier_code) DO UPDATE SET
                  product_url = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.product_url
                    WHEN EXCLUDED.last_http_status IN (404, 410) THEN NULL
                    ELSE COALESCE(EXCLUDED.product_url, shopping_supplier_product_signals.product_url)
                  END,
                  url_status = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.url_status
                    WHEN EXCLUDED.last_http_status IN (404, 410) THEN 'INVALID'
                    WHEN EXCLUDED.product_url IS NULL OR EXCLUDED.product_url = '' THEN 'STALE'
                    ELSE 'ACTIVE'
                  END,
                  lookup_mode = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.lookup_mode
                    ELSE EXCLUDED.lookup_mode
                  END,
                  lookup_mode_source = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.lookup_mode_source
                    ELSE 'INFERRED'
                  END,
                  last_checked_at = EXCLUDED.last_checked_at,
                  last_success_at = CASE
                    WHEN EXCLUDED.product_url IS NULL OR EXCLUDED.product_url = '' THEN shopping_supplier_product_signals.last_success_at
                    ELSE EXCLUDED.last_success_at
                  END,
                  last_http_status = EXCLUDED.last_http_status,
                  last_error_message = EXCLUDED.last_error_message,
                  next_discovery_at = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.next_discovery_at
                    WHEN EXCLUDED.last_http_status IN (404, 410) THEN NOW() + INTERVAL '30 days'
                    WHEN EXCLUDED.product_url IS NOT NULL AND EXCLUDED.product_url <> '' THEN NULL
                    WHEN %s::text = 'NOT_FOUND' THEN NOW() + INTERVAL '30 days'
                    WHEN %s::text = 'ERROR' AND (
                      EXCLUDED.last_error_message ILIKE '%%cloudflare%%'
                      OR EXCLUDED.last_error_message ILIKE '%%captcha%%'
                      OR EXCLUDED.last_error_message ILIKE '%%access denied%%'
                    ) THEN NOW() + INTERVAL '7 days'
                    ELSE shopping_supplier_product_signals.next_discovery_at
                  END,
                  not_found_count = CASE
                    WHEN shopping_supplier_product_signals.manual_override THEN shopping_supplier_product_signals.not_found_count
                    WHEN EXCLUDED.product_url IS NOT NULL AND EXCLUDED.product_url <> '' THEN 0
                    WHEN %s::text = 'NOT_FOUND' THEN shopping_supplier_product_signals.not_found_count + 1
                    ELSE shopping_supplier_product_signals.not_found_count
                  END,
                  updated_at = NOW()
                """,
                (
                    item.product_id,
                    item.supplier_code,
                    item.product_url,
                    item.http_status,
                    item.product_url,
                    item.product_url,
                    inferred_lookup_mode,
                    item.observed_at,
                    item.product_url,
                    item.product_url,
                    item.observed_at,
                    item.http_status,
                    item.notes,
                    item.item_status,
                    item.item_status,
                    item.item_status,
                    item.item_status,
                    item.item_status,
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


def claim_request_by_id(
    conn: psycopg.Connection,
    tenant_id: str,
    run_request_id: str,
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
              AND run_request_id = %s
              AND request_status = 'queued'
            FOR UPDATE SKIP LOCKED
            LIMIT 1
            """,
            (run_request_id,),
        )
        row = cur.fetchone()
        if row is None:
            cur.execute("COMMIT")
            return None

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
        run_request_id=str(row[0]),
        tenant_id=str(row[1]),
        input_mode=str(row[2]),
        input_payload=payload,
        requested_by=str(row[4]),
    )


def fetch_published_run_requested_events(
    conn: psycopg.Connection,
    limit: int,
    tenant_id_filter: str | None,
) -> list[PublishedRunRequestedEvent]:
    if not tenant_id_filter:
        raise ValueError("tenant_id is required for event mode (RLS enforced)")

    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id_filter,))
        cur.execute(
            """
            SELECT
              e.event_id,
              e.tenant_id,
              e.payload_json ->> 'run_request_id' AS run_request_id
            FROM outbox_events e
            JOIN shopping_price_run_requests r
              ON r.tenant_id = e.tenant_id
             AND r.run_request_id = (e.payload_json ->> 'run_request_id')
             AND r.request_status = 'queued'
            WHERE e.event_name = 'shopping.run_requested'
              AND e.status = 'published'
              AND e.tenant_id = current_tenant_id()
            ORDER BY e.created_at ASC
            LIMIT %s
            """,
            (limit,),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    events: list[PublishedRunRequestedEvent] = []
    for row in rows:
        event_id = str(row[0]).strip()
        tenant_id = str(row[1]).strip()
        run_request_id = str(row[2]).strip()
        if event_id and tenant_id and run_request_id:
            events.append(
                PublishedRunRequestedEvent(
                    event_id=event_id,
                    tenant_id=tenant_id,
                    run_request_id=run_request_id,
                )
            )
    return events


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


def update_request_progress(
    conn: psycopg.Connection,
    tenant_id: str,
    run_request_id: str,
    processed_items: int,
    total_items: int,
    current_supplier_code: str | None,
    current_product_id: str | None,
    current_product_label: str | None,
) -> None:
    supplier_code = current_supplier_code.strip().upper() if current_supplier_code else None
    product_id = current_product_id.strip() if current_product_id else None
    product_label = current_product_label.strip() if current_product_label else None
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            UPDATE shopping_price_run_requests
            SET processed_items = %s,
                total_items = %s,
                current_supplier_code = %s,
                current_product_id = %s,
                current_product_label = %s,
                progress_updated_at = NOW(),
                updated_at = NOW()
            WHERE tenant_id = current_tenant_id()
              AND run_request_id = %s
            """,
            (
                max(0, processed_items),
                max(0, total_items),
                supplier_code,
                product_id,
                product_label,
                run_request_id,
            ),
        )
        cur.execute("COMMIT")


class ProgressTracker:
    def __init__(
        self,
        conn: psycopg.Connection,
        tenant_id: str,
        run_request_id: str,
    ) -> None:
        self.conn = conn
        self.tenant_id = tenant_id
        self.run_request_id = run_request_id
        self.total_items = 0
        self.processed_items = 0
        self.current_supplier_code: str | None = None
        self.current_product_id: str | None = None
        self.current_product_label: str | None = None
        self.last_flush = 0.0

    def set_total(self, total_items: int) -> None:
        self.total_items = max(0, total_items)
        self.flush(force=True)

    def mark_item(self, supplier_code: str, product_id: str, product_label: str) -> None:
        self.processed_items += 1
        self.current_supplier_code = supplier_code
        self.current_product_id = product_id
        self.current_product_label = product_label
        self.flush(force=False)

    def flush(self, force: bool) -> None:
        now = time.monotonic()
        if not force:
            if self.processed_items % 5 != 0 and (now - self.last_flush) < 2.0:
                return
        update_request_progress(
            conn=self.conn,
            tenant_id=self.tenant_id,
            run_request_id=self.run_request_id,
            processed_items=self.processed_items,
            total_items=self.total_items,
            current_supplier_code=self.current_supplier_code,
            current_product_id=self.current_product_id,
            current_product_label=self.current_product_label,
        )
        self.last_flush = now

    def finalize(self) -> None:
        self.flush(force=True)


@dataclass(frozen=True)
class RuntimeTask:
    product_id: str
    supplier_code: str
    product_label: str
    product_reference: str
    product_ean: str
    base_price: float
    currency_code: str
    observed_at: str
    signal: SupplierSignal | None


class TokenBucket:
    def __init__(self, rate_per_second: float, capacity: float = 1.0) -> None:
        self.rate_per_second = max(0.01, rate_per_second)
        self.capacity = max(1.0, capacity)
        self.tokens = self.capacity
        self.updated_at = time.monotonic()
        self._lock = threading.Lock()

    def wait_for_token(self) -> None:
        while True:
            sleep_for = 0.0
            with self._lock:
                now = time.monotonic()
                elapsed = max(0.0, now - self.updated_at)
                self.updated_at = now
                self.tokens = min(self.capacity, self.tokens + elapsed * self.rate_per_second)
                if self.tokens >= 1.0:
                    self.tokens -= 1.0
                    return
                missing = 1.0 - self.tokens
                sleep_for = missing / self.rate_per_second
            time.sleep(max(0.001, sleep_for))


def _safe_int(value: Any, default: int, minimum: int, maximum: int) -> int:
    try:
        parsed = int(value)
    except (TypeError, ValueError):
        return default
    if parsed < minimum:
        return minimum
    if parsed > maximum:
        return maximum
    return parsed


def _safe_float_range(value: Any, default: float, minimum: float, maximum: float) -> float:
    try:
        parsed = float(value)
    except (TypeError, ValueError):
        return default
    if parsed < minimum:
        return minimum
    if parsed > maximum:
        return maximum
    return parsed


def _max_workers_for_config(config: SupplierRuntimeConfig) -> int:
    if config.family == "http":
        return _safe_int(config.config_json.get("maxConcurrency"), 4, 1, 16)
    if config.family == "playwright":
        return _safe_int(config.config_json.get("tabs"), 1, 1, 10)
    return 1


def _requests_per_second_for_config(config: SupplierRuntimeConfig) -> float:
    if config.family != "http":
        return 0.0
    return _safe_float_range(config.config_json.get("requestsPerSecond"), 0.0, 0.0, 10.0)


def _default_runtime_error_item(
    task: RuntimeTask,
    family: str,
    note: str,
) -> RunItem:
    return RunItem(
        product_id=task.product_id,
        supplier_code=task.supplier_code,
        item_status="ERROR",
        seller_name=task.supplier_code.lower(),
        channel=family.upper(),
        observed_price=_safe_float(task.base_price, task.base_price),
        currency_code=task.currency_code,
        observed_at=task.observed_at,
        product_url=task.signal.product_url if task.signal is not None else None,
        notes=note,
    )


def execute_runtime_tasks_parallel(
    tasks: list[RuntimeTask],
    runtime_configs: dict[str, SupplierRuntimeConfig],
    on_item_completed: Callable[[RuntimeTask], None] | None = None,
) -> list[RunItem]:
    if len(tasks) == 0:
        return []

    indexed_tasks: list[tuple[int, RuntimeTask]] = list(enumerate(tasks))

    semaphores: dict[str, threading.Semaphore] = {}
    rate_limiters: dict[str, TokenBucket] = {}
    max_workers_sum = 0
    for supplier_code, config in runtime_configs.items():
        supplier_upper = supplier_code.upper()
        cap = _max_workers_for_config(config)
        semaphores[supplier_upper] = threading.Semaphore(cap)
        max_workers_sum += cap
        rps = _requests_per_second_for_config(config)
        if rps > 0.0:
            rate_limiters[supplier_upper] = TokenBucket(rate_per_second=rps, capacity=max(1.0, rps))

    if max_workers_sum <= 0:
        max_workers_sum = 1
    max_workers = min(max_workers_sum, len(tasks), 32)
    max_workers = max(1, max_workers)

    results: list[RunItem | None] = [None] * len(tasks)

    playwright_batches: dict[str, list[tuple[int, RuntimeTask]]] = {}
    regular_tasks: list[tuple[int, RuntimeTask]] = []
    for index, task in indexed_tasks:
        config = runtime_configs.get(task.supplier_code.upper())
        if config is None:
            regular_tasks.append((index, task))
            continue
        if config.family == "playwright":
            strategy = str(config.config_json.get("strategy") or "").strip().lower()
            if strategy == "playwright.pdp_first.v1":
                playwright_batches.setdefault(config.supplier_code.upper(), []).append((index, task))
                continue
        regular_tasks.append((index, task))

    def run_single(task: RuntimeTask) -> RunItem:
        supplier_code = task.supplier_code.upper()
        config = runtime_configs.get(supplier_code)
        if config is None:
            return _default_runtime_error_item(task, "runtime", "runtime_config_missing")

        semaphore = semaphores.get(supplier_code)
        limiter = rate_limiters.get(supplier_code)

        if semaphore is None:
            return _default_runtime_error_item(task, config.family, "runtime_semaphore_missing")

        with semaphore:
            if limiter is not None:
                limiter.wait_for_token()
            try:
                return execute_supplier_runtime(
                    config=config,
                    product_id=task.product_id,
                    product_reference=task.product_reference,
                    product_ean=task.product_ean,
                    base_price=task.base_price,
                    currency_code=task.currency_code,
                    observed_at=task.observed_at,
                    signal=task.signal,
                )
            except Exception as exc:
                return _default_runtime_error_item(task, config.family, f"runtime_exception:{str(exc)[:240]}")

    def run_playwright_batch(
        supplier_code: str,
        batch_tasks: list[tuple[int, RuntimeTask]],
    ) -> list[tuple[int, RunItem]]:
        config = runtime_configs.get(supplier_code)
        if config is None:
            return [(idx, _default_runtime_error_item(task, "runtime", "runtime_config_missing")) for idx, task in batch_tasks]

        tabs = _max_workers_for_config(config)
        batch_inputs: list[tuple[int, LookupInputs, float, SupplierSignal | None]] = []
        for index, task in batch_tasks:
            batch_inputs.append(
                (
                    index,
                    LookupInputs(
                        product_id=task.product_id,
                        product_reference=task.product_reference,
                        product_ean=task.product_ean,
                    ),
                    task.base_price,
                    task.signal,
                )
            )
        items = build_batch_items(config=config, inputs=batch_inputs)
        try:
            batch_results = execute_playwright_pdp_first_batch(config=config, items=items, tabs=tabs)
        except Exception as exc:
            fallback: list[tuple[int, RunItem]] = []
            for index, task in batch_tasks:
                fallback.append(
                    (
                        index,
                        _default_runtime_error_item(task, config.family, f"playwright_batch_exception:{str(exc)[:200]}"),
                    )
                )
            return fallback

        mapped: dict[int, RuntimeTask] = {index: task for index, task in batch_tasks}
        output: list[tuple[int, RunItem]] = []
        for result in batch_results:
            task = mapped.get(result.index)
            if task is None:
                continue
            observation = result.observation
            output.append(
                (
                    result.index,
                    RunItem(
                        product_id=task.product_id,
                        supplier_code=config.supplier_code,
                        item_status=observation.item_status,
                        seller_name=observation.seller_name,
                        channel=observation.channel,
                        observed_price=_safe_float(observation.observed_price, task.base_price),
                        currency_code=task.currency_code.upper(),
                        observed_at=task.observed_at,
                        product_url=task.signal.product_url if task.signal is not None else None,
                        http_status=observation.http_status,
                        elapsed_s=result.elapsed_s,
                        chosen_seller_json={
                            "supplier_code": config.supplier_code,
                            "family": config.family,
                            "strategy": observation.strategy,
                            "lookup_policy": config.lookup_policy,
                            "execution_kind": config.execution_kind,
                            "lookup_term": observation.lookup_term,
                        },
                        notes=observation.notes,
                    ),
                )
            )
        return output

    with ThreadPoolExecutor(max_workers=max_workers, thread_name_prefix="shopping-runtime") as executor:
        pending: dict[Future[RunItem], int] = {}
        batch_futures: dict[Future[list[tuple[int, RunItem]]], list[int]] = {}
        for supplier_code, batch_tasks in playwright_batches.items():
            future = executor.submit(run_playwright_batch, supplier_code, batch_tasks)
            batch_futures[future] = [idx for idx, _ in batch_tasks]
        for index, task in regular_tasks:
            pending[executor.submit(run_single, task)] = index
        for future in as_completed(pending):
            index = pending[future]
            results[index] = future.result()
            if on_item_completed is not None:
                on_item_completed(tasks[index])
        for future in as_completed(batch_futures):
            batch_results = future.result()
            for index, item in batch_results:
                results[index] = item
                if on_item_completed is not None:
                    on_item_completed(tasks[index])

    return [item for item in results if item is not None]


def query_catalog_items(
    conn: psycopg.Connection,
    tenant_id: str,
    product_ids: list[str],
    supplier_codes: list[str],
    progress: ProgressTracker | None = None,
) -> tuple[list[RunItem], int]:
    if len(product_ids) == 0:
        return [], 0
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              p.product_id,
              p.name,
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
    total_items = len(rows) * len(effective_suppliers)
    if progress is not None:
        progress.set_total(total_items)
    identifiers = fetch_catalog_identifiers(conn, tenant_id, [str(row[0]) for row in rows])
    signal_overrides = fetch_supplier_signal_overrides(conn, tenant_id, [str(row[0]) for row in rows], effective_suppliers)
    runtime_configs = fetch_supplier_runtime_configs(conn, tenant_id, effective_suppliers)
    now_iso = utc_now_iso()
    tasks: list[RuntimeTask] = []
    items: list[RunItem] = []
    for row in rows:
        product_id = str(row[0])
        product_label = str(row[1]).strip()
        observed_price = float(row[2])
        currency_code = str(row[3]).upper()
        product_reference, product_ean = identifiers.get(product_id, ("", ""))
        for supplier_code in effective_suppliers:
            runtime_config = runtime_configs.get(supplier_code.upper())
            signal = signal_overrides.get(signal_key(product_id, supplier_code))
            if runtime_config is None:
                items.append(
                    RunItem(
                        product_id=product_id,
                        supplier_code=supplier_code.upper(),
                        item_status="ERROR",
                        seller_name=supplier_code.lower(),
                        channel="RUNTIME",
                        observed_price=_safe_float(observed_price, observed_price),
                        currency_code=currency_code,
                        observed_at=now_iso,
                        product_url=signal.product_url if signal is not None else None,
                        notes="runtime_config_missing",
                    )
                )
                if progress is not None:
                    progress.mark_item(supplier_code.upper(), product_id, product_label or product_id)
                continue
            tasks.append(
                RuntimeTask(
                    product_id=product_id,
                    supplier_code=supplier_code.upper(),
                    product_label=product_label or product_id,
                    product_reference=product_reference,
                    product_ean=product_ean,
                    base_price=_safe_float(observed_price, observed_price),
                    currency_code=currency_code,
                    observed_at=now_iso,
                    signal=signal,
                )
            )
    if progress is None:
        items.extend(execute_runtime_tasks_parallel(tasks, runtime_configs))
    else:
        items.extend(
            execute_runtime_tasks_parallel(
                tasks,
                runtime_configs,
                on_item_completed=lambda task: progress.mark_item(task.supplier_code, task.product_id, task.product_label),
            )
        )
    return items, total_items


def query_xlsx_fallback_items(
    conn: psycopg.Connection,
    tenant_id: str,
    limit: int,
    supplier_codes: list[str],
    progress: ProgressTracker | None = None,
) -> tuple[list[RunItem], int]:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              p.product_id,
              p.name,
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
    total_items = len(rows) * len(effective_suppliers)
    if progress is not None:
        progress.set_total(total_items)
    identifiers = fetch_catalog_identifiers(conn, tenant_id, [str(row[0]) for row in rows])
    signal_overrides = fetch_supplier_signal_overrides(conn, tenant_id, [str(row[0]) for row in rows], effective_suppliers)
    runtime_configs = fetch_supplier_runtime_configs(conn, tenant_id, effective_suppliers)
    now_iso = utc_now_iso()
    tasks: list[RuntimeTask] = []
    items: list[RunItem] = []
    for row in rows:
        product_id = str(row[0])
        product_label = str(row[1]).strip()
        observed_price = float(row[2])
        currency_code = str(row[3]).upper()
        product_reference, product_ean = identifiers.get(product_id, ("", ""))
        for supplier_code in effective_suppliers:
            runtime_config = runtime_configs.get(supplier_code.upper())
            signal = signal_overrides.get(signal_key(product_id, supplier_code))
            if runtime_config is None:
                items.append(
                    RunItem(
                        product_id=product_id,
                        supplier_code=supplier_code.upper(),
                        item_status="ERROR",
                        seller_name=supplier_code.lower(),
                        channel="RUNTIME",
                        observed_price=_safe_float(observed_price, observed_price),
                        currency_code=currency_code,
                        observed_at=now_iso,
                        product_url=signal.product_url if signal is not None else None,
                        notes="runtime_config_missing",
                    )
                )
                if progress is not None:
                    progress.mark_item(supplier_code.upper(), product_id, product_label or product_id)
                continue
            tasks.append(
                RuntimeTask(
                    product_id=product_id,
                    supplier_code=supplier_code.upper(),
                    product_label=product_label or product_id,
                    product_reference=product_reference,
                    product_ean=product_ean,
                    base_price=_safe_float(observed_price, observed_price),
                    currency_code=currency_code,
                    observed_at=now_iso,
                    signal=signal,
                )
            )
    if progress is None:
        items.extend(execute_runtime_tasks_parallel(tasks, runtime_configs))
    else:
        items.extend(
            execute_runtime_tasks_parallel(
                tasks,
                runtime_configs,
                on_item_completed=lambda task: progress.mark_item(task.supplier_code, task.product_id, task.product_label),
            )
        )
    return items, total_items


def infer_lookup_mode(item: RunItem) -> str:
    lookup_term = ""
    if isinstance(item.chosen_seller_json, dict):
        raw_term = item.chosen_seller_json.get("lookup_term")
        if raw_term is not None:
            lookup_term = str(raw_term).strip()

    if lookup_term.isdigit() and len(lookup_term) in (13, 14):
        return "EAN"

    channel = str(item.channel or "").upper()
    if "EAN" in channel:
        return "EAN"

    return "REFERENCE"


def signal_key(product_id: str, supplier_code: str) -> str:
    return f"{product_id}|{supplier_code.upper()}"


def fetch_supplier_signal_overrides(
    conn: psycopg.Connection,
    tenant_id: str,
    product_ids: list[str],
    supplier_codes: list[str],
) -> dict[str, SupplierSignal]:
    if len(product_ids) == 0 or len(supplier_codes) == 0:
        return {}
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              product_id,
              supplier_code,
              product_url,
              lookup_mode,
              manual_override,
              next_discovery_at
            FROM shopping_supplier_product_signals
            WHERE tenant_id = current_tenant_id()
              AND product_id = ANY(%s::text[])
              AND supplier_code = ANY(%s::text[])
            """,
            (product_ids, supplier_codes),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    now = datetime.now(timezone.utc)
    output: dict[str, SupplierSignal] = {}
    for row in rows:
        product_id = str(row[0])
        supplier_code = str(row[1]).upper()
        product_url = str(row[2]).strip() if row[2] is not None else None
        lookup_mode = str(row[3]).upper() if row[3] is not None else "REFERENCE"
        manual_override = bool(row[4])
        next_discovery_at = row[5] if len(row) > 5 else None
        if isinstance(next_discovery_at, datetime) and next_discovery_at.tzinfo is None:
            next_discovery_at = next_discovery_at.replace(tzinfo=timezone.utc)

        allow_url_discovery = True
        if manual_override:
            allow_url_discovery = False
        if isinstance(next_discovery_at, datetime) and next_discovery_at > now:
            allow_url_discovery = False
        output[signal_key(product_id, supplier_code)] = SupplierSignal(
            product_url=product_url if product_url else None,
            lookup_mode=lookup_mode if lookup_mode in {"EAN", "REFERENCE"} else "REFERENCE",
            manual_override=manual_override,
            allow_url_discovery=allow_url_discovery,
        )
    return output


def fetch_supplier_runtime_configs(
    conn: psycopg.Connection,
    tenant_id: str,
    supplier_codes: list[str],
) -> dict[str, SupplierRuntimeConfig]:
    if len(supplier_codes) == 0:
        return {}
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              d.supplier_code,
              d.execution_kind,
              d.lookup_policy,
              m.family,
              m.config_json
            FROM suppliers_directory d
            JOIN supplier_driver_manifests m
              ON m.tenant_id = d.tenant_id
             AND m.supplier_code = d.supplier_code
            WHERE d.tenant_id = current_tenant_id()
              AND d.enabled = TRUE
              AND m.is_active = TRUE
              AND m.validation_status = 'valid'
              AND d.supplier_code = ANY(%s::text[])
            ORDER BY d.supplier_code
            """,
            (supplier_codes,),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    configs: dict[str, SupplierRuntimeConfig] = {}
    for row in rows:
        supplier_code = str(row[0]).strip().upper()
        if supplier_code == "":
            continue
        raw_config = row[4]
        config_json = raw_config if isinstance(raw_config, dict) else {}
        configs[supplier_code] = SupplierRuntimeConfig(
            supplier_code=supplier_code,
            execution_kind=str(row[1]).strip().upper(),
            lookup_policy=str(row[2]).strip().upper(),
            family=str(row[3]).strip().lower(),
            config_json=config_json,
        )
    return configs


def fetch_catalog_identifiers(
    conn: psycopg.Connection,
    tenant_id: str,
    product_ids: list[str],
) -> dict[str, tuple[str, str]]:
    if len(product_ids) == 0:
        return {}
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT
              product_id,
              MAX(CASE WHEN identifier_type = 'reference' THEN identifier_value END) AS reference,
              MAX(CASE WHEN identifier_type = 'ean' THEN identifier_value END) AS ean
            FROM catalog_product_identifiers
            WHERE tenant_id = current_tenant_id()
              AND product_id = ANY(%s::text[])
            GROUP BY product_id
            """,
            (product_ids,),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    out: dict[str, tuple[str, str]] = {}
    for row in rows:
        product_id = str(row[0]).strip()
        reference = str(row[1]).strip() if row[1] is not None else ""
        ean = str(row[2]).strip() if row[2] is not None else ""
        if product_id:
            out[product_id] = (reference, ean)
    return out


def _safe_float(value: Any, default: float) -> float:
    try:
        parsed = float(value)
    except (TypeError, ValueError):
        return default
    if parsed < 0:
        return 0.0
    return round(parsed, 2)


def execute_supplier_runtime(
    config: SupplierRuntimeConfig,
    product_id: str,
    product_reference: str,
    product_ean: str,
    base_price: float,
    currency_code: str,
    observed_at: str,
    signal: SupplierSignal | None,
) -> RunItem:
    started_at = time.perf_counter()
    observation: RuntimeObservation = runtime_execute(
        config=config,
        inputs=LookupInputs(
            product_id=product_id,
            product_reference=product_reference,
            product_ean=product_ean,
        ),
        base_price=base_price,
        signal=signal,
    )
    elapsed_s = round(time.perf_counter() - started_at, 3)

    return RunItem(
        product_id=product_id,
        supplier_code=config.supplier_code,
        item_status=observation.item_status,
        seller_name=observation.seller_name,
        channel=observation.channel,
        observed_price=_safe_float(observation.observed_price, base_price),
        currency_code=currency_code.upper(),
        observed_at=observed_at,
        product_url=signal.product_url if signal is not None else None,
        http_status=observation.http_status,
        elapsed_s=elapsed_s,
        chosen_seller_json={
            "supplier_code": config.supplier_code,
            "family": config.family,
            "strategy": observation.strategy,
            "lookup_policy": config.lookup_policy,
            "execution_kind": config.execution_kind,
            "lookup_term": observation.lookup_term,
        },
        notes=observation.notes,
    )


def fetch_active_valid_supplier_codes(
    conn: psycopg.Connection,
    tenant_id: str,
    requested_supplier_codes: list[str],
) -> list[str]:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
        cur.execute(
            """
            SELECT DISTINCT d.supplier_code
            FROM suppliers_directory d
            JOIN supplier_driver_manifests m
              ON m.tenant_id = d.tenant_id
             AND m.supplier_code = d.supplier_code
            WHERE d.tenant_id = current_tenant_id()
              AND d.enabled = TRUE
              AND m.is_active = TRUE
              AND m.validation_status = 'valid'
              AND (
                cardinality(%s::text[]) = 0
                OR d.supplier_code = ANY(%s::text[])
              )
            ORDER BY d.supplier_code ASC
            """,
            (requested_supplier_codes, requested_supplier_codes),
        )
        rows = cur.fetchall()
        cur.execute("COMMIT")

    return [str(row[0]).strip().upper() for row in rows if str(row[0]).strip() != ""]


def process_claimed_request(
    conn: psycopg.Connection,
    run_request: RunRequest,
    xlsx_fallback_limit: int,
) -> tuple[str, int]:
    run_id = str(uuid4())
    mark_request_running(conn, run_request.tenant_id, run_request.run_request_id, run_id)
    progress = ProgressTracker(conn, run_request.tenant_id, run_request.run_request_id)

    notes = f"requested_by={run_request.requested_by}; input_mode={run_request.input_mode}"
    supplier_codes = parse_supplier_codes(run_request.input_payload)
    effective_supplier_codes = fetch_active_valid_supplier_codes(conn, run_request.tenant_id, supplier_codes)
    if len(effective_supplier_codes) == 0:
        raise ValueError("no active valid supplier manifests available for this tenant/request")
    notes = f"{notes}; suppliers={','.join(effective_supplier_codes)}"
    if run_request.input_mode == "catalog":
        product_ids = parse_catalog_product_ids(run_request.input_payload)
        if len(product_ids) == 0:
            raise ValueError("catalog mode requires at least one catalogProductIds entry")
        items, _ = query_catalog_items(conn, run_request.tenant_id, product_ids, effective_supplier_codes, progress)
    elif run_request.input_mode == "xlsx":
        xlsx_file_path = parse_xlsx_file_path(run_request.input_payload)
        product_ids = parse_catalog_product_ids(run_request.input_payload)
        unresolved_scope = parse_catalog_product_ids(
            {"catalogProductIds": run_request.input_payload.get("unresolvedScopeIdentifiers")}
        )
        ambiguous_scope = parse_catalog_product_ids(
            {"catalogProductIds": run_request.input_payload.get("ambiguousScopeIdentifiers")}
        )
        if len(product_ids) > 0:
            items, _ = query_catalog_items(conn, run_request.tenant_id, product_ids, effective_supplier_codes, progress)
            notes = (
                f"{notes}; xlsx_path={xlsx_file_path or 'not_provided'}; "
                f"mode=xlsx_resolved; resolved={len(product_ids)}; "
                f"unresolved={len(unresolved_scope)}; ambiguous={len(ambiguous_scope)}"
            )
        else:
            items, _ = query_xlsx_fallback_items(
                conn,
                run_request.tenant_id,
                xlsx_fallback_limit,
                effective_supplier_codes,
                progress,
            )
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
    progress.finalize()
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
        try:
            conn.rollback()
        except Exception:
            pass
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


def run_event_once(
    conn: psycopg.Connection,
    worker_id: str,
    xlsx_fallback_limit: int,
    tenant_id_filter: str | None,
) -> bool:
    if not tenant_id_filter:
        raise ValueError("tenant_id is required for event mode")
    events = fetch_published_run_requested_events(conn, limit=1, tenant_id_filter=tenant_id_filter)
    if len(events) == 0:
        log("shopping_event_idle", worker_id=worker_id, tenant_id=tenant_id_filter or "all")
        return False

    event = events[0]
    run_request = claim_request_by_id(conn, event.tenant_id, event.run_request_id, worker_id)
    if run_request is None:
        log(
            "shopping_event_skipped",
            worker_id=worker_id,
            event_id=event.event_id,
            tenant_id=event.tenant_id,
            run_request_id=event.run_request_id,
            reason="run_request_not_queued_or_locked",
        )
        return True

    log(
        "shopping_event_claimed",
        worker_id=worker_id,
        event_id=event.event_id,
        tenant_id=run_request.tenant_id,
        run_request_id=run_request.run_request_id,
        input_mode=run_request.input_mode,
    )
    try:
        run_id, rows_written = process_claimed_request(conn, run_request, xlsx_fallback_limit)
    except Exception as exc:
        try:
            conn.rollback()
        except Exception:
            pass
        mark_request_failed(conn, run_request.tenant_id, run_request.run_request_id, str(exc))
        log(
            "shopping_event_failed",
            worker_id=worker_id,
            event_id=event.event_id,
            tenant_id=run_request.tenant_id,
            run_request_id=run_request.run_request_id,
            error=str(exc),
        )
        return True

    log(
        "shopping_event_completed",
        worker_id=worker_id,
        event_id=event.event_id,
        tenant_id=run_request.tenant_id,
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
    mode = os.getenv("MS_SHOPPING_WORKER_MODE", "").strip().lower() or "queue"
    xlsx_fallback_limit_raw = os.getenv("MS_SHOPPING_XLSX_FALLBACK_LIMIT", "").strip() or "50"
    max_queue_claims_raw = os.getenv("MS_SHOPPING_MAX_QUEUE_CLAIMS", "").strip() or "1"
    keep_alive_raw = os.getenv("MS_SHOPPING_KEEP_ALIVE", "").strip().lower() or "false"
    poll_interval_raw = os.getenv("MS_SHOPPING_POLL_INTERVAL_SECONDS", "").strip() or "2"

    if database_url == "":
        print("Missing required env var: MS_DATABASE_URL", file=sys.stderr)
        return 2
    try:
        xlsx_fallback_limit = max(1, int(xlsx_fallback_limit_raw))
        max_queue_claims = max(1, int(max_queue_claims_raw))
        poll_interval_seconds = max(1, int(poll_interval_raw))
    except ValueError:
        print(
            "MS_SHOPPING_XLSX_FALLBACK_LIMIT, MS_SHOPPING_MAX_QUEUE_CLAIMS, and MS_SHOPPING_POLL_INTERVAL_SECONDS must be integers",
            file=sys.stderr,
        )
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

    if mode not in {"queue", "event"}:
        print("MS_SHOPPING_WORKER_MODE must be queue or event", file=sys.stderr)
        return 2
    if mode == "queue" and tenant_id == "":
        print("Missing required env var for queue mode: MS_TENANT_ID", file=sys.stderr)
        return 2
    if mode == "event" and tenant_id == "":
        print("Missing required env var for event mode: MS_TENANT_ID", file=sys.stderr)
        return 2

    keep_alive = keep_alive_raw in {"1", "true", "yes", "y", "on"}

    log(
        "shopping_worker_start",
        tenant_id=tenant_id or "all",
        worker_id=worker_id,
        mode=mode,
        max_queue_claims=max_queue_claims,
        keep_alive=keep_alive,
        poll_interval_seconds=poll_interval_seconds,
    )
    handled = 0
    try:
        with psycopg.connect(database_url) as conn:
            while True:
                did_handle_any = False
                for _ in range(max_queue_claims):
                    if mode == "queue":
                        did_handle = run_queue_once(
                            conn=conn,
                            tenant_id=tenant_id,
                            worker_id=worker_id,
                            xlsx_fallback_limit=xlsx_fallback_limit,
                        )
                    else:
                        did_handle = run_event_once(
                            conn=conn,
                            worker_id=worker_id,
                            xlsx_fallback_limit=xlsx_fallback_limit,
                            tenant_id_filter=tenant_id,
                        )
                    if not did_handle:
                        break
                    did_handle_any = True
                    handled += 1

                if did_handle_any or not keep_alive:
                    break

                time.sleep(poll_interval_seconds)
    except Exception as exc:  # pragma: no cover - worker boundary log
        log(
            "shopping_worker_error",
            tenant_id=tenant_id or "all",
            worker_id=worker_id,
            mode=mode,
            error=str(exc),
        )
        return 1

    log(
        "shopping_worker_end",
        tenant_id=tenant_id or "all",
        worker_id=worker_id,
        mode=mode,
        handled=handled,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
