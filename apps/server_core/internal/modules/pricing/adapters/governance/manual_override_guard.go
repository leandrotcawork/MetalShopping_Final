package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/pricing/domain"
	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
)

type ManualPriceOverridePolicy struct {
	AllowManualPriceOverride bool `json:"allow_manual_price_override"`
}

type ManualOverrideGuard struct {
	resolver    *policy_resolver.Resolver
	environment string
}

func NewManualOverrideGuard(resolver *policy_resolver.Resolver, environment string) *ManualOverrideGuard {
	return &ManualOverrideGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

func (g *ManualOverrideGuard) ValidateManualOverride(ctx context.Context, tenantID string, originType domain.OriginType) error {
	if g == nil || g.resolver == nil || originType != domain.OriginTypeManual {
		return nil
	}

	raw, found, err := g.resolver.Resolve(bootstrap.PricingManualPriceOverrideKey, policy_resolver.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
		Module:      "pricing",
	})
	if err != nil {
		return fmt.Errorf("resolve pricing manual override policy: %w", err)
	}
	if !found {
		return domain.ErrManualPriceOverrideDisabled
	}

	var policy ManualPriceOverridePolicy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return fmt.Errorf("decode pricing manual override policy: %w", err)
	}
	if !policy.AllowManualPriceOverride {
		return domain.ErrManualPriceOverrideDisabled
	}
	return nil
}
