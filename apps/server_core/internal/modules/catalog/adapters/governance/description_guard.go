package governance

import (
	"context"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

type DescriptionGuard struct {
	resolver    *threshold_resolver.Resolver
	environment string
}

func NewDescriptionGuard(resolver *threshold_resolver.Resolver, environment string) *DescriptionGuard {
	return &DescriptionGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

func (g *DescriptionGuard) ValidateDescription(ctx context.Context, tenantID, description string) error {
	if g == nil || g.resolver == nil || strings.TrimSpace(description) == "" {
		return nil
	}

	value, found, err := g.resolver.Resolve(bootstrap.CatalogMaxDescriptionLengthKey, threshold_resolver.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
		Module:      "catalog",
	})
	if err != nil {
		return fmt.Errorf("resolve catalog description threshold: %w", err)
	}
	if !found {
		return nil
	}
	if float64(len(description)) > value {
		return domain.ErrProductDescriptionTooLong
	}
	return nil
}
