package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/analytics_serving/ports"
)

type Service struct {
	reader ports.HomeReader
}

func NewService(reader ports.HomeReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) GetHome(ctx context.Context, tenantID string, requestedSnapshotID string) (ports.Home, error) {
	return s.reader.GetHome(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(requestedSnapshotID))
}
