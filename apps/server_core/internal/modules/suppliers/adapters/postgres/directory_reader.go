package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/suppliers/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
)

type Repository struct {
	db *sql.DB
}

var (
	ErrDriverManifestNotFound = errors.New("supplier driver manifest not found")
	ErrDriverManifestNotValid = errors.New("supplier driver manifest must be valid before activation")
)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]ports.DirectorySupplier, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT supplier_code, supplier_label, execution_kind, lookup_policy, enabled, updated_at
FROM suppliers_directory
WHERE tenant_id = current_tenant_id()
  AND ($1 = FALSE OR enabled = TRUE)
ORDER BY supplier_code ASC
`

	rows, err := tx.QueryContext(ctx, query, onlyEnabled)
	if err != nil {
		return nil, fmt.Errorf("list suppliers directory: %w", err)
	}
	defer rows.Close()

	items := []ports.DirectorySupplier{}
	for rows.Next() {
		var item ports.DirectorySupplier
		if err := rows.Scan(
			&item.SupplierCode,
			&item.SupplierLabel,
			&item.ExecutionKind,
			&item.LookupPolicy,
			&item.Enabled,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan suppliers directory row: %w", err)
		}
		item.UpdatedAt = item.UpdatedAt.UTC()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suppliers directory rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit suppliers directory read: %w", err)
	}
	return items, nil
}

func (r *Repository) UpsertDirectorySupplier(ctx context.Context, tenantID string, input ports.UpsertDirectorySupplierInput) (ports.DirectorySupplier, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return ports.DirectorySupplier{}, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	const query = `
INSERT INTO suppliers_directory (
  supplier_id,
  tenant_id,
  supplier_code,
  supplier_label,
  execution_kind,
  lookup_policy,
  enabled,
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
  $7
)
ON CONFLICT (tenant_id, supplier_code) DO UPDATE
SET supplier_label = EXCLUDED.supplier_label,
    execution_kind = EXCLUDED.execution_kind,
    lookup_policy = EXCLUDED.lookup_policy,
    enabled = EXCLUDED.enabled,
    updated_at = EXCLUDED.updated_at
RETURNING supplier_code, supplier_label, execution_kind, lookup_policy, enabled, updated_at
`
	var item ports.DirectorySupplier
	if err := tx.QueryRowContext(
		ctx,
		query,
		newSupplierID(),
		input.SupplierCode,
		input.SupplierLabel,
		input.ExecutionKind,
		input.LookupPolicy,
		input.Enabled,
		now,
	).Scan(
		&item.SupplierCode,
		&item.SupplierLabel,
		&item.ExecutionKind,
		&item.LookupPolicy,
		&item.Enabled,
		&item.UpdatedAt,
	); err != nil {
		return ports.DirectorySupplier{}, fmt.Errorf("upsert suppliers directory item: %w", err)
	}
	item.UpdatedAt = item.UpdatedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.DirectorySupplier{}, fmt.Errorf("commit suppliers directory upsert: %w", err)
	}
	return item, nil
}

func (r *Repository) SetDirectorySupplierEnabled(ctx context.Context, tenantID, supplierCode string, enabled bool) (ports.DirectorySupplier, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return ports.DirectorySupplier{}, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	const query = `
UPDATE suppliers_directory
SET enabled = $2,
    updated_at = $3
WHERE tenant_id = current_tenant_id()
  AND supplier_code = $1
RETURNING supplier_code, supplier_label, execution_kind, lookup_policy, enabled, updated_at
`
	var item ports.DirectorySupplier
	if err := tx.QueryRowContext(ctx, query, supplierCode, enabled, now).Scan(
		&item.SupplierCode,
		&item.SupplierLabel,
		&item.ExecutionKind,
		&item.LookupPolicy,
		&item.Enabled,
		&item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ports.DirectorySupplier{}, fmt.Errorf("supplier not found: %s", supplierCode)
		}
		return ports.DirectorySupplier{}, fmt.Errorf("set suppliers directory enabled: %w", err)
	}
	item.UpdatedAt = item.UpdatedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.DirectorySupplier{}, fmt.Errorf("commit suppliers directory enablement: %w", err)
	}
	return item, nil
}

func (r *Repository) ListDriverManifests(ctx context.Context, tenantID, supplierCode string, limit, offset int64) (ports.DriverManifestList, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.DriverManifestList{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const countQuery = `
SELECT COUNT(*)
FROM supplier_driver_manifests
WHERE tenant_id = current_tenant_id()
  AND ($1 = '' OR supplier_code = $1)
