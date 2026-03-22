package readmodel

import (
	"context"
	"strings"
	"time"
)

const (
	defaultProductsPortfolioLimit = 50
	maxProductsPortfolioLimit     = 200
	defaultProductsPortfolioSortKey = "pn_interno"
	defaultProductsPortfolioSortDirection = "asc"
)

type ProductsPortfolioFilter struct {
	Search            string
	BrandNames        []string
	TaxonomyLeaf0Names []string
	Statuses          []string
	SortKey           string
	SortDirection     string
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
	filter.BrandNames = normalizeStringList(filter.BrandNames)
	filter.TaxonomyLeaf0Names = normalizeStringList(filter.TaxonomyLeaf0Names)
	filter.Statuses = normalizeLowerStringList(filter.Statuses)
	filter.SortKey = strings.ToLower(strings.TrimSpace(filter.SortKey))
	filter.SortDirection = strings.ToLower(strings.TrimSpace(filter.SortDirection))

	if filter.Limit <= 0 {
		filter.Limit = defaultProductsPortfolioLimit
	}
	if filter.Limit > maxProductsPortfolioLimit {
		filter.Limit = maxProductsPortfolioLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.SortKey == "" {
		filter.SortKey = defaultProductsPortfolioSortKey
	}
	if filter.SortDirection == "" {
		filter.SortDirection = defaultProductsPortfolioSortDirection
	}

	return filter
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeLowerStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized := strings.ToLower(trimmed)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
