package unit

import (
	"context"
	"errors"
	"testing"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/catalog/ports"
)

type fakeCatalogRepository struct {
	created       domain.Product
	list          []domain.Product
	taxonomyNodes []domain.TaxonomyNode
	taxonomyDefs  []domain.TaxonomyLevelDef
	err           error
}

func (f *fakeCatalogRepository) CreateProduct(_ context.Context, product domain.Product) error {
	if f.err != nil {
		return f.err
	}
	f.created = product
	return nil
}

func (f *fakeCatalogRepository) ListProducts(context.Context, string) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func (f *fakeCatalogRepository) ListTaxonomyNodes(context.Context, string, ports.TaxonomyNodeFilter) ([]domain.TaxonomyNode, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.taxonomyNodes, nil
}

func (f *fakeCatalogRepository) ListTaxonomyLevelDefs(context.Context, string) ([]domain.TaxonomyLevelDef, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.taxonomyDefs, nil
}

type fakeProductCreationGuard struct {
	enabled bool
	err     error
}

func (f *fakeProductCreationGuard) IsProductCreationEnabled(context.Context, string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.enabled, nil
}

func TestCatalogServiceCreatesProduct(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})

	product, err := service.CreateProduct(context.Background(), application.CreateProductCommand{
		TenantID:              "tenant-1",
		SKU:                   "SKU-001",
		Name:                  "Steel Sheet",
		BrandName:             "Acme Steel",
		StockProfileCode:      "standard",
		PrimaryTaxonomyNodeID: "txn_leaf_1",
		Status:                "active",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.ProductID == "" {
		t.Fatal("expected generated product id")
	}
	if repo.created.BrandName != "Acme Steel" {
		t.Fatalf("expected brand Acme Steel, got %q", repo.created.BrandName)
	}
	if repo.created.PrimaryTaxonomyNodeID != "txn_leaf_1" {
		t.Fatalf("expected taxonomy node txn_leaf_1, got %q", repo.created.PrimaryTaxonomyNodeID)
	}
}

func TestCatalogServiceRejectsMissingSKU(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})

	_, err := service.CreateProduct(context.Background(), application.CreateProductCommand{
		TenantID: "tenant-1",
		Name:     "Steel Sheet",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCatalogServiceRejectsGovernanceDisabledCreate(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: false})

	_, err := service.CreateProduct(context.Background(), application.CreateProductCommand{
		TenantID: "tenant-1",
		SKU:      "SKU-001",
		Name:     "Steel Sheet",
		Status:   "active",
	})
	if !errors.Is(err, domain.ErrProductCreationDisabled) {
		t.Fatalf("expected ErrProductCreationDisabled, got %v", err)
	}
}

func TestCatalogServiceListsTaxonomyLevelDefs(t *testing.T) {
	repo := &fakeCatalogRepository{
		taxonomyDefs: []domain.TaxonomyLevelDef{
			{TenantID: "tenant-1", Level: 0, Label: "Department", IsEnabled: true},
		},
	}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true})

	defs, err := service.ListTaxonomyLevelDefs(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(defs) != 1 || defs[0].Label != "Department" {
		t.Fatalf("unexpected taxonomy defs: %+v", defs)
	}
}
