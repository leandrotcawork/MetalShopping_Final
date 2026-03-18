package ports

import (
	"context"

	"metalshopping/server_core/internal/modules/iam/domain"
	inventorydomain "metalshopping/server_core/internal/modules/inventory/domain"
)

type Repository interface {
	CreateProductPosition(ctx context.Context, position inventorydomain.ProductPosition, traceID string) (inventorydomain.ProductPosition, bool, error)
	ListProductPositions(ctx context.Context, tenantID, productID string) ([]inventorydomain.ProductPosition, error)
	GetCurrentProductPosition(ctx context.Context, tenantID, productID string) (inventorydomain.ProductPosition, error)
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID string, permission domain.Permission) (bool, error)
}
