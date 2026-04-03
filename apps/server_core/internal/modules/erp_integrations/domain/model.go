package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// ConnectorType identifies the ERP system being integrated.
type ConnectorType string

const (
	ConnectorTypeSankhya ConnectorType = "sankhya"
)

func (c ConnectorType) IsValid() bool {
	return c == ConnectorTypeSankhya
}

// InstanceStatus represents the operational state of an integration instance.
type InstanceStatus string

const (
	InstanceStatusActive   InstanceStatus = "active"
	InstanceStatusPaused   InstanceStatus = "paused"
	InstanceStatusDisabled InstanceStatus = "disabled"
)

func (s InstanceStatus) IsValid() bool {
	return s == InstanceStatusActive || s == InstanceStatusPaused || s == InstanceStatusDisabled
}

// EntityType identifies which ERP data domain is being synced.
type EntityType string

const (
	EntityTypeProducts  EntityType = "products"
	EntityTypePrices    EntityType = "prices"
	EntityTypeCosts     EntityType = "costs"
	EntityTypeInventory EntityType = "inventory"
	EntityTypeSales     EntityType = "sales"
	EntityTypePurchases EntityType = "purchases"
	EntityTypeCustomers EntityType = "customers"
	EntityTypeSuppliers EntityType = "suppliers"
)

func (e EntityType) IsValid() bool {
	return e == EntityTypeProducts ||
		e == EntityTypePrices ||
		e == EntityTypeCosts ||
		e == EntityTypeInventory ||
		e == EntityTypeSales ||
		e == EntityTypePurchases ||
		e == EntityTypeCustomers ||
		e == EntityTypeSuppliers
}

// RunMode describes how a sync run is triggered and scoped.
type RunMode string

const (
	RunModeBulk        RunMode = "bulk"
	RunModeIncremental RunMode = "incremental"
	RunModeManualRerun RunMode = "manual_rerun"
)

func (r RunMode) IsValid() bool {
	return r == RunModeBulk || r == RunModeIncremental || r == RunModeManualRerun
}

// RunStatus tracks the lifecycle of a sync run.
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusPartial   RunStatus = "partial"
)

func (r RunStatus) IsValid() bool {
	return r == RunStatusPending ||
		r == RunStatusRunning ||
		r == RunStatusCompleted ||
		r == RunStatusFailed ||
		r == RunStatusPartial
}

// ReviewSeverity classifies the urgency of a review item.
type ReviewSeverity string

const (
	ReviewSeverityInfo     ReviewSeverity = "info"
	ReviewSeverityWarning  ReviewSeverity = "warning"
	ReviewSeverityError    ReviewSeverity = "error"
	ReviewSeverityCritical ReviewSeverity = "critical"
)

func (r ReviewSeverity) IsValid() bool {
	return r == ReviewSeverityInfo ||
		r == ReviewSeverityWarning ||
		r == ReviewSeverityError ||
		r == ReviewSeverityCritical
}

// ReviewItemStatus tracks the resolution lifecycle of a review item.
type ReviewItemStatus string

const (
	ReviewItemStatusOpen      ReviewItemStatus = "open"
	ReviewItemStatusResolved  ReviewItemStatus = "resolved"
	ReviewItemStatusDismissed ReviewItemStatus = "dismissed"
)

func (r ReviewItemStatus) IsValid() bool {
	return r == ReviewItemStatusOpen || r == ReviewItemStatusResolved || r == ReviewItemStatusDismissed
}

// PromotionStatus tracks whether a reconciliation result has been promoted
// to the production data store. Matches migration 0031.
type PromotionStatus string

const (
	PromotionStatusPending   PromotionStatus = "pending"
	PromotionStatusPromoting PromotionStatus = "promoting"
	PromotionStatusPromoted  PromotionStatus = "promoted"
	PromotionStatusFailed    PromotionStatus = "failed"
)

func (p PromotionStatus) IsValid() bool {
	return p == PromotionStatusPending ||
		p == PromotionStatusPromoting ||
		p == PromotionStatusPromoted ||
		p == PromotionStatusFailed
}

// ReconciliationClassification describes the outcome of a reconciliation check.
type ReconciliationClassification string

const (
	ClassificationPromotable            ReconciliationClassification = "promotable"
	ClassificationPromotableWithWarning ReconciliationClassification = "promotable_with_warning"
	ClassificationReviewRequired        ReconciliationClassification = "review_required"
	ClassificationRejected              ReconciliationClassification = "rejected"
)

