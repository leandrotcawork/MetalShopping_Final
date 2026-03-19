package ports

import (
	"context"
	"time"
)

type Summary struct {
	TotalRuns     int64
	RunningRuns   int64
	CompletedRuns int64
	FailedRuns    int64
	LastRunAt     *time.Time
}

type Run struct {
	RunID          string
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	ProcessedItems int64
	TotalItems     int64
	Notes          string
}

type RunList struct {
	Rows   []Run
	Offset int64
	Limit  int64
	Total  int64
}

type ProductLatest struct {
	ProductID     string
	RunID         string
	ObservedAt    time.Time
	SellerName    string
	Channel       string
	ObservedPrice float64
	Currency      string
}

type RunListFilter struct {
	Status string
	Limit  int64
	Offset int64
}

type Reader interface {
	GetBootstrap(ctx context.Context, tenantID string) (Bootstrap, error)
	GetSummary(ctx context.Context, tenantID string) (Summary, error)
	ListRuns(ctx context.Context, tenantID string, filter RunListFilter) (RunList, error)
	GetRun(ctx context.Context, tenantID, runID string) (Run, error)
	GetProductLatest(ctx context.Context, tenantID, productID string) (ProductLatest, error)
	GetRunRequest(ctx context.Context, tenantID, runRequestID string) (RunRequest, error)
}

type BootstrapSupplier struct {
	SupplierCode  string
	SupplierLabel string
	ExecutionKind string
	LookupPolicy  string
	Enabled       bool
}

type AdvancedDefaults struct {
	TimeoutSeconds   int64
	HTTPWorkers      int64
	PlaywrightWorker int64
	TopN             int64
}

type Bootstrap struct {
	InputModes       []string
	RunStatuses      []string
	SupportsManual   bool
	AdvancedDefaults AdvancedDefaults
	Suppliers        []BootstrapSupplier
}

type CreateRunRequestInput struct {
	InputMode         string
	CatalogProductIDs []string
	XLSXFilePath      string
	XLSXScopeIDs      []string
	SupplierCodes     []string
	Advanced          AdvancedDefaults
	Notes             string
	RequestedBy       string
}

type RunRequest struct {
	RunRequestID string
	Status       string
	InputMode    string
	RequestedAt  time.Time
	RequestedBy  string
	ClaimedAt    *time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
	WorkerID     *string
	RunID        *string
	ErrorMessage *string

	CatalogProductIDs         []string
	XLSXFilePath              *string
	XLSXScopeIDs              []string
	ResolvedCatalogProductIDs []string
	UnresolvedScopeIDs        []string
	AmbiguousScopeIDs         []string
}

type Writer interface {
	CreateRunRequest(ctx context.Context, tenantID string, input CreateRunRequestInput) (RunRequest, error)
}
