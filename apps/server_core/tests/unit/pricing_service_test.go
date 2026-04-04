package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/pricing/application"
	"metalshopping/server_core/internal/modules/pricing/domain"
)

type fakePricingRepository struct {
	created domain.ProductPrice
	traceID string
	list    []domain.ProductPrice
	current domain.ProductPrice
	applied bool
	err     error
}

func (f *fakePricingRepository) CreateProductPrice(_ context.Context, price domain.ProductPrice, traceID string) (domain.ProductPrice, bool, error) {
	if f.err != nil {
		return domain.ProductPrice{}, false, f.err
	}
	f.created = price
	f.traceID = traceID
	if f.applied {
		return price, true, nil
	}
	return f.current, false, nil
}

func (f *fakePricingRepository) ListProductPrices(context.Context, string, string) ([]domain.ProductPrice, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func (f *fakePricingRepository) GetCurrentProductPrice(context.Context, string, string) (domain.ProductPrice, error) {
	if f.err != nil {
		return domain.ProductPrice{}, f.err
	}
	return f.current, nil
}

type fakeManualOverrideGuard struct {
	err error
}

func (f *fakeManualOverrideGuard) ValidateManualOverride(context.Context, string, domain.OriginType) error {
	return f.err
}

func TestPricingServiceSetsProductPrice(t *testing.T) {
	repo := &fakePricingRepository{applied: true}
	avg := 85.0
	service := application.NewService(repo, &fakeManualOverrideGuard{})

	effectiveFrom := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	price, applied, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:              "tenant-1",
		TraceID:               "trace-pricing-set",
		ProductID:             "prd_1",
		PriceContextCode:      "promo",
		CurrencyCode:          "brl",
		PriceAmount:           120,
		ReplacementCostAmount: 90,
		AverageCostAmount:     &avg,
		PricingStatus:         "active",
		EffectiveFrom:         effectiveFrom,
		OriginType:            "manual",
		ReasonCode:            "initial_price",
		UpdatedBy:             "admin-local",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !applied {
		t.Fatal("expected change to be applied")
	}
	if price.PriceID == "" {
		t.Fatal("expected generated price id")
	}
	if repo.created.ReplacementCostAmount != 90 {
		t.Fatalf("expected replacement cost 90, got %v", repo.created.ReplacementCostAmount)
	}
	if repo.created.AverageCostAmount == nil || *repo.created.AverageCostAmount != 85 {
		t.Fatalf("expected average cost 85, got %v", repo.created.AverageCostAmount)
	}
	if repo.created.CurrencyCode != "BRL" {
		t.Fatalf("expected normalized currency BRL, got %q", repo.created.CurrencyCode)
	}
	if repo.created.PriceContextCode != "promo" {
		t.Fatalf("expected preserved price context promo, got %q", repo.created.PriceContextCode)
	}
	if repo.traceID != "trace-pricing-set" {
		t.Fatalf("expected trace id propagation, got %q", repo.traceID)
	}
}

func TestPricingServiceNoOpsWhenCommercialStateDidNotChange(t *testing.T) {
	avg := 84.5
	repo := &fakePricingRepository{
		applied: false,
		current: domain.ProductPrice{
			PriceID:               "prc_current",
			TenantID:              "tenant-1",
			ProductID:             "prd_1",
			CurrencyCode:          "BRL",
			PriceAmount:           120,
			ReplacementCostAmount: 90,
			AverageCostAmount:     &avg,
			PricingStatus:         domain.PricingStatusActive,
			EffectiveFrom:         time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
			OriginType:            domain.OriginTypeImport,
			ReasonCode:            "existing",
			UpdatedBy:             "integration-worker",
		},
	}
	service := application.NewService(repo, &fakeManualOverrideGuard{})

	price, applied, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:              "tenant-1",
		ProductID:             "prd_1",
		CurrencyCode:          "BRL",
		PriceAmount:           120,
		ReplacementCostAmount: 90,
		AverageCostAmount:     &avg,
		PricingStatus:         "active",
		EffectiveFrom:         time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
		OriginType:            "import",
		ReasonCode:            "rerun_same_price",
		UpdatedBy:             "integration-worker",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if applied {
		t.Fatal("expected no-op when commercial state did not change")
	}
	if price.PriceID != "prc_current" {
		t.Fatalf("expected current price to be returned, got %q", price.PriceID)
	}
}

func TestPricingServiceDefaultsPriceContextCodeWhenOmitted(t *testing.T) {
	repo := &fakePricingRepository{applied: true}
	service := application.NewService(repo, &fakeManualOverrideGuard{})

	_, _, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:              "tenant-1",
		ProductID:             "prd_1",
		CurrencyCode:          "BRL",
		PriceAmount:           120,
		ReplacementCostAmount: 80,
		PricingStatus:         "active",
		EffectiveFrom:         time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
		OriginType:            "manual",
		ReasonCode:            "initial",
		UpdatedBy:             "admin-local",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.created.PriceContextCode != domain.DefaultPriceContextCode {
		t.Fatalf("expected default price context, got %q", repo.created.PriceContextCode)
	}
}

func TestPricingServiceRejectsManualOverrideWhenDisabled(t *testing.T) {
	repo := &fakePricingRepository{}
	service := application.NewService(repo, &fakeManualOverrideGuard{err: domain.ErrManualPriceOverrideDisabled})

	_, _, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:              "tenant-1",
		ProductID:             "prd_1",
		CurrencyCode:          "BRL",
		PriceAmount:           120,
		ReplacementCostAmount: 80,
		PricingStatus:         "active",
		EffectiveFrom:         time.Now().UTC(),
		OriginType:            "manual",
		ReasonCode:            "blocked",
		UpdatedBy:             "admin-local",
	})
	if !errors.Is(err, domain.ErrManualPriceOverrideDisabled) {
		t.Fatalf("expected ErrManualPriceOverrideDisabled, got %v", err)
	}
}
