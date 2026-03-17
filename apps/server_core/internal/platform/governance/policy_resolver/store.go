package policy_resolver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

const loadPolicyValuesSQL = `
SELECT policy_name, scope_type, scope_key, policy_json
FROM governance_policy_values
WHERE is_active = TRUE
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY policy_name ASC, effective_from DESC
`

type valueRecord struct {
	PolicyName string
	ScopeType  config_registry.Scope
	ScopeKey   string
	Value      json.RawMessage
}

func NewPostgresResolver(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (*Resolver, error) {
	values, err := loadScopeValues(ctx, db, registry)
	if err != nil {
		return nil, err
	}
	return NewResolver(registry, values), nil
}

func loadScopeValues(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (map[string]ScopeValues, error) {
	rows, err := db.QueryContext(ctx, loadPolicyValuesSQL)
	if err != nil {
		return nil, fmt.Errorf("load governance policies: %w", err)
	}
	defer rows.Close()

	records := make([]valueRecord, 0, 16)
	for rows.Next() {
		var record valueRecord
		if err := rows.Scan(&record.PolicyName, &record.ScopeType, &record.ScopeKey, &record.Value); err != nil {
			return nil, fmt.Errorf("scan governance policy: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate governance policies: %w", err)
	}

	values := make(map[string]ScopeValues, len(records))
	for _, record := range records {
		key := strings.TrimSpace(record.PolicyName)
		entry, ok := registry.Get(key)
		if !ok {
			return nil, fmt.Errorf("policy not registered in governance registry: %s", key)
		}
		if entry.Kind != config_registry.ArtifactPolicy {
			return nil, fmt.Errorf("governance entry is not a policy: %s", key)
		}

		scopeValues := values[key]
		switch record.ScopeType {
		case config_registry.ScopeGlobal:
			scopeValues.Global = record.Value
		case config_registry.ScopeEnvironment:
			if scopeValues.Environment == nil {
				scopeValues.Environment = map[string]json.RawMessage{}
			}
			scopeValues.Environment[strings.TrimSpace(record.ScopeKey)] = record.Value
		case config_registry.ScopeTenant:
			if scopeValues.Tenant == nil {
				scopeValues.Tenant = map[string]json.RawMessage{}
			}
			scopeValues.Tenant[strings.TrimSpace(record.ScopeKey)] = record.Value
		case config_registry.ScopeModule:
			if scopeValues.Module == nil {
				scopeValues.Module = map[string]json.RawMessage{}
			}
			scopeValues.Module[strings.TrimSpace(record.ScopeKey)] = record.Value
		default:
			return nil, fmt.Errorf("unsupported policy scope type: %s", record.ScopeType)
		}
		values[key] = scopeValues
	}

	return values, nil
}
