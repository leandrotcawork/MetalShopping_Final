package sankhya

import (
	"context"
	"fmt"
	"strings"
	"time"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource/oracle"
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
	UseSID   bool
	Timeout  int
}

// Connector implements erp_runtime.Connector for Sankhya ERP.
type Connector struct {
	extractor     *Extractor
	mapper        *Mapper
	runnerFactory func(context.Context, erp_runtime.ExtractConnection) (dbsource.QueryRunner, error)
}

// New returns a new Sankhya Connector.
func New() *Connector {
	mapper := newMapper()
	return &Connector{
		extractor:     newExtractor(mapper),
		mapper:        mapper,
		runnerFactory: defaultRunnerFactory,
	}
}

// NewWithRunnerFactory is used by tests to inject a fake db source.
func NewWithRunnerFactory(factory func(context.Context, erp_runtime.ExtractConnection) (dbsource.QueryRunner, error)) *Connector {
	connector := New()
	if factory != nil {
		connector.runnerFactory = factory
	}
	return connector
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
	if err != nil {
		return err
	}
	runner, err := c.runnerFactory(ctx, connection)
	if err != nil {
		return err
	}
	return runner.Close()
}

func (c *Connector) Extract(ctx context.Context, req erp_runtime.ExtractRequest) (*erp_runtime.ExtractionResult, error) {
	if isFixtureConnection(req.Connection) {
		return c.extractor.Extract(ctx, req, nil)
	}

	runner, err := c.runnerFactory(ctx, req.Connection)
	if err != nil {
		return nil, fmt.Errorf("sankhya: open query runner: %w", err)
	}
	defer runner.Close() //nolint:errcheck
	return c.extractor.Extract(ctx, req, runner)
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
		Timeout:  connection.ConnectTimeoutSec,
	}
	if hasServiceName {
		cfg.Service = strings.TrimSpace(*connection.ServiceName)
	} else {
		cfg.Service = strings.TrimSpace(*connection.SID)
		cfg.UseSID = true
	}
	return cfg, nil
}

func defaultRunnerFactory(_ context.Context, connection erp_runtime.ExtractConnection) (dbsource.QueryRunner, error) {
	cfg, err := buildConnectionConfig(connection)
	if err != nil {
		return nil, err
	}

	oracleCfg := oracle.Config{
		Host:              cfg.Host,
		Port:              cfg.Port,
		Username:          cfg.Username,
		Password:          cfg.Password,
		ConnectTimeoutSec: cfg.Timeout,
	}
	if cfg.UseSID {
		oracleCfg.SID = &cfg.Service
	} else {
		oracleCfg.ServiceName = &cfg.Service
	}

	runner, err := oracle.NewQueryRunner(oracleCfg)
	if err != nil {
		return nil, err
	}
	return &queryRunnerWithTimeoutDefault{inner: runner}, nil
}

type queryRunnerWithTimeoutDefault struct {
	inner dbsource.QueryRunner
}

func (r *queryRunnerWithTimeoutDefault) Query(ctx context.Context, spec dbsource.QuerySpec, fn func(dbsource.RowReader) error) error {
	if spec.Timeout <= 0 {
		spec.Timeout = 30 * time.Second
	}
	return r.inner.Query(ctx, spec, fn)
}

func (r *queryRunnerWithTimeoutDefault) Close() error {
	return r.inner.Close()
}
