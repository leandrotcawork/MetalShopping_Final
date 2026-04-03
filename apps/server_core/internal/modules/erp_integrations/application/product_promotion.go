package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

const unsupportedPromotionEntityReasonCode = "ERP_PROMOTION_UNSUPPORTED_ENTITY_TYPE"

// ProductPromotion loads normalized product staging data and translates it into
// a canonical catalog write input.
type ProductPromotion struct {
	stagingRepo ports.StagingReader
	runRepo     ports.RunRepository
	writer      ports.ProductWriter
}

// NewProductPromotion constructs a product-only promotion service.
func NewProductPromotion(stagingRepo ports.StagingReader, runRepo ports.RunRepository, writer ports.ProductWriter) *ProductPromotion {
	return &ProductPromotion{
		stagingRepo: stagingRepo,
		runRepo:     runRepo,
		writer:      writer,
	}
}

// PromoteProduct promotes a reconciled ERP product into the canonical catalog.
func (p *ProductPromotion) PromoteProduct(ctx context.Context, result *domain.ReconciliationResult) (string, error) {
	if p == nil || p.stagingRepo == nil || p.runRepo == nil || p.writer == nil {
		return "", fmt.Errorf("product promotion is not configured")
	}
	if result == nil {
		return "", fmt.Errorf("reconciliation result is required")
	}

	staging, err := p.stagingRepo.GetStagingRecord(ctx, result.TenantID, result.StagingID)
	if err != nil {
		return "", fmt.Errorf("load erp staging record %s: %w", result.StagingID, err)
	}
	if staging == nil {
		return "", fmt.Errorf("load erp staging record %s returned no data", result.StagingID)
	}
	if staging.EntityType != domain.EntityTypeProducts {
		return "", fmt.Errorf("staging record %s has unsupported entity type %s", staging.StagingID, staging.EntityType)
	}
	if !strings.EqualFold(strings.TrimSpace(staging.ValidationStatus), "valid") {
		return "", fmt.Errorf("staging record %s is not valid", staging.StagingID)
	}

	input, err := buildProductPromotionInput(staging)
	if err != nil {
		return "", err
	}

	run, err := p.runRepo.Get(ctx, result.TenantID, result.RunID)
	if err != nil {
		return "", fmt.Errorf("load erp sync run %s: %w", result.RunID, err)
	}
	if run == nil {
		return "", fmt.Errorf("load erp sync run %s returned no data", result.RunID)
	}

	traceID := promotionTraceID(result)
	return p.writer.PromoteProduct(ctx, traceID, result, run, input)
}

func buildProductPromotionInput(staging *domain.StagingRecord) (ports.ProductPromotionInput, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(staging.NormalizedJSON, &payload); err != nil {
		return ports.ProductPromotionInput{}, fmt.Errorf("unmarshal normalized product staging %s: %w", staging.StagingID, err)
	}

	sku := firstNonBlank(
		readStringField(payload, "sku"),
		readStringField(payload, "pn_interno"),
	)
	if sku == "" {
		return ports.ProductPromotionInput{}, fmt.Errorf("normalized product staging %s is missing sku", staging.StagingID)
	}

	name := firstNonBlank(
		readStringField(payload, "name"),
		readStringField(payload, "descricao"),
		readStringField(payload, "reference"),
		readStringField(payload, "ean"),
		sku,
	)
	if name == "" {
		return ports.ProductPromotionInput{}, fmt.Errorf("normalized product staging %s is missing name", staging.StagingID)
	}

	status, err := resolveProductStatus(payload)
	if err != nil {
		return ports.ProductPromotionInput{}, fmt.Errorf("normalize product status for staging %s: %w", staging.StagingID, err)
	}

	identifiers, err := buildProductIdentifiers(payload, sku)
	if err != nil {
		return ports.ProductPromotionInput{}, err
	}

	return ports.ProductPromotionInput{
		SKU:                   sku,
		Name:                  name,
		Description:           firstNonBlank(readStringField(payload, "description"), readStringField(payload, "descricao")),
		BrandName:             firstNonBlank(readStringField(payload, "brand_name"), readStringField(payload, "marca")),
		StockProfileCode:      firstNonBlank(readStringField(payload, "stock_profile_code"), readStringField(payload, "tipo_estoque")),
		PrimaryTaxonomyNodeID: firstNonBlank(readStringField(payload, "primary_taxonomy_node_id"), readStringField(payload, "taxonomy_node_id")),
		Status:                status,
		Identifiers:           identifiers,
	}, nil
}

