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

type CreateProductIdentifierInput struct {
	IdentifierType  string
	IdentifierValue string
	SourceSystem    string
	IsPrimary       bool
}

type CreateProductCommand struct {
	TenantID              string
	SKU                   string
	Name                  string
	Description           string
	BrandName             string
	StockProfileCode      string
	PrimaryTaxonomyNodeID string
	Status                string
	Identifiers           []CreateProductIdentifierInput
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
	productID := generateProductID()
	product := domain.Product{
		ProductID:             productID,
		TenantID:              strings.TrimSpace(cmd.TenantID),
		SKU:                   strings.TrimSpace(cmd.SKU),
		Name:                  strings.TrimSpace(cmd.Name),
		Description:           strings.TrimSpace(cmd.Description),
		BrandName:             strings.TrimSpace(cmd.BrandName),
		StockProfileCode:      strings.TrimSpace(cmd.StockProfileCode),
		PrimaryTaxonomyNodeID: strings.TrimSpace(cmd.PrimaryTaxonomyNodeID),
		Status:                status,
		Identifiers:           buildIdentifiers(productID, strings.TrimSpace(cmd.TenantID), now, cmd.Identifiers),
		CreatedAt:             now,
		UpdatedAt:             now,
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

func (s *Service) ListProductIdentifiers(ctx context.Context, tenantID, productID string) ([]domain.ProductIdentifier, error) {
	return s.repo.ListProductIdentifiers(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func (s *Service) ListTaxonomyNodes(ctx context.Context, tenantID string, filter ports.TaxonomyNodeFilter) ([]domain.TaxonomyNode, error) {
	return s.repo.ListTaxonomyNodes(ctx, strings.TrimSpace(tenantID), filter)
}

func (s *Service) ListTaxonomyLevelDefs(ctx context.Context, tenantID string) ([]domain.TaxonomyLevelDef, error) {
	return s.repo.ListTaxonomyLevelDefs(ctx, strings.TrimSpace(tenantID))
}

func buildIdentifiers(productID, tenantID string, now time.Time, inputs []CreateProductIdentifierInput) []domain.ProductIdentifier {
	identifiers := make([]domain.ProductIdentifier, 0, len(inputs))
	for _, input := range inputs {
		identifiers = append(identifiers, domain.ProductIdentifier{
			ProductIdentifierID: generateProductIdentifierID(),
			ProductID:           productID,
			TenantID:            tenantID,
			IdentifierType:      strings.TrimSpace(input.IdentifierType),
			IdentifierValue:     strings.TrimSpace(input.IdentifierValue),
			SourceSystem:        strings.TrimSpace(input.SourceSystem),
			IsPrimary:           input.IsPrimary,
			CreatedAt:           now,
			UpdatedAt:           now,
		})
	}
	return identifiers
}

func generateProductID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "prd_fallback"
	}
	return "prd_" + hex.EncodeToString(buf)
}

func generateProductIdentifierID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "pid_fallback"
	}
	return "pid_" + hex.EncodeToString(buf)
}