func (c ReconciliationClassification) IsValid() bool {
	return c == ClassificationPromotable ||
		c == ClassificationPromotableWithWarning ||
		c == ClassificationReviewRequired ||
		c == ClassificationRejected
}

// ---------------------------------------------------------------------------
// Entity structs
// ---------------------------------------------------------------------------

// IntegrationInstance represents a tenant's configured ERP integration.
type IntegrationInstance struct {
	InstanceID      string
	TenantID        string
	ConnectorType   ConnectorType
	DisplayName     string
	ConnectionRef   string
	EnabledEntities []EntityType
	SyncSchedule    *string
	Status          InstanceStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ValidateForWrite checks that all required fields are present and valid.
func (i *IntegrationInstance) ValidateForWrite() error {
	if strings.TrimSpace(i.TenantID) == "" {
		return ErrEmptyTenantID
	}
	if !i.ConnectorType.IsValid() {
		return ErrInvalidConnectorType
	}
	if strings.TrimSpace(i.DisplayName) == "" {
		return ErrEmptyDisplayName
	}
	if len(i.EnabledEntities) == 0 {
		return ErrEmptyEnabledEntities
	}
	for _, e := range i.EnabledEntities {
		if !e.IsValid() {
			return ErrInvalidEntityType
		}
	}
	if !i.Status.IsValid() {
		return ErrInvalidInstanceStatus
	}
	return nil
}

// SyncRun represents a single execution of an ERP sync operation.
type SyncRun struct {
	RunID          string
	TenantID       string
	InstanceID     string
	ConnectorType  ConnectorType
	RunMode        RunMode
	EntityScope    []EntityType
	Status         RunStatus
	StartedAt      *time.Time
	CompletedAt    *time.Time
	PromotedCount  int
	WarningCount   int
	RejectedCount  int
	ReviewCount    int
	FailureSummary *string
	CursorState    *string // JSON string
	CreatedAt      time.Time
}

// ValidateForCreate checks that all required fields are present and valid
// before a new SyncRun is persisted.
func (r *SyncRun) ValidateForCreate() error {
	if strings.TrimSpace(r.TenantID) == "" {
		return ErrEmptyTenantID
	}
	if strings.TrimSpace(r.InstanceID) == "" {
		return ErrInstanceNotFound
	}
	if !r.ConnectorType.IsValid() {
		return ErrInvalidConnectorType
	}
	if !r.RunMode.IsValid() {
		return ErrInvalidRunMode
	}
	if len(r.EntityScope) == 0 {
		return ErrEmptyEntityScope
	}
	for _, e := range r.EntityScope {
		if !e.IsValid() {
			return ErrInvalidEntityType
		}
	}
	return nil
}

// ReviewItem represents a data record that requires human review before
// it can be promoted to production.
type ReviewItem struct {
	ReviewID             string
	TenantID             string
	InstanceID           string
	ConnectorType        ConnectorType
	EntityType           EntityType
	SourceID             string
	RunID                string
	Severity             ReviewSeverity
	ReasonCode           string
	ProblemSummary       string
	RawPayloadRef        string
	StagingID            string  // FK reference to erp_staging_records
	ReconciliationID     string  // FK reference to erp_reconciliation_results
	StagingSnapshot      *string // optional JSON snapshot of staging payload
	ReconciliationOutput *string // optional JSON snapshot of reconciliation result
	RecommendedAction    string
	ItemStatus           ReviewItemStatus
	ResolvedAt           *time.Time
	ResolvedBy           *string
	CreatedAt            time.Time
}

// ReconciliationResult captures the outcome of reconciling a single staged
// ERP record against the canonical data store.
type ReconciliationResult struct {
	ReconciliationID string
	TenantID         string
	RunID            string
	StagingID        string
	EntityType       EntityType
	SourceID         string
	CanonicalID      *string
	Action           string
	Classification   ReconciliationClassification
	ReasonCode       string
	WarningDetails   *string // JSON string
	ReconciledAt     time.Time
	PromotionStatus  PromotionStatus
}

// StagingRecord captures the normalized ERP payload that is eligible for
// promotion into a canonical domain module.
type StagingRecord struct {
	StagingID        string
	TenantID         string
	RunID            string
	RawID            string
	EntityType       EntityType
	SourceID         string
	NormalizedJSON   json.RawMessage
	ValidationStatus string
	ValidationErrors *string
	NormalizedAt     time.Time
}
