package threshold_resolver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/platform/governance/config_registry"
)

const loadThresholdValuesSQL = `
SELECT threshold_name, scope_type, scope_key, threshold_value
FROM governance_threshold_values
WHERE is_active = TRUE
  AND effective_from <= NOW()
  AND (effective_to IS NULL OR effective_to > NOW())
ORDER BY threshold_name ASC, effective_from DESC
`

type valueRecord struct {
	ThresholdName string
	ScopeType     config_registry.Scope
	ScopeKey      string
	Value         float64
}

func NewPostgresResolver(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (*Resolver, error) {
	values, err := loadScopeValues(ctx, db, registry)
	if err != nil {
		return nil, err
	}
	return NewResolver(registry, values), nil
}

func loadScopeValues(ctx context.Context, db *sql.DB, registry *config_registry.Registry) (map[string]ScopeValues, error) {
	rows, err := db.QueryContext(ctx, loadThresholdValuesSQL)
	if err != nil {
		return nil, fmt.Errorf("load governance thresholds: %w", err)
	}
	defer rows.Close()

	records := make([]valueRecord, 0, 16)
	for rows.Next() {
		var record valueRecord
		if err := rows.Scan(&record.ThresholdName, &record.ScopeType, &record.ScopeKey, &record.Value); err != nil {
			return nil, fmt.Errorf("scan governance threshold: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate governance thresholds: %w", err)
	}

	values := make(map[string]ScopeValues, len(records))
	for _, record := range records {
		key := strings.TrimSpace(record.ThresholdName)
		entry, ok := registry.Get(key)
		if !ok {
			return nil, fmt.Errorf("threshold not registered in governance registry: %s", key)
		}
		if entry.Kind != config_registry.ArtifactThreshold {
			return nil, fmt.Errorf("governance entry is not a threshold: %s", key)
		}

		scopeValues := values[key]
		switch record.ScopeType {
		case config_registry.ScopeGlobal:
			value := record.Value
			scopeValues.Global = &value
		case config_registry.ScopeEnvironment:
			if scopeValues.Environment == nil {
				scopeValues.Environment = map[string]float64{}
			}
			scopeValues.Environment[strings.TrimSpace(record.ScopeKey)] = record.Value
		case config_registry.ScopeTenant:
			if scopeValues.Tenant == nil {
				scopeValues.Tenant = map[string]float64{}
			}
			scopeValues.Tenant[strings.TrimSpace(record.ScopeKey)] = record.Value
		case config_registry.ScopeModule:
			if scopeValues.Module == nil {
				scopeValues.Module = map[string]float64{}
			}
			scopeValues.Module[strings.TrimSpace(record.ScopeKey)] = record.Value
		case config_registry.ScopeEntityProfile:
			if scopeValues.EntityProfile == nil {
				scopeValues.EntityProfile = map[string]float64{}
			}
			scopeValues.EntityProfile[strings.TrimSpace(record.ScopeKey)] = record.Value
		default:
			return nil, fmt.Errorf("unsupported threshold scope type: %s", record.ScopeType)
		}
		values[key] = scopeValues
	}

	return values, nil
}
