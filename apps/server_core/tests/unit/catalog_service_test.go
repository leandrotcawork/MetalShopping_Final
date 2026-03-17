package unit

import (
	"context"
	"errors"
	"testing"

	"metalshopping/server_core/internal/modules/catalog/application"
	"metalshopping/server_core/internal/modules/catalog/domain"
)

type fakeCatalogRepository struct {
	created domain.Product
	list    []domain.Product
	err     error
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
		TenantID: "tenant-1",
		SKU:      "SKU-001",
		Name:     "Steel Sheet",
		Status:   "active",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.ProductID == "" {
		t.Fatal("expected generated product id")
	}
	if repo.created.SKU != "SKU-001" {
		t.Fatalf("expected created SKU-001, got %q", repo.created.SKU)
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
