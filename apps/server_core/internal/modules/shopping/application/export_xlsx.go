package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"metalshopping/server_core/internal/modules/shopping/ports"
)

const (
	exportRootEnv     = "MS_SHOPPING_EXPORT_ROOT"
	exportMaxRowsEnv  = "MS_SHOPPING_EXPORT_MAX_ROWS"
	defaultExportRows = int64(5000)
)

var (
	ErrExportInvalid     = errors.New("shopping export invalid")
	ErrExportTooLarge    = errors.New("shopping export too large")
	ErrExportRootMissing = errors.New("shopping export root missing")
	ErrExportWriteFailed = errors.New("shopping export write failed")
)

func (s *Service) ExportRunReportXlsx(
	ctx context.Context,
	tenantID string,
	runID string,
	input ports.RunExportXlsxInput,
) (ports.RunExportXlsxResult, error) {
	outputFilePath := strings.TrimSpace(input.OutputFilePath)
	if outputFilePath == "" {
		return ports.RunExportXlsxResult{}, fmt.Errorf("%w: outputFilePath is required", ErrExportInvalid)
	}

	exportRoot := strings.TrimSpace(os.Getenv(exportRootEnv))
	if exportRoot == "" {
		return ports.RunExportXlsxResult{}, fmt.Errorf("%w: MS_SHOPPING_EXPORT_ROOT not set", ErrExportRootMissing)
	}

	resolvedPath, err := resolveExportPath(exportRoot, outputFilePath, defaultRunExportFileName(strings.TrimSpace(runID)))
	if err != nil {
		return ports.RunExportXlsxResult{}, fmt.Errorf("%w: %s", ErrExportInvalid, err.Error())
	}

	supplierCodes := normalizeSupplierCodes(input.SupplierCodes)
	maxRows := exportMaxRowsFromEnv()

	exportList, err := s.reader.ListRunItemsForExport(ctx, tenantID, runID, ports.RunExportListFilter{
		SupplierCodes: supplierCodes,
		Limit:         maxRows,
	})
	if err != nil {
		return ports.RunExportXlsxResult{}, err
	}
	if maxRows > 0 && exportList.Total > maxRows {
		return ports.RunExportXlsxResult{}, fmt.Errorf(
			"%w: total rows %d exceeds max %d",
			ErrExportTooLarge,
			exportList.Total,
			maxRows,
		)
	}

	if err := writeRunExportXlsx(resolvedPath, exportList.Rows); err != nil {
		return ports.RunExportXlsxResult{}, fmt.Errorf("%w: %s", ErrExportWriteFailed, err.Error())
	}

	exportedAt := time.Now().UTC()
	return ports.RunExportXlsxResult{
		RunID:          strings.TrimSpace(runID),
		OutputFilePath: resolvedPath,
		ExportedAt:     exportedAt,
		TotalRows:      exportList.Total,
		SupplierCodes:  supplierCodes,
	}, nil
}

func normalizeSupplierCodes(raw []string) []string {
	unique := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, code := range raw {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		unique = append(unique, code)
	}
	return unique
}

func resolveExportPath(root, requested, defaultFileName string) (string, error) {
	cleanRoot := filepath.Clean(strings.TrimSpace(root))
	if cleanRoot == "." || cleanRoot == "" {
		return "", errors.New("export root is invalid")
	}
	absRoot, err := filepath.Abs(cleanRoot)
	if err != nil {
		return "", fmt.Errorf("resolve export root: %w", err)
	}

	trimmedRequested := strings.TrimSpace(requested)
	cleanRequested := filepath.Clean(trimmedRequested)
	if cleanRequested == "." || cleanRequested == "" {
		return "", errors.New("outputFilePath must be a file")
	}

	absOutput := cleanRequested
	if !filepath.IsAbs(cleanRequested) {
		absOutput = filepath.Join(absRoot, cleanRequested)
	}
	absOutput, err = filepath.Abs(absOutput)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	absOutput, err = normalizeExportTarget(absOutput, absRoot, trimmedRequested, defaultFileName)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absRoot, absOutput)
	if err != nil {
		return "", fmt.Errorf("relativize output path: %w", err)
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", errors.New("outputFilePath must be under export root")
	}
	if strings.ToLower(filepath.Ext(absOutput)) != ".xlsx" {
		return "", errors.New("outputFilePath must end with .xlsx")
	}
	return absOutput, nil
}

