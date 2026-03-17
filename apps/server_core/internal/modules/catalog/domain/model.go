package domain

import (
	"fmt"
	"strings"
	"time"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

type Product struct {
	ProductID             string
	TenantID              string
	SKU                   string
	Name                  string
	BrandName             string
	StockProfileCode      string
	PrimaryTaxonomyNodeID string
	Status                ProductStatus
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TaxonomyNode struct {
	TaxonomyNodeID       string
	TenantID             string
	Name                 string
	NameNorm             string
	Code                 string
	ParentTaxonomyNodeID string
	Level                int
	Path                 string
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type TaxonomyLevelDef struct {
	TenantID   string
	Level      int
	Label      string
	ShortLabel string
	IsEnabled  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p Product) ValidateForCreate() error {
	if strings.TrimSpace(p.TenantID) == "" {
		return ErrTenantIDRequired
	}
	if strings.TrimSpace(p.SKU) == "" {
		return ErrSKURequired
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrProductNameRequired
	}
	if !p.Status.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidProductStatus, p.Status)
	}
	return nil
}

func (n TaxonomyNode) Validate() error {
	if strings.TrimSpace(n.TenantID) == "" {
		return ErrTenantIDRequired
	}
	if strings.TrimSpace(n.TaxonomyNodeID) == "" {
		return ErrTaxonomyNodeIDRequired
	}
	if strings.TrimSpace(n.Name) == "" {
		return ErrTaxonomyNodeNameRequired
	}
	if n.Level < 0 {
		return ErrInvalidTaxonomyLevel
	}
	return nil
}

func (d TaxonomyLevelDef) Validate() error {
	if strings.TrimSpace(d.TenantID) == "" {
		return ErrTenantIDRequired
	}
	if d.Level < 0 {
		return ErrInvalidTaxonomyLevel
	}
	if strings.TrimSpace(d.Label) == "" {
		return ErrTaxonomyLevelLabelRequired
	}
	return nil
}

func (s ProductStatus) IsValid() bool {
	return s == ProductStatusActive || s == ProductStatusInactive
}
