package ports

import (
	"context"
	"encoding/json"
	"time"
)

type DirectorySupplier struct {
	SupplierCode  string
	SupplierLabel string
	ExecutionKind string
	LookupPolicy  string
	Enabled       bool
	UpdatedAt     time.Time
}

type DriverManifest struct {
	ManifestID       string
	SupplierCode     string
	VersionNumber    int64
	Family           string
	ConfigJSON       json.RawMessage
	ValidationStatus string
	ValidationErrors json.RawMessage
	IsActive         bool
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DriverManifestList struct {
	Rows   []DriverManifest
	Offset int64
	Limit  int64
	Total  int64
}

type UpsertDirectorySupplierInput struct {
	SupplierCode  string
	SupplierLabel string
	ExecutionKind string
	LookupPolicy  string
	Enabled       bool
}

type CreateDriverManifestInput struct {
	SupplierCode string
	Family       string
	ConfigJSON   json.RawMessage
	CreatedBy    string
}

type Repository interface {
	ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]DirectorySupplier, error)
	UpsertDirectorySupplier(ctx context.Context, tenantID string, input UpsertDirectorySupplierInput) (DirectorySupplier, error)
	SetDirectorySupplierEnabled(ctx context.Context, tenantID, supplierCode string, enabled bool) (DirectorySupplier, error)
	ListDriverManifests(ctx context.Context, tenantID, supplierCode string, limit, offset int64) (DriverManifestList, error)
	CreateDriverManifest(ctx context.Context, tenantID string, input CreateDriverManifestInput) (DriverManifest, error)
}
