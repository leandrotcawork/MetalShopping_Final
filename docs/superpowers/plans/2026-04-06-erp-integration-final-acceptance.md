# ERP Integration Final Acceptance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the ERP Oracle integration live against Sankhya in two gates, then merge `feat/erp-oracle-integration` and close `T-001`.

**Architecture:** Acceptance runs only in the dedicated Oracle worktree on `feat/erp-oracle-integration`. Gate A uses the existing `erp_integrations.auto_promotion` governance policy to keep the run read-only at the canonical-write boundary while still exercising real Oracle extraction, raw persistence, staging, reconciliation, and review behavior. Gate B re-enables auto-promotion and proves the canonical write path for `products`, `prices`, and `inventory` before merge and progress updates.

**Tech Stack:** Go 1.23, PostgreSQL, Oracle via `godror`, PowerShell, static auth mode for local API access, existing ERP HTTP contract and worker runtime.

---

## File Structure Map

### Existing files to inspect or modify

- Inspect: `.env` or local runtime env source in the Oracle worktree
- Inspect: `scripts/start_metalshopping_local.ps1`
- Inspect: `apps/server_core/internal/modules/erp_integrations/application/promotion_consumer.go`
- Inspect: `apps/server_core/internal/modules/erp_integrations/adapters/governance/integration_guard.go`
- Inspect: `apps/integration_worker/cmd/erp-sync/main.go`
- Modify: `docs/PROGRESS.md`
- Modify: `.brain/system-pulse.md`
- Modify: `.brain/session-log.md`
- Modify: `.brain/roadmap.json`

### New files to create

- Create: `docs/ERP_ORACLE_ACCEPTANCE.md`

### Responsibility split

- Runtime env and process startup prove local operator readiness
- ERP HTTP API creates the instance and triggers runs
- Direct SQL checks capture objective evidence from `erp_sync_runs`, `erp_run_entity_steps`, `erp_raw_records`, `erp_staging_records`, and canonical domain tables
- `docs/ERP_ORACLE_ACCEPTANCE.md` records the acceptance evidence and verdict
- `docs/PROGRESS.md` and Nexus brain files record only the final accepted state

---

### Task 1: Prepare the Oracle worktree runtime and acceptance workbook

**Files:**
- Inspect: `.env`
- Inspect: `scripts/start_metalshopping_local.ps1`
- Inspect: `apps/server_core/internal/modules/erp_integrations/adapters/governance/integration_guard.go`
- Inspect: `apps/integration_worker/cmd/erp-sync/main.go`
- Create: `docs/ERP_ORACLE_ACCEPTANCE.md`

- [ ] **Step 1: Confirm the Oracle worktree is the active execution scope**

Run: `git -C .worktrees/erp-oracle-integration status --short`

Expected: only `?? .gocache/` or no meaningful tracked changes

- [ ] **Step 2: Confirm the branch contains the Oracle runtime path and auto-promotion gate**

Run: `rg -n "NO promotion step|AllowAutoPromotion|ERP_SYNC_DATABASE_URL" .worktrees/erp-oracle-integration/apps/integration_worker .worktrees/erp-oracle-integration/apps/server_core`

Expected: matches in:

```text
apps/integration_worker/internal/erp_runtime/runner.go
apps/server_core/internal/modules/erp_integrations/adapters/governance/integration_guard.go
apps/integration_worker/cmd/erp-sync/main.go
```

- [ ] **Step 3: Create the acceptance workbook with the exact evidence sections**

```md
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
```

- [ ] **Step 4: Save the workbook**

```bash
git add docs/ERP_ORACLE_ACCEPTANCE.md
```

- [ ] **Step 5: Commit the acceptance workbook scaffold**

```bash
git commit -m "docs: add ERP Oracle acceptance workbook"
```

---

### Task 2: Verify local runtime prerequisites for live acceptance

**Files:**
- Inspect: `.env`
- Inspect: `.env.example`
- Inspect: `scripts/start_metalshopping_local.ps1`
- Inspect: `docs/KEYCLOAK_LOCAL_BOOTSTRAP.md`

- [ ] **Step 1: Check whether the Oracle worktree already has a local env file**

