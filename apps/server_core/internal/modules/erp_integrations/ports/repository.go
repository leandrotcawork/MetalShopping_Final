package ports

import (
	"context"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
)

// InstanceRepository manages persistence for ERP integration instances.
type InstanceRepository interface {
	Create(ctx context.Context, instance *domain.IntegrationInstance) error
	Get(ctx context.Context, tenantID, instanceID string) (*domain.IntegrationInstance, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.IntegrationInstance, error)
	HasActiveInstance(ctx context.Context, tenantID string) (bool, error)
}

// RunRepository manages persistence for ERP sync runs.
type RunRepository interface {
	Create(ctx context.Context, run *domain.SyncRun) error
	Get(ctx context.Context, tenantID, runID string) (*domain.SyncRun, error)
	List(ctx context.Context, tenantID, instanceID string, limit, offset int) ([]*domain.SyncRun, error)
}

// ReviewRepository manages persistence for ERP review items.
type ReviewRepository interface {
	Create(ctx context.Context, item *domain.ReviewItem) error
	Get(ctx context.Context, tenantID, reviewID string) (*domain.ReviewItem, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ReviewItem, error)
	Resolve(ctx context.Context, tenantID, reviewID string, status domain.ReviewItemStatus, resolvedBy string, resolvedAt time.Time) error
}

// ReconciliationReader provides read access and claim semantics for reconciliation
// results that are candidates for promotion to the canonical data store.
type ReconciliationReader interface {
	// ListPromotableResults returns pending reconciliation results eligible for promotion.
	ListPromotableResults(ctx context.Context, tenantID string, limit int) ([]*domain.ReconciliationResult, error)
	// ClaimForPromotion atomically sets promotion_status = 'promoting' for a single result.
	ClaimForPromotion(ctx context.Context, tenantID, reconciliationID string) error
	// MarkPromoted records a successful promotion and stores the canonical entity ID.
	MarkPromoted(ctx context.Context, tenantID, reconciliationID, canonicalID string) error
	// MarkPromotionFailed records a failed promotion attempt.
	MarkPromotionFailed(ctx context.Context, tenantID, reconciliationID string) error
}

// PermissionChecker verifies tenant-scoped access rights.
type PermissionChecker interface {
	CanManageIntegrations(ctx context.Context, tenantID, principalID string) (bool, error)
}

// IntegrationEnabledGuard enforces the feature flag that gates ERP integration usage.
type IntegrationEnabledGuard interface {
	// CheckEnabled returns domain.ErrIntegrationDisabled if ERP integration is not
	// enabled for the given tenant.
	CheckEnabled(ctx context.Context, tenantID string) error
}

// AutoPromotionGuard enforces the policy that controls automatic promotion of
// reconciled ERP records to the canonical data store.
type AutoPromotionGuard interface {
	// CheckAutoPromotion returns domain.ErrAutoPromotionDisabled if auto-promotion
	// is not permitted for the given tenant.
	CheckAutoPromotion(ctx context.Context, tenantID string) error
}
