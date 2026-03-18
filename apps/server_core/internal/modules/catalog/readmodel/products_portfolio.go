package readmodel

import (
	"context"
	"strings"
	"time"
)

const (
	defaultProductsPortfolioLimit = 50
	maxProductsPortfolioLimit     = 200
)

type ProductsPortfolioFilter struct {
	Search            string
	BrandName         string
	TaxonomyLeaf0Name string
	Status            string
	Limit             int
	Offset            int
}

type ProductsPortfolioItem struct {
	ProductID               string
	SKU                     string
	Name                    string
	Description             *string
	BrandName               *string
	PNInterno               *string
	Reference               *string
	EAN                     *string
	TaxonomyLeafName        *string
	TaxonomyLeaf0Name       *string
	StockProfileCode        *string
	ProductStatus           string
	CurrentPriceAmount      *float64
	ReplacementCostAmount   *float64
	AverageCostAmount       *float64
	CurrencyCode            *string
	OnHandQuantity          *float64
	InventoryPositionStatus *string
	UpdatedAt               time.Time
}

type ProductsPortfolioFilters struct {
	Brands             []string
	TaxonomyLeaf0Names []string
	TaxonomyLeaf0Label string
	Status             []string
}

type ProductsPortfolioPaging struct {
	Offset   int
	Limit    int
	Returned int
	Total    int
}

type ProductsPortfolioResult struct {
	Rows    []ProductsPortfolioItem
	Filters ProductsPortfolioFilters
	Paging  ProductsPortfolioPaging
}

type ProductsPortfolioRepository interface {
	ListProductsPortfolio(context.Context, string, ProductsPortfolioFilter) (ProductsPortfolioResult, error)
}

type ProductsPortfolioService struct {
	repo ProductsPortfolioRepository
}

func NewProductsPortfolioService(repo ProductsPortfolioRepository) *ProductsPortfolioService {
	return &ProductsPortfolioService{repo: repo}
}

func (s *ProductsPortfolioService) ListProductsPortfolio(ctx context.Context, tenantID string, filter ProductsPortfolioFilter) (ProductsPortfolioResult, error) {
	normalized := normalizeProductsPortfolioFilter(filter)
	return s.repo.ListProductsPortfolio(ctx, tenantID, normalized)
}

func normalizeProductsPortfolioFilter(filter ProductsPortfolioFilter) ProductsPortfolioFilter {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.BrandName = strings.TrimSpace(filter.BrandName)
	filter.TaxonomyLeaf0Name = strings.TrimSpace(filter.TaxonomyLeaf0Name)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))

	if filter.Limit <= 0 {
		filter.Limit = defaultProductsPortfolioLimit
	}
	if filter.Limit > maxProductsPortfolioLimit {
		filter.Limit = maxProductsPortfolioLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return filter
}
