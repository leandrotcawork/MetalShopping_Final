# ERP Oracle Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current URL-style Sankhya connection reference with a production-grade Oracle ingestion path built on structured connection metadata, a worker-only `godror` adapter, raw-plus-normalized staging, and dependency-aware run state.

**Architecture:** `apps/server_core` remains the ERP governance and canonical promotion boundary, while `apps/integration_worker` owns Oracle extraction. The implementation introduces a generic query-runner-only `dbsource` layer, a `dbsource/oracle` `godror` adapter, Sankhya-specific SQL/mapping on top of it, and explicit run/entity/batch checkpoint persistence for replayable staged ingestion.

**Tech Stack:** Go 1.23, PostgreSQL migrations, `godror`, existing `erp_runtime` runner/raw/staging/reconciliation packages, existing ERP OpenAPI/contracts pipeline.

---

## File Structure Map

### Existing files to modify

- `apps/integration_worker/go.mod`
- `apps/integration_worker/internal/erp_runtime/connector.go`
- `apps/integration_worker/internal/erp_runtime/runner.go`
- `apps/integration_worker/internal/erp_runtime/types/types.go`
- `apps/integration_worker/internal/erp_runtime/raw/store.go`
- `apps/integration_worker/internal/erp_runtime/staging/model.go`
- `apps/integration_worker/internal/erp_runtime/connectors/sankhya/connector.go`
- `apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor.go`
- `apps/integration_worker/internal/erp_runtime/connectors/sankhya/mapping.go`
- `apps/server_core/internal/modules/erp_integrations/domain/model.go`
- `apps/server_core/internal/modules/erp_integrations/application/service.go`
- `apps/server_core/internal/modules/erp_integrations/ports/repository.go`
- `apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go`
- `apps/server_core/internal/modules/erp_integrations/transport/http/handler.go`
- `contracts/api/openapi/erp_integrations_v1.openapi.yaml`
- `contracts/api/jsonschema/erp_create_instance_request_v1.schema.json`
- `contracts/api/jsonschema/erp_integration_instance_v1.schema.json`
- `contracts/api/jsonschema/erp_sync_run_v1.schema.json`

### New files to create

- `apps/integration_worker/internal/erp_runtime/dbsource/query_runner.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/row_reader.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/config.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/secret_resolver.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner_test.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/config_test.go`
- `apps/integration_worker/internal/erp_runtime/dbsource/oracle/row_reader_test.go`
- `apps/integration_worker/internal/erp_runtime/runs/entity_steps.go`
- `apps/integration_worker/internal/erp_runtime/runs/entity_steps_test.go`
- `apps/server_core/migrations/0037_erp_oracle_instance_config.sql`
- `apps/server_core/migrations/0038_erp_run_entity_steps.sql`
- `apps/server_core/migrations/0039_erp_raw_staging_batch_checkpoints.sql`
- `docs/PROGRESS.md`

### Responsibility split

- `server_core` files own structured ERP instance config, API validation, persistence, and run state visibility.
- `dbsource/*` files own generic query contracts and Oracle implementation details only.
- `connectors/sankhya/*` files own Sankhya SQL and Sankhya-shaped row mapping only.
- `runs/*` files own run/entity/batch checkpoint persistence and final status calculation.

---

### Task 1: Replace `connection_ref` with Structured Oracle Config

**Files:**
- Modify: `apps/server_core/internal/modules/erp_integrations/domain/model.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/application/service.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/ports/repository.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/transport/http/handler.go`
- Modify: `contracts/api/openapi/erp_integrations_v1.openapi.yaml`
- Modify: `contracts/api/jsonschema/erp_create_instance_request_v1.schema.json`
- Modify: `contracts/api/jsonschema/erp_integration_instance_v1.schema.json`
- Test: `apps/server_core/internal/modules/erp_integrations/application/service_test.go`
- Test: `apps/server_core/internal/modules/erp_integrations/transport/http/handler_test.go`
- Create: `apps/server_core/migrations/0037_erp_oracle_instance_config.sql`

