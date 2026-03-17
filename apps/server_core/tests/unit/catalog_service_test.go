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
	traceID       string
	list          []domain.Product
	identifiers   []domain.ProductIdentifier
	taxonomyNodes []domain.TaxonomyNode
	taxonomyDefs  []domain.TaxonomyLevelDef
	err           error
}

func (f *fakeCatalogRepository) CreateProduct(_ context.Context, product domain.Product, traceID string) error {
	if f.err != nil {
		return f.err
	}
	f.created = product
	f.traceID = traceID
	return nil
}

func (f *fakeCatalogRepository) ListProducts(context.Context, string) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func (f *fakeCatalogRepository) ListProductIdentifiers(context.Context, string, string) ([]domain.ProductIdentifier, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identifiers, nil
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

type fakeProductDescriptionGuard struct {
	err error
}

func (f *fakeProductDescriptionGuard) ValidateDescription(context.Context, string, string) error {
	return f.err
}

func TestCatalogServiceCreatesProduct(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})

	product, err := service.CreateProduct(context.Background(), application.CreateProductCommand{
		TenantID:              "tenant-1",
		TraceID:               "trace-catalog-create",
		SKU:                   "SKU-001",
		Name:                  "Steel Sheet",
		Description:           "Galvanized steel sheet for roofing.",
		BrandName:             "Acme Steel",
		StockProfileCode:      "standard",
		PrimaryTaxonomyNodeID: "txn_leaf_1",
		Status:                "active",
		Identifiers: []application.CreateProductIdentifierInput{
			{
				IdentifierType:  "ean",
				IdentifierValue: "789000000001",
				SourceSystem:    "erp",
				IsPrimary:       true,
			},
		},
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
	if repo.created.Description != "Galvanized steel sheet for roofing." {
		t.Fatalf("expected description to be propagated, got %q", repo.created.Description)
	}
	if repo.traceID != "trace-catalog-create" {
		t.Fatalf("expected trace id to be propagated, got %q", repo.traceID)
	}
	if repo.created.PrimaryTaxonomyNodeID != "txn_leaf_1" {
		t.Fatalf("expected taxonomy node txn_leaf_1, got %q", repo.created.PrimaryTaxonomyNodeID)
	}
	if len(repo.created.Identifiers) != 1 || repo.created.Identifiers[0].IdentifierType != "ean" {
		t.Fatalf("expected identifiers to be created, got %+v", repo.created.Identifiers)
	}
}

func TestCatalogServiceRejectsMissingSKU(t *testing.T) {
	repo := &fakeCatalogRepository{}
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})

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
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: false}, &fakeProductDescriptionGuard{})

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
	service := application.NewService(repo, &fakeProductCreationGuard{enabled: true}, &fakeProductDescriptionGuard{})

	defs, err := service.ListTaxonomyLevelDefs(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(defs) != 1 || defs[0].Label != "Department" {
		t.Fatalf("unexpected taxonomy defs: %+v", defs)
	}
}

func TestCatalogServiceListsProductIdentifiers(t *testing.T) {
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

	identifiers, err := service.ListProductIdentifiers(context.Background(), "tenant-1", "prd_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(identifiers) != 1 || identifiers[0].IdentifierValue != "789000000001" {
		t.Fatalf("unexpected identifiers: %+v", identifiers)
	}
}