func buildProductIdentifiers(payload map[string]json.RawMessage, sku string) ([]ports.ProductPromotionIdentifierInput, error) {
	identifiers := make([]ports.ProductPromotionIdentifierInput, 0, 4)
	seen := map[string]struct{}{}
	add := func(identifierType, identifierValue, sourceSystem string, primary bool) {
		identifierType = strings.TrimSpace(identifierType)
		identifierValue = strings.TrimSpace(identifierValue)
		if identifierType == "" || identifierValue == "" {
			return
		}
		key := strings.ToLower(identifierType) + "|" + strings.ToLower(identifierValue)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		identifiers = append(identifiers, ports.ProductPromotionIdentifierInput{
			IdentifierType:  identifierType,
			IdentifierValue: identifierValue,
			SourceSystem:    strings.TrimSpace(sourceSystem),
			IsPrimary:       primary,
		})
	}

	add("pn_interno", sku, "erp", true)
	add("reference", readStringField(payload, "reference"), "erp", false)
	add("ean", readStringField(payload, "ean"), "erp", false)

	rawIdentifiers, ok := payload["identifiers"]
	if !ok || len(rawIdentifiers) == 0 || strings.EqualFold(string(rawIdentifiers), "null") {
		return identifiers, nil
	}

	var stagedIdentifiers []map[string]json.RawMessage
	if err := json.Unmarshal(rawIdentifiers, &stagedIdentifiers); err != nil {
		return nil, fmt.Errorf("unmarshal product identifiers: %w", err)
	}
	for _, stagedIdentifier := range stagedIdentifiers {
		add(
			readStringField(stagedIdentifier, "identifier_type"),
			readStringField(stagedIdentifier, "identifier_value"),
			firstNonBlank(readStringField(stagedIdentifier, "source_system"), "erp"),
			readBoolField(stagedIdentifier, "is_primary"),
		)
	}

	return identifiers, nil
}

func resolveProductStatus(payload map[string]json.RawMessage) (string, error) {
	status := strings.ToLower(strings.TrimSpace(readStringField(payload, "status")))
	if status != "" {
		switch status {
		case "active", "inactive":
			return status, nil
		default:
			return "", fmt.Errorf("invalid product status %q", status)
		}
	}

	if value, ok := readBoolFieldWithOK(payload, "ativo"); ok {
		if value {
			return "active", nil
		}
		return "inactive", nil
	}

	return "active", nil
}

func promotionTraceID(result *domain.ReconciliationResult) string {
	if result == nil {
		return "erp-promotion"
	}
	reconciliationID := strings.TrimSpace(result.ReconciliationID)
	if reconciliationID == "" {
		return "erp-promotion"
	}
	return "erp-promotion:" + reconciliationID
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func readStringField(payload map[string]json.RawMessage, key string) string {
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	value, ok := decodeScalarString(raw)
	if !ok {
		return ""
	}
	return value
}

func readBoolField(payload map[string]json.RawMessage, key string) bool {
	value, _ := readBoolFieldWithOK(payload, key)
	return value
}

func readBoolFieldWithOK(payload map[string]json.RawMessage, key string) (bool, bool) {
	raw, ok := payload[key]
	if !ok {
		return false, false
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, false
	}
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "y", "active", "ativo":
			return true, true
		case "false", "0", "no", "n", "inactive", "inativo":
			return false, true
		default:
			return false, false
		}
	case float64:
		return typed != 0, true
	default:
		return false, false
	}
}

func decodeScalarString(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	case float64:
		return strings.TrimSpace(strings.TrimSuffix(fmt.Sprintf("%.0f", typed), ".0")), true
	case nil:
		return "", false
	default:
		return strings.TrimSpace(fmt.Sprint(typed)), true
	}
}
