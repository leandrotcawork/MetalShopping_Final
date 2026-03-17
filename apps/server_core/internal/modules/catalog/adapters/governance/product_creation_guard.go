package governance

import (
	"context"

	"metalshopping/server_core/internal/modules/catalog/ports"
	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
)

var _ ports.ProductCreationGuard = (*ProductCreationGuard)(nil)

type ProductCreationGuard struct {
	resolver    *feature_flags.Resolver
	environment string
}

func NewProductCreationGuard(resolver *feature_flags.Resolver, environment string) *ProductCreationGuard {
	return &ProductCreationGuard{
		resolver:    resolver,
		environment: environment,
	}
}

func (g *ProductCreationGuard) IsProductCreationEnabled(_ context.Context, tenantID string) (bool, error) {
	return g.resolver.Resolve(bootstrap.CatalogProductCreationEnabledKey, feature_flags.ResolutionContext{
		Environment: g.environment,
		TenantID:    tenantID,
	})
}