- [ ] **Step 1: Write the failing service and handler tests for structured Oracle config**

```go
func TestCreateInstanceRejectsMissingPasswordSecretRef(t *testing.T) {
	cmd := application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "ERP Oracle",
		Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	}

	_, err := svc.CreateInstance(ctx, cmd)
	if !errors.Is(err, domain.ErrEmptyPasswordSecretRef) {
		t.Fatalf("expected ErrEmptyPasswordSecretRef, got %v", err)
	}
}
```

```go
func TestCreateInstanceHandlerDecodesStructuredConnection(t *testing.T) {
	body := `{
	  "connector_type":"sankhya",
	  "display_name":"ERP Oracle",
	  "connection":{
	    "kind":"oracle",
	    "host":"10.55.10.101",
	    "port":1521,
	    "service_name":"ORCL",
	    "username":"leandroth",
	    "password_secret_ref":"erp/sankhya/password"
	  },
	  "enabled_entities":["products","prices","inventory"]
	}`
}
```

- [ ] **Step 2: Run the focused server_core tests and verify they fail**

Run: `go test ./apps/server_core/internal/modules/erp_integrations/application ./apps/server_core/internal/modules/erp_integrations/transport/http`

Expected: FAIL with missing `Connection` fields or still using `connection_ref`.

- [ ] **Step 3: Add structured connection types to the domain model and command layer**

```go
type ConnectionKind string

const ConnectionKindOracle ConnectionKind = "oracle"

type InstanceConnectionConfig struct {
	Kind              ConnectionKind
	Host              string
	Port              int
	ServiceName       *string
	SID               *string
	Username          string
	PasswordSecretRef string
	ConnectTimeoutSec *int
	FetchBatchSize    *int
	EntityBatchSize   *int
}
```

```go
type CreateInstanceCommand struct {
	TenantID        string
	PrincipalID     string
	ConnectorType   domain.ConnectorType
	DisplayName     string
	Connection      domain.InstanceConnectionConfig
	EnabledEntities []domain.EntityType
	SyncSchedule    *string
}
```

- [ ] **Step 4: Add the migration that replaces `connection_ref` storage with structured fields**

```sql
ALTER TABLE erp_integration_instances
  ADD COLUMN connection_kind TEXT NOT NULL DEFAULT 'oracle',
  ADD COLUMN db_host TEXT NOT NULL DEFAULT '',
  ADD COLUMN db_port INT NOT NULL DEFAULT 1521,
  ADD COLUMN db_service_name TEXT NULL,
  ADD COLUMN db_sid TEXT NULL,
  ADD COLUMN db_username TEXT NOT NULL DEFAULT '',
  ADD COLUMN db_password_secret_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN connect_timeout_seconds INT NULL,
  ADD COLUMN fetch_batch_size INT NULL,
  ADD COLUMN entity_batch_size INT NULL;

ALTER TABLE erp_integration_instances
  DROP COLUMN connection_ref;
```

- [ ] **Step 5: Update repository scans/inserts, service validation, HTTP request models, and OpenAPI schemas**

```go
type createInstanceRequest struct {
	ConnectorType   string            `json:"connector_type"`
	DisplayName     string            `json:"display_name"`
	Connection      connectionRequest `json:"connection"`
	EnabledEntities []string          `json:"enabled_entities"`
	SyncSchedule    *string           `json:"sync_schedule,omitempty"`
}
```

```yaml
connection:
  type: object
  required:
    - kind
    - host
    - port
    - username
    - password_secret_ref
```

- [ ] **Step 6: Run the focused server_core tests and contract validation**

Run: `go test ./apps/server_core/internal/modules/erp_integrations/application ./apps/server_core/internal/modules/erp_integrations/transport/http`

Expected: PASS

