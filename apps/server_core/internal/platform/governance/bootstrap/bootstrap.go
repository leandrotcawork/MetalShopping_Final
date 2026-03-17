package bootstrap

import (
	"metalshopping/server_core/internal/platform/governance/config_registry"
)

const (
	CatalogProductCreationEnabledKey = "catalog.product_creation_enabled"
	CatalogMaxDescriptionLengthKey   = "catalog.max_description_length"
	PricingDefaultMarginFloorKey     = "pricing.default_margin_floor"
	IAMAdminRoleAssignmentKey        = "iam.admin_role_assignment"
)

func NewRegistry() *config_registry.Registry {
	registry := config_registry.NewRegistry()
	registry.MustRegister(config_registry.Entry{
		Key:            CatalogProductCreationEnabledKey,
		Kind:           config_registry.ArtifactFeatureFlag,
		BoundedContext: "catalog",
		ValueType:      config_registry.ValueTypeBool,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeFeatureTarget,
		},
		Description: "Controls whether catalog product creation is enabled for the resolved runtime scope.",
	})
	registry.MustRegister(config_registry.Entry{
		Key:            CatalogMaxDescriptionLengthKey,
		Kind:           config_registry.ArtifactThreshold,
		BoundedContext: "catalog",
		ValueType:      config_registry.ValueTypeNumber,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
			config_registry.ScopeEntityProfile,
		},
		Description: "Defines the maximum product description length accepted by catalog writes for the resolved runtime scope.",
	})
	registry.MustRegister(config_registry.Entry{
		Key:            PricingDefaultMarginFloorKey,
		Kind:           config_registry.ArtifactThreshold,
		BoundedContext: "pricing",
		ValueType:      config_registry.ValueTypeNumber,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
		},
		Description: "Defines the default pricing margin floor for the resolved runtime scope.",
	})
	registry.MustRegister(config_registry.Entry{
		Key:            IAMAdminRoleAssignmentKey,
		Kind:           config_registry.ArtifactPolicy,
		BoundedContext: "iam",
		ValueType:      config_registry.ValueTypeJSON,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
		},
		Description: "Defines the policy for administrative IAM role assignment decisions.",
	})
	return registry
}