Run: `Get-ChildItem .worktrees/erp-oracle-integration -Force | Where-Object { $_.Name -like ".env*" } | Select-Object -ExpandProperty Name`

Expected: `.env` appears, or you confirm you must create it from `.env.example`

- [ ] **Step 2: Ensure local API auth can use static mode**

Inspect or set these values in `.worktrees/erp-oracle-integration/.env`:

```dotenv
MS_AUTH_MODE=static
MS_AUTH_STATIC_BEARER_TOKEN=local-dev-token
MS_AUTH_STATIC_SUBJECT_ID=admin-local
MS_AUTH_STATIC_TENANT_ID=bootstrap-local
MS_AUTH_STATIC_EMAIL=admin@metalshopping.local
MS_AUTH_STATIC_NAME=MetalShopping Local Admin
```

- [ ] **Step 3: Ensure Postgres runtime variables are present for server_core**

Use existing local values or populate them explicitly:

```dotenv
APP_ENV=local
APP_PORT=8080
PGHOST=127.0.0.1
PGPORT=5432
PGDATABASE=metalshopping
PGUSER=metalshopping_app
PGPASSWORD=metalshopping_app
PGSSLMODE=disable
```

- [ ] **Step 4: Ensure the worker has a canonical Postgres DSN**

Set this in the shell used to start `erp-sync`:

```powershell
$env:ERP_SYNC_DATABASE_URL = "postgres://metalshopping_app:metalshopping_app@127.0.0.1:5432/metalshopping?sslmode=disable"
```

- [ ] **Step 5: Start `server_core` in the Oracle worktree**

Run:

```powershell
Set-Location "C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalShopping_Final\MetalShopping_Final\.worktrees\erp-oracle-integration"
go run ./apps/server_core/cmd/metalshopping-server
```

Expected: startup logs complete without fatal configuration or migration errors

- [ ] **Step 6: Start `erp-sync` in a second shell in the Oracle worktree**

Run:

```powershell
Set-Location "C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalShopping_Final\MetalShopping_Final\.worktrees\erp-oracle-integration"
$env:ERP_SYNC_DATABASE_URL = "postgres://metalshopping_app:metalshopping_app@127.0.0.1:5432/metalshopping?sslmode=disable"
go run ./apps/integration_worker/cmd/erp-sync
```

Expected: worker logs `erp-sync: database connected` and enters the run-claim loop

- [ ] **Step 7: Verify the ERP API responds under static auth**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "X-Trace-Id" = "erp-acceptance-healthcheck" }
Invoke-RestMethod -Method GET -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances"
```

Expected: HTTP 200 with a JSON list payload

- [ ] **Step 8: Commit only if `.env` was intentionally added to git-ignore-safe tracked files**

If no tracked file changed, do not commit.

---

### Task 3: Force Gate A into read-only mode through the existing governance policy

**Files:**
- Inspect: `apps/server_core/internal/modules/erp_integrations/adapters/governance/integration_guard.go`
- Inspect: `apps/server_core/migrations/0033_erp_governance_defaults.sql`
- Update evidence: `docs/ERP_ORACLE_ACCEPTANCE.md`

- [ ] **Step 1: Verify the policy key and JSON shape**

Run:

```powershell
rg -n "erp_integrations.auto_promotion|allow_auto_promotion" .worktrees/erp-oracle-integration/apps/server_core
```

Expected: the guard and default migration show the key `erp_integrations.auto_promotion` and field `allow_auto_promotion`

- [ ] **Step 2: Set auto-promotion to false for the acceptance tenant**

Run this against the local Postgres database used by `server_core`:

```sql
UPDATE governance_policy_values
SET value_json = '{"allow_auto_promotion": false}'::jsonb,
    updated_at = NOW()
WHERE tenant_id = 'bootstrap-local'
  AND key = 'erp_integrations.auto_promotion';
```

If no row exists, insert one using the repo's existing governance table shape instead of inventing a new table.

- [ ] **Step 3: Verify the policy was applied**

Run:

```sql
SELECT tenant_id, key, value_json
FROM governance_policy_values
WHERE tenant_id = 'bootstrap-local'
  AND key = 'erp_integrations.auto_promotion';
