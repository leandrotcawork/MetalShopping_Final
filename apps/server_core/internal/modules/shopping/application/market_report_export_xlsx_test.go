package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"metalshopping/server_core/internal/modules/shopping/ports"
)

type marketReportExportReaderStub struct {
	productIDs              []string
	suppliers               []ports.MarketReportSupplier
	products                []ports.MarketReportProductRow
	runItems                []ports.MarketReportRunItem
	productIDsErr           error
	productsErr             error
	runItemsErr             error
	suppliersErr            error
	gotProductIDsRunID      string
	gotProductIDsSuppliers  []string
	gotProductsProductIDs   []string
	gotRunItemsProductIDs   []string
	gotRunItemsSupplierCodes []string
}

func (s *marketReportExportReaderStub) GetBootstrap(context.Context, string) (ports.Bootstrap, error) {
	return ports.Bootstrap{}, nil
}

func (s *marketReportExportReaderStub) GetSummary(context.Context, string) (ports.Summary, error) {
	return ports.Summary{}, nil
}

func (s *marketReportExportReaderStub) ListRuns(context.Context, string, ports.RunListFilter) (ports.RunList, error) {
	return ports.RunList{}, nil
}

func (s *marketReportExportReaderStub) GetRun(context.Context, string, string) (ports.Run, error) {
	return ports.Run{}, nil
}

func (s *marketReportExportReaderStub) GetRunItemStatusSummary(context.Context, string, string) (ports.RunItemStatusSummary, error) {
	return ports.RunItemStatusSummary{}, nil
}

func (s *marketReportExportReaderStub) GetRunSupplierItemStatusSummary(context.Context, string, string) (ports.RunSupplierItemStatusSummary, error) {
	return ports.RunSupplierItemStatusSummary{}, nil
}

func (s *marketReportExportReaderStub) ListRunItems(context.Context, string, string, ports.RunItemListFilter) (ports.RunItemList, error) {
	return ports.RunItemList{}, nil
}

func (s *marketReportExportReaderStub) ListRunItemsForExport(context.Context, string, string, ports.RunExportListFilter) (ports.RunExportList, error) {
	return ports.RunExportList{}, nil
}

func (s *marketReportExportReaderStub) ListMarketReportProductIDs(
	_ context.Context,
	_ string,
	runID string,
	supplierCodes []string,
) ([]string, error) {
	s.gotProductIDsRunID = runID
	s.gotProductIDsSuppliers = append([]string{}, supplierCodes...)
	if s.productIDsErr != nil {
		return nil, s.productIDsErr
	}
	return append([]string{}, s.productIDs...), nil
}

func (s *marketReportExportReaderStub) ListMarketReportProducts(
	_ context.Context,
	_ string,
	productIDs []string,
) ([]ports.MarketReportProductRow, error) {
	s.gotProductsProductIDs = append([]string{}, productIDs...)
	if s.productsErr != nil {
		return nil, s.productsErr
	}
	return append([]ports.MarketReportProductRow{}, s.products...), nil
}

func (s *marketReportExportReaderStub) ListMarketReportRunItems(
	_ context.Context,
	_ string,
	_ string,
	productIDs []string,
	supplierCodes []string,
) ([]ports.MarketReportRunItem, error) {
	s.gotRunItemsProductIDs = append([]string{}, productIDs...)
	s.gotRunItemsSupplierCodes = append([]string{}, supplierCodes...)
	if s.runItemsErr != nil {
		return nil, s.runItemsErr
	}
	return append([]ports.MarketReportRunItem{}, s.runItems...), nil
}

func (s *marketReportExportReaderStub) ListMarketReportSuppliers(context.Context, string, []string) ([]ports.MarketReportSupplier, error) {
	if s.suppliersErr != nil {
		return nil, s.suppliersErr
	}
	return append([]ports.MarketReportSupplier{}, s.suppliers...), nil
}

func (s *marketReportExportReaderStub) GetProductLatest(context.Context, string, string) (ports.ProductLatest, error) {
	return ports.ProductLatest{}, nil
}

func (s *marketReportExportReaderStub) GetRunRequest(context.Context, string, string) (ports.RunRequest, error) {
	return ports.RunRequest{}, nil
}

func (s *marketReportExportReaderStub) ListSupplierSignals(context.Context, string, ports.SupplierSignalListFilter) (ports.SupplierSignalList, error) {
	return ports.SupplierSignalList{}, nil
}

func (s *marketReportExportReaderStub) ListManualURLCandidates(context.Context, string, ports.ManualURLCandidateFilter) (ports.ManualURLCandidateList, error) {
	return ports.ManualURLCandidateList{}, nil
}

