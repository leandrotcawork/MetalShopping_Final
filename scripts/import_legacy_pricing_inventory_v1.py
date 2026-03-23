from __future__ import annotations

import argparse
import hashlib
import os
from dataclasses import dataclass
from datetime import datetime, timezone
from decimal import Decimal
from typing import Any

import psycopg


def _norm_text(value: Any) -> str:
    return ("" if value is None else str(value)).strip()


def _sha_id(prefix: str, payload: str, length: int = 24) -> str:
    digest = hashlib.sha256(payload.encode("utf-8")).hexdigest()[:length]
    return f"{prefix}_{digest}"


def _to_decimal(value: Any) -> Decimal:
    raw = _norm_text(value)
    if raw == "":
        return Decimal("0")
    return Decimal(raw).quantize(Decimal("0.0001"))


def _non_negative_decimal(value: Decimal) -> Decimal:
    if value < 0:
        return Decimal("0.0000")
    return value


def _to_datetime(value: Any) -> datetime | None:
    if value is None:
        return None
    if isinstance(value, datetime):
        if value.tzinfo is None:
            return value.replace(tzinfo=timezone.utc)
        return value.astimezone(timezone.utc)
    return None


def _now_utc() -> datetime:
    return datetime.now(timezone.utc)


@dataclass(frozen=True)
class LegacyPricingInventoryRow:
    pn_interno: str
    preco_interno: Decimal
    custo_variavel: Decimal
    custo_medio: Decimal
    estoque_disponivel: Decimal
    dt_compra: datetime | None
    dt_venda: datetime | None
    updated_at: datetime


def _current_dsn_from_env() -> str:
    dsn = _norm_text(os.getenv("MS_DATABASE_URL"))
    if dsn != "":
        return dsn
    host = _norm_text(os.getenv("PGHOST")) or "127.0.0.1"
    port = _norm_text(os.getenv("PGPORT")) or "5432"
    database = _norm_text(os.getenv("PGDATABASE")) or "metalshopping"
    user = _norm_text(os.getenv("PGUSER")) or "metalshopping_app"
    password = _norm_text(os.getenv("PGPASSWORD")) or "metalshopping_app_oscar"
    sslmode = _norm_text(os.getenv("PGSSLMODE")) or "disable"
    return f"postgres://{user}:{password}@{host}:{port}/{database}?sslmode={sslmode}"


def _legacy_dsn_from_env() -> str:
    dsn = _norm_text(os.getenv("MS_LEGACY_DATABASE_URL"))
    if dsn != "":
        return dsn
    return "postgres://metalshopping_app:metalshopping_app_oscar@127.0.0.1:5432/metalshopping_db?sslmode=disable"


def _set_tenant(cur: psycopg.Cursor, tenant_id: str) -> None:
    cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))


def _fetch_legacy_rows(conn: psycopg.Connection) -> list[LegacyPricingInventoryRow]:
    sql = """
        SELECT
          pn_interno,
          preco_interno,
          custo_variavel,
          custo_medio,
          estoque_disponivel,
          dt_compra,
          dt_venda,
          updated_at
        FROM metalshopping.product_erp
        ORDER BY pn_interno ASC
    """
    rows: list[LegacyPricingInventoryRow] = []
    with conn.cursor() as cur:
        cur.execute(sql)
        for record in cur.fetchall():
            rows.append(
                LegacyPricingInventoryRow(
                    pn_interno=_norm_text(record[0]),
                    preco_interno=_to_decimal(record[1]),
                    custo_variavel=_to_decimal(record[2]),
                    custo_medio=_to_decimal(record[3]),
                    estoque_disponivel=_to_decimal(record[4]),
                    dt_compra=_to_datetime(record[5]),
                    dt_venda=_to_datetime(record[6]),
                    updated_at=_to_datetime(record[7]) or _now_utc(),
                )
            )
    return rows


