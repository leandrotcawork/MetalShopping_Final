package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"metalshopping/server_core/internal/modules/shopping/ports"
)

func (s *Service) ExportMarketReportXlsx(
	ctx context.Context,
	tenantID string,
	runID string,
	input ports.MarketReportExportXlsxInput,
) (ports.MarketReportExportXlsxResult, error) {
	outputFilePath := strings.TrimSpace(input.OutputFilePath)
	if outputFilePath == "" {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: outputFilePath is required", ErrExportInvalid)
	}

	exportRoot := strings.TrimSpace(os.Getenv(exportRootEnv))
	if exportRoot == "" {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: MS_SHOPPING_EXPORT_ROOT not set", ErrExportRootMissing)
	}

	resolvedPath, err := resolveExportPath(exportRoot, outputFilePath, defaultMarketReportFileName(strings.TrimSpace(runID)))
	if err != nil {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: %s", ErrExportInvalid, err.Error())
	}

	supplierCodes := normalizeSupplierCodes(input.SupplierCodes)
	if len(supplierCodes) == 0 {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: supplierCodes is required", ErrExportInvalid)
	}

	productIDs, err := s.reader.ListMarketReportProductIDs(
		ctx,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(runID),
		supplierCodes,
	)
	if err != nil {
		return ports.MarketReportExportXlsxResult{}, err
	}
	if len(productIDs) == 0 {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: no products found for selected suppliers", ErrExportInvalid)
	}

	maxRows := exportMaxRowsFromEnv()
	if maxRows > 0 && int64(len(productIDs)) > maxRows {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf(
			"%w: total products %d exceeds max %d",
			ErrExportTooLarge,
			len(productIDs),
			maxRows,
		)
	}

	suppliers, err := s.reader.ListMarketReportSuppliers(ctx, strings.TrimSpace(tenantID), supplierCodes)
	if err != nil {
		return ports.MarketReportExportXlsxResult{}, err
	}
	supplierLabels := mapSupplierLabels(suppliers)
	supplierHeaders := buildSupplierHeaders(supplierCodes, supplierLabels)

	products, err := s.reader.ListMarketReportProducts(ctx, strings.TrimSpace(tenantID), productIDs)
	if err != nil {
		return ports.MarketReportExportXlsxResult{}, err
	}

	runItems, err := s.reader.ListMarketReportRunItems(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(runID), productIDs, supplierCodes)
	if err != nil {
		return ports.MarketReportExportXlsxResult{}, err
	}

	priceLookup := buildMarketReportPriceLookup(runItems)
	orderedProducts := orderMarketReportProducts(productIDs, products)

	if err := writeMarketReportXlsx(resolvedPath, orderedProducts, supplierCodes, supplierHeaders, priceLookup); err != nil {
		return ports.MarketReportExportXlsxResult{}, fmt.Errorf("%w: %s", ErrExportWriteFailed, err.Error())
	}

	return ports.MarketReportExportXlsxResult{
		RunID:          strings.TrimSpace(runID),
		OutputFilePath: resolvedPath,
		ExportedAt:     time.Now().UTC(),
		TotalProducts:  int64(len(orderedProducts)),
		SupplierCodes:  supplierCodes,
	}, nil
}

func defaultMarketReportFileName(runID string) string {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return "shopping_market_report.xlsx"
	}
	return fmt.Sprintf("shopping_market_report_%s.xlsx", runID)
}

func mapSupplierLabels(items []ports.MarketReportSupplier) map[string]string {
	labels := make(map[string]string, len(items))
	for _, item := range items {
		code := strings.ToUpper(strings.TrimSpace(item.SupplierCode))
		label := strings.TrimSpace(item.SupplierLabel)
		if code == "" {
			continue
		}
		labels[code] = label
	}
	return labels
}

func buildSupplierHeaders(codes []string, labels map[string]string) []string {
	base := make([]string, len(codes))
	counts := map[string]int{}
	for idx, code := range codes {
		label := strings.TrimSpace(labels[code])
		if label == "" {
			label = code
		}
		base[idx] = label
		counts[strings.ToLower(label)]++
	}

	headers := make([]string, len(codes))
	for idx, code := range codes {
		label := base[idx]
		if counts[strings.ToLower(label)] > 1 {
			headers[idx] = fmt.Sprintf("%s (%s)", label, code)
		} else {
			headers[idx] = label
		}
	}
	return headers
}

func buildMarketReportPriceLookup(items []ports.MarketReportRunItem) map[string]map[string]float64 {
	result := map[string]map[string]float64{}
	for _, item := range items {
		if strings.ToUpper(strings.TrimSpace(item.ItemStatus)) != "OK" {
			continue
		}
		productID := strings.TrimSpace(item.ProductID)
		supplierCode := strings.ToUpper(strings.TrimSpace(item.SupplierCode))
		if productID == "" || supplierCode == "" {
			continue
		}
		bySupplier, exists := result[productID]
		if !exists {
			bySupplier = map[string]float64{}
			result[productID] = bySupplier
		}
		bySupplier[supplierCode] = item.ObservedPrice
	}
	return result
}

func orderMarketReportProducts(productIDs []string, products []ports.MarketReportProductRow) []ports.MarketReportProductRow {
	lookup := make(map[string]ports.MarketReportProductRow, len(products))
	for _, row := range products {
		lookup[row.ProductID] = row
	}

	ordered := make([]ports.MarketReportProductRow, 0, len(products))
	for _, productID := range productIDs {
		if row, ok := lookup[productID]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}

func writeMarketReportXlsx(
	path string,
	products []ports.MarketReportProductRow,
	supplierCodes []string,
	supplierHeaders []string,
	priceLookup map[string]map[string]float64,
) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	sheetName := "Market Report"
	defaultSheet := file.GetSheetName(file.GetActiveSheetIndex())
	if defaultSheet != sheetName {
		if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
			return fmt.Errorf("rename sheet: %w", err)
		}
	}

	headers := []string{
		"SKU",
		"PN Interno",
		"Reference",
		"EAN",
		"Produto",
		"Marca",
		"Grupo",
		"Nosso Preco",
		"Custo Reposicao",
		"Custo Medio",
	}
	headers = append(headers, supplierHeaders...)

	if err := file.SetSheetRow(sheetName, "A1", &headers); err != nil {
		return fmt.Errorf("write header row: %w", err)
	}

	for idx, row := range products {
		values := []any{
			row.SKU,
			optionalString(row.PNInterno),
			optionalString(row.Reference),
			optionalString(row.EAN),
			row.ProductLabel,
			optionalString(row.BrandName),
			optionalString(row.TaxonomyLeaf0Name),
			optionalFloat64(row.PriceAmount),
			optionalFloat64(row.ReplacementCostAmount),
			optionalFloat64(row.AverageCostAmount),
		}

		bySupplier := priceLookup[row.ProductID]
		for _, supplierCode := range supplierCodes {
			if bySupplier == nil {
				values = append(values, "")
				continue
			}
			price, ok := bySupplier[supplierCode]
			if !ok {
				values = append(values, "")
				continue
			}
			values = append(values, price)
		}

		cell := fmt.Sprintf("A%d", idx+2)
		if err := file.SetSheetRow(sheetName, cell, &values); err != nil {
			return fmt.Errorf("write data row: %w", err)
		}
	}

	if err := file.SaveAs(path); err != nil {
		return fmt.Errorf("save xlsx: %w", err)
	}

	return nil
}
