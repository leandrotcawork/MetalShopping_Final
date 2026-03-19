package ports

import "context"

type DirectorySupplier struct {
	SupplierCode  string
	SupplierLabel string
	ExecutionKind string
	LookupPolicy  string
	Enabled       bool
}

type DirectoryReader interface {
	ListDirectory(ctx context.Context, tenantID string, onlyEnabled bool) ([]DirectorySupplier, error)
}