`
	var total int64
	if err := tx.QueryRowContext(ctx, countQuery, supplierCode).Scan(&total); err != nil {
		return ports.DriverManifestList{}, fmt.Errorf("count supplier driver manifests: %w", err)
	}

	const listQuery = `
SELECT
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
FROM supplier_driver_manifests
WHERE tenant_id = current_tenant_id()
  AND ($1 = '' OR supplier_code = $1)
ORDER BY updated_at DESC, supplier_code ASC, version_number DESC
LIMIT $2 OFFSET $3
`
	rows, err := tx.QueryContext(ctx, listQuery, supplierCode, limit, offset)
	if err != nil {
		return ports.DriverManifestList{}, fmt.Errorf("list supplier driver manifests: %w", err)
	}
	defer rows.Close()

	items := make([]ports.DriverManifest, 0, limit)
	for rows.Next() {
		item, scanErr := scanDriverManifest(rows)
		if scanErr != nil {
			return ports.DriverManifestList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.DriverManifestList{}, fmt.Errorf("iterate supplier driver manifests rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.DriverManifestList{}, fmt.Errorf("commit supplier driver manifests list: %w", err)
	}
	return ports.DriverManifestList{
		Rows:   items,
		Offset: offset,
		Limit:  limit,
		Total:  total,
	}, nil
}

func (r *Repository) GetDriverManifest(ctx context.Context, tenantID, manifestID string) (ports.DriverManifest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ports.DriverManifest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
FROM supplier_driver_manifests
WHERE tenant_id = current_tenant_id()
  AND manifest_id = $1
LIMIT 1
`
	row := tx.QueryRowContext(ctx, query, manifestID)
	item, err := scanDriverManifest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.DriverManifest{}, ErrDriverManifestNotFound
		}
		return ports.DriverManifest{}, fmt.Errorf("get supplier driver manifest: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("commit supplier driver manifest read: %w", err)
	}
	return item, nil
}

func (r *Repository) CreateDriverManifest(ctx context.Context, tenantID string, input ports.CreateDriverManifestInput) (ports.DriverManifest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return ports.DriverManifest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const nextVersionQuery = `
SELECT COALESCE(MAX(version_number), 0) + 1
FROM supplier_driver_manifests
WHERE tenant_id = current_tenant_id()
  AND supplier_code = $1
`
	var nextVersion int64
	if err := tx.QueryRowContext(ctx, nextVersionQuery, input.SupplierCode).Scan(&nextVersion); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("resolve next supplier manifest version: %w", err)
	}

	now := time.Now().UTC()
	const createQuery = `
INSERT INTO supplier_driver_manifests (
  manifest_id,
  tenant_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
)
VALUES (
  $1,
  current_tenant_id(),
  $2,
  $3,
  $4,
  $5::jsonb,
  'pending',
  '[]'::jsonb,
  FALSE,
  $6,
  $7,
  $7
)
RETURNING
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
`
	var item ports.DriverManifest
	if err := tx.QueryRowContext(
		ctx,
		createQuery,
		newManifestID(input.SupplierCode, nextVersion),
		input.SupplierCode,
		nextVersion,
		input.Family,
		string(input.ConfigJSON),
		input.CreatedBy,
		now,
	).Scan(
		&item.ManifestID,
		&item.SupplierCode,
		&item.VersionNumber,
		&item.Family,
		&item.ConfigJSON,
		&item.ValidationStatus,
		&item.ValidationErrors,
		&item.IsActive,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("create supplier driver manifest: %w", err)
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	if err := tx.Commit(); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("commit supplier driver manifest create: %w", err)
	}
	return item, nil
}

