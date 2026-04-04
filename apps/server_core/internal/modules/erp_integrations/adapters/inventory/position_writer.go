package inventory

import (
	"context"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	inventoryapp "metalshopping/server_core/internal/modules/inventory/application"
	inventorydomain "metalshopping/server_core/internal/modules/inventory/domain"
)

type positionService interface {
	SetProductPosition(ctx context.Context, cmd inventoryapp.SetProductPositionCommand) (inventorydomain.ProductPosition, bool, error)
}

// Writer promotes ERP inventory rows into the canonical inventory module.
type Writer struct {
	service positionService
}

var _ ports.InventoryWriter = (*Writer)(nil)

func NewWriter(service positionService) *Writer {
	return &Writer{service: service}
}

func (w *Writer) PromoteInventory(ctx context.Context, traceID string, result *domain.ReconciliationResult, run *domain.SyncRun, input ports.InventoryPromotionInput) (string, error) {
	if w == nil || w.service == nil {
		return "", fmt.Errorf("inventory promotion writer is not configured")
	}
	if result == nil {
		return "", fmt.Errorf("reconciliation result is required")
	}
	if run == nil {
		return "", fmt.Errorf("sync run is required")
	}
	tenantID := strings.TrimSpace(result.TenantID)
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	productID := strings.TrimSpace(input.ProductID)
	if productID == "" {
		return "", fmt.Errorf("product_id is required")
	}

	position, _, err := w.service.SetProductPosition(ctx, inventoryapp.SetProductPositionCommand{
		TenantID:           tenantID,
		TraceID:            strings.TrimSpace(traceID),
		ProductID:          productID,
		SourceCompanyCode:  strings.TrimSpace(input.SourceCompanyCode),
		SourceLocationCode: strings.TrimSpace(input.SourceLocationCode),
		OnHandQuantity:     input.OnHandQuantity,
		LastPurchaseAt:     input.LastPurchaseAt,
		LastSaleAt:         input.LastSaleAt,
		PositionStatus:     defaultString(strings.TrimSpace(input.PositionStatus), "active"),
		EffectiveFrom:      input.EffectiveFrom.UTC(),
		EffectiveTo:        input.EffectiveTo,
		OriginType:         defaultString(strings.TrimSpace(input.OriginType), "import"),
		OriginRef:          strings.TrimSpace(input.OriginRef),
		ReasonCode:         defaultString(strings.TrimSpace(input.ReasonCode), "ERP_INVENTORY_PROMOTED"),
		UpdatedBy:          defaultString(strings.TrimSpace(input.UpdatedBy), "erp_integrations"),
	})
	if err != nil {
		return "", fmt.Errorf("promote inventory for product %s: %w", productID, err)
	}
	if position.PositionID == "" {
		return "", fmt.Errorf("promote inventory for product %s returned empty position id", productID)
	}
	return position.PositionID, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
