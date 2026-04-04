package review

import (
	"strings"
	"testing"

	reconciliation_pkg "metalshopping/integration_worker/internal/erp_runtime/reconciliation"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

func TestReviewContextForDuplicateSecondaryIdentifiers(t *testing.T) {
	t.Parallel()

	warningDetails := `{"reason_code":"ERP_PRODUCT_IDENTIFIER_CONFLICT","blocking_scope":"product_prices_inventory","blocked_entities":["products","prices","inventory"],"entity_type":"products","source_id":"SKU-1","staging_id":"stage-1","run_id":"run-1","conflicts":[{"field":"ean","value":"7891234567890","normalized_value":"7891234567890","staging_ids":["stage-1","stage-2"],"source_ids":["SKU-1","SKU-2"]}]}`
	result := &reconciliation_pkg.ReconciliationResult{
		Classification: reconciliation_pkg.ClassificationReviewRequired,
		ReasonCode:     "ERP_PRODUCT_IDENTIFIER_CONFLICT",
		WarningDetails: &warningDetails,
	}

	severity, recommendedAction, problemSummary := reviewContextForResult(result)

	if severity != "warning" {
		t.Fatalf("expected warning severity, got %s", severity)
	}
	if recommendedAction != "review duplicate secondary identifiers in the source ERP and reprocess" {
		t.Fatalf("unexpected recommended action: %s", recommendedAction)
	}
	if !strings.Contains(problemSummary, "Duplicate EAN value") {
		t.Fatalf("expected duplicate EAN summary, got %s", problemSummary)
	}
	if !strings.Contains(problemSummary, "products, prices, and inventory promotion") {
		t.Fatalf("expected downstream block scope in summary, got %s", problemSummary)
	}

	parsed, ok := parseDuplicateReviewWarningDetails(result.WarningDetails)
	if !ok {
		t.Fatal("expected warning details to parse")
	}
	if parsed.BlockingScope != "product_prices_inventory" {
		t.Fatalf("expected blocking scope product_prices_inventory, got %s", parsed.BlockingScope)
	}
	if len(parsed.BlockedEntities) != 3 {
		t.Fatalf("expected 3 blocked entities, got %v", parsed.BlockedEntities)
	}
}

func TestMarshalReconciliationOutputPreservesDuplicateWarningDetails(t *testing.T) {
	t.Parallel()

	warningDetails := `{"reason_code":"ERP_PRODUCT_IDENTIFIER_CONFLICT","blocking_scope":"product_prices_inventory","blocked_entities":["products","prices","inventory"],"entity_type":"products","source_id":"SKU-1","staging_id":"stage-1","run_id":"run-1","conflicts":[{"field":"manufacturer_reference","value":"FAB-1","normalized_value":"fab-1","staging_ids":["stage-1","stage-2"],"source_ids":["SKU-1","SKU-2"]}]}`
	result := &reconciliation_pkg.ReconciliationResult{
		ReconciliationID: "rec-1",
		EntityType:       types.EntityTypeProducts,
		SourceID:         "SKU-1",
		Action:           "skip",
		Classification:   reconciliation_pkg.ClassificationReviewRequired,
		ReasonCode:       "ERP_PRODUCT_IDENTIFIER_CONFLICT",
		WarningDetails:   &warningDetails,
	}

	output, err := marshalReconciliationOutput(result)
	if err != nil {
		t.Fatalf("marshalReconciliationOutput returned error: %v", err)
	}
	if output == nil || !strings.Contains(*output, `"blocking_scope":"product_prices_inventory"`) {
		t.Fatalf("expected blocking scope in reconciliation output, got %v", output)
	}
	if output == nil || !strings.Contains(*output, `"field":"manufacturer_reference"`) {
		t.Fatalf("expected manufacturer reference conflict in reconciliation output, got %v", output)
	}
}
