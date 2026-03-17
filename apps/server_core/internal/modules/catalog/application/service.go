package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/modules/catalog/ports"
)

type CreateProductCommand struct {
	TenantID string
	SKU      string
	Name     string
	Status   string
}

type Service struct {
	repo               ports.Repository
	productCreateGuard ports.ProductCreationGuard
	now                func() time.Time
}

func NewService(repo ports.Repository, productCreateGuard ports.ProductCreationGuard) *Service {
	return &Service{
		repo:               repo,
		productCreateGuard: productCreateGuard,
		now:                func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateProduct(ctx context.Context, cmd CreateProductCommand) (domain.Product, error) {
	status := domain.ProductStatus(strings.ToLower(strings.TrimSpace(cmd.Status)))
	if status == "" {
		status = domain.ProductStatusActive
	}

	now := s.now()
	product := domain.Product{
		ProductID: generateProductID(),
		TenantID:  strings.TrimSpace(cmd.TenantID),
		SKU:       strings.TrimSpace(cmd.SKU),
		Name:      strings.TrimSpace(cmd.Name),
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := product.ValidateForCreate(); err != nil {
		return domain.Product{}, err
	}
	if s.productCreateGuard != nil {
		enabled, err := s.productCreateGuard.IsProductCreationEnabled(ctx, product.TenantID)
		if err != nil {
			return domain.Product{}, err
		}
		if !enabled {
			return domain.Product{}, domain.ErrProductCreationDisabled
		}
	}

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (s *Service) ListProducts(ctx context.Context, tenantID string) ([]domain.Product, error) {
	return s.repo.ListProducts(ctx, strings.TrimSpace(tenantID))
}

func generateProductID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "prd_fallback"
	}
	return "prd_" + hex.EncodeToString(buf)
}
