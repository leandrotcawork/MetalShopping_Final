package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/home/ports"
)

type Service struct {
	reader ports.SummaryReader
}

func NewService(reader ports.SummaryReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
	return s.reader.GetSummary(ctx, strings.TrimSpace(tenantID))
}
