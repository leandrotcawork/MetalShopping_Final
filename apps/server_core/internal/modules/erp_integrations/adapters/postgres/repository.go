package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

// ---------------------------------------------------------------------------
// Shared base
// ---------------------------------------------------------------------------

// base holds the shared dependencies for all per-interface repository types.
type base struct {
	db          *sql.DB
	outboxStore *outbox.Store
}

// ---------------------------------------------------------------------------
// Concrete repository types
// ---------------------------------------------------------------------------

// InstanceRepo implements ports.InstanceRepository.
type InstanceRepo struct{ base }

// RunRepo implements ports.RunRepository.
type RunRepo struct{ base }

// ReviewRepo implements ports.ReviewRepository.
type ReviewRepo struct{ base }

// ReconciliationRepo implements ports.ReconciliationReader.
type ReconciliationRepo struct{ base }

// Compile-time interface assertions.
var _ ports.InstanceRepository = (*InstanceRepo)(nil)
var _ ports.RunRepository = (*RunRepo)(nil)
var _ ports.ReviewRepository = (*ReviewRepo)(nil)
var _ ports.ReconciliationReader = (*ReconciliationRepo)(nil)

// ---------------------------------------------------------------------------
// Repos container and constructor
// ---------------------------------------------------------------------------

// Repos bundles all four repository types returned by NewRepos.
type Repos struct {
	Instances       *InstanceRepo
	Runs            *RunRepo
	Reviews         *ReviewRepo
	Reconciliations *ReconciliationRepo
}

// NewRepos constructs all four repository types sharing a single db connection
// and outboxStore. outboxStore may be a no-op implementation if outbox
// publishing is not required for this deployment.
func NewRepos(db *sql.DB, outboxStore *outbox.Store) *Repos {
	b := base{db: db, outboxStore: outboxStore}
	return &Repos{
		Instances:       &InstanceRepo{b},
		Runs:            &RunRepo{b},
		Reviews:         &ReviewRepo{b},
		Reconciliations: &ReconciliationRepo{b},
	}
}

// ---------------------------------------------------------------------------
// InstanceRepo — implements ports.InstanceRepository
// ---------------------------------------------------------------------------

func (r *InstanceRepo) Create(ctx context.Context, instance *domain.IntegrationInstance) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, instance.TenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	enabledEntities := make([]string, len(instance.EnabledEntities))
	for i, e := range instance.EnabledEntities {
		enabledEntities[i] = string(e)
	}

	const insertSQL = `
INSERT INTO erp_integration_instances (
  instance_id,
  tenant_id,
  connector_type,
  display_name,
  connection_ref,
  enabled_entities,
  sync_schedule,
  status,
  created_at,
  updated_at
)
VALUES (
  $1,
  current_tenant_id(),
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		instance.InstanceID,
		string(instance.ConnectorType),
		instance.DisplayName,
		instance.ConnectionRef,
		pgtype.FlatArray[string](enabledEntities),
		nullableText(instance.SyncSchedule),
		string(instance.Status),
		instance.CreatedAt,
		instance.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert erp integration instance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp integration instance create: %w", err)
	}
	return nil
}

func (r *InstanceRepo) Get(ctx context.Context, tenantID, instanceID string) (*domain.IntegrationInstance, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT instance_id, tenant_id, connector_type, display_name, connection_ref,
       enabled_entities, sync_schedule, status, created_at, updated_at
FROM erp_integration_instances
WHERE tenant_id = current_tenant_id()
  AND instance_id = $1
`
	row := tx.QueryRowContext(ctx, querySQL, instanceID)
	item, err := scanInstance(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInstanceNotFound
		}
		return nil, fmt.Errorf("get erp integration instance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp integration instance get: %w", err)
	}
	return item, nil
}

func (r *InstanceRepo) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.IntegrationInstance, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT instance_id, tenant_id, connector_type, display_name, connection_ref,
       enabled_entities, sync_schedule, status, created_at, updated_at
FROM erp_integration_instances
WHERE tenant_id = current_tenant_id()
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := tx.QueryContext(ctx, querySQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list erp integration instances: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.IntegrationInstance, 0, limit)
	for rows.Next() {
		item, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate erp integration instances: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp integration instance list: %w", err)
	}
	return items, nil
}