type marketReportExportWriterStub struct{}

func (marketReportExportWriterStub) CreateRunRequest(context.Context, string, string, ports.CreateRunRequestInput) (ports.RunRequest, error) {
	return ports.RunRequest{}, errors.New("unexpected CreateRunRequest call")
}

func (marketReportExportWriterStub) UpsertSupplierSignal(context.Context, string, ports.UpsertSupplierSignalInput) (ports.SupplierSignal, error) {
	return ports.SupplierSignal{}, errors.New("unexpected UpsertSupplierSignal call")
}

func TestExportMarketReportXlsxDerivesProductsFromRun(t *testing.T) {
	exportRoot := t.TempDir()
	t.Setenv(exportRootEnv, exportRoot)

	reader := &marketReportExportReaderStub{
		productIDs: []string{"prd-2", "prd-1"},
		suppliers: []ports.MarketReportSupplier{
			{SupplierCode: "SUP-1", SupplierLabel: "Supplier 1"},
		},
		products: []ports.MarketReportProductRow{
			{ProductID: "prd-2", SKU: "SKU-2", ProductLabel: "Produto 2"},
			{ProductID: "prd-1", SKU: "SKU-1", ProductLabel: "Produto 1"},
		},
		runItems: []ports.MarketReportRunItem{
			{ProductID: "prd-2", SupplierCode: "SUP-1", ItemStatus: "OK", ObservedPrice: 20},
			{ProductID: "prd-1", SupplierCode: "SUP-1", ItemStatus: "OK", ObservedPrice: 10},
		},
	}
	service := NewService(reader, marketReportExportWriterStub{})

	result, err := service.ExportMarketReportXlsx(context.Background(), "tenant-1", "run-1", ports.MarketReportExportXlsxInput{
		SupplierCodes:  []string{"SUP-1"},
		OutputFilePath: "exports",
	})
	if err != nil {
		t.Fatalf("ExportMarketReportXlsx returned error: %v", err)
	}

	expectedProductIDs := []string{"prd-2", "prd-1"}
	if !reflect.DeepEqual(reader.gotProductIDsSuppliers, []string{"SUP-1"}) {
		t.Fatalf("ListMarketReportProductIDs supplierCodes mismatch: got %v", reader.gotProductIDsSuppliers)
	}
	if reader.gotProductIDsRunID != "run-1" {
		t.Fatalf("ListMarketReportProductIDs runID mismatch: got %q", reader.gotProductIDsRunID)
	}
	if !reflect.DeepEqual(reader.gotProductsProductIDs, expectedProductIDs) {
		t.Fatalf("ListMarketReportProducts productIDs mismatch: got %v want %v", reader.gotProductsProductIDs, expectedProductIDs)
	}
	if !reflect.DeepEqual(reader.gotRunItemsProductIDs, expectedProductIDs) {
		t.Fatalf("ListMarketReportRunItems productIDs mismatch: got %v want %v", reader.gotRunItemsProductIDs, expectedProductIDs)
	}
	if !reflect.DeepEqual(reader.gotRunItemsSupplierCodes, []string{"SUP-1"}) {
		t.Fatalf("ListMarketReportRunItems supplierCodes mismatch: got %v", reader.gotRunItemsSupplierCodes)
	}
	if result.TotalProducts != 2 {
		t.Fatalf("TotalProducts mismatch: got %d want 2", result.TotalProducts)
	}
	if !strings.HasPrefix(result.OutputFilePath, filepath.Join(exportRoot, "exports")) {
		t.Fatalf("output path mismatch: got %q", result.OutputFilePath)
	}
	if _, err := os.Stat(result.OutputFilePath); err != nil {
		t.Fatalf("expected export file to exist: %v", err)
	}
}

func TestExportMarketReportXlsxRejectsRunsWithoutDerivedProducts(t *testing.T) {
	t.Setenv(exportRootEnv, t.TempDir())

	reader := &marketReportExportReaderStub{
		productIDs: []string{},
		suppliers: []ports.MarketReportSupplier{
			{SupplierCode: "SUP-1", SupplierLabel: "Supplier 1"},
		},
	}
	service := NewService(reader, marketReportExportWriterStub{})

	_, err := service.ExportMarketReportXlsx(context.Background(), "tenant-1", "run-1", ports.MarketReportExportXlsxInput{
		SupplierCodes:  []string{"SUP-1"},
		OutputFilePath: "exports",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrExportInvalid) {
		t.Fatalf("expected ErrExportInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "no products found for selected suppliers") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
