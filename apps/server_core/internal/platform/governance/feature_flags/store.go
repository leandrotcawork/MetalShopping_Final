package feature_flags

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

const loadFeatureFlagValuesSQL = `
SELECT flag_name, scope_type, scope_key, flag_value
FROM governance_feature_flag_values
WHERE is_active = TRUE
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY flag_name ASC, effective_from DESC
`

type valueRecord struct {
	FlagName  string
	ScopeType config_registry.Scope
	ScopeKey  string
	Value     bool
}

type TestOnlyValueRecord = valueRecord

func NewPostgresResolver(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (*Resolver, error) {
	values, err := loadScopeValues(ctx, db, registry)
	if err != nil {
		return nil, err
	}
	return NewResolver(registry, values), nil
}

func loadScopeValues(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (map[string]ScopeValues, error) {
	rows, err := db.QueryContext(ctx, loadFeatureFlagValuesSQL)
	if err != nil {
		return nil, fmt.Errorf("load governance feature flags: %w", err)
	}
	defer rows.Close()

	records := make([]valueRecord, 0, 16)
	for rows.Next() {
		var record valueRecord
		if err := rows.Scan(&record.FlagName, &record.ScopeType, &record.ScopeKey, &record.Value); err != nil {
			return nil, fmt.Errorf("scan governance feature flag: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate governance feature flags: %w", err)
	}

	values, err := scopeValuesFromRecords(registry, records)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func scopeValuesFromRecords(registry *config_registry.Registry, records []valueRecord) (map[string]ScopeValues, error) {
	values := make(map[string]ScopeValues, len(records))
	for _, record := range records {
		flagName := strings.TrimSpace(record.FlagName)
		entry, ok := registry.Get(flagName)
		if !ok {
			return nil, fmt.Errorf("feature flag not registered in governance registry: %s", flagName)
		}
		if entry.Kind != config_registry.ArtifactFeatureFlag {
			return nil, fmt.Errorf("governance entry is not a feature flag: %s", flagName)
		}

		scopeKey := strings.TrimSpace(record.ScopeKey)
		scopeValues := values[flagName]
		switch record.ScopeType {
		case config_registry.ScopeGlobal:
			value := record.Value
			scopeValues.Global = &value
		case config_registry.ScopeEnvironment:
			if scopeValues.Environment == nil {
				scopeValues.Environment = map[string]bool{}
			}
			scopeValues.Environment[scopeKey] = record.Value
		case config_registry.ScopeTenant:
			if scopeValues.Tenant == nil {
				scopeValues.Tenant = map[string]bool{}
			}
			scopeValues.Tenant[scopeKey] = record.Value
		case config_registry.ScopeFeatureTarget:
			if scopeValues.FeatureTarget == nil {
				scopeValues.FeatureTarget = map[string]bool{}
			}
			scopeValues.FeatureTarget[scopeKey] = record.Value
		default:
			return nil, fmt.Errorf("unsupported feature flag scope type: %s", record.ScopeType)
		}
		values[flagName] = scopeValues
	}
	return values, nil
}

func TestOnlyScopeValuesFromRecords(registry *config_registry.Registry, records []TestOnlyValueRecord) (map[string]ScopeValues, error) {
	return scopeValuesFromRecords(registry, records)
}
