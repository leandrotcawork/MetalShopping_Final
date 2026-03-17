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
	err     error
}

func (f *fakePricingRepository) CreateProductPrice(_ context.Context, price domain.ProductPrice, traceID string) error {
	if f.err != nil {
		return f.err
	}
	f.created = price
	f.traceID = traceID
	return nil
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

type fakeMarginFloorResolver struct {
	value float64
	err   error
}

func (f *fakeMarginFloorResolver) ResolveMarginFloor(context.Context, string, float64) (float64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.value, nil
}

func TestPricingServiceSetsProductPrice(t *testing.T) {
	repo := &fakePricingRepository{}
	service := application.NewService(repo, &fakeManualOverrideGuard{}, &fakeMarginFloorResolver{value: 15})

	effectiveFrom := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)
	price, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:        "tenant-1",
		TraceID:         "trace-pricing-set",
		ProductID:       "prd_1",
		CurrencyCode:    "brl",
		PriceAmount:     120,
		CostBasisAmount: 90,
		PricingStatus:   "active",
		EffectiveFrom:   effectiveFrom,
		OriginType:      "manual",
		ReasonCode:      "initial_price",
		UpdatedBy:       "admin-local",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if price.PriceID == "" {
		t.Fatal("expected generated price id")
	}
	if repo.created.MarginFloorValue != 15 {
		t.Fatalf("expected resolved margin floor 15, got %v", repo.created.MarginFloorValue)
	}
	if repo.created.CurrencyCode != "BRL" {
		t.Fatalf("expected normalized currency BRL, got %q", repo.created.CurrencyCode)
	}
	if repo.traceID != "trace-pricing-set" {
		t.Fatalf("expected trace id propagation, got %q", repo.traceID)
	}
}

func TestPricingServiceRejectsPriceBelowMarginFloor(t *testing.T) {
	repo := &fakePricingRepository{}
	service := application.NewService(repo, &fakeManualOverrideGuard{}, &fakeMarginFloorResolver{value: 30})

	_, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:        "tenant-1",
		ProductID:       "prd_1",
		CurrencyCode:    "BRL",
		PriceAmount:     100,
		CostBasisAmount: 90,
		PricingStatus:   "active",
		EffectiveFrom:   time.Now().UTC(),
		OriginType:      "manual",
		ReasonCode:      "too_low",
		UpdatedBy:       "admin-local",
	})
	if !errors.Is(err, domain.ErrPriceBelowMarginFloor) {
		t.Fatalf("expected ErrPriceBelowMarginFloor, got %v", err)
	}
}

func TestPricingServiceRejectsManualOverrideWhenDisabled(t *testing.T) {
	repo := &fakePricingRepository{}
	service := application.NewService(repo, &fakeManualOverrideGuard{err: domain.ErrManualPriceOverrideDisabled}, &fakeMarginFloorResolver{value: 10})

	_, err := service.SetProductPrice(context.Background(), application.SetProductPriceCommand{
		TenantID:        "tenant-1",
		ProductID:       "prd_1",
		CurrencyCode:    "BRL",
		PriceAmount:     120,
		CostBasisAmount: 80,
		PricingStatus:   "active",
		EffectiveFrom:   time.Now().UTC(),
		OriginType:      "manual",
		ReasonCode:      "blocked",
		UpdatedBy:       "admin-local",
	})
	if !errors.Is(err, domain.ErrManualPriceOverrideDisabled) {
		t.Fatalf("expected ErrManualPriceOverrideDisabled, got %v", err)
	}
}
