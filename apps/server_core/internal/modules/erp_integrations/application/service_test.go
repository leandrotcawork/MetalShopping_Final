package application_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	erpPostgres "metalshopping/server_core/internal/modules/erp_integrations/adapters/postgres"
	"metalshopping/server_core/internal/modules/erp_integrations/application"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

// ---------------------------------------------------------------------------
// In-memory stubs
// ---------------------------------------------------------------------------

type stubInstanceRepo struct {
	instances  []*domain.IntegrationInstance
	activeFlag bool
	createErr  error
	getErr     error
}

func (r *stubInstanceRepo) Create(_ context.Context, instance *domain.IntegrationInstance) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.instances = append(r.instances, instance)
	return nil
}

func (r *stubInstanceRepo) Get(_ context.Context, _, instanceID string) (*domain.IntegrationInstance, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, i := range r.instances {
		if i.InstanceID == instanceID {
			return i, nil
		}
	}
	return nil, domain.ErrInstanceNotFound
}

func (r *stubInstanceRepo) List(_ context.Context, _ string, limit, offset int) ([]*domain.IntegrationInstance, error) {
	end := offset + limit
	if offset >= len(r.instances) {
		return []*domain.IntegrationInstance{}, nil
	}
	if end > len(r.instances) {
		end = len(r.instances)
	}
	return r.instances[offset:end], nil
}

func (r *stubInstanceRepo) HasActiveInstance(_ context.Context, _ string) (bool, error) {
	return r.activeFlag, nil
}

type stubRunRepo struct {
	runs      []*domain.SyncRun
	createErr error
}

func (r *stubRunRepo) Create(_ context.Context, run *domain.SyncRun) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.runs = append(r.runs, run)
	return nil
}

func (r *stubRunRepo) Get(_ context.Context, _, runID string) (*domain.SyncRun, error) {
	for _, run := range r.runs {
		if run.RunID == runID {
			return run, nil
		}
	}
	return nil, domain.ErrRunNotFound
}

func (r *stubRunRepo) List(_ context.Context, _, _ string, _, _ int) ([]*domain.SyncRun, error) {
	return r.runs, nil
}

type scriptStepKind string

const (
	stepBegin    scriptStepKind = "begin"
	stepExec     scriptStepKind = "exec"
	stepQuery    scriptStepKind = "query"
	stepCommit   scriptStepKind = "commit"
	stepRollback scriptStepKind = "rollback"
)

type scriptStep struct {
	kind         scriptStepKind
	query        string
	args         []any
	rows         [][]driver.Value
	rowsAffected *int64
	assert       func(*testing.T, string, []driver.NamedValue)
}

type scriptState struct {
	t     *testing.T
	mu    sync.Mutex
	steps []scriptStep
	pos   int
}

