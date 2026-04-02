package runs

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// RunClaim holds the details of a claimed pending run.
type RunClaim struct {
	RunID         string
	TenantID      string
	InstanceID    string
	ConnectorType string
	RunMode       string
	EntityScope   []string
}

// Ledger manages run state in erp_sync_runs.
type Ledger struct {
	db *sql.DB
}

// NewLedger constructs a Ledger backed by the given DB connection.
func NewLedger(db *sql.DB) *Ledger {
	return &Ledger{db: db}
}

// ClaimPendingRun atomically claims a pending run using SELECT ... FOR UPDATE SKIP LOCKED.
// Returns nil, nil if no pending run is available.
func (l *Ledger) ClaimPendingRun(ctx context.Context) (*RunClaim, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	const selectQ = `
SELECT run_id, tenant_id, instance_id, connector_type, run_mode, entity_scope
FROM erp_sync_runs
WHERE status = 'pending'
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED`

	row := tx.QueryRowContext(ctx, selectQ)

	var claim RunClaim
	// entity_scope is TEXT[] in Postgres. The pgx/v5 stdlib adapter does not
	// automatically decode TEXT[] into []string via database/sql Scan. Instead we
	// scan the raw Postgres array literal (e.g. "{val1,val2}") as a plain string
	// and decode it with parsePGTextArray.
	var entityScopeRaw string
	err = row.Scan(
		&claim.RunID,
		&claim.TenantID,
		&claim.InstanceID,
		&claim.ConnectorType,
		&claim.RunMode,
		&entityScopeRaw,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	claim.EntityScope = parsePGTextArray(entityScopeRaw)

	const updateQ = `
UPDATE erp_sync_runs
SET status = 'running', started_at = NOW()
WHERE run_id = $1`

	if _, err := tx.ExecContext(ctx, updateQ, claim.RunID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &claim, nil
}

// MarkCompleted finalises a run as fully successful.
func (l *Ledger) MarkCompleted(ctx context.Context, runID string, promoted, warnings, rejected, reviews int) error {
	const q = `
UPDATE erp_sync_runs
SET status = 'completed',
    completed_at = NOW(),
    promoted_count = $2,
    warning_count  = $3,
    rejected_count = $4,
    review_count   = $5
WHERE run_id = $1`
	_, err := l.db.ExecContext(ctx, q, runID, promoted, warnings, rejected, reviews)
	return err
}

// MarkPartial finalises a run where some entities succeeded and some failed.
func (l *Ledger) MarkPartial(ctx context.Context, runID, failureSummary string, promoted, warnings, rejected, reviews int) error {
	const q = `
UPDATE erp_sync_runs
SET status = 'partial',
    completed_at    = NOW(),
    failure_summary = $2,
    promoted_count  = $3,
    warning_count   = $4,
    rejected_count  = $5,
    review_count    = $6
WHERE run_id = $1`
	_, err := l.db.ExecContext(ctx, q, runID, failureSummary, promoted, warnings, rejected, reviews)
	return err
}

// MarkFailed finalises a run as fully failed.
func (l *Ledger) MarkFailed(ctx context.Context, runID, failureSummary string) error {
	const q = `
UPDATE erp_sync_runs
SET status = 'failed',
    completed_at    = NOW(),
    failure_summary = $2
WHERE run_id = $1`
	_, err := l.db.ExecContext(ctx, q, runID, failureSummary)
	return err
}

// SaveCursor persists the cursor state for an in-progress or completed run.
func (l *Ledger) SaveCursor(ctx context.Context, runID string, cursorJSON string) error {
	const q = `
UPDATE erp_sync_runs
SET cursor_state = $2::jsonb
WHERE run_id = $1`
	_, err := l.db.ExecContext(ctx, q, runID, cursorJSON)
	return err
}

// parsePGTextArray decodes a Postgres TEXT[] literal of the form {val1,val2,...}
// into a Go string slice. Quoted elements (e.g. {"a b","c,d"}) are handled by
// stripping surrounding double-quotes from each element. This covers all entity
// scope values, which are simple lowercase identifiers without embedded commas.
func parsePGTextArray(raw string) []string {
	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		result = append(result, p)
	}
	return result
}
