package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type PricingStatus string

const (
	PricingStatusDraft    PricingStatus = "draft"
	PricingStatusActive   PricingStatus = "active"
	PricingStatusInactive PricingStatus = "inactive"
)

type OriginType string

const (
	OriginTypeManual OriginType = "manual"
	OriginTypePolicy OriginType = "policy"
	OriginTypeImport OriginType = "import"
)

type ProductPrice struct {
	PriceID          string
	TenantID         string
	ProductID        string
	CurrencyCode     string
	PriceAmount      float64
	CostBasisAmount  float64
	MarginFloorValue float64
	PricingStatus    PricingStatus
	EffectiveFrom    time.Time
	EffectiveTo      *time.Time
	OriginType       OriginType
	OriginRef        string
	ReasonCode       string
	UpdatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

func (p ProductPrice) ValidateForWrite() error {
	if strings.TrimSpace(p.PriceID) == "" {
		return ErrPriceIDRequired
	}
	if strings.TrimSpace(p.TenantID) == "" {
		return ErrTenantIDRequired
	}
	if strings.TrimSpace(p.ProductID) == "" {
		return ErrProductIDRequired
	}
	if strings.TrimSpace(p.CurrencyCode) == "" {
		return ErrCurrencyCodeRequired
	}
	if !currencyCodePattern.MatchString(strings.TrimSpace(p.CurrencyCode)) {
		return fmt.Errorf("%w: %s", ErrInvalidCurrencyCode, p.CurrencyCode)
	}
	if p.PriceAmount < 0 {
		return ErrPriceAmountInvalid
	}
	if p.CostBasisAmount < 0 {
		return ErrCostBasisAmountInvalid
	}
	if p.MarginFloorValue < 0 {
		return ErrMarginFloorValueInvalid
	}
	if !p.PricingStatus.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidPricingStatus, p.PricingStatus)
	}
	if !p.OriginType.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidOriginType, p.OriginType)
	}
	if strings.TrimSpace(p.ReasonCode) == "" {
		return ErrReasonCodeRequired
	}
	if strings.TrimSpace(p.UpdatedBy) == "" {
		return ErrUpdatedByRequired
	}
	if p.EffectiveFrom.IsZero() {
		return ErrEffectiveFromRequired
	}
	if p.EffectiveTo != nil && !p.EffectiveTo.After(p.EffectiveFrom) {
		return ErrInvalidEffectiveWindow
	}
	return nil
}

func (s PricingStatus) IsValid() bool {
	return s == PricingStatusDraft || s == PricingStatusActive || s == PricingStatusInactive
}

func (o OriginType) IsValid() bool {
	return o == OriginTypeManual || o == OriginTypePolicy || o == OriginTypeImport
}

func (p ProductPrice) MarginPercent() float64 {
	if p.PriceAmount <= 0 {
		if p.CostBasisAmount == 0 {
			return 0
		}
		return -100
	}
	return ((p.PriceAmount - p.CostBasisAmount) / p.PriceAmount) * 100
}