func newScriptedDB(t *testing.T, steps ...scriptStep) (*sql.DB, *scriptState) {
	t.Helper()

	state := &scriptState{t: t, steps: steps}
	db := sql.OpenDB(&scriptConnector{state: state})
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

type scriptConnector struct {
	state *scriptState
}

func (c *scriptConnector) Connect(context.Context) (driver.Conn, error) {
	return &scriptConn{state: c.state}, nil
}

func (c *scriptConnector) Driver() driver.Driver {
	return scriptDriver{}
}

type scriptDriver struct{}

func (scriptDriver) Open(string) (driver.Conn, error) {
	return nil, fmt.Errorf("open not supported")
}

type scriptConn struct {
	state *scriptState
}

func (c *scriptConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}

func (c *scriptConn) Close() error { return nil }

func (c *scriptConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *scriptConn) Ping(context.Context) error { return nil }

func (c *scriptConn) CheckNamedValue(*driver.NamedValue) error { return nil }

func (c *scriptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.state.expect(stepBegin, "", nil)
	return &scriptTx{state: c.state}, nil
}

func (c *scriptConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	step := c.state.expect(stepExec, query, args)
	rowsAffected := int64(1)
	if step.rowsAffected != nil {
		rowsAffected = *step.rowsAffected
	}
	return driver.RowsAffected(rowsAffected), nil
}

func (c *scriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	step := c.state.expect(stepQuery, query, args)
	return &scriptRows{rows: step.rows}, nil
}

type scriptTx struct {
	state *scriptState
}

func (tx *scriptTx) Commit() error {
	tx.state.expect(stepCommit, "", nil)
	return nil
}

func (tx *scriptTx) Rollback() error {
	tx.state.expect(stepRollback, "", nil)
	return nil
}

type scriptRows struct {
	rows [][]driver.Value
	pos  int
}

func (r *scriptRows) Columns() []string {
	if len(r.rows) == 0 {
		return nil
	}
	cols := make([]string, len(r.rows[0]))
	for i := range cols {
		cols[i] = fmt.Sprintf("col_%d", i)
	}
	return cols
}

func (r *scriptRows) Close() error { return nil }

func (r *scriptRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

func (s *scriptState) expect(kind scriptStepKind, query string, args []driver.NamedValue) scriptStep {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.steps) {
		s.t.Fatalf("unexpected %s operation with query %q", kind, query)
	}

	step := s.steps[s.pos]
	if step.kind != kind {
		s.t.Fatalf("step %d: expected %s, got %s", s.pos, step.kind, kind)
	}
	if step.query != "" && !strings.Contains(query, step.query) {
		s.t.Fatalf("step %d: expected query to contain %q, got %q", s.pos, step.query, query)
	}
	if step.assert != nil {
		step.assert(s.t, query, args)
	} else if len(step.args) > 0 {
		if len(args) != len(step.args) {
			s.t.Fatalf("step %d: expected %d args, got %d", s.pos, len(step.args), len(args))
		}
		for i := range step.args {
			if fmt.Sprint(args[i].Value) != fmt.Sprint(step.args[i]) {
				s.t.Fatalf("step %d: arg %d expected %v, got %v", s.pos, i, step.args[i], args[i].Value)
			}
		}
	}

	s.pos++
	return step
}

func (s *scriptState) done() {
	s.t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos != len(s.steps) {
		s.t.Fatalf("expected %d scripted operations, consumed %d", len(s.steps), s.pos)
	}
}

type stubReviewRepo struct {
	items      []*domain.ReviewItem
	getErr     error
	resolveErr error
}

func (r *stubReviewRepo) Create(_ context.Context, item *domain.ReviewItem) error {
	r.items = append(r.items, item)
	return nil
}

func (r *stubReviewRepo) Get(_ context.Context, _, reviewID string) (*domain.ReviewItem, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.ReviewID == reviewID {
			return item, nil
		}
	}
	return nil, domain.ErrReviewItemNotFound
}

func (r *stubReviewRepo) List(_ context.Context, _ string, _, _ int) ([]*domain.ReviewItem, error) {
	return r.items, nil
}

func (r *stubReviewRepo) Resolve(_ context.Context, _, _ string, status domain.ReviewItemStatus, resolvedBy string, resolvedAt time.Time) error {
	if r.resolveErr != nil {
		return r.resolveErr
	}
	for _, item := range r.items {
		item.ItemStatus = status
		item.ResolvedBy = &resolvedBy
		item.ResolvedAt = &resolvedAt
	}
	return nil
}

type stubEnabledGuard struct {
	err error
}

func (g *stubEnabledGuard) CheckEnabled(_ context.Context, _ string) error {
	return g.err
}

type stubPermChecker struct {
	allowed bool
	err     error
}

func (c *stubPermChecker) CanManageIntegrations(_ context.Context, _, _ string) (bool, error) {
	return c.allowed, c.err
}

// buildService constructs a Service with all stubs wired. The outboxStore is
// nil (no-op) since unit tests don't require outbox persistence.
func buildService(
	instanceRepo ports.InstanceRepository,
	runRepo ports.RunRepository,
	reviewRepo ports.ReviewRepository,
	guard ports.IntegrationEnabledGuard,
	perm ports.PermissionChecker,
) *application.Service {
	return application.NewService(instanceRepo, runRepo, reviewRepo, guard, perm, nil)
}

func stringPtr(v string) *string { return &v }

func stringPtr(v string) *string { return &v }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateInstance_Success(t *testing.T) {
	instanceRepo := &stubInstanceRepo{}
	svc := buildService(
		instanceRepo,
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	cmd := application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	}

	instance, err := svc.CreateInstance(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if instance == nil {
		t.Fatal("expected instance, got nil")
	}
	if instance.TenantID != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", instance.TenantID)
	}
	if instance.Status != domain.InstanceStatusActive {
		t.Errorf("expected status active, got %s", instance.Status)
	}
	if len(instanceRepo.instances) != 1 {
		t.Errorf("expected 1 persisted instance, got %d", len(instanceRepo.instances))
	}
	if got := instanceRepo.instances[0].Connection.Kind; got != domain.ConnectionKindOracle {
		t.Fatalf("expected oracle connection kind, got %s", got)
	}
}

