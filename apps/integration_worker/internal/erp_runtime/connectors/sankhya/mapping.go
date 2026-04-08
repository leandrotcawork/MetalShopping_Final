package sankhya

import (
	"fmt"
	"strings"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
)

// FieldMapping maps a Sankhya source field to a MetalShopping canonical field name.
type FieldMapping struct {
	SourceField string
	TargetField string
	Required    bool
}

// entityMappings defines field-level mappings per entity.
var entityMappings = map[erp_runtime.EntityType][]FieldMapping{
	erp_runtime.EntityTypeProducts: {
		{SourceField: "CODPROD", TargetField: "erp_product_id", Required: true},
		{SourceField: "DESCRPROD", TargetField: "name", Required: true},
		{SourceField: "MARCA", TargetField: "brand_name", Required: false},
		{SourceField: "REFERENCIA", TargetField: "ean", Required: false},
		{SourceField: "REFFORN", TargetField: "manufacturer_reference", Required: false},
		{SourceField: "ATIVO", TargetField: "source_status", Required: false},
		{SourceField: "CODVOL", TargetField: "unit", Required: false},
		{SourceField: "CODGRUPOPROD", TargetField: "taxonomy_source_code", Required: false},
		{SourceField: "AD_STATUS", TargetField: "canonical_status_hint", Required: false},
		{SourceField: "AD_COMPETITIVO", TargetField: "competitive_flag", Required: false},
	},
	erp_runtime.EntityTypePrices: {
		{SourceField: "NUTAB", TargetField: "source_table_id", Required: true},
		{SourceField: "CODTAB", TargetField: "source_table_code", Required: true},
		{SourceField: "NOMETAB", TargetField: "source_table_name", Required: false},
		{SourceField: "DTVIGOR", TargetField: "effective_at", Required: false},
		{SourceField: "CODPROD", TargetField: "product_source_id", Required: true},
		{SourceField: "VLRVENDA", TargetField: "sale_price", Required: true},
	},
	erp_runtime.EntityTypeCosts: {
		{SourceField: "CODPROD", TargetField: "source_id", Required: true},
		{SourceField: "VLRCUSTO_REP", TargetField: "replacement_cost", Required: false},
		{SourceField: "VLRCUSTOMEDIO", TargetField: "average_cost", Required: false},
	},
	erp_runtime.EntityTypeInventory: {
		{SourceField: "CODPROD", TargetField: "product_source_id", Required: true},
		{SourceField: "CODEMP", TargetField: "company_code", Required: true},
		{SourceField: "CODLOCAL", TargetField: "location_code", Required: true},
		{SourceField: "ESTOQUE", TargetField: "raw_quantity", Required: true},
		{SourceField: "RESERVADO", TargetField: "reserved_quantity", Required: false},
		{SourceField: "RAW_AVAILABLE_POSITION", TargetField: "raw_available_position", Required: false},
	},
	erp_runtime.EntityTypeSales: {
		{SourceField: "NUNOTA", TargetField: "source_id", Required: true},
		{SourceField: "CODPARC", TargetField: "partner_source_id", Required: true},
		{SourceField: "DTNEG", TargetField: "transaction_date", Required: true},
		{SourceField: "VLRNOTA", TargetField: "total_amount", Required: true},
		{SourceField: "TIPMOV", TargetField: "movement_type", Required: false},
	},
	erp_runtime.EntityTypePurchases: {
		{SourceField: "NUNOTA", TargetField: "source_id", Required: true},
		{SourceField: "CODPARC", TargetField: "partner_source_id", Required: true},
		{SourceField: "DTNEG", TargetField: "transaction_date", Required: true},
		{SourceField: "VLRNOTA", TargetField: "total_amount", Required: true},
		{SourceField: "TIPMOV", TargetField: "movement_type", Required: false},
	},
	erp_runtime.EntityTypeCustomers: {
		{SourceField: "CODPARC", TargetField: "source_id", Required: true},
		{SourceField: "NOMEPARC", TargetField: "trade_name", Required: true},
		{SourceField: "CGC_CPF", TargetField: "tax_id", Required: false},
		{SourceField: "EMAIL", TargetField: "email", Required: false},
		{SourceField: "CLIENTE", TargetField: "is_customer", Required: false},
	},
	erp_runtime.EntityTypeSuppliers: {
		{SourceField: "CODPARC", TargetField: "source_id", Required: true},
		{SourceField: "NOMEPARC", TargetField: "trade_name", Required: true},
		{SourceField: "CGC_CPF", TargetField: "tax_id", Required: false},
		{SourceField: "EMAIL", TargetField: "email", Required: false},
		{SourceField: "FORNECEDOR", TargetField: "is_supplier", Required: false},
	},
}

// Mapper provides field mapping metadata for Sankhya entities.
type Mapper struct{}

func newMapper() *Mapper { return &Mapper{} }

// MappingsFor returns the field mappings for the given entity.
func (m *Mapper) MappingsFor(entity erp_runtime.EntityType) []FieldMapping {
	return entityMappings[entity]
}

// MapRow maps a source row into a Sankhya payload and stable source ID.
func (m *Mapper) MapRow(entity erp_runtime.EntityType, row dbsource.RowReader, sourceIDKeys []string) (map[string]any, string, error) {
	mappings := m.MappingsFor(entity)
	if len(mappings) == 0 {
		return nil, "", fmt.Errorf("sankhya mapper: no mappings for entity %q", entity)
	}

	payload := make(map[string]any, len(mappings))
	for _, field := range mappings {
		value, err := row.NullString(field.SourceField)
		if err != nil {
			if field.Required {
				return nil, "", fmt.Errorf("sankhya mapper: read required field %q: %w", field.SourceField, err)
			}
			continue
		}
		if value == nil {
			if field.Required {
				return nil, "", fmt.Errorf("sankhya mapper: required field %q is null", field.SourceField)
			}
			payload[field.SourceField] = nil
			continue
		}
		trimmed := strings.TrimSpace(*value)
		if field.Required && trimmed == "" {
			return nil, "", fmt.Errorf("sankhya mapper: required field %q is empty", field.SourceField)
		}
		payload[field.SourceField] = trimmed
	}

	sourceID, err := sourceIDFromPayload(payload, sourceIDKeys)
	if err != nil {
		return nil, "", err
	}
	return payload, sourceID, nil
}

func sourceIDFromPayload(payload map[string]any, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("sankhya mapper: source ID keys must not be empty")
	}

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value, ok := payload[key]
		if !ok || value == nil {
			return "", fmt.Errorf("sankhya mapper: missing source ID field %q", key)
		}
		got := strings.TrimSpace(fmt.Sprint(value))
		if got == "" || got == "<nil>" {
			return "", fmt.Errorf("sankhya mapper: empty source ID field %q", key)
		}
		parts = append(parts, got)
	}
	return strings.Join(parts, ":"), nil
}
