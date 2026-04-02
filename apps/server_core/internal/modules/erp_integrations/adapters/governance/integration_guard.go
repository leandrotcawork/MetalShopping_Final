package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
)

// ---------------------------------------------------------------------------
// IntegrationEnabledGuard
// ---------------------------------------------------------------------------

// IntegrationEnabledGuard gates ERP integration operations behind a feature flag.
type IntegrationEnabledGuard struct {
	resolver    *feature_flags.Resolver
	environment string
}

// NewIntegrationEnabledGuard constructs an IntegrationEnabledGuard.
func NewIntegrationEnabledGuard(resolver *feature_flags.Resolver, environment string) *IntegrationEnabledGuard {
	return &IntegrationEnabledGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

// CheckEnabled returns domain.ErrIntegrationDisabled if ERP integration is not
// enabled for the given tenant.
func (g *IntegrationEnabledGuard) CheckEnabled(_ context.Context, tenantID string) error {
	enabled, err := g.resolver.Resolve(bootstrap.ERPIntegrationEnabledKey, feature_flags.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
	})
	if err != nil {
		return fmt.Errorf("resolve erp integration enabled flag: %w", err)
	}
	if !enabled {
		return domain.ErrIntegrationDisabled
	}
	return nil
}

// ---------------------------------------------------------------------------
// AutoPromotionGuard
// ---------------------------------------------------------------------------

// autoPromotionPolicy is the JSON schema for the ERP auto-promotion policy.
type autoPromotionPolicy struct {
	AllowAutoPromotion bool `json:"allow_auto_promotion"`
}

// AutoPromotionGuard gates automatic promotion of reconciled ERP records behind
// a tenant-scoped policy.
type AutoPromotionGuard struct {
	resolver    *policy_resolver.Resolver
	environment string
}

// NewAutoPromotionGuard constructs an AutoPromotionGuard.
func NewAutoPromotionGuard(resolver *policy_resolver.Resolver, environment string) *AutoPromotionGuard {
	return &AutoPromotionGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

// CheckAutoPromotion returns domain.ErrAutoPromotionDisabled if the auto-promotion
// policy does not permit automatic promotion for the given tenant.
func (g *AutoPromotionGuard) CheckAutoPromotion(_ context.Context, tenantID string) error {
	raw, found, err := g.resolver.Resolve(bootstrap.ERPAutoPromotionKey, policy_resolver.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
		Module:      "erp_integrations",
	})
	if err != nil {
		return fmt.Errorf("resolve erp auto-promotion policy: %w", err)
	}
	if !found {
		return domain.ErrAutoPromotionDisabled
	}

	var policy autoPromotionPolicy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return fmt.Errorf("decode erp auto-promotion policy: %w", err)
	}
	if !policy.AllowAutoPromotion {
		return domain.ErrAutoPromotionDisabled
	}
	return nil
}
