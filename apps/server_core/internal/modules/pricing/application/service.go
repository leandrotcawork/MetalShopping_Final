package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/pricing/domain"
	"metalshopping/server_core/internal/modules/pricing/ports"
)

type SetProductPriceCommand struct {
	TenantID              string
	TraceID               string
	ProductID             string
	CurrencyCode          string
	PriceAmount           float64
	ReplacementCostAmount float64
	AverageCostAmount     *float64
	PricingStatus         string
	EffectiveFrom         time.Time
	EffectiveTo           *time.Time
	OriginType            string
	OriginRef             string
	ReasonCode            string
	UpdatedBy             string
}

type Service struct {
	repo                ports.Repository
	manualOverrideGuard ports.ManualPriceOverrideGuard
	now                 func() time.Time
}

func NewService(repo ports.Repository, manualOverrideGuard ports.ManualPriceOverrideGuard) *Service {
	return &Service{
		repo:                repo,
		manualOverrideGuard: manualOverrideGuard,
		now:                 func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetProductPrice(ctx context.Context, cmd SetProductPriceCommand) (domain.ProductPrice, bool, error) {
	status := domain.PricingStatus(strings.ToLower(strings.TrimSpace(cmd.PricingStatus)))
	if status == "" {
		status = domain.PricingStatusActive
	}

	originType := domain.OriginType(strings.ToLower(strings.TrimSpace(cmd.OriginType)))
	if originType == "" {
		originType = domain.OriginTypeManual
	}

	if s.manualOverrideGuard != nil {
		if err := s.manualOverrideGuard.ValidateManualOverride(ctx, strings.TrimSpace(cmd.TenantID), originType); err != nil {
			return domain.ProductPrice{}, false, err
		}
	}

	now := s.now()
	price := domain.ProductPrice{
		PriceID:               generatePriceID(),
		TenantID:              strings.TrimSpace(cmd.TenantID),
		ProductID:             strings.TrimSpace(cmd.ProductID),
		CurrencyCode:          strings.ToUpper(strings.TrimSpace(cmd.CurrencyCode)),
		PriceAmount:           cmd.PriceAmount,
		ReplacementCostAmount: cmd.ReplacementCostAmount,
		AverageCostAmount:     normalizeFloatPointer(cmd.AverageCostAmount),
		PricingStatus:         status,
		EffectiveFrom:         cmd.EffectiveFrom.UTC(),
		EffectiveTo:           normalizeTimePointer(cmd.EffectiveTo),
		OriginType:            originType,
		OriginRef:             strings.TrimSpace(cmd.OriginRef),
		ReasonCode:            strings.TrimSpace(cmd.ReasonCode),
		UpdatedBy:             strings.TrimSpace(cmd.UpdatedBy),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := price.ValidateForWrite(); err != nil {
		return domain.ProductPrice{}, false, err
	}
	storedPrice, applied, err := s.repo.CreateProductPrice(ctx, price, strings.TrimSpace(cmd.TraceID))
	if err != nil {
		return domain.ProductPrice{}, false, err
	}
	return storedPrice, applied, nil
}

func (s *Service) ListProductPrices(ctx context.Context, tenantID, productID string) ([]domain.ProductPrice, error) {
	return s.repo.ListProductPrices(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func (s *Service) GetCurrentProductPrice(ctx context.Context, tenantID, productID string) (domain.ProductPrice, error) {
	return s.repo.GetCurrentProductPrice(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func normalizeFloatPointer(value *float64) *float64 {
	if value == nil {
		return nil
	}
	normalized := *value
	return &normalized
}

func generatePriceID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "prc_fallback"
	}
	return "prc_" + hex.EncodeToString(buf)
}
