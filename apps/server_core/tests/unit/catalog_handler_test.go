package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
	cataloghttp "metalshopping/server_core/internal/modules/catalog/transport/http"
	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

type fakePermissionChecker struct {
	allowed bool
	err     error
}

func (f *fakePermissionChecker) HasPermission(context.Context, string, iamdomain.Permission) (bool, error) {
	return f.allowed, f.err
}

func TestCatalogHandlerCreatesProduct(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})
	handler := cataloghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products", strings.NewReader(`{"sku":"SKU-001","name":"Steel Sheet","status":"active"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestCatalogHandlerListsProducts(t *testing.T) {
	repo := &fakeCatalogRepository{
		list: []domain.Product{
			{
				ProductID: "prd_1",
				TenantID:  "tenant-1",
				SKU:       "SKU-001",
				Name:      "Steel Sheet",
				Status:    domain.ProductStatusActive,
			},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})
	handler := cataloghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCatalogHandlerRejectsForbiddenUser(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})
	handler := cataloghttp.NewHandler(service, &fakePermissionChecker{allowed: false})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "viewer-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCatalogHandlerRejectsGovernanceDisabledCreate(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: false})
	handler := cataloghttp.NewHandler(service, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products", strings.NewReader(`{"sku":"SKU-001","name":"Steel Sheet","status":"active"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
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
