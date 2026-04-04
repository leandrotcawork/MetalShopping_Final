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

// InventoryPromotion loads normalized ERP inventory staging data and translates
// it into a canonical inventory write input.
type InventoryPromotion struct {
	stagingRepo   ports.StagingReader
	runRepo       ports.RunRepository
	productLookup ports.ProductLookup
	writer        ports.InventoryWriter
}

// NewInventoryPromotion constructs an inventory promotion service.
func NewInventoryPromotion(
	stagingRepo ports.StagingReader,
	runRepo ports.RunRepository,
	productLookup ports.ProductLookup,
	writer ports.InventoryWriter,
) *InventoryPromotion {
	return &InventoryPromotion{
		stagingRepo:   stagingRepo,
		runRepo:       runRepo,
		productLookup: productLookup,
		writer:        writer,
	}
}

// PromoteInventory promotes a reconciled ERP inventory position into the
// canonical inventory module.
func (p *InventoryPromotion) PromoteInventory(ctx context.Context, result *domain.ReconciliationResult) (string, error) {
	if p == nil || p.stagingRepo == nil || p.runRepo == nil || p.productLookup == nil || p.writer == nil {
		return "", ErrInventoryPromotionNotConfigured
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
	if staging.EntityType != domain.EntityTypeInventory {
		return "", fmt.Errorf("staging record %s has unsupported entity type %s", staging.StagingID, staging.EntityType)
	}
	if !strings.EqualFold(strings.TrimSpace(staging.ValidationStatus), "valid") {
		return "", fmt.Errorf("staging record %s is not valid", staging.StagingID)
	}

	input, err := buildInventoryPromotionInput(staging)
	if err != nil {
		return "", err
	}

	productID, found, err := p.productLookup.FindProductIDBySKU(ctx, result.TenantID, input.ProductSourceID)
	if err != nil {
		return "", fmt.Errorf("lookup canonical product %s for inventory promotion: %w", input.ProductSourceID, err)
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
	return p.writer.PromoteInventory(ctx, traceID, result, run, ports.InventoryPromotionInput{
		ProductID:          productID,
		SourceCompanyCode:  input.SourceCompanyCode,
		SourceLocationCode: input.SourceLocationCode,
		OnHandQuantity:     input.OnHandQuantity,
		LastPurchaseAt:     input.LastPurchaseAt,
		LastSaleAt:         input.LastSaleAt,
		PositionStatus:     input.PositionStatus,
		EffectiveFrom:      input.EffectiveFrom,
		EffectiveTo:        input.EffectiveTo,
		OriginType:         input.OriginType,
		OriginRef:          input.OriginRef,
		ReasonCode:         input.ReasonCode,
		UpdatedBy:          input.UpdatedBy,
	})
}

type inventoryPromotionStagingInput struct {
	ProductSourceID    string
	SourceCompanyCode  string
	SourceLocationCode string
	OnHandQuantity     float64
	LastPurchaseAt     *time.Time
	LastSaleAt         *time.Time
	PositionStatus     string
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	OriginType         string
	OriginRef          string
	ReasonCode         string
	UpdatedBy          string
}

func buildInventoryPromotionInput(staging *domain.StagingRecord) (inventoryPromotionStagingInput, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(staging.NormalizedJSON, &payload); err != nil {
		return inventoryPromotionStagingInput{}, fmt.Errorf("unmarshal normalized inventory staging %s: %w", staging.StagingID, err)
	}

	productSourceID := firstNonBlank(readStringField(payload, "product_source_id"), readStringField(payload, "CODPROD"))
	if productSourceID == "" {
		return inventoryPromotionStagingInput{}, fmt.Errorf("normalized inventory staging %s is missing product source id", staging.StagingID)
	}
	sourceCompanyCode := firstNonBlank(readStringField(payload, "source_company_code"), readStringField(payload, "CODEMP"))
	sourceLocationCode := firstNonBlank(readStringField(payload, "source_location_code"), readStringField(payload, "CODLOCAL"))
	if sourceCompanyCode == "" || sourceLocationCode == "" {
		return inventoryPromotionStagingInput{}, fmt.Errorf("normalized inventory staging %s is missing source location fields", staging.StagingID)
	}

	effectiveFrom := staging.NormalizedAt.UTC()
	if parsed, ok := readTimeField(payload, "effective_from"); ok && parsed != nil && !parsed.IsZero() {
		effectiveFrom = parsed.UTC()
	}

	onHandQuantity, ok := readFloatFieldValue(payload, "raw_quantity")
	if !ok {
		onHandQuantity, ok = readFloatFieldValue(payload, "ESTOQUE")
		if !ok {
			onHandQuantity = 0
		}
	}

	lastPurchaseAt, _ := readTimeField(payload, "last_purchase_at")
	lastSaleAt, _ := readTimeField(payload, "last_sale_at")

	return inventoryPromotionStagingInput{
		ProductSourceID:    productSourceID,
		SourceCompanyCode:  sourceCompanyCode,
		SourceLocationCode: sourceLocationCode,
		OnHandQuantity:     onHandQuantity,
		LastPurchaseAt:     lastPurchaseAt,
		LastSaleAt:         lastSaleAt,
		PositionStatus:     firstNonBlank(readStringField(payload, "position_status"), "active"),
		EffectiveFrom:      effectiveFrom,
		EffectiveTo:        nil,
		OriginType:         firstNonBlank(readStringField(payload, "origin_type"), "import"),
		OriginRef:          firstNonBlank(readStringField(payload, "origin_ref"), staging.SourceID),
		ReasonCode:         firstNonBlank(readStringField(payload, "reason_code"), "ERP_INVENTORY_PROMOTED"),
		UpdatedBy:          firstNonBlank(readStringField(payload, "updated_by"), "erp_integrations"),
	}, nil
}
