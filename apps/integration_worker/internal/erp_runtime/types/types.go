package types

import "time"

// EntityType mirrors the 8 ERP entity types.
type EntityType string

const (
	EntityTypeProducts  EntityType = "products"
	EntityTypePrices    EntityType = "prices"
	EntityTypeCosts     EntityType = "costs"
	EntityTypeInventory EntityType = "inventory"
	EntityTypeSales     EntityType = "sales"
	EntityTypePurchases EntityType = "purchases"
	EntityTypeCustomers EntityType = "customers"
	EntityTypeSuppliers EntityType = "suppliers"
)

// SyncStrategy describes how data is fetched from the ERP.
type SyncStrategy string

const (
	SyncStrategySnapshot    SyncStrategy = "snapshot"
	SyncStrategyIncremental SyncStrategy = "incremental"
)

// EntityCapability describes a connector's support for a specific entity and sync strategy.
type EntityCapability struct {
	Entity   EntityType
	Strategy SyncStrategy
}

// RawRecord is a single record returned by a connector's Extract call.
type RawRecord struct {
	SourceID        string
	ConnectorType   string
	EntityType      EntityType
	PayloadJSON     []byte
	PayloadHash     string
	SourceTimestamp *time.Time
	CursorValue     *string
}

// ExtractionResult holds the output of a single Extract page.
type ExtractionResult struct {
	Records    []*RawRecord
	NextCursor *string
	HasMore    bool
}

// ErrorClass classifies extraction and mapping errors.
type ErrorClass string

const (
	ErrorClassTransient      ErrorClass = "transient"
	ErrorClassSourceData     ErrorClass = "source_data_error"
	ErrorClassMapping        ErrorClass = "mapping_error"
	ErrorClassReconciliation ErrorClass = "reconciliation_error"
	ErrorClassPlatform       ErrorClass = "platform_error"
)

// ExtractConnection carries structured source connection details.
type ExtractConnection struct {
	Kind              string
	Host              string
	Port              int
	ServiceName       string
	SID               string
	Username          string
	PasswordSecretRef string
	ConnectTimeoutSec int
	FetchBatchSize    int
	EntityBatchSize   int
}

// ExtractRequest is the input to a connector's Extract call.
type ExtractRequest struct {
	TenantID   string
	RunID      string
	Entity     EntityType
	Cursor     *string
	Connection ExtractConnection
}
