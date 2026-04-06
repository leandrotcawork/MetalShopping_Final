package erp_runtime

import (
	"context"

	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Re-export all shared types so callers import only erp_runtime.
type EntityType = types.EntityType
type SyncStrategy = types.SyncStrategy
type EntityCapability = types.EntityCapability
type RawRecord = types.RawRecord
type ExtractionResult = types.ExtractionResult
type ErrorClass = types.ErrorClass
type ExtractConnection = types.ExtractConnection
type ExtractRequest = types.ExtractRequest

// Entity type constants.
const (
	EntityTypeProducts  = types.EntityTypeProducts
	EntityTypePrices    = types.EntityTypePrices
	EntityTypeCosts     = types.EntityTypeCosts
	EntityTypeInventory = types.EntityTypeInventory
	EntityTypeSales     = types.EntityTypeSales
	EntityTypePurchases = types.EntityTypePurchases
	EntityTypeCustomers = types.EntityTypeCustomers
	EntityTypeSuppliers = types.EntityTypeSuppliers
)

// SyncStrategy constants.
const (
	SyncStrategySnapshot    = types.SyncStrategySnapshot
	SyncStrategyIncremental = types.SyncStrategyIncremental
)

// ErrorClass constants.
const (
	ErrorClassTransient      = types.ErrorClassTransient
	ErrorClassSourceData     = types.ErrorClassSourceData
	ErrorClassMapping        = types.ErrorClassMapping
	ErrorClassReconciliation = types.ErrorClassReconciliation
	ErrorClassPlatform       = types.ErrorClassPlatform
)

// Connector is the interface all ERP connectors must implement.
type Connector interface {
	// Type returns the connector type identifier (e.g., "sankhya").
	Type() string
	// Capabilities returns the entities this connector supports
	Capabilities() []EntityCapability
	// ValidateConnection checks structured connectivity inputs (does not extract)
	ValidateConnection(ctx context.Context, connection ExtractConnection) error
	// Extract fetches a page of raw records for the given entity
	Extract(ctx context.Context, req ExtractRequest) (*ExtractionResult, error)
	// ClassifyError classifies an extraction or mapping error
	ClassifyError(err error) ErrorClass
}
