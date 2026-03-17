package ports

import (
	"context"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/modules/pricing/domain"
)

type Repository interface {
	CreateProductPrice(ctx context.Context, price domain.ProductPrice, traceID string) error
	ListProductPrices(ctx context.Context, tenantID, productID string) ([]domain.ProductPrice, error)
	GetCurrentProductPrice(ctx context.Context, tenantID, productID string) (domain.ProductPrice, error)
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID string, permission iamdomain.Permission) (bool, error)
}

type ManualPriceOverrideGuard interface {
	ValidateManualOverride(ctx context.Context, tenantID string, originType domain.OriginType) error
}

type MarginFloorResolver interface {
	ResolveMarginFloor(ctx context.Context, tenantID string, explicitMarginFloor float64) (float64, error)
}
