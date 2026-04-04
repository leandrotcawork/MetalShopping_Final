package pricing

import (
	"context"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	pricingapp "metalshopping/server_core/internal/modules/pricing/application"
	pricingdomain "metalshopping/server_core/internal/modules/pricing/domain"
)

type recordingPriceService struct {
	cmd   pricingapp.SetProductPriceCommand
	price pricingdomain.ProductPrice
	err   error
}

func (r *recordingPriceService) SetProductPrice(_ context.Context, cmd pricingapp.SetProductPriceCommand) (pricingdomain.ProductPrice, bool, error) {
	r.cmd = cmd
	if r.err != nil {
		return pricingdomain.ProductPrice{}, false, r.err
	}
	if r.price.PriceID == "" {
		r.price = pricingdomain.ProductPrice{PriceID: "prc_1"}
	}
	return r.price, true, nil
}

type recordingPriceContextResolver struct {
	tenantID        string
	sourceSystem    string
	sourceTableCode string
	resolved        string
	found           bool
	err             error
}

func (r *recordingPriceContextResolver) ResolvePriceContextCode(_ context.Context, tenantID, sourceSystem, sourceTableCode string) (string, bool, error) {
	r.tenantID = tenantID
	r.sourceSystem = sourceSystem
	r.sourceTableCode = sourceTableCode
	if r.err != nil {
		return "", false, r.err
	}
	return r.resolved, r.found, nil
}

func TestWriterPromotePriceUsesResolverAndCanonicalContext(t *testing.T) {
	service := &recordingPriceService{}
	resolver := &recordingPriceContextResolver{resolved: "promotion", found: true}
	writer := NewWriter(service, resolver)

	productID, err := writer.PromotePrice(context.Background(), "trace_1", &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
	}, &domain.SyncRun{
		RunID:      "run_1",
		TenantID:   "tenant-1",
		InstanceID: "inst_1",
	}, ports.PricePromotionInput{
		ProductID:             "prd_1",
		SourceSystem:          "sankhya",
		SourceTableCode:       "17",
		SourceTableID:         "5002",
		SourceTableName:       "Promocao Abril",
		CurrencyCode:          "BRL",
		PriceAmount:           84.5,
		ReplacementCostAmount: 0,
		EffectiveFrom:         time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		OriginType:            "import",
		ReasonCode:            "ERP_PRICE_PROMOTED",
		UpdatedBy:             "erp_integrations",
	})
	if err != nil {
		t.Fatalf("PromotePrice error: %v", err)
	}
	if productID != "prc_1" {
		t.Fatalf("expected prc_1, got %s", productID)
	}
	if resolver.tenantID != "tenant-1" || resolver.sourceSystem != "sankhya" || resolver.sourceTableCode != "17" {
		t.Fatalf("expected resolver to receive tenant/system/table code, got %+v", resolver)
	}
	if service.cmd.ProductID != "prd_1" {
		t.Fatalf("expected product id prd_1, got %s", service.cmd.ProductID)
	}
	if service.cmd.PriceContextCode != "promotion" {
		t.Fatalf("expected resolved price context promotion, got %s", service.cmd.PriceContextCode)
	}
	if service.cmd.CurrencyCode != "BRL" {
		t.Fatalf("expected currency BRL, got %s", service.cmd.CurrencyCode)
	}
}

func TestWriterPromotePriceRejectsMissingMapping(t *testing.T) {
	writer := NewWriter(&recordingPriceService{}, &recordingPriceContextResolver{found: false})

	_, err := writer.PromotePrice(context.Background(), "trace_1", &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
	}, &domain.SyncRun{
		RunID:      "run_1",
		TenantID:   "tenant-1",
		InstanceID: "inst_1",
	}, ports.PricePromotionInput{
		ProductID:       "prd_1",
		SourceSystem:    "sankhya",
		SourceTableCode: "99",
		CurrencyCode:    "BRL",
		PriceAmount:     10,
		EffectiveFrom:   time.Now().UTC(),
		OriginType:      "import",
		ReasonCode:      "ERP_PRICE_PROMOTED",
		UpdatedBy:       "erp_integrations",
	})
	if err == nil {
		t.Fatal("expected missing mapping error")
	}
	if err != ports.ErrPriceContextMappingNotFound {
		t.Fatalf("expected ErrPriceContextMappingNotFound, got %v", err)
	}
}
