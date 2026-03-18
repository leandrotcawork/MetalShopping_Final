package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Role string

const (
	RoleAdmin            Role = "admin"
	RoleTenantAdmin      Role = "tenant_admin"
	RoleCatalogManager   Role = "catalog_manager"
	RoleInventoryManager Role = "inventory_manager"
	RolePricingManager   Role = "pricing_manager"
	RoleSalesManager     Role = "sales_manager"
	RoleAnalyst          Role = "analyst"
	RoleAutomationOwner  Role = "automation_owner"
	RoleViewer           Role = "viewer"
)

type Permission string

const (
	PermIAMManageRoles         Permission = "iam:manage_roles"
	PermIAMReadRoles           Permission = "iam:read_roles"
	PermCatalogRead            Permission = "catalog:read"
	PermCatalogWrite           Permission = "catalog:write"
	PermInventoryRead          Permission = "inventory:read"
	PermInventoryWrite         Permission = "inventory:write"
	PermPricingRead            Permission = "pricing:read"
	PermPricingWrite           Permission = "pricing:write"
	PermSalesRead              Permission = "sales:read"
	PermCRMRead                Permission = "crm:read"
	PermCRMWrite               Permission = "crm:write"
	PermAutomationRead         Permission = "automation:read"
	PermAutomationManage       Permission = "automation:manage"
	PermAnalyticsServingRead   Permission = "analytics_serving:read"
	PermMarketIntelligenceRead Permission = "market_intelligence:read"
)

type RoleAssignment struct {
	UserID      string
	DisplayName string
	Role        Role
	AssignedBy  string
	AssignedAt  time.Time
}

type Authorizer interface {
	Can(role Role, permission Permission) bool
}

var allowedRoles = []Role{
	RoleAdmin,
	RoleTenantAdmin,
	RoleCatalogManager,
	RoleInventoryManager,
	RolePricingManager,
	RoleSalesManager,
	RoleAnalyst,
	RoleAutomationOwner,
	RoleViewer,
}

func ParseRole(raw string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(raw)))
	if !role.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidRole, raw)
	}
	return role, nil
}

func (r Role) IsValid() bool {
	return slices.Contains(allowedRoles, r)
}

func (a RoleAssignment) Validate() error {
	if strings.TrimSpace(a.UserID) == "" {
		return ErrUserIDRequired
	}
	if !a.Role.IsValid() {
		return ErrInvalidRole
	}
	return nil
}
