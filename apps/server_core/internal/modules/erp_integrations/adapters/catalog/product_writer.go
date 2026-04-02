package catalog

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	erpevents "metalshopping/server_core/internal/modules/erp_integrations/events"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type ProductWriter struct {
	db          *sql.DB
	outboxStore *outbox.Store
}

var _ ports.ProductWriter = (*ProductWriter)(nil)

func NewProductWriter(db *sql.DB, outboxStore *outbox.Store) *ProductWriter {
	return &ProductWriter{db: db, outboxStore: outboxStore}
}

func (w *ProductWriter) PromoteProduct(ctx context.Context, traceID string, result *domain.ReconciliationResult, input ports.ProductPromotionInput) (string, error) {
	if w == nil || w.db == nil {
		return "", fmt.Errorf("product promotion writer is not configured")
	}
	if w.outboxStore == nil {
		return "", fmt.Errorf("product promotion outbox store is required")
	}
	if result == nil {
		return "", fmt.Errorf("reconciliation result is required")
	}
	tenantID := strings.TrimSpace(result.TenantID)
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	sku := strings.TrimSpace(input.SKU)
	name := strings.TrimSpace(input.Name)
	if sku == "" {
		return "", fmt.Errorf("sku is required")
	}
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return "", fmt.Errorf("invalid product status: %s", status)
	}

	tx, err := pgdb.BeginTenantTx(ctx, w.db, tenantID, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	if !result.ReconciledAt.IsZero() {
		now = result.ReconciledAt.UTC()
	}

	productID := generateProductID()
	const upsertProductSQL = `
INSERT INTO catalog_products (
  product_id,
  tenant_id,
  sku,
  name,
  description,
  brand_name,
  stock_profile_code,
  primary_taxonomy_node_id,
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
  $9,
  $10
)
ON CONFLICT (tenant_id, sku) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  brand_name = EXCLUDED.brand_name,
  stock_profile_code = EXCLUDED.stock_profile_code,
  primary_taxonomy_node_id = EXCLUDED.primary_taxonomy_node_id,
  status = EXCLUDED.status,
  updated_at = EXCLUDED.updated_at
RETURNING product_id
`
	if err := tx.QueryRowContext(
		ctx,
		upsertProductSQL,
		productID,
		sku,
		name,
		nullableText(input.Description),
		nullableText(input.BrandName),
		nullableText(input.StockProfileCode),
		nullableText(input.PrimaryTaxonomyNodeID),
		status,
		now,
		now,
	).Scan(&productID); err != nil {
		return "", fmt.Errorf("upsert catalog product for promotion: %w", err)
	}

	const upsertIdentifierSQL = `
INSERT INTO catalog_product_identifiers (
  product_identifier_id,
  product_id,
  tenant_id,
  identifier_type,
  identifier_value,
  source_system,
  is_primary,
  created_at,
  updated_at
)
VALUES (
  $1,
  $2,
  current_tenant_id(),
  $3,
  $4,
  NULLIF($5, ''),
  $6,
  $7,
  $8
)
ON CONFLICT (tenant_id, identifier_type, identifier_value) DO NOTHING
`
	for _, identifier := range input.Identifiers {
		identifierType := strings.TrimSpace(identifier.IdentifierType)
		identifierValue := strings.TrimSpace(identifier.IdentifierValue)
		if identifierType == "" || identifierValue == "" {
			continue
		}
		if _, err := tx.ExecContext(
			ctx,
			upsertIdentifierSQL,
			generateProductIdentifierID(),
			productID,
			identifierType,
			identifierValue,
			strings.TrimSpace(identifier.SourceSystem),
			identifier.IsPrimary,
			now,
			now,
		); err != nil {
			return "", fmt.Errorf("upsert catalog product identifier for promotion: %w", err)
		}
	}

	promoted := *result
	promoted.CanonicalID = &productID
	record, err := erpevents.NewEntityPromotedOutboxRecord(&promoted, traceID, now)
	if err != nil {
		return "", fmt.Errorf("build erp entity promoted outbox record: %w", err)
	}
	if err := w.outboxStore.AppendInTx(ctx, tx, []outbox.Record{record}); err != nil {
		return "", fmt.Errorf("append erp entity promoted outbox record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit catalog product promotion: %w", err)
	}

	return productID, nil
}

func generateProductID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "prd_fallback"
	}
	return "prd_" + hex.EncodeToString(buf)
}

func generateProductIdentifierID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "pid_fallback"
	}
	return "pid_" + hex.EncodeToString(buf)
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
