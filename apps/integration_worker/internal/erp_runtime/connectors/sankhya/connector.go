package sankhya

import (
	"context"
	"fmt"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

const ConnectorType = "sankhya"

// Connector implements erp_runtime.Connector for Sankhya ERP.
type Connector struct {
	extractor *Extractor
	mapper    *Mapper
}

// New returns a new Sankhya Connector.
func New() *Connector {
	return &Connector{
		extractor: newExtractor(),
		mapper:    newMapper(),
	}
}

func (c *Connector) Type() string { return ConnectorType }

func (c *Connector) Capabilities() []erp_runtime.EntityCapability {
	return []erp_runtime.EntityCapability{
		{Entity: erp_runtime.EntityTypeProducts, Strategy: erp_runtime.SyncStrategySnapshot},
		{Entity: erp_runtime.EntityTypePrices, Strategy: erp_runtime.SyncStrategySnapshot},
		{Entity: erp_runtime.EntityTypeCosts, Strategy: erp_runtime.SyncStrategySnapshot},
		{Entity: erp_runtime.EntityTypeInventory, Strategy: erp_runtime.SyncStrategySnapshot},
		{Entity: erp_runtime.EntityTypeSales, Strategy: erp_runtime.SyncStrategyIncremental},
		{Entity: erp_runtime.EntityTypePurchases, Strategy: erp_runtime.SyncStrategyIncremental},
		{Entity: erp_runtime.EntityTypeCustomers, Strategy: erp_runtime.SyncStrategySnapshot},
		{Entity: erp_runtime.EntityTypeSuppliers, Strategy: erp_runtime.SyncStrategySnapshot},
	}
}

func (c *Connector) ValidateConnection(ctx context.Context, connectionRef string) error {
	// v1 stub: Sankhya API authentication would happen here.
	// Connection format: "sankhya://<host>:<port>?serviceKey=<key>"
	// For v1, just validate the connectionRef is non-empty.
	if connectionRef == "" {
		return fmt.Errorf("connectionRef must not be empty")
	}
	return nil // v1 stub — real connection test deferred
}

func (c *Connector) Extract(ctx context.Context, req erp_runtime.ExtractRequest) (*erp_runtime.ExtractionResult, error) {
	return c.extractor.Extract(ctx, req)
}

func (c *Connector) ClassifyError(err error) erp_runtime.ErrorClass {
	if err == nil {
		return erp_runtime.ErrorClassPlatform
	}
	// Add more specific cases as real Sankhya API errors are encountered.
	return erp_runtime.ErrorClassSourceData
}