```

Expected: `{"allow_auto_promotion": false}`

- [ ] **Step 4: Record the Gate A policy value in the acceptance workbook**

```md
- auto_promotion policy value: `{"allow_auto_promotion": false}`
```

- [ ] **Step 5: Commit only if you had to change tracked repo files**

If this task only used SQL and workbook edits are pending with later evidence, do not commit yet.

---

### Task 4: Create the live ERP instance and run Gate A

**Files:**
- Inspect: `contracts/api/jsonschema/erp_create_instance_request_v1.schema.json`
- Inspect: `contracts/api/jsonschema/erp_trigger_run_request_v1.schema.json`
- Update evidence: `docs/ERP_ORACLE_ACCEPTANCE.md`

- [ ] **Step 1: Create or confirm the ERP instance payload for Sankhya Oracle**

Use this request body, adjusting only the secret ref if your environment uses a different convention:

```json
{
  "connector_type": "sankhya",
  "display_name": "Sankhya Oracle Acceptance",
  "connection": {
    "kind": "oracle",
    "host": "10.55.10.101",
    "port": 1521,
    "service_name": "ORCL",
    "username": "leandroth",
    "password_secret_ref": "erp/sankhya/password"
  },
  "enabled_entities": ["products", "prices", "inventory"]
}
```

- [ ] **Step 2: Create the instance through the ERP API**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "Content-Type" = "application/json"; "X-Trace-Id" = "erp-acceptance-create-instance" }
$body = @'
{
  "connector_type": "sankhya",
  "display_name": "Sankhya Oracle Acceptance",
  "connection": {
    "kind": "oracle",
    "host": "10.55.10.101",
    "port": 1521,
    "service_name": "ORCL",
    "username": "leandroth",
    "password_secret_ref": "erp/sankhya/password"
  },
  "enabled_entities": ["products", "prices", "inventory"]
}
'@
Invoke-RestMethod -Method POST -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances" -Body $body
```

Expected: HTTP 201 with `instance_id`

- [ ] **Step 3: Trigger the Gate A run**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "Content-Type" = "application/json"; "X-Trace-Id" = "erp-acceptance-gate-a-run" }
$body = @'
{
  "entity_scope": ["products", "prices", "inventory"]
}
'@
Invoke-RestMethod -Method POST -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances/{instance_id}/runs" -Body $body
```

Expected: HTTP 202 with `run_id`

- [ ] **Step 4: Poll the run until it reaches a terminal state**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "X-Trace-Id" = "erp-acceptance-gate-a-poll" }
do {
  $run = Invoke-RestMethod -Method GET -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances/{instance_id}/runs/{run_id}"
  $run.status
  Start-Sleep -Seconds 3
} while ($run.status -in @("pending", "running"))
$run | ConvertTo-Json -Depth 10
```

Expected: terminal status is `completed` or `partial`; `failed` means Gate A did not pass

- [ ] **Step 5: Record API evidence in the workbook**

```md
- instance_id:
- tenant_id: `bootstrap-local`
- run_id:
- API create/list/get results:
```

- [ ] **Step 6: If Gate A fails, stop here and open a debugging branch task list instead of proceeding**

Expected: no Gate B work begins until Gate A passes

---

### Task 5: Capture Gate A database and log evidence

**Files:**
- Update evidence: `docs/ERP_ORACLE_ACCEPTANCE.md`

- [ ] **Step 1: Query the run row**

Run:

```sql
SELECT run_id, tenant_id, instance_id, status, failure_summary, promoted_count, warning_count, rejected_count, review_count
FROM erp_sync_runs
WHERE run_id = '{run_id}';
```

Expected: one row, terminal status, no DSN or password in `failure_summary`

- [ ] **Step 2: Query entity-step rows**

Run:

```sql
SELECT entity_type, status, batch_ordinal, source_cursor, failure_summary
FROM erp_run_entity_steps
WHERE run_id = '{run_id}'
ORDER BY entity_type, batch_ordinal;
```

Expected: rows for `products`, `prices`, and `inventory`