def _fetch_product_lookup(conn: psycopg.Connection, tenant_id: str) -> dict[str, str]:
    sql = """
        SELECT product_id, identifier_value
        FROM catalog_product_identifiers
        WHERE tenant_id = current_tenant_id()
          AND identifier_type = 'pn_interno'
    """
    lookup: dict[str, str] = {}
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        cur.execute(sql)
        for product_id, identifier_value in cur.fetchall():
            token = _norm_text(identifier_value)
            if token != "":
                lookup[token] = _norm_text(product_id)
        cur.execute("COMMIT")
    return lookup


def _upsert_pricing_rows(
    conn: psycopg.Connection,
    *,
    tenant_id: str,
    mapped_rows: list[tuple[str, LegacyPricingInventoryRow]],
    updated_by: str,
) -> int:
    if len(mapped_rows) == 0:
        return 0

    sql = """
        INSERT INTO pricing_product_prices (
          price_id,
          tenant_id,
          product_id,
          currency_code,
          price_amount,
          replacement_cost_amount,
          average_cost_amount,
          pricing_status,
          effective_from,
          effective_to,
          origin_type,
          origin_ref,
          reason_code,
          updated_by
        )
        VALUES (
          %s, current_tenant_id(), %s, 'BRL', %s, %s, %s, 'active', %s, NULL,
          'import', %s, 'legacy_migration_v1', %s
        )
        ON CONFLICT (tenant_id, product_id) WHERE effective_to IS NULL
        DO UPDATE SET
          price_amount = EXCLUDED.price_amount,
          replacement_cost_amount = EXCLUDED.replacement_cost_amount,
          average_cost_amount = EXCLUDED.average_cost_amount,
          pricing_status = EXCLUDED.pricing_status,
          effective_from = EXCLUDED.effective_from,
          origin_type = EXCLUDED.origin_type,
          origin_ref = EXCLUDED.origin_ref,
          reason_code = EXCLUDED.reason_code,
          updated_by = EXCLUDED.updated_by,
          updated_at = NOW()
        WHERE pricing_product_prices.origin_type = 'import'
           OR pricing_product_prices.origin_ref = EXCLUDED.origin_ref
    """

    affected = 0
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        for product_id, row in mapped_rows:
            price_id = _sha_id("price", f"{tenant_id}|{product_id}|legacy_pricing_inventory_v1")
            cur.execute(
                sql,
                (
                    price_id,
                    product_id,
                    row.preco_interno,
                    row.custo_variavel,
                    row.custo_medio,
                    row.updated_at,
                    "legacy:metalshopping_db.product_erp",
                    updated_by,
                ),
            )
            affected += int(cur.rowcount or 0)
        cur.execute("COMMIT")
    return affected


def _upsert_inventory_rows(
    conn: psycopg.Connection,
    *,
    tenant_id: str,
    mapped_rows: list[tuple[str, LegacyPricingInventoryRow]],
    updated_by: str,
) -> tuple[int, int]:
    if len(mapped_rows) == 0:
        return 0, 0

    sql = """
        INSERT INTO inventory_product_positions (
          position_id,
          tenant_id,
          product_id,
          on_hand_quantity,
          last_purchase_at,
          last_sale_at,
          position_status,
          effective_from,
          effective_to,
          origin_type,
          origin_ref,
          reason_code,
          updated_by
        )
        VALUES (
          %s, current_tenant_id(), %s, %s, %s, %s, 'active', %s, NULL,
          'import', %s, 'legacy_migration_v1', %s
        )
        ON CONFLICT (tenant_id, product_id) WHERE effective_to IS NULL
        DO UPDATE SET
          on_hand_quantity = EXCLUDED.on_hand_quantity,
          last_purchase_at = EXCLUDED.last_purchase_at,
          last_sale_at = EXCLUDED.last_sale_at,
          position_status = EXCLUDED.position_status,
          effective_from = EXCLUDED.effective_from,
          origin_type = EXCLUDED.origin_type,
          origin_ref = EXCLUDED.origin_ref,
          reason_code = EXCLUDED.reason_code,
          updated_by = EXCLUDED.updated_by,
          updated_at = NOW()
        WHERE inventory_product_positions.origin_type = 'import'
           OR inventory_product_positions.origin_ref = EXCLUDED.origin_ref
    """

    affected = 0
    normalized_negative_stock = 0
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        for product_id, row in mapped_rows:
            position_id = _sha_id("pos", f"{tenant_id}|{product_id}|legacy_pricing_inventory_v1")
            on_hand_quantity = _non_negative_decimal(row.estoque_disponivel)
            if on_hand_quantity != row.estoque_disponivel:
                normalized_negative_stock += 1
            cur.execute(
                sql,
                (
                    position_id,
                    product_id,
                    on_hand_quantity,
                    row.dt_compra,
                    row.dt_venda,
                    row.updated_at,
                    "legacy:metalshopping_db.product_erp",
                    updated_by,
                ),
            )
            affected += int(cur.rowcount or 0)
        cur.execute("COMMIT")
    return affected, normalized_negative_stock


