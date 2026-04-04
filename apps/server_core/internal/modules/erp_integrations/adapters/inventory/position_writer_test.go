package inventory

import (
	"context"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	inventoryapp "metalshopping/server_core/internal/modules/inventory/application"
	inventorydomain "metalshopping/server_core/internal/modules/inventory/domain"
)

type recordingPositionService struct {
	cmd      inventoryapp.SetProductPositionCommand
	position inventorydomain.ProductPosition
	err      error
}

func (r *recordingPositionService) SetProductPosition(_ context.Context, cmd inventoryapp.SetProductPositionCommand) (inventorydomain.ProductPosition, bool, error) {
	r.cmd = cmd
	if r.err != nil {
		return inventorydomain.ProductPosition{}, false, r.err
	}
	if r.position.PositionID == "" {
		r.position = inventorydomain.ProductPosition{PositionID: "pos_1"}
	}
	return r.position, true, nil
}

func TestWriterPromoteInventoryUsesERPSourceLocationFields(t *testing.T) {
	service := &recordingPositionService{}
	writer := NewWriter(service)

	positionID, err := writer.PromoteInventory(context.Background(), "trace_1", &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
	}, &domain.SyncRun{
		RunID:      "run_1",
		TenantID:   "tenant-1",
		InstanceID: "inst_1",
	}, ports.InventoryPromotionInput{
		ProductID:          "prd_1",
		SourceCompanyCode:  "1",
		SourceLocationCode: "10101",
		OnHandQuantity:     -4,
		PositionStatus:     "active",
		EffectiveFrom:      time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		OriginType:         "import",
		ReasonCode:         "ERP_INVENTORY_PROMOTED",
		UpdatedBy:          "erp_integrations",
	})
	if err != nil {
		t.Fatalf("PromoteInventory error: %v", err)
	}
	if positionID != "pos_1" {
		t.Fatalf("expected pos_1, got %s", positionID)
	}
	if service.cmd.ProductID != "prd_1" {
		t.Fatalf("expected product id prd_1, got %s", service.cmd.ProductID)
	}
	if service.cmd.SourceCompanyCode != "1" || service.cmd.SourceLocationCode != "10101" {
		t.Fatalf("expected ERP-native location fields, got %+v", service.cmd)
	}
	if service.cmd.OnHandQuantity != -4 {
		t.Fatalf("expected raw quantity -4, got %v", service.cmd.OnHandQuantity)
	}
}
