package unit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/inventory/application"
	"metalshopping/server_core/internal/modules/inventory/domain"
	inventoryhttp "metalshopping/server_core/internal/modules/inventory/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

func TestInventoryHandlerSetsProductPosition(t *testing.T) {
	repo := &fakeInventoryRepository{applied: true}
	service := application.NewService(repo)
	handler := inventoryhttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/products/prd_1/positions", strings.NewReader(`{"on_hand_quantity":42,"last_purchase_at":"2026-03-10T08:00:00Z","last_sale_at":"2026-03-16T09:00:00Z","position_status":"active","effective_from":"2026-03-17T12:00:00Z","origin_type":"import","reason_code":"erp_sync"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "inventory-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("X-Change-Applied") != "true" {
		t.Fatalf("expected X-Change-Applied true, got %q", rr.Header().Get("X-Change-Applied"))
	}
	if !strings.Contains(rr.Body.String(), `"on_hand_quantity":42`) {
		t.Fatalf("expected quantity in response, got %s", rr.Body.String())
	}
}

func TestInventoryHandlerReturns200WhenPositionDidNotChange(t *testing.T) {
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
	handler := inventoryhttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/products/prd_1/positions", strings.NewReader(`{"on_hand_quantity":42,"last_purchase_at":"2026-03-10T08:00:00Z","last_sale_at":"2026-03-16T09:00:00Z","position_status":"active","effective_from":"2026-03-18T12:00:00Z","origin_type":"import","reason_code":"rerun_same_position"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "inventory-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("X-Change-Applied") != "false" {
		t.Fatalf("expected X-Change-Applied false, got %q", rr.Header().Get("X-Change-Applied"))
	}
	if !strings.Contains(rr.Body.String(), `"position_id":"pos_current"`) {
		t.Fatalf("expected current position to be returned, got %s", rr.Body.String())
	}
}

func TestInventoryHandlerListsProductPositions(t *testing.T) {
	repo := &fakeInventoryRepository{
		list: []domain.ProductPosition{
			{
				PositionID:     "pos_1",
				TenantID:       "tenant-1",
				ProductID:      "prd_1",
				OnHandQuantity: 42,
				PositionStatus: domain.PositionStatusActive,
				EffectiveFrom:  time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
				OriginType:     domain.OriginTypeImport,
				ReasonCode:     "erp_sync",
				UpdatedBy:      "inventory-sync",
			},
		},
	}
	service := application.NewService(repo)
	handler := inventoryhttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/products/prd_1/positions", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "inventory-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"position_id":"pos_1"`) {
		t.Fatalf("expected position id in response, got %s", rr.Body.String())
	}
}

func TestInventoryHandlerGetsCurrentProductPosition(t *testing.T) {
	repo := &fakeInventoryRepository{
		current: domain.ProductPosition{
			PositionID:     "pos_2",
			TenantID:       "tenant-1",
			ProductID:      "prd_1",
			OnHandQuantity: 21,
			PositionStatus: domain.PositionStatusActive,
			EffectiveFrom:  time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
			OriginType:     domain.OriginTypeImport,
			ReasonCode:     "current_position",
			UpdatedBy:      "inventory-sync",
		},
	}
	service := application.NewService(repo)
	handler := inventoryhttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/products/prd_1/positions/current", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "inventory-admin", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"on_hand_quantity":21`) {
		t.Fatalf("expected current quantity in response, got %s", rr.Body.String())
	}
}
