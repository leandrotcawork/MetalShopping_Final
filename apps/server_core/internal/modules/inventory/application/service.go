package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/inventory/domain"
	"metalshopping/server_core/internal/modules/inventory/ports"
)

type SetProductPositionCommand struct {
	TenantID           string
	TraceID            string
	ProductID          string
	SourceCompanyCode  string
	SourceLocationCode string
	OnHandQuantity     float64
	LastPurchaseAt     *time.Time
	LastSaleAt         *time.Time
	PositionStatus     string
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	OriginType         string
	OriginRef          string
	ReasonCode         string
	UpdatedBy          string
}

type Service struct {
	repo ports.Repository
	now  func() time.Time
}

func NewService(repo ports.Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetProductPosition(ctx context.Context, cmd SetProductPositionCommand) (domain.ProductPosition, bool, error) {
	status := domain.PositionStatus(strings.ToLower(strings.TrimSpace(cmd.PositionStatus)))
	if status == "" {
		status = domain.PositionStatusActive
	}

	originType := domain.OriginType(strings.ToLower(strings.TrimSpace(cmd.OriginType)))
	if originType == "" {
		originType = domain.OriginTypeImport
	}

	now := s.now()
	position := domain.ProductPosition{
		PositionID:         generatePositionID(),
		TenantID:           strings.TrimSpace(cmd.TenantID),
		ProductID:          strings.TrimSpace(cmd.ProductID),
		SourceCompanyCode:  strings.TrimSpace(cmd.SourceCompanyCode),
		SourceLocationCode: strings.TrimSpace(cmd.SourceLocationCode),
		OnHandQuantity:     cmd.OnHandQuantity,
		LastPurchaseAt:     normalizeTimePointer(cmd.LastPurchaseAt),
		LastSaleAt:         normalizeTimePointer(cmd.LastSaleAt),
		PositionStatus:     status,
		EffectiveFrom:      cmd.EffectiveFrom.UTC(),
		EffectiveTo:        normalizeTimePointer(cmd.EffectiveTo),
		OriginType:         originType,
		OriginRef:          strings.TrimSpace(cmd.OriginRef),
		ReasonCode:         strings.TrimSpace(cmd.ReasonCode),
		UpdatedBy:          strings.TrimSpace(cmd.UpdatedBy),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := position.ValidateForWrite(); err != nil {
		return domain.ProductPosition{}, false, err
	}
	return s.repo.CreateProductPosition(ctx, position, strings.TrimSpace(cmd.TraceID))
}

func (s *Service) ListProductPositions(ctx context.Context, tenantID, productID string) ([]domain.ProductPosition, error) {
	return s.repo.ListProductPositions(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func (s *Service) GetCurrentProductPosition(ctx context.Context, tenantID, productID string) (domain.ProductPosition, error) {
	return s.repo.GetCurrentProductPosition(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func generatePositionID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "pos_fallback"
	}
	return "pos_" + hex.EncodeToString(buf)
}
