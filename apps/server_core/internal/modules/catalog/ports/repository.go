package ports

import (
	"context"

	"metalshopping/server_core/internal/modules/catalog/domain"
	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
)

type Repository interface {
	CreateProduct(ctx context.Context, product domain.Product) error
	ListProducts(ctx context.Context, tenantID string) ([]domain.Product, error)
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID string, permission iamdomain.Permission) (bool, error)
}

type ProductCreationGuard interface {
	IsProductCreationEnabled(ctx context.Context, tenantID string) (bool, error)
}