- [ ] **Step 3: Query raw evidence**

Run:

```sql
SELECT entity_type, COUNT(*) AS row_count, MIN(batch_ordinal) AS min_batch, MAX(batch_ordinal) AS max_batch
FROM erp_raw_records
WHERE run_id = '{run_id}'
GROUP BY entity_type
ORDER BY entity_type;
```

Expected: non-zero rows for extracted entities and populated `batch_ordinal`

- [ ] **Step 4: Query staging evidence**

Run:

```sql
SELECT entity_type, validation_status, COUNT(*) AS row_count
FROM erp_staging_records
WHERE run_id = '{run_id}'
GROUP BY entity_type, validation_status
ORDER BY entity_type, validation_status;
```

Expected: staging rows exist and validation outcome is visible

- [ ] **Step 5: Query reconciliation/review evidence**

Run:

```sql
SELECT classification, entity_type, COUNT(*) AS row_count
FROM erp_reconciliation_results
WHERE run_id = '{run_id}'
GROUP BY classification, entity_type
ORDER BY entity_type, classification;
```

And:

```sql
SELECT entity_type, item_status, COUNT(*) AS row_count
FROM erp_review_items
WHERE tenant_id = 'bootstrap-local'
  AND instance_id = '{instance_id}'
GROUP BY entity_type, item_status
ORDER BY entity_type, item_status;
```

Expected: Gate A shows extraction/reconciliation evidence; review rows are acceptable because canonical promotion is intentionally blocked

- [ ] **Step 6: Check for secret leakage in logs and API payload captures**

Run:

```powershell
rg -n "10\\.55\\.10\\.101:1521|Leandrocruz04|password=" .worktrees/erp-oracle-integration -g"!**/.gocache/**"
```

Expected: no hits in tracked logs or captured output files; manually inspect the live terminal output as well

- [ ] **Step 7: Record Gate A verdict in the workbook**

```md
### Verdict
- pass/fail:
- notes:
```

- [ ] **Step 8: Commit the workbook once Gate A evidence is recorded**

```bash
git add docs/ERP_ORACLE_ACCEPTANCE.md
git commit -m "docs: record ERP Oracle Gate A acceptance evidence"
```

---

### Task 6: Re-enable auto-promotion and execute Gate B

**Files:**
- Update evidence: `docs/ERP_ORACLE_ACCEPTANCE.md`

- [ ] **Step 1: Re-enable auto-promotion for the acceptance tenant**

Run:

```sql
UPDATE governance_policy_values
SET value_json = '{"allow_auto_promotion": true}'::jsonb,
    updated_at = NOW()
WHERE tenant_id = 'bootstrap-local'
  AND key = 'erp_integrations.auto_promotion';
```

- [ ] **Step 2: Verify the policy value**

Run:

```sql
SELECT tenant_id, key, value_json
FROM governance_policy_values
WHERE tenant_id = 'bootstrap-local'
  AND key = 'erp_integrations.auto_promotion';
```

Expected: `{"allow_auto_promotion": true}`

