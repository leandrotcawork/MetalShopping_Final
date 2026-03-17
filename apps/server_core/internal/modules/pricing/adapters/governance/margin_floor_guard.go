package governance

import (
	"context"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

type MarginFloorGuard struct {
	resolver    *threshold_resolver.Resolver
	environment string
}

func NewMarginFloorGuard(resolver *threshold_resolver.Resolver, environment string) *MarginFloorGuard {
	return &MarginFloorGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

func (g *MarginFloorGuard) ResolveMarginFloor(ctx context.Context, tenantID string, explicitMarginFloor float64) (float64, error) {
	value := explicitMarginFloor
	if g == nil || g.resolver == nil {
		return value, nil
	}

	resolved, found, err := g.resolver.Resolve(bootstrap.PricingDefaultMarginFloorKey, threshold_resolver.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
		Module:      "pricing",
	})
	if err != nil {
		return 0, fmt.Errorf("resolve pricing margin floor threshold: %w", err)
	}
	if !found {
		return value, nil
	}
	if resolved > value {
		return resolved, nil
	}
	return value, nil
}
