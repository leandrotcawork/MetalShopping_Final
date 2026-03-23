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

type RunItemStatusCount struct {
	ItemStatus string
	Total      int64
}

type RunItemStatusSummary struct {
	RunID      string
	TotalItems int64
	Rows       []RunItemStatusCount
}

type RunSupplierItemStatusCount struct {
	SupplierCode string
	Total        int64
	Ok           int64
	NotFound     int64
	Ambiguous    int64
	Error        int64
}

type RunSupplierItemStatusSummary struct {
	RunID          string
	TotalSuppliers int64
	Rows           []RunSupplierItemStatusCount
}

type RunItem struct {
	RunItemID      string
	RunID          string
	ProductID      string
	ProductLabel   string
	PNInterno      *string
	Reference      *string
	SupplierCode   string
	ItemStatus     string
	ObservedPrice  float64
	Currency       string
	ObservedAt     time.Time
	SellerName     string
	Channel        string
	ProductURL     *string
	HTTPStatus     *int64
	ElapsedSeconds *float64
	LookupTerm     *string
	Notes          *string
}

type RunItemList struct {
	Rows   []RunItem
	Offset int64
	Limit  int64
	Total  int64
}

type RunExportRow struct {
	RunItemID         string
	RunID             string
	ProductID         string
	SKU               string
	PNInterno         *string
	Reference         *string
	EAN               *string
	ProductLabel      string
	BrandName         *string
	TaxonomyLeaf0Name *string
	SupplierCode      string
	ItemStatus        string
	ObservedPrice     float64
	Currency          string
	ObservedAt        time.Time
	SellerName        string
	Channel           string
	ProductURL        *string
	LookupTerm        *string
	HTTPStatus        *int64
	ElapsedSeconds    *float64
	Notes             *string
}

type RunExportList struct {
	Rows  []RunExportRow
	Total int64
}

type RunExportListFilter struct {
	SupplierCodes []string
	Limit         int64
}

type RunExportXlsxInput struct {
	SupplierCodes  []string
	OutputFilePath string
}

type RunExportXlsxResult struct {
	RunID          string
	OutputFilePath string
	ExportedAt     time.Time
	TotalRows      int64
	SupplierCodes  []string
}

type MarketReportProductRow struct {
	ProductID             string
	SKU                   string
	PNInterno             *string
	Reference             *string
	EAN                   *string
	ProductLabel          string
	BrandName             *string
	TaxonomyLeaf0Name     *string
	PriceAmount           *float64
	ReplacementCostAmount *float64
	AverageCostAmount     *float64
	CurrencyCode          *string
}

type MarketReportRunItem struct {
	ProductID     string
	SupplierCode  string
	ItemStatus    string
	ObservedPrice float64
}

type MarketReportSupplier struct {
	SupplierCode  string
	SupplierLabel string
}

type MarketReportExportXlsxInput struct {
	SupplierCodes  []string
	ProductIDs     []string
	OutputFilePath string
}

type MarketReportExportXlsxResult struct {
	RunID          string
	OutputFilePath string
	ExportedAt     time.Time
	TotalProducts  int64
	SupplierCodes  []string
}

