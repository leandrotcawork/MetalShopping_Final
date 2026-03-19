package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/shopping/ports"
)

type Service struct {
	reader ports.Reader
}

func NewService(reader ports.Reader) *Service {
	return &Service{reader: reader}
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
	return s.reader.GetSummary(ctx, strings.TrimSpace(tenantID))
}

func (s *Service) ListRuns(ctx context.Context, tenantID string, filter ports.RunListFilter) (ports.RunList, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	return s.reader.ListRuns(ctx, strings.TrimSpace(tenantID), filter)
}

func (s *Service) GetRun(ctx context.Context, tenantID, runID string) (ports.Run, error) {
	return s.reader.GetRun(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(runID))
}

func (s *Service) GetProductLatest(ctx context.Context, tenantID, productID string) (ports.ProductLatest, error) {
	return s.reader.GetProductLatest(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}
