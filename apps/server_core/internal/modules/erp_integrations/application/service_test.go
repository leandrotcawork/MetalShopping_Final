package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/application"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
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

type stubReviewRepo struct {
	items     []*domain.ReviewItem
	getErr    error
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
		TenantID:        "tenant-1",
		PrincipalID:     "user-1",
		ConnectorType:   domain.ConnectorTypeSankhya,
		DisplayName:     "Test ERP",
		ConnectionRef:   "ref-001",
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
		TenantID:        "tenant-1",
		PrincipalID:     "user-1",
		ConnectorType:   domain.ConnectorTypeSankhya,
		DisplayName:     "Test ERP",
		ConnectionRef:   "ref-001",
		EnabledEntities: []domain.EntityType{domain.EntityTypeProducts},
	})
	if !errors.Is(err, domain.ErrActiveInstanceExists) {
		t.Errorf("expected ErrActiveInstanceExists, got %v", err)
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
		TenantID:        "tenant-1",
		PrincipalID:     "user-1",
		ConnectorType:   domain.ConnectorTypeSankhya,
		DisplayName:     "Test ERP",
		ConnectionRef:   "ref-001",
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
				ConnectionRef:   "ref",
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
			ConnectionRef:   "ref",
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
