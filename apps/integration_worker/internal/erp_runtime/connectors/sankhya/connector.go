package sankhya

import (
	"context"
	"fmt"
	"strings"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

const ConnectorType = "sankhya"

const defaultOraclePort = 1521

// ConnectionConfig captures the validated Sankhya endpoint credentials.
type ConnectionConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Service  string
}

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

func (c *Connector) ValidateConnection(ctx context.Context, connection erp_runtime.ExtractConnection) error {
	_, err := buildConnectionConfig(connection)
	return err
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

func buildConnectionConfig(connection erp_runtime.ExtractConnection) (*ConnectionConfig, error) {
	host := strings.TrimSpace(connection.Host)
	if host == "" {
		return nil, fmt.Errorf("connection host must not be empty")
	}
	if connection.Port <= 0 {
		return nil, fmt.Errorf("connection port must be a positive integer")
	}

	hasServiceName := connection.ServiceName != nil && strings.TrimSpace(*connection.ServiceName) != ""
	hasSID := connection.SID != nil && strings.TrimSpace(*connection.SID) != ""
	switch {
	case hasServiceName && hasSID:
		return nil, fmt.Errorf("connection must set exactly one of service_name or sid")
	case !hasServiceName && !hasSID:
		return nil, fmt.Errorf("connection must set exactly one of service_name or sid")
	}

	username := strings.TrimSpace(connection.Username)
	if username == "" {
		return nil, fmt.Errorf("connection username must not be empty")
	}
	password := strings.TrimSpace(connection.PasswordSecretRef)
	if password == "" {
		return nil, fmt.Errorf("connection password_secret_ref must not be empty")
	}

	port := connection.Port
	if port == 0 {
		port = defaultOraclePort
	}

	cfg := &ConnectionConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
	if hasServiceName {
		cfg.Service = strings.TrimSpace(*connection.ServiceName)
	} else {
		cfg.Service = strings.TrimSpace(*connection.SID)
	}
	return cfg, nil
}
