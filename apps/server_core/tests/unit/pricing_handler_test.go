package unit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/pricing/application"
	"metalshopping/server_core/internal/modules/pricing/domain"
	pricinghttp "metalshopping/server_core/internal/modules/pricing/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

func TestPricingHandlerSetsProductPrice(t *testing.T) {
	repo := &fakePricingRepository{applied: true}
	service := application.NewService(repo, &fakeManualOverrideGuard{})
	handler := pricinghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/products/prd_1/prices", strings.NewReader(`{"currency_code":"BRL","price_amount":120,"replacement_cost_amount":90,"average_cost_amount":84.5,"pricing_status":"active","effective_from":"2026-03-17T12:00:00Z","origin_type":"manual","reason_code":"initial_price"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "pricing-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("X-Change-Applied") != "true" {
		t.Fatalf("expected X-Change-Applied true, got %q", rr.Header().Get("X-Change-Applied"))
	}
	if !strings.Contains(rr.Body.String(), `"currency_code":"BRL"`) {
		t.Fatalf("expected currency code in response, got %s", rr.Body.String())
	}
}

func TestPricingHandlerReturns200WhenPriceDidNotChange(t *testing.T) {
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
	handler := pricinghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/products/prd_1/prices", strings.NewReader(`{"currency_code":"BRL","price_amount":120,"replacement_cost_amount":90,"average_cost_amount":84.5,"pricing_status":"active","effective_from":"2026-03-18T12:00:00Z","origin_type":"import","reason_code":"rerun_same_price"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "pricing-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("X-Change-Applied") != "false" {
		t.Fatalf("expected X-Change-Applied false, got %q", rr.Header().Get("X-Change-Applied"))
	}
	if !strings.Contains(rr.Body.String(), `"price_id":"prc_current"`) {
		t.Fatalf("expected current price to be returned, got %s", rr.Body.String())
	}
}

func TestPricingHandlerListsProductPrices(t *testing.T) {
	repo := &fakePricingRepository{
		list: []domain.ProductPrice{
			{
				PriceID:               "prc_1",
				TenantID:              "tenant-1",
				ProductID:             "prd_1",
				CurrencyCode:          "BRL",
				PriceAmount:           120,
				ReplacementCostAmount: 90,
				PricingStatus:         domain.PricingStatusActive,
				EffectiveFrom:         time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
				OriginType:            domain.OriginTypeManual,
				ReasonCode:            "initial_price",
				UpdatedBy:             "pricing-admin",
			},
		},
	}
	service := application.NewService(repo, &fakeManualOverrideGuard{})
	handler := pricinghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pricing/products/prd_1/prices", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "pricing-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"price_id":"prc_1"`) {
		t.Fatalf("expected price id in response, got %s", rr.Body.String())
	}
}

func TestPricingHandlerGetsCurrentProductPrice(t *testing.T) {
	repo := &fakePricingRepository{
		current: domain.ProductPrice{
			PriceID:               "prc_2",
			TenantID:              "tenant-1",
			ProductID:             "prd_1",
			CurrencyCode:          "BRL",
			PriceAmount:           130,
			ReplacementCostAmount: 95,
			PricingStatus:         domain.PricingStatusActive,
			EffectiveFrom:         time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
			OriginType:            domain.OriginTypeManual,
			ReasonCode:            "current_price",
			UpdatedBy:             "pricing-admin",
		},
	}
	service := application.NewService(repo, &fakeManualOverrideGuard{})
	handler := pricinghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pricing/products/prd_1/prices/current", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "pricing-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"price_amount":130`) {
		t.Fatalf("expected current price in response, got %s", rr.Body.String())
	}
}

func TestPricingHandlerRejectsGovernanceBlockedWrite(t *testing.T) {
	repo := &fakePricingRepository{}
	service := application.NewService(repo, &fakeManualOverrideGuard{err: domain.ErrManualPriceOverrideDisabled})
	handler := pricinghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/products/prd_1/prices", strings.NewReader(`{"currency_code":"BRL","price_amount":120,"replacement_cost_amount":90,"pricing_status":"active","effective_from":"2026-03-17T12:00:00Z","origin_type":"manual","reason_code":"blocked"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "pricing-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "GOVERNANCE_DISABLED") {
		t.Fatalf("expected GOVERNANCE_DISABLED response, got %s", rr.Body.String())
	}
}
