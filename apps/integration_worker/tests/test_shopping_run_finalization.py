from __future__ import annotations

import copy
import json
import pathlib
import sys
import unittest
from unittest import mock

ROOT = pathlib.Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

import shopping_price_worker as worker


class FakeCursor:
    def __init__(self, connection: "FakeConnection") -> None:
        self.connection = connection

    def __enter__(self) -> "FakeCursor":
        return self

    def __exit__(self, exc_type, exc, tb) -> bool:
        return False

    def execute(self, query: str, params: tuple[object, ...] | None = None) -> None:
        text = query.strip()
        state = self.connection._pending

        if text == "BEGIN":
            self.connection._pending = copy.deepcopy(self.connection._committed)
            return
        if text == "COMMIT":
            self.connection._committed = copy.deepcopy(self.connection._pending)
            return
        if text == "ROLLBACK":
            self.connection._pending = copy.deepcopy(self.connection._committed)
            return
        if text.startswith("SELECT set_config"):
            return
        if state is None:
            raise AssertionError("transaction not started")
        if text.startswith("INSERT INTO shopping_price_runs"):
            run_id, run_status, started_at, finished_at, processed_items, total_items, notes = params or ()
            state["shopping_price_runs"][str(run_id)] = {
                "run_id": str(run_id),
                "run_status": str(run_status),
                "started_at": str(started_at),
                "finished_at": finished_at,
                "processed_items": int(processed_items),
                "total_items": int(total_items),
                "notes": str(notes),
            }
            return
        if text.startswith("UPDATE shopping_price_run_requests"):
            run_request_id = params[-1]
            row = state["shopping_price_run_requests"].setdefault(str(run_request_id), {})
            if "SET request_status = 'running'" in text:
                row["request_status"] = "running"
                row["started_at"] = str(params[0])
                row["run_id"] = str(params[1])
                return
            if "SET request_status = 'failed'" in text:
                row["request_status"] = "failed"
                row["error_message"] = str(params[0])
                row["finished_at"] = "now"
                return
            request_status = params[0]
            row["request_status"] = str(request_status)
            row["finished_at"] = "now"
            return
        if text.startswith("INSERT INTO outbox_events"):
            event_id, aggregate_type, aggregate_id, event_name, event_version, tenant_id, trace_id, idempotency_key, payload_json = params or ()
            state["outbox_events"][str(idempotency_key)] = {
                "event_id": str(event_id),
                "aggregate_type": str(aggregate_type),
                "aggregate_id": str(aggregate_id),
                "event_name": str(event_name),
                "event_version": str(event_version),
                "tenant_id": str(tenant_id),
                "trace_id": str(trace_id),
                "idempotency_key": str(idempotency_key),
                "payload_json": str(payload_json),
                "status": "pending",
            }
            return
        raise AssertionError(f"unexpected query: {text}")


class FakeConnection:
    def __init__(self) -> None:
        self._committed = {
            "shopping_price_runs": {},
            "shopping_price_run_requests": {},
            "outbox_events": {},
        }
        self._pending = copy.deepcopy(self._committed)

    def cursor(self) -> FakeCursor:
        return FakeCursor(self)

    @property
    def committed(self) -> dict[str, dict[str, object]]:
        return self._committed


class ShoppingRunFinalizationTest(unittest.TestCase):
    def _run_request(self) -> worker.RunRequest:
        return worker.RunRequest(
            run_request_id="req_123",
            tenant_id="tenant-1",
            input_mode="catalog",
            input_payload={},
            requested_by="buyer@example.com",
        )

    def test_finalize_run_request_commits_run_row_and_outbox(self) -> None:
        conn = FakeConnection()
        request = self._run_request()

        worker.finalize_run_request(
            conn=conn,  # type: ignore[arg-type]
            tenant_id=request.tenant_id,
            run_request=request,
            run_id="run_123",
            run_status="completed",
            started_at="2026-04-02T10:00:00Z",
            finished_at="2026-04-02T10:01:00Z",
            notes="requested_by=buyer@example.com; input_mode=catalog",
            items=[],
            rows_written=0,
            total_items=0,
            error_message=None,
        )

        self.assertIn("run_123", conn.committed["shopping_price_runs"])
        self.assertEqual("completed", conn.committed["shopping_price_runs"]["run_123"]["run_status"])
        self.assertIn(
            worker.run_completed_idempotency_key("run_123"),
            conn.committed["outbox_events"],
        )
        payload = json.loads(
            conn.committed["outbox_events"][worker.run_completed_idempotency_key("run_123")]["payload_json"]
        )
        self.assertEqual("run_123", payload["run_id"])
        self.assertEqual("completed", payload["run_status"])
        self.assertEqual("completed", conn.committed["shopping_price_run_requests"]["req_123"]["request_status"])

    def test_finalize_run_request_rolls_back_when_outbox_write_fails(self) -> None:
        conn = FakeConnection()
        request = self._run_request()

        with mock.patch.object(
            worker,
            "append_run_completed_outbox_in_tx",
            side_effect=RuntimeError("outbox unavailable"),
        ):
            with self.assertRaises(RuntimeError):
                worker.finalize_run_request(
                    conn=conn,  # type: ignore[arg-type]
                    tenant_id=request.tenant_id,
                    run_request=request,
                    run_id="run_456",
                    run_status="completed",
                    started_at="2026-04-02T10:00:00Z",
                    finished_at="2026-04-02T10:01:00Z",
                    notes="requested_by=buyer@example.com; input_mode=catalog",
                    items=[],
                    rows_written=0,
                    total_items=0,
                    error_message=None,
                )

        self.assertNotIn("run_456", conn.committed["shopping_price_runs"])
        self.assertNotIn(
            worker.run_completed_idempotency_key("run_456"),
            conn.committed["outbox_events"],
        )
        self.assertNotIn("req_123", conn.committed["shopping_price_run_requests"])

    def test_run_completed_idempotency_key_is_stable(self) -> None:
        self.assertEqual(
            worker.run_completed_idempotency_key("run_123"),
            worker.run_completed_idempotency_key("run_123"),
        )
        self.assertEqual(
            "shopping.run_completed:run_123",
            worker.run_completed_idempotency_key("run_123"),
        )


if __name__ == "__main__":
    unittest.main()
