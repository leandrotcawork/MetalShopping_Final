package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
)

type AdminRoleAssignmentPolicy struct {
	AllowAdminRoleAssignment bool `json:"allow_admin_role_assignment"`
}

type AdminPolicyGuard struct {
	resolver    *policy_resolver.Resolver
	environment string
}

func NewAdminPolicyGuard(resolver *policy_resolver.Resolver, environment string) *AdminPolicyGuard {
	return &AdminPolicyGuard{
		resolver:    resolver,
		environment: strings.TrimSpace(environment),
	}
}

func (g *AdminPolicyGuard) ValidateAdminRoleAssignment(ctx context.Context, tenantID string, role domain.Role) error {
	if g == nil || g.resolver == nil || role != domain.RoleAdmin {
		return nil
	}

	raw, found, err := g.resolver.Resolve(bootstrap.IAMAdminRoleAssignmentKey, policy_resolver.ResolutionContext{
		Environment: g.environment,
		TenantID:    strings.TrimSpace(tenantID),
		Module:      "iam",
	})
	if err != nil {
		return fmt.Errorf("resolve iam admin role assignment policy: %w", err)
	}
	if !found {
		return domain.ErrAdminRoleAssignmentDisabled
	}

	var policy AdminRoleAssignmentPolicy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return fmt.Errorf("decode iam admin role assignment policy: %w", err)
	}
	if !policy.AllowAdminRoleAssignment {
		return domain.ErrAdminRoleAssignmentDisabled
	}
	return nil
}
