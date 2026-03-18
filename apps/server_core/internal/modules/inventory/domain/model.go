package domain

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

type PositionStatus string

const (
	PositionStatusActive   PositionStatus = "active"
	PositionStatusInactive PositionStatus = "inactive"
)

type OriginType string

const (
	OriginTypeManual         OriginType = "manual"
	OriginTypeImport         OriginType = "import"
	OriginTypeReconciliation OriginType = "reconciliation"
)

type ProductPosition struct {
	PositionID     string
	TenantID       string
	ProductID      string
	OnHandQuantity float64
	LastPurchaseAt *time.Time
	LastSaleAt     *time.Time
	PositionStatus PositionStatus
	EffectiveFrom  time.Time
	EffectiveTo    *time.Time
	OriginType     OriginType
	OriginRef      string
	ReasonCode     string
	UpdatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var productIDPattern = regexp.MustCompile(`^prd_[a-z0-9]+$`)

func (p ProductPosition) ValidateForWrite() error {
	if strings.TrimSpace(p.PositionID) == "" {
		return ErrPositionIDRequired
	}
	if strings.TrimSpace(p.TenantID) == "" {
		return ErrTenantIDRequired
	}
	if strings.TrimSpace(p.ProductID) == "" {
		return ErrProductIDRequired
	}
	if !productIDPattern.MatchString(strings.ToLower(strings.TrimSpace(p.ProductID))) {
		return fmt.Errorf("%w: %s", ErrProductIDRequired, p.ProductID)
	}
	if p.OnHandQuantity < 0 {
		return ErrOnHandQuantityInvalid
	}
	if !p.PositionStatus.IsValid() {
		return ErrInvalidPositionStatus
	}
	if p.EffectiveFrom.IsZero() {
		return ErrEffectiveFromRequired
	}
	if p.EffectiveTo != nil && !p.EffectiveTo.After(p.EffectiveFrom) {
		return ErrInvalidEffectiveWindow
	}
	if !p.OriginType.IsValid() {
		return ErrInvalidOriginType
	}
	if strings.TrimSpace(p.ReasonCode) == "" {
		return ErrReasonCodeRequired
	}
	if strings.TrimSpace(p.UpdatedBy) == "" {
		return ErrUpdatedByRequired
	}
	return nil
}

func (s PositionStatus) IsValid() bool {
	return s == PositionStatusActive || s == PositionStatusInactive
}

func (o OriginType) IsValid() bool {
	return o == OriginTypeManual || o == OriginTypeImport || o == OriginTypeReconciliation
}

func (p ProductPosition) HasSameOperationalState(other ProductPosition) bool {
	return strings.EqualFold(strings.TrimSpace(p.TenantID), strings.TrimSpace(other.TenantID)) &&
		strings.EqualFold(strings.TrimSpace(p.ProductID), strings.TrimSpace(other.ProductID)) &&
		sameRoundedNumber(p.OnHandQuantity, other.OnHandQuantity) &&
		sameOptionalTime(p.LastPurchaseAt, other.LastPurchaseAt) &&
		sameOptionalTime(p.LastSaleAt, other.LastSaleAt) &&
		p.PositionStatus == other.PositionStatus &&
		sameOptionalTime(p.EffectiveTo, other.EffectiveTo)
}

func sameRoundedNumber(a, b float64) bool {
	return math.Abs(a-b) < 0.00005
}

func sameOptionalTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.UTC().Equal(b.UTC())
}
