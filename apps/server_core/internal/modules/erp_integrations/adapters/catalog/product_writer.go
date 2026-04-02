package catalog

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	catalogdomain "metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	erpevents "metalshopping/server_core/internal/modules/erp_integrations/events"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	pgdb "metalshopping/server_core/internal/platform/db/postgres"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type catalogProductTxWriter interface {
	CreateProductInTx(ctx context.Context, tx *sql.Tx, product catalogdomain.Product, traceID string) error
}

type ProductWriter struct {
	db          *sql.DB
	outboxStore *outbox.Store
	catalog     catalogProductTxWriter
}

var _ ports.ProductWriter = (*ProductWriter)(nil)

func NewProductWriter(db *sql.DB, outboxStore *outbox.Store, catalogWriter catalogProductTxWriter) *ProductWriter {
	return &ProductWriter{db: db, outboxStore: outboxStore, catalog: catalogWriter}
}

func (w *ProductWriter) PromoteProduct(ctx context.Context, traceID string, result *domain.ReconciliationResult, run *domain.SyncRun, input ports.ProductPromotionInput) (string, error) {
	if w == nil || w.db == nil {
		return "", fmt.Errorf("product promotion writer is not configured")
	}
	if w.outboxStore == nil {
		return "", fmt.Errorf("product promotion outbox store is required")
	}
	if w.catalog == nil {
		return "", fmt.Errorf("product promotion catalog repository is required")
	}
	if result == nil {
		return "", fmt.Errorf("reconciliation result is required")
	}
	if run == nil {
		return "", fmt.Errorf("sync run is required")
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
	product := catalogdomain.Product{
		ProductID:             productID,
		TenantID:              tenantID,
		SKU:                   sku,
		Name:                  name,
		Description:           strings.TrimSpace(input.Description),
		BrandName:             strings.TrimSpace(input.BrandName),
		StockProfileCode:      strings.TrimSpace(input.StockProfileCode),
		PrimaryTaxonomyNodeID: strings.TrimSpace(input.PrimaryTaxonomyNodeID),
		Status:                catalogdomain.ProductStatus(status),
		Identifiers:           buildCatalogIdentifiers(productID, tenantID, now, input.Identifiers),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := product.ValidateForCreate(); err != nil {
		return "", err
	}
	if err := w.catalog.CreateProductInTx(ctx, tx, product, traceID); err != nil {
		return "", fmt.Errorf("create catalog product for promotion: %w", err)
	}

	promoted := *result
	promoted.CanonicalID = &productID
	record, err := erpevents.NewEntityPromotedOutboxRecord(&promoted, run, traceID, now)
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

func buildCatalogIdentifiers(productID, tenantID string, now time.Time, inputs []ports.ProductPromotionIdentifierInput) []catalogdomain.ProductIdentifier {
	identifiers := make([]catalogdomain.ProductIdentifier, 0, len(inputs))
	for _, input := range inputs {
		identifierType := strings.TrimSpace(input.IdentifierType)
		identifierValue := strings.TrimSpace(input.IdentifierValue)
		if identifierType == "" || identifierValue == "" {
			continue
		}
		identifiers = append(identifiers, catalogdomain.ProductIdentifier{
			ProductIdentifierID: generateProductIdentifierID(),
			ProductID:           productID,
			TenantID:            tenantID,
			IdentifierType:      identifierType,
			IdentifierValue:     identifierValue,
			SourceSystem:        strings.TrimSpace(input.SourceSystem),
			IsPrimary:           input.IsPrimary,
			CreatedAt:           now,
			UpdatedAt:           now,
		})
	}
	return identifiers
}
