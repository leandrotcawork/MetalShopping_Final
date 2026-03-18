package bootstrap

import (
	"metalshopping/server_core/internal/platform/governance/config_registry"
)

const (
	CatalogProductCreationEnabledKey     = "catalog.product_creation_enabled"
	CatalogMaxDescriptionLengthKey       = "catalog.max_description_length"
	PricingManualPriceOverrideKey        = "pricing.manual_price_override"
	IAMAdminRoleAssignmentKey            = "iam.admin_role_assignment"
	AuthWebSessionEnabledKey             = "auth.web_session_enabled"
	AuthSessionIdleTimeoutMinutesKey     = "auth.session_idle_timeout_minutes"
	AuthSessionAbsoluteTimeoutMinutesKey = "auth.session_absolute_timeout_minutes"
)

func NewRegistry() *config_registry.Registry {
	registry := config_registry.NewRegistry()
	registry.MustRegister(config_registry.Entry{
		Key:            AuthWebSessionEnabledKey,
		Kind:           config_registry.ArtifactFeatureFlag,
		BoundedContext: "auth",
		ValueType:      config_registry.ValueTypeBool,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeFeatureTarget,
		},
		Description: "Controls whether the backend-owned authenticated web session surface is enabled for the resolved runtime scope.",
	})
	registry.MustRegister(config_registry.Entry{
		Key:            AuthSessionIdleTimeoutMinutesKey,
		Kind:           config_registry.ArtifactThreshold,
		BoundedContext: "auth",
		ValueType:      config_registry.ValueTypeNumber,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
			config_registry.ScopeEntityProfile,
		},
		Description: "Defines the idle timeout in minutes for backend-owned authenticated web sessions.",
	})
	registry.MustRegister(config_registry.Entry{
		Key:            AuthSessionAbsoluteTimeoutMinutesKey,
		Kind:           config_registry.ArtifactThreshold,
		BoundedContext: "auth",
		ValueType:      config_registry.ValueTypeNumber,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
			config_registry.ScopeEntityProfile,
		},
		Description: "Defines the absolute timeout in minutes for backend-owned authenticated web sessions.",
	})
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
		Key:            PricingManualPriceOverrideKey,
		Kind:           config_registry.ArtifactPolicy,
		BoundedContext: "pricing",
		ValueType:      config_registry.ValueTypeJSON,
		Scopes: []config_registry.Scope{
			config_registry.ScopeGlobal,
			config_registry.ScopeEnvironment,
			config_registry.ScopeTenant,
			config_registry.ScopeModule,
		},
		Description: "Defines the policy for manual product price override decisions.",
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
