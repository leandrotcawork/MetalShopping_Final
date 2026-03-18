package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
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
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products", strings.NewReader(`{"sku":"SKU-001","name":"Steel Sheet","description":"Galvanized steel sheet","brand_name":"Acme","stock_profile_code":"standard","primary_taxonomy_node_id":"txn_leaf_1","status":"active"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"brand_name":"Acme"`) {
		t.Fatalf("expected brand_name in response, got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"description":"Galvanized steel sheet"`) {
		t.Fatalf("expected description in response, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerCreatesProductWithIdentifiers(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/products", strings.NewReader(`{"sku":"SKU-002","name":"Fastener","status":"active","identifiers":[{"identifier_type":"ean","identifier_value":"789000000002","source_system":"erp","is_primary":true}]}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"identifier_type":"ean"`) {
		t.Fatalf("expected identifier in response, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerListsProducts(t *testing.T) {
	repo := &fakeCatalogRepository{
		list: []domain.Product{
			{
				ProductID:             "prd_1",
				TenantID:              "tenant-1",
				SKU:                   "SKU-001",
				Name:                  "Steel Sheet",
				Description:           "Galvanized steel sheet",
				BrandName:             "Acme",
				StockProfileCode:      "standard",
				PrimaryTaxonomyNodeID: "txn_leaf_1",
				Status:                domain.ProductStatusActive,
			},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

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
	if !strings.Contains(rr.Body.String(), `"primary_taxonomy_node_id":"txn_leaf_1"`) {
		t.Fatalf("expected taxonomy node in response, got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"description":"Galvanized steel sheet"`) {
		t.Fatalf("expected description in response, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerListsProductsPortfolio(t *testing.T) {
	repo := &fakeCatalogRepository{
		portfolio: catalogreadmodel.ProductsPortfolioResult{
			Rows: []catalogreadmodel.ProductsPortfolioItem{
				{
					ProductID:     "prd_1",
					SKU:           "SKU-001",
					Name:          "Steel Sheet",
					ProductStatus: "active",
					UpdatedAt:     time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC),
				},
			},
			Filters: catalogreadmodel.ProductsPortfolioFilters{
				Brands:             []string{"Acme"},
				TaxonomyLeaf0Names: []string{"Roofing"},
				TaxonomyLeaf0Label: "Departamento",
				Status:             []string{"active"},
			},
			Paging: catalogreadmodel.ProductsPortfolioPaging{
				Offset:   0,
				Limit:    50,
				Returned: 1,
				Total:    1,
			},
		},
	}
	description := "Galvanized steel sheet"
	brand := "Acme"
	currentPrice := 125.5
	replacementCost := 89.75
	onHandQuantity := 37.0
	repo.portfolio.Rows[0].Description = &description
	repo.portfolio.Rows[0].BrandName = &brand
	repo.portfolio.Rows[0].CurrentPriceAmount = &currentPrice
	repo.portfolio.Rows[0].ReplacementCostAmount = &replacementCost
	repo.portfolio.Rows[0].OnHandQuantity = &onHandQuantity

	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	portfolioService := catalogreadmodel.NewProductsPortfolioService(repo)
	handler := cataloghttp.NewHandler(service, portfolioService, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/portfolio?search=SKU&limit=25", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "viewer-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"rows"`) || !strings.Contains(rr.Body.String(), `"filters"`) {
		t.Fatalf("expected products portfolio payload, got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"current_price_amount":125.5`) {
		t.Fatalf("expected current price in response, got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"taxonomy_leaf0_label":"Departamento"`) {
		t.Fatalf("expected taxonomy level label in response, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerListsTaxonomyLevels(t *testing.T) {
	repo := &fakeCatalogRepository{
		taxonomyDefs: []domain.TaxonomyLevelDef{
			{
				TenantID:   "tenant-1",
				Level:      0,
				Label:      "Department",
				ShortLabel: "Dept",
				IsEnabled:  true,
			},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/taxonomy/levels", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"label":"Department"`) {
		t.Fatalf("expected taxonomy level label, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerListsProductIdentifiers(t *testing.T) {
	repo := &fakeCatalogRepository{
		identifiers: []domain.ProductIdentifier{
			{
				ProductIdentifierID: "pid_1",
				ProductID:           "prd_1",
				TenantID:            "tenant-1",
				IdentifierType:      "ean",
				IdentifierValue:     "789000000001",
				SourceSystem:        "erp",
				IsPrimary:           true,
			},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/products/prd_1/identifiers", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"identifier_value":"789000000001"`) {
		t.Fatalf("expected identifier value in response, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerListsTaxonomyNodes(t *testing.T) {
	repo := &fakeCatalogRepository{
		taxonomyNodes: []domain.TaxonomyNode{
			{
				TaxonomyNodeID: "txn_leaf_1",
				TenantID:       "tenant-1",
				Name:           "Screws",
				Level:          2,
				IsActive:       true,
			},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/taxonomy/nodes?level=2", nil)
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-local", TenantID: "tenant-1"}))
	req = req.WithContext(tenancy_runtime.WithTenant(req.Context(), tenancy_runtime.Tenant{ID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"taxonomy_node_id":"txn_leaf_1"`) {
		t.Fatalf("expected taxonomy node id, got %s", rr.Body.String())
	}
}

func TestCatalogHandlerRejectsForbiddenUser(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: false})

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
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: false}, &fakeProductDescriptionGuard{})
	handler := cataloghttp.NewHandler(service, nil, &fakePermissionChecker{allowed: true})

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
