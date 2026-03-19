package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/shopping/ports"
)

type Service struct {
	reader ports.Reader
	writer ports.Writer
}

func NewService(reader ports.Reader, writer ports.Writer) *Service {
	return &Service{reader: reader, writer: writer}
}

func (s *Service) GetBootstrap(ctx context.Context, tenantID string) (ports.Bootstrap, error) {
	return s.reader.GetBootstrap(ctx, strings.TrimSpace(tenantID))
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (ports.Summary, error) {
	return s.reader.GetSummary(ctx, strings.TrimSpace(tenantID))
}

func (s *Service) ListRuns(ctx context.Context, tenantID string, filter ports.RunListFilter) (ports.RunList, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	return s.reader.ListRuns(ctx, strings.TrimSpace(tenantID), filter)
}

func (s *Service) GetRun(ctx context.Context, tenantID, runID string) (ports.Run, error) {
	return s.reader.GetRun(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(runID))
}

func (s *Service) GetProductLatest(ctx context.Context, tenantID, productID string) (ports.ProductLatest, error) {
	return s.reader.GetProductLatest(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(productID))
}

func (s *Service) GetRunRequest(ctx context.Context, tenantID, runRequestID string) (ports.RunRequest, error) {
	return s.reader.GetRunRequest(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(runRequestID))
}

func (s *Service) CreateRunRequest(ctx context.Context, tenantID string, input ports.CreateRunRequestInput) (ports.RunRequest, error) {
	input.InputMode = strings.ToLower(strings.TrimSpace(input.InputMode))
	input.XLSXFilePath = strings.TrimSpace(input.XLSXFilePath)
	input.Notes = strings.TrimSpace(input.Notes)
	input.RequestedBy = strings.TrimSpace(input.RequestedBy)

	filteredProductIDs := make([]string, 0, len(input.CatalogProductIDs))
	for _, productID := range input.CatalogProductIDs {
		productID = strings.TrimSpace(productID)
		if productID == "" {
			continue
		}
		filteredProductIDs = append(filteredProductIDs, productID)
	}
	input.CatalogProductIDs = filteredProductIDs

	filteredSuppliers := make([]string, 0, len(input.SupplierCodes))
	for _, supplierCode := range input.SupplierCodes {
		supplierCode = strings.ToUpper(strings.TrimSpace(supplierCode))
		if supplierCode == "" {
			continue
		}
		filteredSuppliers = append(filteredSuppliers, supplierCode)
	}
	input.SupplierCodes = filteredSuppliers

	return s.writer.CreateRunRequest(ctx, strings.TrimSpace(tenantID), input)
}
