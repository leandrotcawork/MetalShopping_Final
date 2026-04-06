# ERP Oracle Acceptance

## Scope

- Branch: `feat/erp-oracle-integration`
- Worktree: `.worktrees/erp-oracle-integration`
- Entity scope: `products`, `prices`, `inventory`

## Gate A - Read-only live proof

### Runtime
- server_core start command:
- erp-sync start command:
- auth mode:

### Inputs
- instance_id:
- tenant_id:
- run_id:
- auto_promotion policy value:

### Evidence
- API create/list/get results:
- `erp_sync_runs` row summary:
- `erp_run_entity_steps` row summary:
- `erp_raw_records` row summary:
- `erp_staging_records` row summary:
- review/reconciliation summary:
- secret leakage check:

### Verdict
- pass/fail:
- notes:

## Gate B - Canonical write proof

### Inputs
- run_id:
- auto_promotion policy value:

### Evidence
- API get run result:
- reconciliation summary:
- canonical catalog evidence:
- canonical pricing evidence:
- canonical inventory evidence:
- rerun/idempotency notes:

### Verdict
- pass/fail:
- notes:
