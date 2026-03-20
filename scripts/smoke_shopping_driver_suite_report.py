from __future__ import annotations

import argparse
import json
from typing import Any

import psycopg


def _query_summary(
    *,
    database_url: str,
    tenant_id: str,
    run_request_id: str,
    supplier_code: str,
) -> dict[str, Any]:
    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cur:
            cur.execute("BEGIN")
            cur.execute("SELECT set_config('app.tenant_id', %s, true)", (tenant_id,))
            cur.execute(
                """
                SELECT
                  run_request_id,
                  request_status,
                  input_mode,
                  run_id,
                  requested_at,
                  finished_at
                FROM shopping_price_run_requests
                WHERE tenant_id = current_tenant_id()
                  AND run_request_id = %s
                """,
                (run_request_id,),
            )
            request_row = cur.fetchone()
            if request_row is None:
                cur.execute("ROLLBACK")
                raise RuntimeError(f"run_request_id not found: {run_request_id}")

            run_id = str(request_row[3]) if request_row[3] is not None else ""
            status_counts: list[tuple[Any, ...]] = []
            channel_counts: list[tuple[Any, ...]] = []
            samples: list[tuple[Any, ...]] = []
            total_items = 0

            if run_id != "":
                cur.execute(
                    """
                    SELECT item_status, COUNT(*)
                    FROM shopping_price_run_items
                    WHERE tenant_id = current_tenant_id()
                      AND run_id = %s
                      AND supplier_code = %s
                    GROUP BY item_status
                    ORDER BY item_status
                    """,
                    (run_id, supplier_code),
                )
                status_counts = cur.fetchall()

                cur.execute(
                    """
                    SELECT channel, item_status, COUNT(*)
                    FROM shopping_price_run_items
                    WHERE tenant_id = current_tenant_id()
                      AND run_id = %s
                      AND supplier_code = %s
                    GROUP BY channel, item_status
                    ORDER BY channel, item_status
                    """,
                    (run_id, supplier_code),
                )
                channel_counts = cur.fetchall()

                cur.execute(
                    """
                    SELECT
                      COUNT(*)
                    FROM shopping_price_run_items
                    WHERE tenant_id = current_tenant_id()
                      AND run_id = %s
                      AND supplier_code = %s
                    """,
                    (run_id, supplier_code),
                )
                count_row = cur.fetchone()
                total_items = int(count_row[0]) if count_row is not None else 0

                cur.execute(
                    """
                    SELECT
                      product_id,
                      observed_price::text,
                      currency_code,
                      COALESCE(http_status::text, '') AS http_status,
                      item_status,
                      COALESCE(chosen_seller_json->>'lookup_term', '') AS lookup_term
                    FROM shopping_price_run_items
                    WHERE tenant_id = current_tenant_id()
                      AND run_id = %s
                      AND supplier_code = %s
                    ORDER BY observed_at DESC
                    LIMIT 5
                    """,
                    (run_id, supplier_code),
                )
                samples = cur.fetchall()

            cur.execute("COMMIT")

    return {
        "run_request_id": str(request_row[0]),
        "request_status": str(request_row[1]),
        "input_mode": str(request_row[2]),
        "run_id": run_id,
        "requested_at": str(request_row[4]),
        "finished_at": str(request_row[5]) if request_row[5] is not None else None,
        "supplier_code": supplier_code,
        "total_items": total_items,
        "status_counts": [{"item_status": str(row[0]), "count": int(row[1])} for row in status_counts],
        "channel_status_counts": [
            {"channel": str(row[0]), "item_status": str(row[1]), "count": int(row[2])}
            for row in channel_counts
        ],
        "samples": [
            {
                "product_id": str(row[0]),
                "observed_price": str(row[1]),
                "currency_code": str(row[2]),
                "http_status": str(row[3]) if row[3] is not None else "",
                "item_status": str(row[4]),
                "lookup_term": str(row[5]),
            }
            for row in samples
        ],
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Collect Shopping driver smoke suite evidence.")
    parser.add_argument("--database-url", required=True)
    parser.add_argument("--tenant-id", required=True)
    parser.add_argument("--run-request-id", required=True)
    parser.add_argument("--supplier-code", required=True)
    args = parser.parse_args()

    summary = _query_summary(
        database_url=args.database_url,
        tenant_id=args.tenant_id,
        run_request_id=args.run_request_id,
        supplier_code=args.supplier_code.strip().upper(),
    )
    print(json.dumps(summary, ensure_ascii=True))


if __name__ == "__main__":
    main()
