package sankhya

import (
	"context"
	"testing"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

func stringPtr(s string) *string {
	return &s
}

func loadFixtureCount(t *testing.T, name string) int {
	t.Helper()

	rows, err := loadFixtureRows(name)
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return len(rows)
}

func TestExtractorProductsSnapshotFixture(t *testing.T) {
	t.Parallel()

	expectedRows := loadFixtureCount(t, "products_fixture.json")
	if expectedRows == 0 {
		t.Fatal("expected products fixture to contain at least one discovered row")
	}

	extractor := newExtractor()
	got, err := extractor.Extract(context.Background(), erp_runtime.ExtractRequest{
		Entity: erp_runtime.EntityTypeProducts,
		Connection: erp_runtime.ExtractConnection{
			Kind: "oracle",
			Host: "fixture",
		},
	})
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	if len(got.Records) != expectedRows {
		t.Fatalf("expected %d product records from fixture shape, got %d", expectedRows, len(got.Records))
	}
	if got.Records[0].ConnectorType != ConnectorType {
		t.Fatalf("expected connector type %q, got %q", ConnectorType, got.Records[0].ConnectorType)
	}
	if got.Records[0].SourceID == "" {
		t.Fatal("expected the first extracted record to carry a source ID")
	}
	if len(got.Records[0].PayloadJSON) == 0 {
		t.Fatal("expected extracted payload JSON to be non-empty")
	}
}

func TestExtractorPricesSnapshotAllowsNullDTVIGOR(t *testing.T) {
	t.Parallel()

	expectedRows := loadFixtureCount(t, "prices_fixture.json")
	if expectedRows < 3 {
		t.Fatalf("expected prices fixture to contain the null DTVIGOR row, got %d rows", expectedRows)
	}

	extractor := newExtractor()
	got, err := extractor.Extract(context.Background(), erp_runtime.ExtractRequest{
		Entity: erp_runtime.EntityTypePrices,
		Connection: erp_runtime.ExtractConnection{
			Kind: "oracle",
			Host: "fixture",
		},
	})
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	if len(got.Records) != expectedRows {
		t.Fatalf("expected %d price records from fixture shape, got %d", expectedRows, len(got.Records))
	}
	if got.Records[2].SourceID != "5003:10002:99" {
		t.Fatalf("expected fallback price source id 5003:10002:99, got %q", got.Records[2].SourceID)
	}
	if len(got.Records[2].PayloadJSON) == 0 {
		t.Fatal("expected null DTVIGOR price payload to be non-empty")
	}
}
