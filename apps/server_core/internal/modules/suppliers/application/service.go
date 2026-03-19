package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/suppliers/ports"
)

type Service struct {
	reader ports.DirectoryReader
}

func NewService(reader ports.DirectoryReader) *Service {
	return &Service{reader: reader}
}

func (s *Service) ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]ports.DirectorySupplier, error) {
	return s.reader.ListDirectory(ctx, strings.TrimSpace(tenantID), onlyEnabled)
}
