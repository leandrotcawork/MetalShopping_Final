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
- run_id: `run_f31e6145f962335b`
- auto_promotion policy value: `{"allow_auto_promotion": true}` (tenant `tenant_default`)

### Evidence
- API get run result: `completed` (promoted `488420`, review `9855`, rejected `0`, warning `0`)
- reconciliation summary:
  - `products`: promotable `30270`, review_required `9855`
  - `prices`: promotable `430794`
  - `inventory`: promotable `27356`
- canonical catalog evidence: `catalog_products` shows recent rows for tenant `tenant_default` (sample: `prd_bef69d8b692b6e2eb2426993`, `sku=31704`, `name=FECH.E7/94/1952 BAN.CA MANI`, `brand=IMAB`, `status=active`)
- canonical pricing evidence: `pricing_product_prices` shows current rows (sample: `price_eb0a026c0367fc5dc3b2aaf5`, `product_id=prd_bef69d8b692b6e2eb2426993`, `currency=BRL`, `price_amount=182.3100`, `pricing_status=active`)
- canonical inventory evidence: `inventory_product_positions` shows current rows (sample: `pos_eb0a026c0367fc5dc3b2aaf5`, `product_id=prd_bef69d8b692b6e2eb2426993`, `on_hand=11.0000`, `position_status=active`)
- rerun/idempotency notes: Gate B run completed in one batch per entity; no retries observed.

### Verdict
- pass/fail: pass
- notes: Gate B completed with canonical writes present in catalog, pricing, and inventory for tenant `tenant_default`.