func normalizeExportTarget(absOutput, absRoot, requested, defaultFileName string) (string, error) {
	extension := strings.ToLower(filepath.Ext(absOutput))
	if extension == ".xlsx" {
		return absOutput, nil
	}
	if extension != "" {
		return "", errors.New("outputFilePath must end with .xlsx")
	}

	looksLikeDirectory := strings.HasSuffix(requested, "\\") || strings.HasSuffix(requested, "/")
	if samePath(absOutput, absRoot) || looksLikeDirectory || pathExistsAsDirectory(absOutput) {
		fileName := strings.TrimSpace(defaultFileName)
		if fileName == "" {
			fileName = "shopping_export.xlsx"
		}
		return filepath.Join(absOutput, fileName), nil
	}

	return absOutput + ".xlsx", nil
}

func pathExistsAsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func defaultRunExportFileName(runID string) string {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return "shopping_run_export.xlsx"
	}
	return fmt.Sprintf("shopping_run_%s.xlsx", runID)
}

func exportMaxRowsFromEnv() int64 {
	raw := strings.TrimSpace(os.Getenv(exportMaxRowsEnv))
	if raw == "" {
		return defaultExportRows
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return defaultExportRows
	}
	return value
}

func writeRunExportXlsx(path string, rows []ports.RunExportRow) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	sheetName := "Run Export"
	defaultSheet := file.GetSheetName(file.GetActiveSheetIndex())
	if defaultSheet != sheetName {
		if err := file.SetSheetName(defaultSheet, sheetName); err != nil {
			return fmt.Errorf("rename sheet: %w", err)
		}
	}

	headers := []string{
		"Run ID",
		"Run Item ID",
		"Product ID",
		"SKU",
		"PN Interno",
		"Reference",
		"EAN",
		"Product Name",
		"Brand",
		"Group",
		"Supplier Code",
		"Item Status",
		"Observed Price",
		"Currency",
		"Observed At",
		"Seller Name",
		"Channel",
		"Product URL",
		"Lookup Term",
		"HTTP Status",
		"Elapsed Seconds",
		"Notes",
	}

	if err := file.SetSheetRow(sheetName, "A1", &headers); err != nil {
		return fmt.Errorf("write header row: %w", err)
	}

	for idx, row := range rows {
		record := []any{
			row.RunID,
			row.RunItemID,
			row.ProductID,
			row.SKU,
			optionalString(row.PNInterno),
			optionalString(row.Reference),
			optionalString(row.EAN),
			row.ProductLabel,
			optionalString(row.BrandName),
			optionalString(row.TaxonomyLeaf0Name),
			row.SupplierCode,
			row.ItemStatus,
			row.ObservedPrice,
			row.Currency,
			row.ObservedAt.UTC().Format(time.RFC3339),
			row.SellerName,
			row.Channel,
			optionalString(row.ProductURL),
			optionalString(row.LookupTerm),
			optionalInt64(row.HTTPStatus),
			optionalFloat64(row.ElapsedSeconds),
			optionalString(row.Notes),
		}
		cell := fmt.Sprintf("A%d", idx+2)
		if err := file.SetSheetRow(sheetName, cell, &record); err != nil {
			return fmt.Errorf("write data row: %w", err)
		}
	}

	if err := file.SaveAs(path); err != nil {
		return fmt.Errorf("save xlsx: %w", err)
	}

	return nil
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalInt64(value *int64) any {
	if value == nil {
		return ""
	}
	return *value
}

func optionalFloat64(value *float64) any {
	if value == nil {
		return ""
	}
	return *value
}