func (r *Repository) ValidateDriverManifest(ctx context.Context, tenantID, manifestID string, validationErrors []ports.ValidationError) (ports.DriverManifest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return ports.DriverManifest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	status := "valid"
	if len(validationErrors) > 0 {
		status = "invalid"
	}
	rawErrors, err := json.Marshal(validationErrors)
	if err != nil {
		return ports.DriverManifest{}, fmt.Errorf("marshal supplier manifest validation errors: %w", err)
	}
	now := time.Now().UTC()

	const query = `
UPDATE supplier_driver_manifests
SET validation_status = $2,
    validation_errors_json = $3::jsonb,
    updated_at = $4
WHERE tenant_id = current_tenant_id()
  AND manifest_id = $1
RETURNING
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
`
	row := tx.QueryRowContext(ctx, query, manifestID, status, string(rawErrors), now)
	item, err := scanDriverManifest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.DriverManifest{}, ErrDriverManifestNotFound
		}
		return ports.DriverManifest{}, fmt.Errorf("validate supplier driver manifest: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("commit supplier driver manifest validation: %w", err)
	}
	return item, nil
}

func (r *Repository) ActivateDriverManifest(ctx context.Context, tenantID, manifestID string) (ports.DriverManifest, error) {
	tx, err := pgdb.BeginTenantTx(ctx, r.db, tenantID, nil)
	if err != nil {
		return ports.DriverManifest{}, err
	}
	defer func() { _ = tx.Rollback() }()

	const loadQuery = `
SELECT
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
FROM supplier_driver_manifests
WHERE tenant_id = current_tenant_id()
  AND manifest_id = $1
LIMIT 1
`
	currentRow := tx.QueryRowContext(ctx, loadQuery, manifestID)
	current, err := scanDriverManifest(currentRow)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.DriverManifest{}, ErrDriverManifestNotFound
		}
		return ports.DriverManifest{}, fmt.Errorf("load supplier driver manifest for activation: %w", err)
	}
	if strings.TrimSpace(current.ValidationStatus) != "valid" {
		return ports.DriverManifest{}, ErrDriverManifestNotValid
	}

	now := time.Now().UTC()
	const deactivateQuery = `
UPDATE supplier_driver_manifests
SET is_active = FALSE,
    updated_at = $2
WHERE tenant_id = current_tenant_id()
  AND supplier_code = $1
  AND manifest_id <> $3
  AND is_active = TRUE
`
	if _, err := tx.ExecContext(ctx, deactivateQuery, current.SupplierCode, now, current.ManifestID); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("deactivate previous supplier manifests: %w", err)
	}

	const activateQuery = `
UPDATE supplier_driver_manifests
SET is_active = TRUE,
    updated_at = $2
WHERE tenant_id = current_tenant_id()
  AND manifest_id = $1
RETURNING
  manifest_id,
  supplier_code,
  version_number,
  family,
  config_json,
  validation_status,
  validation_errors_json,
  is_active,
  created_by,
  created_at,
  updated_at
`
	activatedRow := tx.QueryRowContext(ctx, activateQuery, manifestID, now)
	activated, err := scanDriverManifest(activatedRow)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ports.DriverManifest{}, ErrDriverManifestNotFound
		}
		return ports.DriverManifest{}, fmt.Errorf("activate supplier manifest: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("commit supplier manifest activation: %w", err)
	}
	return activated, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDriverManifest(s scanner) (ports.DriverManifest, error) {
	var item ports.DriverManifest
	if err := s.Scan(
		&item.ManifestID,
		&item.SupplierCode,
		&item.VersionNumber,
		&item.Family,
		&item.ConfigJSON,
		&item.ValidationStatus,
		&item.ValidationErrors,
		&item.IsActive,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return ports.DriverManifest{}, fmt.Errorf("scan supplier driver manifest: %w", err)
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

func newSupplierID() string {
	return "sup_" + randomHex(8)
}

func newManifestID(supplierCode string, version int64) string {
	prefix := strings.ToLower(strings.TrimSpace(supplierCode))
	if prefix == "" {
		prefix = "unknown"
	}
	return fmt.Sprintf("manifest_%s_v%d_%s", prefix, version, randomHex(4))
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}