func (r *InstanceRepo) HasActiveInstance(ctx context.Context, tenantID string) (bool, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT COUNT(*)
FROM erp_integration_instances
WHERE tenant_id = current_tenant_id()
  AND status = 'active'
`
	var count int
	if err := tx.QueryRowContext(ctx, querySQL).Scan(&count); err != nil {
		return false, fmt.Errorf("count active erp integration instances: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit erp integration instance has-active: %w", err)
	}
	return count > 0, nil
}

// ---------------------------------------------------------------------------
// RunRepo — implements ports.RunRepository
// ---------------------------------------------------------------------------

func (r *RunRepo) Create(ctx context.Context, run *domain.SyncRun) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, run.TenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	entityScope := make([]string, len(run.EntityScope))
	for i, e := range run.EntityScope {
		entityScope[i] = string(e)
	}

	const insertSQL = `
INSERT INTO erp_sync_runs (
  run_id,
  tenant_id,
  instance_id,
  connector_type,
  run_mode,
  entity_scope,
  status,
  started_at,
  completed_at,
  promoted_count,
  warning_count,
  rejected_count,
  review_count,
  failure_summary,
  cursor_state,
  created_at
)
VALUES (
  $1,
  current_tenant_id(),
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		run.RunID,
		run.InstanceID,
		string(run.ConnectorType),
		string(run.RunMode),
		pgtype.FlatArray[string](entityScope),
		string(run.Status),
		nullableTimePtr(run.StartedAt),
		nullableTimePtr(run.CompletedAt),
		run.PromotedCount,
		run.WarningCount,
		run.RejectedCount,
		run.ReviewCount,
		run.FailureSummary,
		run.CursorState,
		run.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert erp sync run: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp sync run create: %w", err)
	}
	return nil
}

