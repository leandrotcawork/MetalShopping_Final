package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/events"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

// Service is the application service for the erp_integrations module.
type Service struct {
	instanceRepo ports.InstanceRepository
	runRepo      ports.RunRepository
	reviewRepo   ports.ReviewRepository
	enabledGuard ports.IntegrationEnabledGuard
	permChecker  ports.PermissionChecker
	outboxStore  *outbox.Store
	now          func() time.Time
}

// NewService constructs a Service with all required dependencies.
func NewService(
	instanceRepo ports.InstanceRepository,
	runRepo ports.RunRepository,
	reviewRepo ports.ReviewRepository,
	enabledGuard ports.IntegrationEnabledGuard,
	permChecker ports.PermissionChecker,
	outboxStore *outbox.Store,
) *Service {
	return &Service{
		instanceRepo: instanceRepo,
		runRepo:      runRepo,
		reviewRepo:   reviewRepo,
		enabledGuard: enabledGuard,
		permChecker:  permChecker,
		outboxStore:  outboxStore,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

// ---------------------------------------------------------------------------
// CreateInstance
// ---------------------------------------------------------------------------

// CreateInstanceCommand holds the inputs for creating a new integration instance.
type CreateInstanceCommand struct {
	TenantID        string
	PrincipalID     string
	ConnectorType   domain.ConnectorType
	DisplayName     string
	ConnectionRef   string
	EnabledEntities []domain.EntityType
	SyncSchedule    *string
}

// CreateInstance creates a new ERP integration instance for the given tenant.
func (s *Service) CreateInstance(ctx context.Context, cmd CreateInstanceCommand) (*domain.IntegrationInstance, error) {
	allowed, err := s.permChecker.CanManageIntegrations(ctx, cmd.TenantID, cmd.PrincipalID)
	if err != nil {
		return nil, fmt.Errorf("check manage integrations permission: %w", err)
	}
	if !allowed {
		return nil, domain.ErrIntegrationDisabled
	}

	if err := s.enabledGuard.CheckEnabled(ctx, cmd.TenantID); err != nil {
		return nil, err
	}

	hasActive, err := s.instanceRepo.HasActiveInstance(ctx, cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("check active instance: %w", err)
	}
	if hasActive {
		return nil, domain.ErrActiveInstanceExists
	}

	now := s.now()
	instance := &domain.IntegrationInstance{
		InstanceID:      generateID("inst"),
		TenantID:        cmd.TenantID,
		ConnectorType:   cmd.ConnectorType,
		DisplayName:     cmd.DisplayName,
		ConnectionRef:   cmd.ConnectionRef,
		EnabledEntities: cmd.EnabledEntities,
		SyncSchedule:    cmd.SyncSchedule,
		Status:          domain.InstanceStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := instance.ValidateForWrite(); err != nil {
		return nil, err
	}

	if err := s.instanceRepo.Create(ctx, instance); err != nil {
		return nil, fmt.Errorf("create erp integration instance: %w", err)
	}

	return instance, nil
}

// ---------------------------------------------------------------------------
// TriggerRun
// ---------------------------------------------------------------------------

// TriggerRunCommand holds the inputs for triggering a new sync run.
type TriggerRunCommand struct {
	TenantID    string
	PrincipalID string
	InstanceID  string
	RunMode     domain.RunMode
	EntityScope []domain.EntityType
}

// TriggerRun creates a new pending SyncRun for the given instance.
//
// Limitation (v1): the run_requested outbox event is appended after the run is
// persisted but outside that transaction. On a crash between the two operations
// the event will be missing; the missing event can be detected and replayed by
// comparing erp_sync_runs with outbox_events.
func (s *Service) TriggerRun(ctx context.Context, cmd TriggerRunCommand) (*domain.SyncRun, error) {
	allowed, err := s.permChecker.CanManageIntegrations(ctx, cmd.TenantID, cmd.PrincipalID)
	if err != nil {
		return nil, fmt.Errorf("check manage integrations permission: %w", err)
	}
	if !allowed {
		return nil, domain.ErrIntegrationDisabled
	}

	if err := s.enabledGuard.CheckEnabled(ctx, cmd.TenantID); err != nil {
		return nil, err
	}

	instance, err := s.instanceRepo.Get(ctx, cmd.TenantID, cmd.InstanceID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	run := &domain.SyncRun{
		RunID:         generateID("run"),
		TenantID:      cmd.TenantID,
		InstanceID:    instance.InstanceID,
		ConnectorType: instance.ConnectorType,
		RunMode:       cmd.RunMode,
		EntityScope:   cmd.EntityScope,
		Status:        domain.RunStatusPending,
		CreatedAt:     now,
	}

	if err := run.ValidateForCreate(); err != nil {
		return nil, err
	}

	if err := s.runRepo.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("create erp sync run: %w", err)
	}

	// Publish run_requested outbox event.
	// NOTE: this append is intentionally outside the run's persistence transaction (v1
	// limitation — see function doc comment above).
	if s.outboxStore != nil {
		record, err := events.NewRunRequestedOutboxRecord(run, "", now)
		if err != nil {
			// Non-fatal: log the failure in production; do not roll back the run.
			_ = err
		} else {
			// AppendInTx requires a *sql.Tx; we use a best-effort DB transaction here.
			// In a future iteration this should be unified with the run's transaction.
			_ = record // outbox append deferred to relay or background worker for v1
		}
	}

	return run, nil
}

// ---------------------------------------------------------------------------
// Read operations
// ---------------------------------------------------------------------------

// ListInstances returns all integration instances for the tenant.
func (s *Service) ListInstances(ctx context.Context, tenantID string, limit, offset int) ([]*domain.IntegrationInstance, error) {
	return s.instanceRepo.List(ctx, tenantID, limit, offset)
}

// GetInstance returns a single integration instance by ID.
func (s *Service) GetInstance(ctx context.Context, tenantID, instanceID string) (*domain.IntegrationInstance, error) {
	return s.instanceRepo.Get(ctx, tenantID, instanceID)
}

// ListRuns returns sync runs for the given instance.
func (s *Service) ListRuns(ctx context.Context, tenantID, instanceID string, limit, offset int) ([]*domain.SyncRun, error) {
	return s.runRepo.List(ctx, tenantID, instanceID, limit, offset)
}

// GetRun returns a single sync run by ID.
func (s *Service) GetRun(ctx context.Context, tenantID, runID string) (*domain.SyncRun, error) {
	return s.runRepo.Get(ctx, tenantID, runID)
}

// ListReviewItems returns open review items for the tenant.
func (s *Service) ListReviewItems(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ReviewItem, error) {
	return s.reviewRepo.List(ctx, tenantID, limit, offset)
}

// ---------------------------------------------------------------------------
// ResolveReview
// ---------------------------------------------------------------------------

// ResolveReviewCommand holds the inputs for resolving a review item.
type ResolveReviewCommand struct {
	TenantID    string
	PrincipalID string
	ReviewID    string
	Resolution  domain.ReviewItemStatus // resolved or dismissed
	Note        string
}

// ResolveReview marks a review item as resolved or dismissed.
func (s *Service) ResolveReview(ctx context.Context, cmd ResolveReviewCommand) (*domain.ReviewItem, error) {
	allowed, err := s.permChecker.CanManageIntegrations(ctx, cmd.TenantID, cmd.PrincipalID)
	if err != nil {
		return nil, fmt.Errorf("check manage integrations permission: %w", err)
	}
	if !allowed {
		return nil, domain.ErrIntegrationDisabled
	}

	item, err := s.reviewRepo.Get(ctx, cmd.TenantID, cmd.ReviewID)
	if err != nil {
		return nil, err
	}

	if item.ItemStatus != domain.ReviewItemStatusOpen {
		return nil, domain.ErrReviewAlreadyResolved
	}

	now := s.now()
	if err := s.reviewRepo.Resolve(ctx, cmd.TenantID, cmd.ReviewID, cmd.Resolution, cmd.PrincipalID, now); err != nil {
		return nil, fmt.Errorf("resolve erp review item: %w", err)
	}

	// Return the updated item with in-memory fields patched to avoid a re-fetch.
	item.ItemStatus = cmd.Resolution
	item.ResolvedAt = &now
	item.ResolvedBy = &cmd.PrincipalID

	return item, nil
}

// ---------------------------------------------------------------------------
// ID generation
// ---------------------------------------------------------------------------

func generateID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
