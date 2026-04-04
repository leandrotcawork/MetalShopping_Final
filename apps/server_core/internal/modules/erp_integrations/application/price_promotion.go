package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

// PricePromotion loads normalized ERP price staging data and translates it
// into a canonical pricing write input.
type PricePromotion struct {
	stagingRepo   ports.StagingReader
	runRepo       ports.RunRepository
	productLookup ports.ProductLookup
	writer        ports.PriceWriter
}

// NewPricePromotion constructs a price promotion service.
func NewPricePromotion(
	stagingRepo ports.StagingReader,
	runRepo ports.RunRepository,
	productLookup ports.ProductLookup,
	writer ports.PriceWriter,
) *PricePromotion {
	return &PricePromotion{
		stagingRepo:   stagingRepo,
		runRepo:       runRepo,
		productLookup: productLookup,
		writer:        writer,
	}
}

// PromotePrice promotes a reconciled ERP price into the canonical pricing
// module.
func (p *PricePromotion) PromotePrice(ctx context.Context, result *domain.ReconciliationResult) (string, error) {
	if p == nil || p.stagingRepo == nil || p.runRepo == nil || p.productLookup == nil || p.writer == nil {
		return "", ErrPricePromotionNotConfigured
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
	if staging.EntityType != domain.EntityTypePrices {
		return "", fmt.Errorf("staging record %s has unsupported entity type %s", staging.StagingID, staging.EntityType)
	}
	if !strings.EqualFold(strings.TrimSpace(staging.ValidationStatus), "valid") {
		return "", fmt.Errorf("staging record %s is not valid", staging.StagingID)
	}

	input, err := buildPricePromotionInput(staging)
	if err != nil {
		return "", err
	}

	productID, found, err := p.productLookup.FindProductIDBySKU(ctx, result.TenantID, input.ProductSourceID)
	if err != nil {
		return "", fmt.Errorf("lookup canonical product %s for price promotion: %w", input.ProductSourceID, err)
	}
	if !found || strings.TrimSpace(productID) == "" {
		return "", ErrRelatedProductNotPromoted
	}

	run, err := p.runRepo.Get(ctx, result.TenantID, result.RunID)
	if err != nil {
		return "", fmt.Errorf("load erp sync run %s: %w", result.RunID, err)
	}
	if run == nil {
		return "", fmt.Errorf("load erp sync run %s returned no data", result.RunID)
	}

	traceID := promotionTraceID(result)
	sourceSystem := sourceSystemForRun(run)
	return p.writer.PromotePrice(ctx, traceID, result, run, ports.PricePromotionInput{
		ProductID:             productID,
		SourceSystem:          sourceSystem,
		SourceTableID:         input.SourceTableID,
		SourceTableCode:       input.SourceTableCode,
		SourceTableName:       input.SourceTableName,
		CurrencyCode:          input.CurrencyCode,
		PriceAmount:           input.PriceAmount,
		ReplacementCostAmount: input.ReplacementCostAmount,
		AverageCostAmount:     input.AverageCostAmount,
		PricingStatus:         input.PricingStatus,
		EffectiveFrom:         input.EffectiveFrom,
		EffectiveTo:           input.EffectiveTo,
		OriginType:            input.OriginType,
		OriginRef:             input.OriginRef,
		ReasonCode:            input.ReasonCode,
		UpdatedBy:             input.UpdatedBy,
	})
}

type pricePromotionStagingInput struct {
	ProductSourceID       string
	SourceTableID         string
	SourceTableCode       string
	SourceTableName       string
	CurrencyCode          string
	PriceAmount           float64
	ReplacementCostAmount float64
	AverageCostAmount     *float64
	PricingStatus         string
	EffectiveFrom         time.Time
	EffectiveTo           *time.Time
	OriginType            string
	OriginRef             string
	ReasonCode            string
	UpdatedBy             string
}

func buildPricePromotionInput(staging *domain.StagingRecord) (pricePromotionStagingInput, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(staging.NormalizedJSON, &payload); err != nil {
		return pricePromotionStagingInput{}, fmt.Errorf("unmarshal normalized price staging %s: %w", staging.StagingID, err)
	}

	productSourceID := firstNonBlank(readStringField(payload, "product_source_id"), readStringField(payload, "CODPROD"))
	if productSourceID == "" {
		return pricePromotionStagingInput{}, fmt.Errorf("normalized price staging %s is missing product source id", staging.StagingID)
	}

	sourceTableID := firstNonBlank(readStringField(payload, "source_table_id"), readStringField(payload, "NUTAB"))
	sourceTableCode := firstNonBlank(readStringField(payload, "source_table_code"), readStringField(payload, "CODTAB"))
	if sourceTableCode == "" {
		return pricePromotionStagingInput{}, fmt.Errorf("normalized price staging %s is missing source table code", staging.StagingID)
	}

	effectiveFrom, ok := readTimeField(payload, "effective_at")
	if !ok || effectiveFrom == nil || effectiveFrom.IsZero() {
		fallback := staging.NormalizedAt
		effectiveFrom = &fallback
	}

	priceAmount, ok := readFloatFieldValue(payload, "sale_price")
	if !ok {
		priceAmount, ok = readFloatFieldValue(payload, "VLRVENDA")
		if !ok {
			priceAmount = 0
		}
	}

	return pricePromotionStagingInput{
		ProductSourceID:       productSourceID,
		SourceTableID:         sourceTableID,
		SourceTableCode:       sourceTableCode,
		SourceTableName:       firstNonBlank(readStringField(payload, "source_table_name"), readStringField(payload, "NOMETAB")),
		CurrencyCode:          firstNonBlank(readStringField(payload, "currency_code"), "BRL"),
		PriceAmount:           priceAmount,
		ReplacementCostAmount: readFloatField(payload, "replacement_cost_amount"),
		AverageCostAmount:     readOptionalFloatField(payload, "average_cost_amount"),
		PricingStatus:         firstNonBlank(readStringField(payload, "pricing_status"), "active"),
		EffectiveFrom:         effectiveFrom.UTC(),
		EffectiveTo:           nil,
		OriginType:            firstNonBlank(readStringField(payload, "origin_type"), "import"),
		OriginRef:             firstNonBlank(readStringField(payload, "origin_ref"), staging.SourceID),
		ReasonCode:            firstNonBlank(readStringField(payload, "reason_code"), "ERP_PRICE_PROMOTED"),
		UpdatedBy:             firstNonBlank(readStringField(payload, "updated_by"), "erp_integrations"),
	}, nil
}

func readOptionalFloatField(payload map[string]json.RawMessage, key string) *float64 {
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	if strings.EqualFold(string(raw), "null") {
		return nil
	}
	value, ok := readFloatFieldValue(payload, key)
	if !ok {
		return nil
	}
	return &value
}

func readFloatFieldValue(payload map[string]json.RawMessage, key string) (float64, bool) {
	raw, ok := payload[key]
	if !ok {
		return 0, false
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case string:
		parsed, err := parseFloatString(typed)
		if err != nil {
			return 0, false
		}
		return parsed, true
	case bool:
		if typed {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func sourceSystemForRun(run *domain.SyncRun) string {
	if run == nil {
		return "sankhya"
	}
	sourceSystem := strings.TrimSpace(string(run.ConnectorType))
	if sourceSystem == "" {
		return "sankhya"
	}
	return sourceSystem
}
