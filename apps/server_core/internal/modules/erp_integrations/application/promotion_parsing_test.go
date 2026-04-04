package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

type stubPromotionStagingRepo struct {
	staging *domain.StagingRecord
}

func (r *stubPromotionStagingRepo) GetStagingRecord(_ context.Context, _, _ string) (*domain.StagingRecord, error) {
	return r.staging, nil
}

type stubPromotionRunRepo struct {
	run *domain.SyncRun
}

func (r *stubPromotionRunRepo) Create(context.Context, *domain.SyncRun) error {
	return nil
}

func (r *stubPromotionRunRepo) Get(_ context.Context, _, _ string) (*domain.SyncRun, error) {
	return r.run, nil
}

func (r *stubPromotionRunRepo) List(_ context.Context, _, _ string, _, _ int) ([]*domain.SyncRun, error) {
	if r.run == nil {
		return nil, nil
	}
	return []*domain.SyncRun{r.run}, nil
}

type stubPromotionProductLookup struct {
	productID string
	found     bool
}

func (r *stubPromotionProductLookup) FindProductIDBySKU(_ context.Context, _, _ string) (string, bool, error) {
	return r.productID, r.found, nil
}

type stubPromotionPriceWriter struct {
	input ports.PricePromotionInput
}

func (w *stubPromotionPriceWriter) PromotePrice(_ context.Context, _ string, _ *domain.ReconciliationResult, _ *domain.SyncRun, input ports.PricePromotionInput) (string, error) {
	w.input = input
	return "prc_1", nil
}

func TestBuildPricePromotionInputPreservesZeroSalePrice(t *testing.T) {
	t.Parallel()

	staging := &domain.StagingRecord{
		StagingID:    "stg_price_1",
		NormalizedAt: time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC),
		NormalizedJSON: json.RawMessage(`{
			"product_source_id":"SKU-1",
			"source_table_code":"17",
			"sale_price":0,
			"VLRVENDA":99.99,
			"currency_code":"BRL"
		}`),
	}

	input, err := buildPricePromotionInput(staging)
	if err != nil {
		t.Fatalf("buildPricePromotionInput returned error: %v", err)
	}
	if input.PriceAmount != 0 {
		t.Fatalf("expected zero sale_price to be preserved, got %v", input.PriceAmount)
	}
}

func TestBuildInventoryPromotionInputPreservesZeroRawQuantity(t *testing.T) {
	t.Parallel()

	staging := &domain.StagingRecord{
		StagingID:    "stg_inventory_1",
		NormalizedAt: time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC),
		NormalizedJSON: json.RawMessage(`{
			"product_source_id":"SKU-1",
			"source_company_code":"1",
			"source_location_code":"10101",
			"raw_quantity":0,
			"ESTOQUE":12
		}`),
	}

	input, err := buildInventoryPromotionInput(staging)
	if err != nil {
		t.Fatalf("buildInventoryPromotionInput returned error: %v", err)
	}
	if input.OnHandQuantity != 0 {
		t.Fatalf("expected zero raw_quantity to be preserved, got %v", input.OnHandQuantity)
	}
}

func TestPromotePriceUsesRunConnectorTypeAsSourceSystem(t *testing.T) {
	t.Parallel()

	staging := &domain.StagingRecord{
		StagingID:        "stg_price_1",
		TenantID:         "tenant-1",
		RunID:            "run-1",
		EntityType:       domain.EntityTypePrices,
		SourceID:         "src-1",
		NormalizedAt:     time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC),
		ValidationStatus: "valid",
		NormalizedJSON: json.RawMessage(`{
			"product_source_id":"SKU-1",
			"source_table_code":"17",
			"sale_price":84.5,
			"currency_code":"BRL",
			"pricing_status":"active",
			"origin_type":"import",
			"reason_code":"ERP_PRICE_PROMOTED",
			"updated_by":"erp_integrations"
		}`),
	}

	writer := &stubPromotionPriceWriter{}
	promotion := &PricePromotion{
		stagingRepo: &stubPromotionStagingRepo{staging: staging},
		runRepo: &stubPromotionRunRepo{run: &domain.SyncRun{
			RunID:         "run-1",
			TenantID:      "tenant-1",
			ConnectorType: domain.ConnectorType("oracle"),
		}},
		productLookup: &stubPromotionProductLookup{productID: "prd_1", found: true},
		writer:        writer,
	}

	canonicalID, err := promotion.PromotePrice(context.Background(), &domain.ReconciliationResult{
		ReconciliationID: "rec-1",
		TenantID:         "tenant-1",
		RunID:            "run-1",
		StagingID:        "stg_price_1",
		SourceID:         "src-1",
		EntityType:       domain.EntityTypePrices,
		PromotionStatus:  domain.PromotionStatusPromoting,
	})
	if err != nil {
		t.Fatalf("PromotePrice returned error: %v", err)
	}
	if canonicalID != "prc_1" {
		t.Fatalf("expected prc_1 canonical id, got %s", canonicalID)
	}
	if writer.input.SourceSystem != "oracle" {
		t.Fatalf("expected source system oracle, got %s", writer.input.SourceSystem)
	}
}
