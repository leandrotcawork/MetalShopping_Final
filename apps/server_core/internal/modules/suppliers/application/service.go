package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"metalshopping/server_core/internal/modules/suppliers/ports"
	platformsuppliers "metalshopping/server_core/internal/platform/suppliers"
)

type Service struct {
	repo     ports.Repository
	registry *platformsuppliers.Registry
}

func NewService(repo ports.Repository, registry *platformsuppliers.Registry) *Service {
	return &Service{
		repo:     repo,
		registry: registry,
	}
}

func (s *Service) ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]ports.DirectorySupplier, error) {
	return s.repo.ListDirectory(ctx, strings.TrimSpace(tenantID), onlyEnabled)
}

func (s *Service) UpsertDirectorySupplier(ctx context.Context, tenantID string, input ports.UpsertDirectorySupplierInput) (ports.DirectorySupplier, error) {
	input.SupplierCode = strings.TrimSpace(input.SupplierCode)
	input.SupplierLabel = strings.TrimSpace(input.SupplierLabel)
	input.ExecutionKind = strings.ToUpper(strings.TrimSpace(input.ExecutionKind))
	input.LookupPolicy = strings.ToUpper(strings.TrimSpace(input.LookupPolicy))

	switch {
	case input.SupplierCode == "":
		return ports.DirectorySupplier{}, errors.New("supplierCode is required")
	case input.SupplierLabel == "":
		return ports.DirectorySupplier{}, errors.New("supplierLabel is required")
	case input.ExecutionKind != "HTTP" && input.ExecutionKind != "PLAYWRIGHT":
		return ports.DirectorySupplier{}, errors.New("executionKind must be HTTP or PLAYWRIGHT")
	case input.LookupPolicy != "EAN_FIRST" && input.LookupPolicy != "REFERENCE_FIRST":
		return ports.DirectorySupplier{}, errors.New("lookupPolicy must be EAN_FIRST or REFERENCE_FIRST")
	}

	return s.repo.UpsertDirectorySupplier(ctx, strings.TrimSpace(tenantID), input)
}

func (s *Service) SetDirectorySupplierEnabled(ctx context.Context, tenantID, supplierCode string, enabled bool) (ports.DirectorySupplier, error) {
	supplierCode = strings.TrimSpace(supplierCode)
	if supplierCode == "" {
		return ports.DirectorySupplier{}, errors.New("supplierCode is required")
	}
	return s.repo.SetDirectorySupplierEnabled(ctx, strings.TrimSpace(tenantID), supplierCode, enabled)
}

func (s *Service) ListDriverManifests(ctx context.Context, tenantID, supplierCode string, limit, offset int64) (ports.DriverManifestList, error) {
	supplierCode = strings.TrimSpace(supplierCode)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListDriverManifests(ctx, strings.TrimSpace(tenantID), supplierCode, limit, offset)
}

func (s *Service) CreateDriverManifest(ctx context.Context, tenantID string, input ports.CreateDriverManifestInput) (ports.DriverManifest, error) {
	input.SupplierCode = strings.TrimSpace(input.SupplierCode)
	input.Family = strings.TrimSpace(input.Family)
	input.CreatedBy = strings.TrimSpace(input.CreatedBy)
	if input.SupplierCode == "" {
		return ports.DriverManifest{}, errors.New("supplierCode is required")
	}
	if input.Family == "" {
		return ports.DriverManifest{}, errors.New("family is required")
	}
	if input.CreatedBy == "" {
		return ports.DriverManifest{}, errors.New("createdBy is required")
	}
	if len(input.ConfigJSON) == 0 {
		input.ConfigJSON = json.RawMessage(`{}`)
	}
	if !json.Valid(input.ConfigJSON) {
		return ports.DriverManifest{}, errors.New("config must be valid JSON")
	}
	return s.repo.CreateDriverManifest(ctx, strings.TrimSpace(tenantID), input)
}

func (s *Service) ValidateDriverManifest(ctx context.Context, tenantID, manifestID string) (ports.DriverManifest, error) {
	manifestID = strings.TrimSpace(manifestID)
	if manifestID == "" {
		return ports.DriverManifest{}, errors.New("manifestId is required")
	}
	if s.registry == nil {
		return ports.DriverManifest{}, errors.New("driver family registry is required")
	}

	current, err := s.repo.GetDriverManifest(ctx, strings.TrimSpace(tenantID), manifestID)
	if err != nil {
		return ports.DriverManifest{}, err
	}

	validationErrors, err := s.registry.Validate(current.Family, current.ConfigJSON)
	if err != nil {
		return ports.DriverManifest{}, err
	}
	normalizedErrors := make([]ports.ValidationError, 0, len(validationErrors))
	for _, item := range validationErrors {
		normalizedErrors = append(normalizedErrors, ports.ValidationError{
			Code:    item.Code,
			Field:   item.Field,
			Message: item.Message,
		})
	}
	return s.repo.ValidateDriverManifest(ctx, strings.TrimSpace(tenantID), manifestID, normalizedErrors)
}

func (s *Service) ActivateDriverManifest(ctx context.Context, tenantID, manifestID string) (ports.DriverManifest, error) {
	manifestID = strings.TrimSpace(manifestID)
	if manifestID == "" {
		return ports.DriverManifest{}, errors.New("manifestId is required")
	}
	return s.repo.ActivateDriverManifest(ctx, strings.TrimSpace(tenantID), manifestID)
}
