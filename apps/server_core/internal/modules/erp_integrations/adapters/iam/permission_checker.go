package iam

import (
	"context"

	iamdomain "metalshopping/server_core/internal/modules/iam/domain"
)

// authorizationService matches the subset of iamapp.AuthorizationService used here.
type authorizationService interface {
	HasPermission(ctx context.Context, userID string, permission iamdomain.Permission) (bool, error)
}

// PermissionChecker adapts the IAM authorization service to the
// erp_integrations ports.PermissionChecker interface.
//
// CanManageIntegrations returns true when the principal holds
// PermIntegrationsWrite. The tenantID parameter is accepted for interface
// compliance but permission resolution is currently user-scoped (the IAM
// module does not yet support per-tenant permission overrides).
type PermissionChecker struct {
	authz authorizationService
}

// NewPermissionChecker constructs a PermissionChecker wrapping the given IAM
// authorization service.
func NewPermissionChecker(authz authorizationService) *PermissionChecker {
	return &PermissionChecker{authz: authz}
}

// CanManageIntegrations returns true if principalID holds the
// integrations:write permission.
func (c *PermissionChecker) CanManageIntegrations(ctx context.Context, _ string, principalID string) (bool, error) {
	return c.authz.HasPermission(ctx, principalID, iamdomain.PermIntegrationsWrite)
}