def _count_current_rows(conn: psycopg.Connection, tenant_id: str) -> tuple[int, int]:
    with conn.cursor() as cur:
        cur.execute("BEGIN")
        _set_tenant(cur, tenant_id)
        cur.execute("SELECT COUNT(*) FROM pricing_product_prices WHERE tenant_id = current_tenant_id() AND effective_to IS NULL")
        pricing_count = int(cur.fetchone()[0])
        cur.execute("SELECT COUNT(*) FROM inventory_product_positions WHERE tenant_id = current_tenant_id() AND effective_to IS NULL")
        inventory_count = int(cur.fetchone()[0])
        cur.execute("COMMIT")
    return pricing_count, inventory_count


def main() -> None:
    parser = argparse.ArgumentParser(description="Import legacy pricing and inventory into the current MetalShopping schema.")
    parser.add_argument("--tenant-id", default=os.getenv("MS_TENANT_ID", "tenant_default"))
    parser.add_argument("--legacy-dsn", default=_legacy_dsn_from_env())
    parser.add_argument("--current-dsn", default=_current_dsn_from_env())
    parser.add_argument("--updated-by", default="system:legacy_pricing_inventory_import_v1")
    parser.add_argument("--apply", action="store_true", help="Write pricing and inventory rows into the current database.")
    args = parser.parse_args()

    with psycopg.connect(args.legacy_dsn) as legacy_conn, psycopg.connect(args.current_dsn) as current_conn:
        legacy_rows = _fetch_legacy_rows(legacy_conn)
        product_lookup = _fetch_product_lookup(current_conn, args.tenant_id)

        mapped_rows: list[tuple[str, LegacyPricingInventoryRow]] = []
        unmatched_pn: list[str] = []
        for row in legacy_rows:
            product_id = product_lookup.get(row.pn_interno)
            if product_id is None:
                unmatched_pn.append(row.pn_interno)
                continue
            mapped_rows.append((product_id, row))

        print(
            {
                "legacy_rows": len(legacy_rows),
                "mapped_rows": len(mapped_rows),
                "unmatched_rows": len(unmatched_pn),
                "tenant_id": args.tenant_id,
                "mode": "apply" if args.apply else "dry_run",
            }
        )

        if len(unmatched_pn) > 0:
            print({"unmatched_sample": unmatched_pn[:10]})

        if not args.apply:
            return

        pricing_affected = _upsert_pricing_rows(
            current_conn,
            tenant_id=args.tenant_id,
            mapped_rows=mapped_rows,
            updated_by=args.updated_by,
        )
        inventory_affected, normalized_negative_stock = _upsert_inventory_rows(
            current_conn,
            tenant_id=args.tenant_id,
            mapped_rows=mapped_rows,
            updated_by=args.updated_by,
        )
        pricing_count, inventory_count = _count_current_rows(current_conn, args.tenant_id)
        print(
            {
                "pricing_affected": pricing_affected,
                "inventory_affected": inventory_affected,
                "inventory_negative_stock_normalized": normalized_negative_stock,
                "pricing_open_rows": pricing_count,
                "inventory_open_rows": inventory_count,
            }
        )


if __name__ == "__main__":
    main()