func (r *RunRepo) Get(ctx context.Context, tenantID, runID string) (*domain.SyncRun, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT run_id, tenant_id, instance_id, connector_type, run_mode, entity_scope,
       status, started_at, completed_at, promoted_count, warning_count, rejected_count,
       review_count, failure_summary, cursor_state, created_at
FROM erp_sync_runs
WHERE tenant_id = current_tenant_id()
  AND run_id = $1
`
	row := tx.QueryRowContext(ctx, querySQL, runID)
	item, err := scanRun(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRunNotFound
		}
		return nil, fmt.Errorf("get erp sync run: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp sync run get: %w", err)
	}
	return item, nil
}

func (r *RunRepo) List(ctx context.Context, tenantID, instanceID string, limit, offset int) ([]*domain.SyncRun, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT run_id, tenant_id, instance_id, connector_type, run_mode, entity_scope,
       status, started_at, completed_at, promoted_count, warning_count, rejected_count,
       review_count, failure_summary, cursor_state, created_at
FROM erp_sync_runs
WHERE tenant_id = current_tenant_id()
  AND instance_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`
	rows, err := tx.QueryContext(ctx, querySQL, instanceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list erp sync runs: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.SyncRun, 0, limit)
	for rows.Next() {
		item, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate erp sync runs: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp sync run list: %w", err)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// ReviewRepo — implements ports.ReviewRepository
// ---------------------------------------------------------------------------

func (r *ReviewRepo) Create(ctx context.Context, item *domain.ReviewItem) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, item.TenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const insertSQL = `
INSERT INTO erp_review_items (
  review_id,
  tenant_id,
  instance_id,
  connector_type,
  entity_type,
  source_id,
  run_id,
  severity,
  reason_code,
  problem_summary,
  raw_id,
  staging_id,
  reconciliation_id,
  recommended_action,
  item_status,
  resolved_at,
  resolved_by,
  created_at
)
VALUES (
  $1,
  current_tenant_id(),
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17
)
`
	if _, err := tx.ExecContext(
		ctx,
		insertSQL,
		item.ReviewID,
		item.InstanceID,
		string(item.ConnectorType),
		string(item.EntityType),
		item.SourceID,
		item.RunID,
		string(item.Severity),
		item.ReasonCode,
		item.ProblemSummary,
		item.RawPayloadRef,
		item.StagingID,
		item.ReconciliationID,
		item.RecommendedAction,
		string(item.ItemStatus),
		item.ResolvedAt,
		item.ResolvedBy,
		item.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert erp review item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp review item create: %w", err)
	}
	return nil
}

func (r *ReviewRepo) Get(ctx context.Context, tenantID, reviewID string) (*domain.ReviewItem, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT review_id, tenant_id, instance_id, connector_type, entity_type, source_id,
       run_id, severity, reason_code, problem_summary, raw_id, staging_id,
       reconciliation_id, recommended_action, item_status, resolved_at, resolved_by, created_at
FROM erp_review_items
WHERE tenant_id = current_tenant_id()
  AND review_id = $1
`
	row := tx.QueryRowContext(ctx, querySQL, reviewID)
	item, err := scanReviewItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReviewItemNotFound
		}
		return nil, fmt.Errorf("get erp review item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp review item get: %w", err)
	}
	return item, nil
}

func (r *ReviewRepo) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ReviewItem, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT review_id, tenant_id, instance_id, connector_type, entity_type, source_id,
       run_id, severity, reason_code, problem_summary, raw_id, staging_id,
       reconciliation_id, recommended_action, item_status, resolved_at, resolved_by, created_at
FROM erp_review_items
WHERE tenant_id = current_tenant_id()
  AND item_status = 'open'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := tx.QueryContext(ctx, querySQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list erp review items: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.ReviewItem, 0, limit)
	for rows.Next() {
		item, err := scanReviewItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate erp review items: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp review item list: %w", err)
	}
	return items, nil
}

func (r *ReviewRepo) Resolve(ctx context.Context, tenantID, reviewID string, status domain.ReviewItemStatus, resolvedBy string, resolvedAt time.Time) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const updateSQL = `
UPDATE erp_review_items
SET item_status  = $1,
    resolved_by  = $2,
    resolved_at  = $3
WHERE tenant_id = current_tenant_id()
  AND review_id = $4
`
	result, err := tx.ExecContext(ctx, updateSQL, string(status), resolvedBy, resolvedAt.UTC(), reviewID)
	if err != nil {
		return fmt.Errorf("resolve erp review item: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("resolve erp review item rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrReviewItemNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp review item resolve: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ReconciliationRepo — implements ports.ReconciliationReader
// ---------------------------------------------------------------------------

func (r *ReconciliationRepo) ListPromotableResults(ctx context.Context, tenantID string, limit int) ([]*domain.ReconciliationResult, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const querySQL = `
SELECT reconciliation_id, tenant_id, run_id, staging_id, entity_type, source_id,
       canonical_id, action, classification, reason_code, warning_details,
       reconciled_at, promotion_status
FROM erp_reconciliation_results
WHERE tenant_id = current_tenant_id()
  AND classification IN ('promotable', 'promotable_with_warning')
  AND promotion_status = 'pending'
LIMIT $1
`
	rows, err := tx.QueryContext(ctx, querySQL, limit)
	if err != nil {
		return nil, fmt.Errorf("list promotable erp reconciliation results: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.ReconciliationResult, 0, limit)
	for rows.Next() {
		item, err := scanReconciliationResult(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promotable erp reconciliation results: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit erp reconciliation results list: %w", err)
	}
	return items, nil
}

func (r *ReconciliationRepo) ClaimForPromotion(ctx context.Context, tenantID, reconciliationID string) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const updateSQL = `
UPDATE erp_reconciliation_results
SET promotion_status = 'promoting'
WHERE tenant_id = current_tenant_id()
  AND reconciliation_id = $1
  AND promotion_status = 'pending'
`
	result, err := tx.ExecContext(ctx, updateSQL, reconciliationID)
	if err != nil {
		return fmt.Errorf("claim erp reconciliation result for promotion: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("claim erp reconciliation result rows affected: %w", err)
	}
	if n == 0 {
		// Already claimed or not found — treat as a no-op conflict.
		_ = tx.Rollback()
		return nil
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp reconciliation claim: %w", err)
	}
	return nil
}

func (r *ReconciliationRepo) MarkPromoted(ctx context.Context, tenantID, reconciliationID, canonicalID string) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const updateSQL = `
UPDATE erp_reconciliation_results
SET promotion_status = 'promoted',
    canonical_id     = $1
WHERE tenant_id = current_tenant_id()
  AND reconciliation_id = $2
`
	if _, err := tx.ExecContext(ctx, updateSQL, canonicalID, reconciliationID); err != nil {
		return fmt.Errorf("mark erp reconciliation result promoted: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp reconciliation mark promoted: %w", err)
	}
	return nil
}

func (r *ReconciliationRepo) MarkPromotionFailed(ctx context.Context, tenantID, reconciliationID string) error {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const updateSQL = `
UPDATE erp_reconciliation_results
SET promotion_status = 'failed'
WHERE tenant_id = current_tenant_id()
  AND reconciliation_id = $1
`
	if _, err := tx.ExecContext(ctx, updateSQL, reconciliationID); err != nil {
		return fmt.Errorf("mark erp reconciliation result promotion failed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit erp reconciliation mark failed: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Scan helpers
// ---------------------------------------------------------------------------

type scanner interface {
	Scan(dest ...any) error
}

func scanInstance(s scanner) (*domain.IntegrationInstance, error) {
	var item domain.IntegrationInstance
	var connectorType string
	var status string
	var enabledEntities pgtype.FlatArray[string]
	var syncSchedule sql.NullString

	if err := s.Scan(
		&item.InstanceID,
		&item.TenantID,
		&connectorType,
		&item.DisplayName,
		&item.ConnectionRef,
		&enabledEntities,
		&syncSchedule,
		&status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.ConnectorType = domain.ConnectorType(connectorType)
	item.Status = domain.InstanceStatus(status)
	item.EnabledEntities = make([]domain.EntityType, len(enabledEntities))
	for i, e := range enabledEntities {
		item.EnabledEntities[i] = domain.EntityType(e)
	}
	if syncSchedule.Valid {
		s := syncSchedule.String
		item.SyncSchedule = &s
	}
	return &item, nil
}

func scanRun(s scanner) (*domain.SyncRun, error) {
	var item domain.SyncRun
	var connectorType, runMode, status string
	var entityScope pgtype.FlatArray[string]
	var startedAt, completedAt sql.NullTime
	var failureSummary, cursorState sql.NullString

	if err := s.Scan(
		&item.RunID,
		&item.TenantID,
		&item.InstanceID,
		&connectorType,
		&runMode,
		&entityScope,
		&status,
		&startedAt,
		&completedAt,
		&item.PromotedCount,
		&item.WarningCount,
		&item.RejectedCount,
		&item.ReviewCount,
		&failureSummary,
		&cursorState,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.ConnectorType = domain.ConnectorType(connectorType)
	item.RunMode = domain.RunMode(runMode)
	item.Status = domain.RunStatus(status)
	item.EntityScope = make([]domain.EntityType, len(entityScope))
	for i, e := range entityScope {
		item.EntityScope[i] = domain.EntityType(e)
	}
	if startedAt.Valid {
		t := startedAt.Time.UTC()
		item.StartedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time.UTC()
		item.CompletedAt = &t
	}
	if failureSummary.Valid {
		item.FailureSummary = &failureSummary.String
	}
	if cursorState.Valid {
		item.CursorState = &cursorState.String
	}
	return &item, nil
}

func scanReviewItem(s scanner) (*domain.ReviewItem, error) {
	var item domain.ReviewItem
	var connectorType, entityType, severity, itemStatus string
	var resolvedAt sql.NullTime
	var resolvedBy sql.NullString

	if err := s.Scan(
		&item.ReviewID,
		&item.TenantID,
		&item.InstanceID,
		&connectorType,
		&entityType,
		&item.SourceID,
		&item.RunID,
		&severity,
		&item.ReasonCode,
		&item.ProblemSummary,
		&item.RawPayloadRef,
		&item.StagingID,
		&item.ReconciliationID,
		&item.RecommendedAction,
		&itemStatus,
		&resolvedAt,
		&resolvedBy,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.ConnectorType = domain.ConnectorType(connectorType)
	item.EntityType = domain.EntityType(entityType)
	item.Severity = domain.ReviewSeverity(severity)
	item.ItemStatus = domain.ReviewItemStatus(itemStatus)
	if resolvedAt.Valid {
		t := resolvedAt.Time.UTC()
		item.ResolvedAt = &t
	}
	if resolvedBy.Valid {
		item.ResolvedBy = &resolvedBy.String
	}
	return &item, nil
}

func scanReconciliationResult(s scanner) (*domain.ReconciliationResult, error) {
	var item domain.ReconciliationResult
	var entityType, classification, promotionStatus string
	var canonicalID sql.NullString
	var warningDetails sql.NullString

	if err := s.Scan(
		&item.ReconciliationID,
		&item.TenantID,
		&item.RunID,
		&item.StagingID,
		&entityType,
		&item.SourceID,
		&canonicalID,
		&item.Action,
		&classification,
		&item.ReasonCode,
		&warningDetails,
		&item.ReconciledAt,
		&promotionStatus,
	); err != nil {
		return nil, err
	}
	item.EntityType = domain.EntityType(entityType)
	item.Classification = domain.ReconciliationClassification(classification)
	item.PromotionStatus = domain.PromotionStatus(promotionStatus)
	if canonicalID.Valid {
		item.CanonicalID = &canonicalID.String
	}
	if warningDetails.Valid {
		item.WarningDetails = &warningDetails.String
	}
	return &item, nil
}

// ---------------------------------------------------------------------------
// SQL null helpers
// ---------------------------------------------------------------------------

func nullableText(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimePtr(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}
