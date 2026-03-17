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
	ProductID string
	TenantID  string
	SKU       string
	Name      string
	Status    ProductStatus
	CreatedAt time.Time
	UpdatedAt time.Time
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

func (s ProductStatus) IsValid() bool {
	return s == ProductStatusActive || s == ProductStatusInactive
}