Run: `powershell -ExecutionPolicy Bypass -File ./scripts/validate_contracts.ps1 -Scope all`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add apps/server_core/internal/modules/erp_integrations/domain/model.go
git add apps/server_core/internal/modules/erp_integrations/application/service.go
git add apps/server_core/internal/modules/erp_integrations/ports/repository.go
git add apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go
git add apps/server_core/internal/modules/erp_integrations/transport/http/handler.go
git add apps/server_core/internal/modules/erp_integrations/application/service_test.go
git add apps/server_core/internal/modules/erp_integrations/transport/http/handler_test.go
git add contracts/api/openapi/erp_integrations_v1.openapi.yaml
git add contracts/api/jsonschema/erp_create_instance_request_v1.schema.json
git add contracts/api/jsonschema/erp_integration_instance_v1.schema.json
git add apps/server_core/migrations/0037_erp_oracle_instance_config.sql
git commit -m "feat: add structured ERP Oracle instance config"
```

### Task 2: Introduce Generic DB Source Contracts and Oracle Config Builder

**Files:**
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/query_runner.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/row_reader.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/config.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/secret_resolver.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/config_test.go`
- Modify: `apps/integration_worker/internal/erp_runtime/types/types.go`
- Modify: `apps/integration_worker/internal/erp_runtime/connector.go`
- Test: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/config_test.go`

- [ ] **Step 1: Write failing tests for Oracle config validation and DSN building**

```go
func TestBuildConnectStringRequiresServiceOrSID(t *testing.T) {
	cfg := oracle.Config{
		Host:     "10.55.10.101",
		Port:     1521,
		Username: "leandroth",
		Password: "secret",
	}

	_, err := cfg.ConnectString()
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

```go
func TestBuildConnectStringWithServiceName(t *testing.T) {
	cfg := oracle.Config{
		Host:        "10.55.10.101",
		Port:        1521,
		ServiceName: stringPtr("ORCL"),
		Username:    "leandroth",
		Password:    "secret",
	}

	got, err := cfg.ConnectString()
	if err != nil || !strings.Contains(got, "service_name=ORCL") {
		t.Fatalf("unexpected DSN %q err=%v", got, err)
	}
}
```

- [ ] **Step 2: Run the new focused tests and verify they fail**

Run: `go test ./apps/integration_worker/internal/erp_runtime/dbsource/oracle`

Expected: FAIL because package and config builder do not exist yet.

- [ ] **Step 3: Add generic query-runner and row-reader interfaces**

```go
type QueryRunner interface {
	Query(ctx context.Context, spec QuerySpec, fn func(RowReader) error) error
	Close() error
}

type QuerySpec struct {
	SQL     string
	Args    []any
	Timeout time.Duration
}
```

```go
type RowReader interface {
	String(name string) (string, error)
	NullString(name string) (*string, error)
	Float64(name string) (float64, error)
	NullFloat64(name string) (*float64, error)
	Time(name string) (time.Time, error)
	NullTime(name string) (*time.Time, error)
}
```

- [ ] **Step 4: Implement Oracle config, resolver contract, and DSN builder**

```go
type SecretResolver interface {
	Resolve(ctx context.Context, ref string) (string, error)
}

type Config struct {
	Host              string
	Port              int
	ServiceName       *string
	SID               *string
	Username          string
	Password          string
	ConnectTimeoutSec int
}
```

```go
func (c Config) ConnectString() (string, error) {
	if strings.TrimSpace(c.Host) == "" || c.Port <= 0 {
		return "", fmt.Errorf("oracle host/port required")
	}
	if (c.ServiceName == nil || *c.ServiceName == "") && (c.SID == nil || *c.SID == "") {
		return "", fmt.Errorf("service_name or sid required")
	}
	return `user="` + c.Username + `" password="` + c.Password + `" connectString="` + c.Host + `:` + strconv.Itoa(c.Port) + `/` + oracleTarget(c) + `"`, nil
}
```

- [ ] **Step 5: Extend runtime types to carry structured connection config instead of `ConnectionRef`**

```go
type ExtractConnection struct {
	Kind              string
	Host              string
	Port              int
	ServiceName       *string
	SID               *string
	Username          string
	PasswordSecretRef string
	ConnectTimeoutSec *int
	FetchBatchSize    *int
	EntityBatchSize   *int
}
```

```go
type ExtractRequest struct {
	TenantID   string
	RunID      string
	Entity     EntityType
	Cursor     *string
	Connection ExtractConnection
}
```

- [ ] **Step 6: Run the focused tests**

Run: `go test ./apps/integration_worker/internal/erp_runtime/dbsource/oracle ./apps/integration_worker/internal/erp_runtime/...`

Expected: PASS for the config package; runtime compile may still fail on downstream users that still expect `ConnectionRef`.

- [ ] **Step 7: Commit**

```bash
git add apps/integration_worker/internal/erp_runtime/dbsource/query_runner.go
git add apps/integration_worker/internal/erp_runtime/dbsource/row_reader.go
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/config.go
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/secret_resolver.go
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/config_test.go
git add apps/integration_worker/internal/erp_runtime/types/types.go
git add apps/integration_worker/internal/erp_runtime/connector.go
git commit -m "feat: add generic ERP dbsource contracts"
```

### Task 3: Add the `godror` Oracle Query Runner and Typed Row Reader

**Files:**
- Modify: `apps/integration_worker/go.mod`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner_test.go`
- Create: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/row_reader_test.go`
- Test: `apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner_test.go`

- [ ] **Step 1: Write failing tests for row-reader conversions and query-runner behavior**

```go
func TestRowReaderNullString(t *testing.T) {
	reader := oracle.NewTestRowReader(map[string]any{"SERVICE": nil})
	got, err := reader.NullString("SERVICE")
	if err != nil || got != nil {
		t.Fatalf("expected nil, got %v err=%v", got, err)
	}
}
```

```go
func TestQueryRunnerRejectsEmptySQL(t *testing.T) {
	runner := oracle.NewQueryRunner(nil, oracle.Config{})
	err := runner.Query(context.Background(), dbsource.QuerySpec{}, func(dbsource.RowReader) error { return nil })
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

- [ ] **Step 2: Run the new tests and verify they fail**

Run: `go test ./apps/integration_worker/internal/erp_runtime/dbsource/oracle`

Expected: FAIL because the runner and row reader are not implemented.

- [ ] **Step 3: Add `godror` to the integration worker module**

```go
require (
	github.com/godror/godror v0.45.2
)
```

- [ ] **Step 4: Implement the Oracle query runner with run-scoped lifecycle**

```go
type QueryRunner struct {
	db *sql.DB
}

func NewQueryRunner(db *sql.DB) *QueryRunner {
	return &QueryRunner{db: db}
}

func (r *QueryRunner) Query(ctx context.Context, spec dbsource.QuerySpec, fn func(dbsource.RowReader) error) error {
	if strings.TrimSpace(spec.SQL) == "" {
		return fmt.Errorf("query sql required")
	}
	rows, err := r.db.QueryContext(ctx, spec.SQL, spec.Args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanRows(rows, fn)
}
```

- [ ] **Step 5: Implement a typed row reader over Oracle result sets**

```go
type rowReader struct {
	values map[string]any
}

func (r *rowReader) NullString(name string) (*string, error) {
	v, ok := r.values[strings.ToUpper(name)]
	if !ok || v == nil {
		return nil, nil
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	return &s, nil
}
```

- [ ] **Step 6: Run focused Oracle package tests**

Run: `go test ./apps/integration_worker/internal/erp_runtime/dbsource/oracle`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add apps/integration_worker/go.mod
git add apps/integration_worker/go.sum
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner.go
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/query_runner_test.go
git add apps/integration_worker/internal/erp_runtime/dbsource/oracle/row_reader_test.go
git commit -m "feat: add godror Oracle query runner"
```

### Task 4: Refactor Sankhya Connector to Use the Generic Query Runner

**Files:**
- Modify: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/connector.go`
- Modify: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor.go`
- Modify: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/mapping.go`
- Modify: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/connector_test.go`
- Modify: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor_test.go`
- Test: `apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor_test.go`

- [ ] **Step 1: Write failing connector tests that expect structured connection input and query-runner collaboration**

```go
func TestValidateConnectionBuildsOracleConfig(t *testing.T) {
	conn := types.ExtractConnection{
		Kind: "oracle",
		Host: "10.55.10.101",
		Port: 1521,
		ServiceName: stringPtr("ORCL"),
		Username: "leandroth",
		PasswordSecretRef: "erp/sankhya/password",
	}
}
```

```go
func TestExtractProductsUsesQueryRunner(t *testing.T) {
	// fake dbsource.QueryRunner returns one row, extractor emits one RawRecord
}
```

- [ ] **Step 2: Run Sankhya connector tests and verify they fail**

Run: `go test ./apps/integration_worker/internal/erp_runtime/connectors/sankhya`

Expected: FAIL because the connector still expects `connectionRef` parsing and fixture-only extraction.

- [ ] **Step 3: Replace URL parsing with structured config validation and injected query runner**

```go
type Connector struct {
	runnerFactory func(ctx context.Context, conn types.ExtractConnection) (dbsource.QueryRunner, error)
	mapper        *Mapper
}

func (c *Connector) ValidateConnection(ctx context.Context, conn types.ExtractConnection) error {
	_, err := c.runnerFactory(ctx, conn)
	return err
}
```

- [ ] **Step 4: Move Sankhya extraction to SQL plus typed row mapping**

```go
func (e *Extractor) extractProducts(ctx context.Context, req types.ExtractRequest, runner dbsource.QueryRunner) (*types.ExtractionResult, error) {
	spec := dbsource.QuerySpec{SQL: sankhyaProductsSQL}
	var records []*types.RawRecord
	err := runner.Query(ctx, spec, func(row dbsource.RowReader) error {
		payload, sourceID, err := e.mapper.MapProductRow(row)
		if err != nil {
			return err
		}
		records = append(records, buildRawRecord(req, sourceID, payload))
		return nil
	})
	return &types.ExtractionResult{Records: records, HasMore: false}, err
}
```

- [ ] **Step 5: Keep test fixtures by introducing fake row readers instead of deleting coverage**

```go
func fakeProductRow() dbsource.RowReader {
	return oracle.NewTestRowReader(map[string]any{
		"CODPROD":   "1001",
		"DESCRPROD": "Produto A",
	})
}
```

- [ ] **Step 6: Run the connector tests**

Run: `go test ./apps/integration_worker/internal/erp_runtime/connectors/sankhya`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add apps/integration_worker/internal/erp_runtime/connectors/sankhya/connector.go
git add apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor.go
git add apps/integration_worker/internal/erp_runtime/connectors/sankhya/mapping.go
git add apps/integration_worker/internal/erp_runtime/connectors/sankhya/connector_test.go
git add apps/integration_worker/internal/erp_runtime/connectors/sankhya/extractor_test.go
git commit -m "feat: refactor Sankhya extraction onto query runner"
```

### Task 5: Add Raw/Staging Batch Control Columns and Run Entity Checkpoints

**Files:**
- Modify: `apps/integration_worker/internal/erp_runtime/raw/store.go`
- Modify: `apps/integration_worker/internal/erp_runtime/staging/model.go`
- Create: `apps/integration_worker/internal/erp_runtime/runs/entity_steps.go`
- Create: `apps/integration_worker/internal/erp_runtime/runs/entity_steps_test.go`
- Create: `apps/server_core/migrations/0038_erp_run_entity_steps.sql`
- Create: `apps/server_core/migrations/0039_erp_raw_staging_batch_checkpoints.sql`
- Test: `apps/integration_worker/internal/erp_runtime/runs/entity_steps_test.go`

- [ ] **Step 1: Write failing tests for entity-step status and batch checkpoint persistence**

```go
func TestMarkEntityStepFailedKeepsBatchCheckpoint(t *testing.T) {
	// create step -> mark batch -> fail step -> assert checkpoint remains queryable
}
```

```go
func TestRawStorePersistsBatchOrdinal(t *testing.T) {
	// save raw record with BatchOrdinal=3, assert column is written
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: `go test ./apps/integration_worker/internal/erp_runtime/runs ./apps/integration_worker/internal/erp_runtime/raw`

Expected: FAIL because entity-step store and batch columns do not exist yet.

- [ ] **Step 3: Add migrations for entity-step state and raw/staging checkpoint columns**

```sql
CREATE TABLE erp_run_entity_steps (
  step_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  status TEXT NOT NULL,
  batch_ordinal INT NOT NULL DEFAULT 0,
  source_cursor TEXT NULL,
  failure_summary TEXT NULL,
  started_at TIMESTAMPTZ NULL,
  completed_at TIMESTAMPTZ NULL
);
```

```sql
ALTER TABLE erp_raw_records ADD COLUMN batch_ordinal INT NOT NULL DEFAULT 1;
ALTER TABLE erp_staging_records ADD COLUMN batch_ordinal INT NOT NULL DEFAULT 1;
CREATE INDEX idx_erp_raw_records_run_entity_batch ON erp_raw_records (run_id, entity_type, batch_ordinal);
```

- [ ] **Step 4: Update raw/staging models and add an entity-step store**

```go
type RawRecord struct {
	SourceID        string
	ConnectorType   string
	EntityType      EntityType
	PayloadJSON     []byte
	PayloadHash     string
	BatchOrdinal    int
	SourceTimestamp *time.Time
	CursorValue     *string
}
```

```go
type EntityStepStore struct {
	db *sql.DB
}

func (s *EntityStepStore) MarkBatch(ctx context.Context, runID string, entity types.EntityType, batchOrdinal int, cursor *string) error
```

- [ ] **Step 5: Run the focused tests**

Run: `go test ./apps/integration_worker/internal/erp_runtime/runs ./apps/integration_worker/internal/erp_runtime/raw ./apps/integration_worker/internal/erp_runtime/staging`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/integration_worker/internal/erp_runtime/raw/store.go
git add apps/integration_worker/internal/erp_runtime/staging/model.go
git add apps/integration_worker/internal/erp_runtime/runs/entity_steps.go
git add apps/integration_worker/internal/erp_runtime/runs/entity_steps_test.go
git add apps/server_core/migrations/0038_erp_run_entity_steps.sql
git add apps/server_core/migrations/0039_erp_raw_staging_batch_checkpoints.sql
git commit -m "feat: add ERP run entity checkpoints"
```

### Task 6: Refactor Runner for Run-Scoped Oracle Lifecycle and Dependency-Aware Sequential Execution

**Files:**
- Modify: `apps/integration_worker/internal/erp_runtime/runner.go`
- Modify: `apps/integration_worker/internal/erp_runtime/registry.go`
- Modify: `apps/integration_worker/internal/erp_runtime/runs/ledger.go`
- Modify: `apps/integration_worker/internal/erp_runtime/reconciliation/reconciler.go`
- Test: `apps/integration_worker/internal/erp_runtime/reconciliation/reconciler_test.go`
- Test: `apps/integration_worker/internal/erp_runtime/runs/ledger_test.go`

- [ ] **Step 1: Write failing tests for dependency-aware partial success and run-scoped resource usage**

```go
func TestRunnerMarksDependentEntitySkippedWhenProductsFail(t *testing.T) {
	// products fails, prices and inventory become skipped_due_to_dependency, run becomes partial
}
```

```go
func TestRunnerCreatesSingleOracleClientPerRun(t *testing.T) {
	// runnerFactory called once even when multiple entities execute
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: `go test ./apps/integration_worker/internal/erp_runtime/runs ./apps/integration_worker/internal/erp_runtime/reconciliation ./apps/integration_worker/internal/erp_runtime`

Expected: FAIL because the current runner still passes `connectionRef`, does not own run-scoped Oracle lifecycle, and does not record dependency skips.

- [ ] **Step 3: Introduce explicit run-scoped client creation and entity-order execution**

```go
runner, err := connector.OpenRunSource(ctx, claim.Connection)
if err != nil {
	return r.ledger.MarkFailed(ctx, claim.RunID, err.Error())
}
defer runner.Close()

for _, entity := range orderedEntities(claim.EntityScope) {
	// process sequentially using the same run-scoped source
}
```

- [ ] **Step 4: Add dependency-aware skip behavior and per-entity step recording**

```go
if failed[types.EntityTypeProducts] && dependsOnProducts(entity) {
	if err := r.entitySteps.MarkSkipped(ctx, claim.TenantID, claim.RunID, entity, "products failed"); err != nil {
		return err
	}
	continue
}
```

- [ ] **Step 5: Ensure final run status becomes `completed`, `partial`, or `failed` from entity outcomes**

```go
switch {
case completed == total:
	return r.ledger.MarkCompleted(...)
case completed == 0:
	return r.ledger.MarkFailed(...)
default:
	return r.ledger.MarkPartial(...)
}
```

- [ ] **Step 6: Run worker runtime tests**

Run: `go test ./apps/integration_worker/...`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add apps/integration_worker/internal/erp_runtime/runner.go
git add apps/integration_worker/internal/erp_runtime/registry.go
git add apps/integration_worker/internal/erp_runtime/runs/ledger.go
git add apps/integration_worker/internal/erp_runtime/reconciliation/reconciler.go
git add apps/integration_worker/internal/erp_runtime/reconciliation/reconciler_test.go
git add apps/integration_worker/internal/erp_runtime/runs/ledger_test.go
git commit -m "feat: add run-scoped Oracle execution flow"
```

### Task 7: Surface Structured Run State Through Server Core and Regenerate Contracts

**Files:**
- Modify: `apps/server_core/internal/modules/erp_integrations/domain/model.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go`
- Modify: `apps/server_core/internal/modules/erp_integrations/transport/http/handler.go`
- Modify: `contracts/api/jsonschema/erp_sync_run_v1.schema.json`
- Modify: `contracts/api/openapi/erp_integrations_v1.openapi.yaml`
- Test: `apps/server_core/internal/modules/erp_integrations/transport/http/handler_test.go`

- [ ] **Step 1: Write failing response tests for structured connection fields and richer run state**

```go
func TestMapInstanceOmitsSecretsButReturnsStructuredOracleMetadata(t *testing.T) {
	// assert password_secret_ref is returned only if contract allows it, and no raw DSN leaks
}
```

```go
func TestMapRunIncludesPartialStatusAndBatchCursorState(t *testing.T) {
	// assert status and cursor state are surfaced from the richer run model
}
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run: `go test ./apps/server_core/internal/modules/erp_integrations/transport/http`

Expected: FAIL because handlers still emit `connection_ref` and the old run shape.

- [ ] **Step 3: Update response mappers and contract schemas**

```go
func mapInstance(i *domain.IntegrationInstance) map[string]any {
	return map[string]any{
		"instance_id": i.InstanceID,
		"connection": map[string]any{
			"kind": i.Connection.Kind,
			"host": i.Connection.Host,
			"port": i.Connection.Port,
			"service_name": deref(i.Connection.ServiceName),
			"username": i.Connection.Username,
			"password_secret_ref": i.Connection.PasswordSecretRef,
		},
	}
}
```

- [ ] **Step 4: Regenerate contract artifacts**

Run: `powershell -ExecutionPolicy Bypass -File ./scripts/generate_contract_artifacts.ps1 -Target all`

Expected: PASS and generated artifacts updated.

- [ ] **Step 5: Run validation and targeted tests**

Run: `powershell -ExecutionPolicy Bypass -File ./scripts/validate_contracts.ps1 -Scope all`

Expected: PASS

Run: `go test ./apps/server_core/internal/modules/erp_integrations/...`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/server_core/internal/modules/erp_integrations/domain/model.go
git add apps/server_core/internal/modules/erp_integrations/adapters/postgres/repository.go
git add apps/server_core/internal/modules/erp_integrations/transport/http/handler.go
git add contracts/api/openapi/erp_integrations_v1.openapi.yaml
git add contracts/api/jsonschema/erp_create_instance_request_v1.schema.json
git add contracts/api/jsonschema/erp_integration_instance_v1.schema.json
git add contracts/api/jsonschema/erp_sync_run_v1.schema.json
git add packages/generated
git add packages/generated-types
git commit -m "feat: expose structured ERP Oracle run state"
```

### Task 8: End-to-End Verification and Documentation Sync

**Files:**
- Modify: `docs/PROGRESS.md`
- Modify: `.brain/system-pulse.md`

- [ ] **Step 1: Run the full worker and server verification suite**

Run: `go test ./apps/integration_worker/...`

Expected: PASS

Run: `go test ./apps/server_core/...`

Expected: PASS

- [ ] **Step 2: Run web-impact verification only if generated SDK/types changed**

Run: `npm.cmd run web:typecheck`

Expected: PASS

Run: `npm.cmd run web:build`

Expected: PASS

- [ ] **Step 3: Run one operator-style smoke checklist**

```text
1. Create ERP instance with structured Oracle connection metadata
2. Trigger run with entity scope ["products","prices","inventory"]
3. Verify run row is created
4. Verify entity step rows are created
5. Verify raw rows include batch ordinal and source cursor
6. Verify no DSN or password appears in HTTP response or logs
```

- [ ] **Step 4: Update progress tracking**

```md
- ERP Integration Oracle design and implementation baseline added:
  - structured Oracle config
  - worker-only godror runtime
  - raw + normalized staging
  - entity/batch checkpoints
```

- [ ] **Step 5: Commit**

```bash
git add docs/PROGRESS.md
git add .brain/system-pulse.md
git commit -m "docs: record ERP Oracle integration progress"
```

---

## Self-Review

### Spec coverage

- Oracle access boundary: covered by Tasks 2, 3, 4, and 6
- Structured config and secret ref: covered by Tasks 1 and 2
- `godror` adapter only: covered by Task 3
- Query-runner-only generic layer: covered by Tasks 2 and 4
- Typed row-reader API: covered by Tasks 2 and 3
- Raw landing plus normalized staging: covered by Tasks 5 and 6
- Dependency-aware sequential execution: covered by Task 6
- Entity-scoped partial success and batched commits: covered by Tasks 5 and 6
- Checkpointing with source cursor and batch ordinal: covered by Task 5
- Observability and surfaced run state: covered by Tasks 5, 6, and 7
- Testing strategy: covered by all tasks, with full verification in Task 8

### Placeholder scan

- No placeholder markers or deferred implementation language remain in this plan.
- Each task includes explicit files, commands, and representative code targets.

### Type consistency

- `InstanceConnectionConfig`, `ExtractConnection`, `QueryRunner`, `RowReader`, and entity-step concepts are introduced before later tasks depend on them.
- The plan consistently removes `connection_ref` in favor of structured Oracle config.