func TestCreateInstance_RejectsMissingPasswordSecretRef(t *testing.T) {
	svc := buildService(
		&stubInstanceRepo{},
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.CreateInstance(context.Background(), application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:        domain.ConnectionKindOracle,
			Host:        "10.55.10.101",
			Port:        1521,
			ServiceName: stringPtr("ORCL"),
			Username:    "leandroth",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrEmptyPasswordSecretRef) {
		t.Fatalf("expected ErrEmptyPasswordSecretRef, got %v", err)
	}
}

func TestCreateInstance_RejectsServiceNameAndSIDTogether(t *testing.T) {
	svc := buildService(
		&stubInstanceRepo{},
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.CreateInstance(context.Background(), application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			SID:               stringPtr("ORCLSID"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrInvalidOracleConnectionTarget) {
		t.Fatalf("expected ErrInvalidOracleConnectionTarget, got %v", err)
	}
}

func TestCreateInstance_AlreadyActive(t *testing.T) {
	svc := buildService(
		&stubInstanceRepo{activeFlag: true},
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.CreateInstance(context.Background(), application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrActiveInstanceExists) {
		t.Errorf("expected ErrActiveInstanceExists, got %v", err)
	}
}

func TestCreateInstance_MissingPasswordSecretRef(t *testing.T) {
	svc := buildService(
		&stubInstanceRepo{},
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.CreateInstance(context.Background(), application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:        domain.ConnectionKindOracle,
			Host:        "10.55.10.101",
			Port:        1521,
			ServiceName: stringPtr("ORCL"),
			Username:    "leandroth",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrEmptyPasswordSecretRef) {
		t.Fatalf("expected ErrEmptyPasswordSecretRef, got %v", err)
	}
}

func TestCreateInstance_IntegrationDisabled(t *testing.T) {
	svc := buildService(
		&stubInstanceRepo{},
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{err: domain.ErrIntegrationDisabled},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.CreateInstance(context.Background(), application.CreateInstanceCommand{
		TenantID:      "tenant-1",
		PrincipalID:   "user-1",
		ConnectorType: domain.ConnectorTypeSankhya,
		DisplayName:   "Test ERP",
		Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrIntegrationDisabled) {
		t.Errorf("expected ErrIntegrationDisabled, got %v", err)
	}
}

func TestTriggerRun_Success(t *testing.T) {
	// Seed an instance that the service will look up.
	instanceRepo := &stubInstanceRepo{
		instances: []*domain.IntegrationInstance{
			{
				InstanceID:      "inst_abc",
				TenantID:        "tenant-1",
				ConnectorType:   domain.ConnectorTypeSankhya,
				DisplayName:     "Test",
				Connection: domain.InstanceConnectionConfig{
					Kind:              domain.ConnectionKindOracle,
					Host:              "10.55.10.101",
					Port:              1521,
					ServiceName:       stringPtr("ORCL"),
					Username:          "leandroth",
					PasswordSecretRef: "erp/sankhya/password",
				},
				EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
				Status:          domain.InstanceStatusActive,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}
	runRepo := &stubRunRepo{}
	svc := buildService(
		instanceRepo,
		runRepo,
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	run, err := svc.TriggerRun(context.Background(), application.TriggerRunCommand{
		TenantID:    "tenant-1",
		PrincipalID: "user-1",
		InstanceID:  "inst_abc",
		RunMode:     domain.RunModeBulk,
		EntityScope: []domain.EntityType{domain.EntityTypeProducts},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run == nil {
		t.Fatal("expected run, got nil")
	}
	if run.Status != domain.RunStatusPending {
		t.Errorf("expected status pending, got %s", run.Status)
	}
	if len(runRepo.runs) != 1 {
		t.Errorf("expected 1 persisted run, got %d", len(runRepo.runs))
	}
}

func TestTriggerRun_AppendsOutboxInSameTransaction(t *testing.T) {
	db, state := newScriptedDB(t,
		scriptStep{kind: stepBegin},
		scriptStep{kind: stepExec, query: "SELECT set_config('app.tenant_id', $1, true)", args: []any{"tenant-1"}},
		scriptStep{kind: stepExec, query: "INSERT INTO erp_sync_runs"},
		scriptStep{
			kind:  stepExec,
			query: "INSERT INTO outbox_events",
			assert: func(t *testing.T, _ string, args []driver.NamedValue) {
				t.Helper()
				if len(args) != 15 {
					t.Fatalf("expected 15 outbox args, got %d", len(args))
				}
				if got := fmt.Sprint(args[3].Value); got != "erp_integrations.run_requested" {
					t.Fatalf("expected run_requested event, got %s", got)
				}
				if got := fmt.Sprint(args[5].Value); got != "tenant-1" {
					t.Fatalf("expected tenant-1 on outbox row, got %s", got)
				}
				runID := fmt.Sprint(args[2].Value)
				wantKey := "erp_run_requested:" + runID
				if got := fmt.Sprint(args[7].Value); got != wantKey {
					t.Fatalf("expected idempotency key %s, got %s", wantKey, got)
				}
			},
		},
		scriptStep{kind: stepCommit},
	)

	repos := erpPostgres.NewRepos(db, outbox.NewStore(db))
	instanceRepo := &stubInstanceRepo{
		instances: []*domain.IntegrationInstance{
			{
				InstanceID:      "inst_abc",
				TenantID:        "tenant-1",
				ConnectorType:   domain.ConnectorTypeSankhya,
				DisplayName:     "Test",
				Connection: domain.InstanceConnectionConfig{
					Kind:              domain.ConnectionKindOracle,
					Host:              "10.55.10.101",
					Port:              1521,
					ServiceName:       stringPtr("ORCL"),
					Username:          "leandroth",
					PasswordSecretRef: "erp/sankhya/password",
				},
				EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
				Status:          domain.InstanceStatusActive,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		},
	}
	svc := application.NewService(
		instanceRepo,
		repos.Runs,
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
		nil,
	)

	run, err := svc.TriggerRun(context.Background(), application.TriggerRunCommand{
		TenantID:    "tenant-1",
		PrincipalID: "user-1",
		InstanceID:  "inst_abc",
		RunMode:     domain.RunModeBulk,
		EntityScope: []domain.EntityType{domain.EntityTypeProducts},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run == nil {
		t.Fatal("expected run, got nil")
	}
	state.done()
}

func TestResolveReview_AlreadyResolved(t *testing.T) {
	now := time.Now()
	resolvedBy := "someone"
	reviewRepo := &stubReviewRepo{
		items: []*domain.ReviewItem{
			{
				ReviewID:   "rev_001",
				TenantID:   "tenant-1",
				ItemStatus: domain.ReviewItemStatusResolved,
				ResolvedAt: &now,
				ResolvedBy: &resolvedBy,
			},
		},
	}
	svc := buildService(
		&stubInstanceRepo{},
		&stubRunRepo{},
		reviewRepo,
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	_, err := svc.ResolveReview(context.Background(), application.ResolveReviewCommand{
		TenantID:    "tenant-1",
		PrincipalID: "user-1",
		ReviewID:    "rev_001",
		Resolution:  domain.ReviewItemStatusResolved,
	})
	if !errors.Is(err, domain.ErrReviewAlreadyResolved) {
		t.Errorf("expected ErrReviewAlreadyResolved, got %v", err)
	}
}

func TestListInstances_Pagination(t *testing.T) {
	// Seed 5 instances.
	instances := make([]*domain.IntegrationInstance, 5)
	for i := range instances {
			instances[i] = &domain.IntegrationInstance{
				InstanceID:      "inst_" + string(rune('a'+i)),
				TenantID:        "tenant-1",
				ConnectorType:   domain.ConnectorTypeSankhya,
				DisplayName:     "Test",
				Connection: domain.InstanceConnectionConfig{
					Kind:              domain.ConnectionKindOracle,
					Host:              "10.55.10.101",
					Port:              1521,
					ServiceName:       stringPtr("ORCL"),
					Username:          "leandroth",
					PasswordSecretRef: "erp/sankhya/password",
				},
				EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
				Status:          domain.InstanceStatusActive,
				CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}
	instanceRepo := &stubInstanceRepo{instances: instances}
	svc := buildService(
		instanceRepo,
		&stubRunRepo{},
		&stubReviewRepo{},
		&stubEnabledGuard{},
		&stubPermChecker{allowed: true},
	)

	// Request page 2 with limit=2, offset=2 — should return instances[2:4].
	page, err := svc.ListInstances(context.Background(), "tenant-1", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 2 {
		t.Errorf("expected 2 items, got %d", len(page))
	}
	if page[0].InstanceID != instances[2].InstanceID {
		t.Errorf("expected first item to be %s, got %s", instances[2].InstanceID, page[0].InstanceID)
	}

	// Offset beyond end returns empty slice.
	page2, err := svc.ListInstances(context.Background(), "tenant-1", 10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page2) != 0 {
		t.Errorf("expected 0 items for out-of-range offset, got %d", len(page2))
	}
}