type RunItemListFilter struct {
	SupplierCode string
	ItemStatus   string
	Limit        int64
	Offset       int64
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

type SupplierSignal struct {
	ProductID        string
	SupplierCode     string
	ProductURL       *string
	URLStatus        string
	LookupMode       string
	LookupModeSource string
	ManualOverride   bool
	LastCheckedAt    *time.Time
	LastSuccessAt    *time.Time
	LastHTTPStatus   *int64
	LastErrorMessage *string
	NextDiscoveryAt  *time.Time
	NotFoundCount    int64
	UpdatedAt        time.Time
}

type SupplierSignalListFilter struct {
	SupplierCode string
	ProductID    string
	Limit        int64
	Offset       int64
}

type SupplierSignalList struct {
	Rows   []SupplierSignal
	Offset int64
	Limit  int64
	Total  int64
}

type ManualURLCandidate struct {
	ProductID         string
	SupplierCode      string
	SKU               string
	PNInterno         *string
	Reference         *string
	EAN               *string
	Name              string
	BrandName         *string
	TaxonomyLeaf0Name *string
	ProductURL        *string
	URLStatus         string
	LookupMode        string
	LookupModeSource  string
	ManualOverride    bool
	LastCheckedAt     *time.Time
	LastSuccessAt     *time.Time
	LastHTTPStatus    *int64
	LastErrorMessage  *string
	NextDiscoveryAt   *time.Time
	NotFoundCount     int64
	UpdatedAt         time.Time
}

type ManualURLCandidateFilter struct {
	SupplierCode      string
	Search            string
	BrandName         string
	TaxonomyLeaf0Name string
	IncludeExisting   bool
	OnlyWithURL       bool
	Limit             int64
	Offset            int64
}

type ManualURLCandidateList struct {
	Rows   []ManualURLCandidate
	Offset int64
	Limit  int64
	Total  int64
}

type UpsertSupplierSignalInput struct {
	ProductID      string
	SupplierCode   string
	ProductURL     *string
	URLStatus      *string
	LookupMode     *string
	ManualOverride *bool
	UpdatedBy      string
}

type Reader interface {
	GetBootstrap(ctx context.Context, tenantID string) (Bootstrap, error)
	GetSummary(ctx context.Context, tenantID string) (Summary, error)
	ListRuns(ctx context.Context, tenantID string, filter RunListFilter) (RunList, error)
	GetRun(ctx context.Context, tenantID, runID string) (Run, error)
	GetRunItemStatusSummary(ctx context.Context, tenantID, runID string) (RunItemStatusSummary, error)
	GetRunSupplierItemStatusSummary(ctx context.Context, tenantID, runID string) (RunSupplierItemStatusSummary, error)
	ListRunItems(ctx context.Context, tenantID, runID string, filter RunItemListFilter) (RunItemList, error)
	ListRunItemsForExport(ctx context.Context, tenantID, runID string, filter RunExportListFilter) (RunExportList, error)
	ListMarketReportProducts(ctx context.Context, tenantID string, productIDs []string) ([]MarketReportProductRow, error)
	ListMarketReportRunItems(ctx context.Context, tenantID, runID string, productIDs []string, supplierCodes []string) ([]MarketReportRunItem, error)
	ListMarketReportSuppliers(ctx context.Context, tenantID string, supplierCodes []string) ([]MarketReportSupplier, error)
	GetProductLatest(ctx context.Context, tenantID, productID string) (ProductLatest, error)
	GetRunRequest(ctx context.Context, tenantID, runRequestID string) (RunRequest, error)
	ListSupplierSignals(ctx context.Context, tenantID string, filter SupplierSignalListFilter) (SupplierSignalList, error)
	ListManualURLCandidates(ctx context.Context, tenantID string, filter ManualURLCandidateFilter) (ManualURLCandidateList, error)
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
	RunRequestID        string
	Status              string
	InputMode           string
	RequestedAt         time.Time
	RequestedBy         string
	ClaimedAt           *time.Time
	StartedAt           *time.Time
	FinishedAt          *time.Time
	WorkerID            *string
	RunID               *string
	ErrorMessage        *string
	TotalItems          *int64
	ProcessedItems      *int64
	CurrentSupplierCode *string
	CurrentProductID    *string
	CurrentProductLabel *string
	ProgressUpdatedAt   *time.Time

	CatalogProductIDs         []string
	XLSXFilePath              *string
	XLSXScopeIDs              []string
	ResolvedCatalogProductIDs []string
	UnresolvedScopeIDs        []string
	AmbiguousScopeIDs         []string
}

type Writer interface {
	CreateRunRequest(ctx context.Context, tenantID, traceID string, input CreateRunRequestInput) (RunRequest, error)
	UpsertSupplierSignal(ctx context.Context, tenantID string, input UpsertSupplierSignalInput) (SupplierSignal, error)
}
