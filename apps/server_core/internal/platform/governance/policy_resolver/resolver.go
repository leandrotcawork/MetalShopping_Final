package policy_resolver

import (
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

type ScopeValues struct {
	Global      json.RawMessage
	Environment map[string]json.RawMessage
	Tenant      map[string]json.RawMessage
	Module      map[string]json.RawMessage
}

type ResolutionContext struct {
	Environment string
	TenantID    string
	Module      string
}

type Resolver struct {
	registry *config_registry.Registry
	values   map[string]ScopeValues
}

func NewResolver(registry *config_registry.Registry, values map[string]ScopeValues) *Resolver {
	return &Resolver{registry: registry, values: values}
}

func (r *Resolver) Resolve(key string, ctx ResolutionContext) (json.RawMessage, bool, error) {
	entry, ok := r.registry.Get(key)
	if !ok {
		return nil, false, fmt.Errorf("policy not registered: %s", key)
	}
	if entry.Kind != config_registry.ArtifactPolicy {
		return nil, false, fmt.Errorf("governance entry is not a policy: %s", key)
	}

	scopeValues, ok := r.values[strings.TrimSpace(key)]
	if !ok {
		return nil, false, nil
	}
	if value, ok := lookup(scopeValues.Module, ctx.Module); ok {
		return value, true, nil
	}
	if value, ok := lookup(scopeValues.Tenant, ctx.TenantID); ok {
		return value, true, nil
	}
	if value, ok := lookup(scopeValues.Environment, ctx.Environment); ok {
		return value, true, nil
	}
	if len(scopeValues.Global) > 0 {
		return scopeValues.Global, true, nil
	}

	return nil, false, nil
}

func lookup(values map[string]json.RawMessage, key string) (json.RawMessage, bool) {
	if len(values) == 0 {
		return nil, false
	}
	value, ok := values[strings.TrimSpace(key)]
	return value, ok && len(value) > 0
}
