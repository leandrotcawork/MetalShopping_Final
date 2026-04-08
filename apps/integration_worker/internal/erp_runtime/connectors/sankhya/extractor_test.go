package sankhya

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
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

	extractor := newExtractor(newMapper())
	got, err := extractor.Extract(context.Background(), erp_runtime.ExtractRequest{
		Entity: erp_runtime.EntityTypeProducts,
		Connection: erp_runtime.ExtractConnection{
			Kind: "oracle",
			Host: "fixture",
		},
	}, nil)
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

	extractor := newExtractor(newMapper())
	got, err := extractor.Extract(context.Background(), erp_runtime.ExtractRequest{
		Entity: erp_runtime.EntityTypePrices,
		Connection: erp_runtime.ExtractConnection{
			Kind: "oracle",
			Host: "fixture",
		},
	}, nil)
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

func TestExtractProductsUsesQueryRunner(t *testing.T) {
	t.Parallel()

	runner := &fakeQueryRunner{
		rows: []dbsource.RowReader{
			&fakeRowReader{
				values: map[string]any{
					"CODPROD":        "1001",
					"DESCRPROD":      "Produto A",
					"MARCA":          "Marca A",
					"REFERENCIA":     "7891234567890",
					"REFFORN":        "REF-A",
					"ATIVO":          "S",
					"CODVOL":         "UN",
					"CODGRUPOPROD":   "100",
					"AD_STATUS":      "ATIVO",
					"AD_COMPETITIVO": "S",
				},
			},
		},
	}
	extractor := newExtractor(newMapper())

	got, err := extractor.Extract(context.Background(), erp_runtime.ExtractRequest{
		Entity: erp_runtime.EntityTypeProducts,
		Connection: erp_runtime.ExtractConnection{
			Kind:              "oracle",
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
	}, runner)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if runner.calledSQL == "" {
		t.Fatal("expected extractor to call query runner")
	}
	if len(got.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(got.Records))
	}
	if got.Records[0].SourceID != "1001" {
		t.Fatalf("expected source id 1001, got %q", got.Records[0].SourceID)
	}
}

type fakeQueryRunner struct {
	rows      []dbsource.RowReader
	calledSQL string
}

func (f *fakeQueryRunner) Query(_ context.Context, spec dbsource.QuerySpec, fn func(dbsource.RowReader) error) error {
	f.calledSQL = spec.SQL
	for _, row := range f.rows {
		if err := fn(row); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeQueryRunner) Close() error { return nil }

type fakeRowReader struct {
	values map[string]any
}

func (f *fakeRowReader) String(name string) (string, error) {
	value, err := f.NullString(name)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", errors.New("value is null")
	}
	return *value, nil
}

func (f *fakeRowReader) NullString(name string) (*string, error) {
	for key, value := range f.values {
		if strings.EqualFold(key, name) {
			if value == nil {
				return nil, nil
			}
			got := strings.TrimSpace(fmt.Sprint(value))
			return &got, nil
		}
	}
	return nil, errors.New("column not found")
}

func (f *fakeRowReader) Float64(name string) (float64, error) {
	value, err := f.NullFloat64(name)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, errors.New("value is null")
	}
	return *value, nil
}

func (f *fakeRowReader) NullFloat64(name string) (*float64, error) {
	return nil, errors.New("not used")
}

func (f *fakeRowReader) Time(name string) (time.Time, error) {
	value, err := f.NullTime(name)
	if err != nil {
		return time.Time{}, err
	}
	if value == nil {
		return time.Time{}, errors.New("value is null")
	}
	return *value, nil
}

func (f *fakeRowReader) NullTime(name string) (*time.Time, error) {
	return nil, errors.New("not used")
}
