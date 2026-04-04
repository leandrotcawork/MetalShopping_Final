package unit

import (
	"context"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/inventory/application"
	"metalshopping/server_core/internal/modules/inventory/domain"
)

type fakeInventoryRepository struct {
	created domain.ProductPosition
	traceID string
	list    []domain.ProductPosition
	current domain.ProductPosition
	applied bool
	err     error
}

func (f *fakeInventoryRepository) CreateProductPosition(_ context.Context, position domain.ProductPosition, traceID string) (domain.ProductPosition, bool, error) {
	if f.err != nil {
		return domain.ProductPosition{}, false, f.err
	}
	f.created = position
	f.traceID = traceID
	if f.applied {
		return position, true, nil
	}
	return f.current, false, nil
}

func (f *fakeInventoryRepository) ListProductPositions(context.Context, string, string) ([]domain.ProductPosition, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func (f *fakeInventoryRepository) GetCurrentProductPosition(context.Context, string, string) (domain.ProductPosition, error) {
	if f.err != nil {
		return domain.ProductPosition{}, f.err
	}
	return f.current, nil
}

func TestInventoryServiceSetsProductPosition(t *testing.T) {
	repo := &fakeInventoryRepository{applied: true}
	service := application.NewService(repo)

	effectiveFrom := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	lastPurchase := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	lastSale := time.Date(2026, 3, 16, 9, 0, 0, 0, time.UTC)
	position, applied, err := service.SetProductPosition(context.Background(), application.SetProductPositionCommand{
		TenantID:       "tenant-1",
		TraceID:        "trace-inventory-set",
		ProductID:      "prd_1",
		OnHandQuantity: 42,
		LastPurchaseAt: &lastPurchase,
		LastSaleAt:     &lastSale,
		PositionStatus: "active",
		EffectiveFrom:  effectiveFrom,
		OriginType:     "import",
		ReasonCode:     "erp_sync",
		UpdatedBy:      "inventory-sync",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !applied {
		t.Fatal("expected inventory change to be applied")
	}
	if position.PositionID == "" {
		t.Fatal("expected generated position id")
	}
	if repo.created.OnHandQuantity != 42 {
		t.Fatalf("expected on_hand_quantity 42, got %v", repo.created.OnHandQuantity)
	}
	if repo.traceID != "trace-inventory-set" {
		t.Fatalf("expected trace id propagation, got %q", repo.traceID)
	}
	if repo.created.SourceCompanyCode != "" || repo.created.SourceLocationCode != "" {
		t.Fatalf("expected empty source dimensions by default, got %q/%q", repo.created.SourceCompanyCode, repo.created.SourceLocationCode)
	}
}

func TestInventoryServiceNoOpsWhenOperationalStateDidNotChange(t *testing.T) {
	lastPurchase := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	lastSale := time.Date(2026, 3, 16, 9, 0, 0, 0, time.UTC)
	repo := &fakeInventoryRepository{
		applied: false,
		current: domain.ProductPosition{
			PositionID:     "pos_current",
			TenantID:       "tenant-1",
			ProductID:      "prd_1",
			OnHandQuantity: 42,
			LastPurchaseAt: &lastPurchase,
			LastSaleAt:     &lastSale,
			PositionStatus: domain.PositionStatusActive,
			EffectiveFrom:  time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
			OriginType:     domain.OriginTypeImport,
			ReasonCode:     "existing",
			UpdatedBy:      "inventory-sync",
		},
	}
	service := application.NewService(repo)

	position, applied, err := service.SetProductPosition(context.Background(), application.SetProductPositionCommand{
		TenantID:       "tenant-1",
		ProductID:      "prd_1",
		OnHandQuantity: 42,
		LastPurchaseAt: &lastPurchase,
		LastSaleAt:     &lastSale,
		PositionStatus: "active",
		EffectiveFrom:  time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
		OriginType:     "import",
		ReasonCode:     "rerun_same_position",
		UpdatedBy:      "inventory-sync",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if applied {
		t.Fatal("expected no-op when inventory state did not change")
	}
	if position.PositionID != "pos_current" {
		t.Fatalf("expected current position to be returned, got %q", position.PositionID)
	}
}

func TestInventoryServiceDefaultsSourceDimensionsWhenOmitted(t *testing.T) {
	repo := &fakeInventoryRepository{applied: true}
	service := application.NewService(repo)

	_, _, err := service.SetProductPosition(context.Background(), application.SetProductPositionCommand{
		TenantID:       "tenant-1",
		ProductID:      "prd_1",
		OnHandQuantity: 42,
		PositionStatus: "active",
		EffectiveFrom:  time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
		OriginType:     "import",
		ReasonCode:     "initial",
		UpdatedBy:      "inventory-sync",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.created.SourceCompanyCode != "" || repo.created.SourceLocationCode != "" {
		t.Fatalf("expected empty source dimensions, got %q/%q", repo.created.SourceCompanyCode, repo.created.SourceLocationCode)
	}
}
