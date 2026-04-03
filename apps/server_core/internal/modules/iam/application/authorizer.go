package application

import "metalshopping/server_core/internal/modules/iam/domain"

type StaticAuthorizer struct {
	policy map[domain.Role]map[domain.Permission]bool
}

func NewStaticAuthorizer() *StaticAuthorizer {
	return &StaticAuthorizer{
		policy: map[domain.Role]map[domain.Permission]bool{
			domain.RoleAdmin: {
				domain.PermIAMManageRoles:          true,
				domain.PermIAMReadRoles:            true,
				domain.PermCatalogRead:             true,
				domain.PermCatalogWrite:            true,
				domain.PermInventoryRead:           true,
				domain.PermInventoryWrite:          true,
				domain.PermPricingRead:             true,
				domain.PermPricingWrite:            true,
				domain.PermSalesRead:               true,
				domain.PermCRMRead:                 true,
				domain.PermCRMWrite:                true,
				domain.PermAutomationRead:          true,
				domain.PermAutomationManage:        true,
				domain.PermAnalyticsServingRead:    true,
				domain.PermMarketIntelligenceRead:  true,
				domain.PermIntegrationsRead:        true,
				domain.PermIntegrationsWrite:       true,
			},
			domain.RoleTenantAdmin: {
				domain.PermIAMManageRoles:          true,
				domain.PermIAMReadRoles:            true,
				domain.PermCatalogRead:             true,
				domain.PermInventoryRead:           true,
				domain.PermInventoryWrite:          true,
				domain.PermPricingRead:             true,
				domain.PermSalesRead:               true,
				domain.PermCRMRead:                 true,
				domain.PermAutomationRead:          true,
				domain.PermAnalyticsServingRead:    true,
				domain.PermMarketIntelligenceRead:  true,
				domain.PermIntegrationsRead:        true,
				domain.PermIntegrationsWrite:       true,
			},
			domain.RoleCatalogManager: {
				domain.PermCatalogRead:            true,
				domain.PermCatalogWrite:           true,
				domain.PermInventoryRead:          true,
				domain.PermAnalyticsServingRead:   true,
				domain.PermMarketIntelligenceRead: true,
			},
			domain.RoleInventoryManager: {
				domain.PermCatalogRead:            true,
				domain.PermInventoryRead:          true,
				domain.PermInventoryWrite:         true,
				domain.PermAnalyticsServingRead:   true,
				domain.PermMarketIntelligenceRead: true,
			},
			domain.RolePricingManager: {
				domain.PermCatalogRead:            true,
				domain.PermInventoryRead:          true,
				domain.PermPricingRead:            true,
				domain.PermPricingWrite:           true,
				domain.PermAnalyticsServingRead:   true,
				domain.PermMarketIntelligenceRead: true,
			},
			domain.RoleSalesManager: {
				domain.PermCatalogRead:          true,
				domain.PermInventoryRead:        true,
				domain.PermSalesRead:            true,
				domain.PermCRMRead:              true,
				domain.PermCRMWrite:             true,
				domain.PermAnalyticsServingRead: true,
			},
			domain.RoleAnalyst: {
				domain.PermCatalogRead:            true,
				domain.PermInventoryRead:          true,
				domain.PermPricingRead:            true,
				domain.PermSalesRead:              true,
				domain.PermCRMRead:                true,
				domain.PermAutomationRead:         true,
				domain.PermAnalyticsServingRead:   true,
				domain.PermMarketIntelligenceRead: true,
			},
			domain.RoleAutomationOwner: {
				domain.PermCatalogRead:          true,
				domain.PermInventoryRead:        true,
				domain.PermAutomationRead:       true,
				domain.PermAutomationManage:     true,
				domain.PermAnalyticsServingRead: true,
			},
			domain.RoleViewer: {
				domain.PermCatalogRead:            true,
				domain.PermInventoryRead:          true,
				domain.PermPricingRead:            true,
				domain.PermSalesRead:              true,
				domain.PermCRMRead:                true,
				domain.PermAutomationRead:         true,
				domain.PermAnalyticsServingRead:   true,
				domain.PermMarketIntelligenceRead: true,
			},
		},
	}
}

func (a *StaticAuthorizer) Can(role domain.Role, permission domain.Permission) bool {
	rolePolicy, ok := a.policy[role]
	if !ok {
		return false
	}
	return rolePolicy[permission]
}
