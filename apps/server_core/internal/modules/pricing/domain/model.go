package domain

import (
	"fmt"
	"math"
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

const DefaultPriceContextCode = "default"

type ProductPrice struct {
	PriceID               string
	TenantID              string
	ProductID             string
	PriceContextCode      string
	CurrencyCode          string
	PriceAmount           float64
	ReplacementCostAmount float64
	AverageCostAmount     *float64
	PricingStatus         PricingStatus
	EffectiveFrom         time.Time
	EffectiveTo           *time.Time
	OriginType            OriginType
	OriginRef             string
	ReasonCode            string
	UpdatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
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
	if p.ReplacementCostAmount < 0 {
		return ErrReplacementCostAmountInvalid
	}
	if p.AverageCostAmount != nil && *p.AverageCostAmount < 0 {
		return ErrAverageCostAmountInvalid
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

func NormalizePriceContextCode(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return DefaultPriceContextCode
	}
	return normalized
}

func (s PricingStatus) IsValid() bool {
	return s == PricingStatusDraft || s == PricingStatusActive || s == PricingStatusInactive
}

func (o OriginType) IsValid() bool {
	return o == OriginTypeManual || o == OriginTypePolicy || o == OriginTypeImport
}

func (p ProductPrice) HasSameCommercialState(other ProductPrice) bool {
	return strings.EqualFold(strings.TrimSpace(p.TenantID), strings.TrimSpace(other.TenantID)) &&
		strings.EqualFold(strings.TrimSpace(p.ProductID), strings.TrimSpace(other.ProductID)) &&
		NormalizePriceContextCode(p.PriceContextCode) == NormalizePriceContextCode(other.PriceContextCode) &&
		strings.EqualFold(strings.TrimSpace(p.CurrencyCode), strings.TrimSpace(other.CurrencyCode)) &&
		sameRoundedNumber(p.PriceAmount, other.PriceAmount) &&
		sameRoundedNumber(p.ReplacementCostAmount, other.ReplacementCostAmount) &&
		sameOptionalRoundedNumber(p.AverageCostAmount, other.AverageCostAmount) &&
		p.PricingStatus == other.PricingStatus &&
		sameOptionalTime(p.EffectiveTo, other.EffectiveTo)
}

func sameRoundedNumber(a, b float64) bool {
	return math.Abs(a-b) < 0.00005
}

func sameOptionalRoundedNumber(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return sameRoundedNumber(*a, *b)
}

func sameOptionalTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.UTC().Equal(b.UTC())
}
