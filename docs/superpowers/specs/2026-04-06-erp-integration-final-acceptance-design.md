# ERP Integration Final Acceptance Design

## Objective

Close `T-001 0.1 ERP Integration` with a controlled final acceptance flow that proves:

- live Oracle connectivity to Sankhya
- real extraction for `products`, `prices`, and `inventory`
- persisted run, entity-step, raw, and staging evidence
- absence of credential or DSN leakage
- successful canonical promotion in a second gated run

This tranche is complete only after the feature branch is live-validated, merged, and project state is updated.

## Acceptance Strategy

Use a two-stage acceptance model.

### Gate A: Read-only live proof

Run the Oracle-backed ERP sync against the real Sankhya database with entity scope:

- `products`
- `prices`
- `inventory`

This gate must prove extraction and run-state behavior without allowing canonical promotion writes.

Success criteria:

- Oracle connection succeeds with the structured instance config
- the worker claims and executes the run
- all requested entities create entity-step records
- raw rows are persisted with `batch_ordinal`
- staging rows are persisted as expected
- run status and `cursor_state` are visible through the ERP API
- logs and HTTP responses do not expose DSN or password values

### Gate B: Canonical write proof

After Gate A passes, run the same entity scope again with canonical promotion enabled.

Success criteria:

- `products` promote into canonical catalog paths
- `prices` promote into canonical pricing paths
- `inventory` promotes into canonical inventory paths
- reconciliation and review behavior remain coherent for non-promotable records
- rerun behavior stays replay-safe and does not duplicate canonical state improperly
- final run status is coherent with the actual entity outcomes

## Execution Model

All acceptance work happens in `feat/erp-oracle-integration` and its dedicated worktree. `main` is not used for live acceptance execution.

Execution order:

1. Inspect runtime prerequisites
2. Confirm structured Oracle instance configuration
3. Run Gate A and collect evidence
4. Fix defects found in Gate A, if any
5. Run Gate B and collect evidence
6. Fix defects found in Gate B, if any
7. Merge the branch only after both gates pass
8. Update `docs/PROGRESS.md` and Nexus brain
9. Mark `T-001` done

## Components Involved

### Branch/runtime scope

- `apps/integration_worker/cmd/erp-sync`
- `apps/integration_worker/internal/erp_runtime/*`
- `apps/integration_worker/internal/erp_runtime/connectors/sankhya/*`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/*`
- `apps/server_core/internal/modules/erp_integrations/*`

### Evidence stores

- `erp_sync_runs`
- `erp_run_entity_steps`
- `erp_raw_records`
- `erp_staging_records`
- canonical catalog/pricing/inventory tables as applicable for Gate B

## Data Flow

### Gate A

1. `server_core` stores the ERP instance with structured Oracle metadata.
2. A run is created for `products`, `prices`, and `inventory`.
3. `erp-sync` claims the run.
4. The Sankhya connector opens Oracle access through the `godror` query runner.
5. Extraction persists raw and staging evidence plus entity checkpoints.
6. Promotion remains disabled for this gate.
7. Evidence is reviewed in API responses, database state, and logs.

### Gate B

1. The same branch/runtime is used.
2. Promotion is enabled.
3. The same entity scope is executed again.
4. Canonical modules receive real promotions.
5. Reconciliation and review flows handle non-promotable records.
6. Final acceptance is based on canonical writes plus run evidence.

## Failure Handling

Failures are handled by gate, not by intuition.

### If Gate A fails

Treat the problem as one of these until proven otherwise:

- Oracle connectivity or credentials
- query-runner/driver behavior
- Sankhya SQL compatibility
- row mapping
- raw/staging/checkpoint persistence
- worker runtime orchestration

Do not move to canonical-write debugging before Gate A passes.

### If Gate B fails

Treat Oracle extraction as already proven and focus on:

- normalization
- canonical promotion
- reconciliation logic
- review-item generation
- idempotency or replay behavior

## Security Requirements

Acceptance must verify:

- no password value is returned in API payloads beyond governed metadata already allowed by contract
- no DSN string leaks into logs
- no raw credential string is printed in worker failures

## Testing and Evidence Requirements

Before merge, the branch must have:

- `go test -count=1 ./apps/integration_worker/...`
- `go test -count=1 ./apps/server_core/...`
- `powershell -ExecutionPolicy Bypass -File ./scripts/validate_contracts.ps1 -Scope all`
- `powershell -ExecutionPolicy Bypass -File ./scripts/generate_contract_artifacts.ps1 -Target all`

And the live acceptance evidence must include:

- one successful Gate A run record
- one successful Gate B run record
- entity-step evidence for `products`, `prices`, and `inventory`
- raw/staging evidence from Gate A
- canonical write evidence from Gate B

## Merge Criteria

The branch may be merged only if all of the following are true:

- Gate A passed
- Gate B passed
- required automated verification passed
- no secret leakage was observed
- `docs/PROGRESS.md` was updated
- Nexus brain was updated accurately

## Scope Boundary

This acceptance design closes the ERP integration tranche only.

It does not start Analytics Home or any Layer 1 work. Layer 1 may begin only after this acceptance flow passes and `T-001` is formally closed.
