package pricing

import (
	"context"
	"fmt"
	"strings"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	pricingapp "metalshopping/server_core/internal/modules/pricing/application"
	pricingdomain "metalshopping/server_core/internal/modules/pricing/domain"
)

type priceService interface {
	SetProductPrice(ctx context.Context, cmd pricingapp.SetProductPriceCommand) (pricingdomain.ProductPrice, bool, error)
}

// Writer promotes ERP price rows into the canonical pricing module.
type Writer struct {
	service  priceService
	resolver ports.PriceContextResolver
}

var _ ports.PriceWriter = (*Writer)(nil)

func NewWriter(service priceService, resolver ports.PriceContextResolver) *Writer {
	return &Writer{service: service, resolver: resolver}
}

func (w *Writer) PromotePrice(ctx context.Context, traceID string, result *domain.ReconciliationResult, run *domain.SyncRun, input ports.PricePromotionInput) (string, error) {
	if w == nil || w.service == nil {
		return "", fmt.Errorf("price promotion writer is not configured")
	}
	if w.resolver == nil {
		return "", fmt.Errorf("price context resolver is required")
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

	priceContextCode, found, err := w.resolver.ResolvePriceContextCode(ctx, tenantID, input.SourceSystem, input.SourceTableCode)
	if err != nil {
		return "", fmt.Errorf("resolve price context for source table %s: %w", input.SourceTableCode, err)
	}
	if !found || strings.TrimSpace(priceContextCode) == "" {
		return "", ports.ErrPriceContextMappingNotFound
	}

	price, _, err := w.service.SetProductPrice(ctx, pricingapp.SetProductPriceCommand{
		TenantID:              tenantID,
		TraceID:               strings.TrimSpace(traceID),
		ProductID:             productID,
		PriceContextCode:      priceContextCode,
		CurrencyCode:          defaultString(strings.TrimSpace(input.CurrencyCode), "BRL"),
		PriceAmount:           input.PriceAmount,
		ReplacementCostAmount: input.ReplacementCostAmount,
		AverageCostAmount:     input.AverageCostAmount,
		PricingStatus:         defaultString(strings.TrimSpace(input.PricingStatus), "active"),
		EffectiveFrom:         input.EffectiveFrom.UTC(),
		EffectiveTo:           input.EffectiveTo,
		OriginType:            defaultString(strings.TrimSpace(input.OriginType), "import"),
		OriginRef:             strings.TrimSpace(input.OriginRef),
		ReasonCode:            defaultString(strings.TrimSpace(input.ReasonCode), "ERP_PRICE_PROMOTED"),
		UpdatedBy:             defaultString(strings.TrimSpace(input.UpdatedBy), "erp_integrations"),
	})
	if err != nil {
		return "", fmt.Errorf("promote price for product %s: %w", productID, err)
	}
	if price.PriceID == "" {
		return "", fmt.Errorf("promote price for product %s returned empty price id", productID)
	}
	return price.PriceID, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