- [ ] **Step 3: Trigger the Gate B run using the same instance and entity scope**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "Content-Type" = "application/json"; "X-Trace-Id" = "erp-acceptance-gate-b-run" }
$body = @'
{
  "entity_scope": ["products", "prices", "inventory"]
}
'@
Invoke-RestMethod -Method POST -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances/{instance_id}/runs" -Body $body
```

Expected: HTTP 202 with new `run_id`

- [ ] **Step 4: Poll Gate B to terminal state**

Run:

```powershell
$headers = @{ Authorization = "Bearer local-dev-token"; "X-Trace-Id" = "erp-acceptance-gate-b-poll" }
do {
  $run = Invoke-RestMethod -Method GET -Headers $headers -Uri "http://127.0.0.1:8080/api/v1/erp/instances/{instance_id}/runs/{run_id}"
  $run.status
  Start-Sleep -Seconds 3
} while ($run.status -in @("pending", "running"))
$run | ConvertTo-Json -Depth 10
```

Expected: `completed` or acceptable `partial` with explainable review behavior; unexplained `failed` means Gate B did not pass

- [ ] **Step 5: Query canonical catalog evidence**

Run queries that prove promoted products now exist using the source identifiers observed in Gate A/Gate B. At minimum:

```sql
SELECT product_id, name, manufacturer_reference
FROM catalog_products
WHERE tenant_id = 'bootstrap-local'
ORDER BY created_at DESC
FETCH FIRST 20 ROWS ONLY;
```

- [ ] **Step 6: Query canonical pricing evidence**

```sql
SELECT product_id, price_type, amount, effective_at
FROM pricing_product_prices
WHERE tenant_id = 'bootstrap-local'
ORDER BY effective_at DESC
FETCH FIRST 20 ROWS ONLY;
```

- [ ] **Step 7: Query canonical inventory evidence**

```sql
SELECT product_id, location_code, on_hand_quantity, updated_at
FROM inventory_product_positions
WHERE tenant_id = 'bootstrap-local'
ORDER BY updated_at DESC
FETCH FIRST 20 ROWS ONLY;
```

- [ ] **Step 8: Record Gate B evidence and verdict**

```md
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
```

- [ ] **Step 9: Commit the workbook with Gate B evidence**

```bash
git add docs/ERP_ORACLE_ACCEPTANCE.md
git commit -m "docs: record ERP Oracle Gate B acceptance evidence"
```

---

### Task 7: Run final verification and merge the Oracle branch

**Files:**
- Modify: `docs/PROGRESS.md`
- Modify: `.brain/system-pulse.md`
- Modify: `.brain/session-log.md`
- Modify: `.brain/roadmap.json`

- [ ] **Step 1: Run the required automated verification in the Oracle worktree**

Run:

```powershell
go test -count=1 ./apps/integration_worker/...
go test -count=1 ./apps/server_core/...
powershell -ExecutionPolicy Bypass -File ./scripts/validate_contracts.ps1 -Scope all
powershell -ExecutionPolicy Bypass -File ./scripts/generate_contract_artifacts.ps1 -Target all
```

Expected: all commands pass

- [ ] **Step 2: Update `docs/PROGRESS.md` with the completed acceptance milestone**

Add a line under `## Done` similar to:

```md
- ERP Oracle integration live acceptance completed on `feat/erp-oracle-integration` with Gate A read-only proof and Gate B canonical write proof for `products`, `prices`, and `inventory`
```

- [ ] **Step 3: Update the Nexus brain to reflect the real completed state**

Apply `nexus-checkpoint` only after Gate B and all verification pass so:

```text
T-001 -> done
Phase 1 -> completed
Phase 2 -> in_progress
```

- [ ] **Step 4: Commit docs and brain updates**

```bash
git add docs/PROGRESS.md
git add .brain/system-pulse.md
git add .brain/session-log.md
git add .brain/roadmap.json
git commit -m "docs: close ERP Oracle acceptance and update progress"
```

- [ ] **Step 5: Merge the feature branch into main only after the branch is clean**

Run from the main worktree:

```bash
git status --short
git merge --ff-only feat/erp-oracle-integration
```

Expected: fast-forward merge succeeds; if `git status --short` shows unrelated tracked changes on `main`, stop and isolate them before merge

- [ ] **Step 6: Verify main now contains the Oracle runtime commits**

Run:

```bash
git log --oneline --decorate -12
```

Expected: the Oracle branch commits now appear on `main`

- [ ] **Step 7: Commit only if the merge itself created a commit**

If `--ff-only` was used successfully, there is no merge commit to create.

---

## Self-Review

### Spec coverage

- Gate A read-only proof: covered by Tasks 2 through 5
- Gate B canonical write proof: covered by Task 6
- live Oracle + Sankhya validation: covered by Tasks 2 through 6
- no secret leakage: covered by Task 5
- merge and project-state closure: covered by Task 7

### Placeholder scan

- No `TODO`, `TBD`, or “implement later” markers remain
- Each operator action includes explicit commands, payloads, or SQL

### Type consistency

- The plan consistently uses the existing policy key `erp_integrations.auto_promotion`
- The plan consistently uses entity scope `products`, `prices`, `inventory`
- The plan consistently treats `server_core` promotion as separate from worker extraction
