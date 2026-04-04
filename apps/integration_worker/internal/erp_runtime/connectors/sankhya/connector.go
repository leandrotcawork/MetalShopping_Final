package sankhya

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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

func (c *Connector) ValidateConnection(ctx context.Context, connectionRef string) error {
	_, err := parseConnectionRef(connectionRef)
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

func parseConnectionRef(connectionRef string) (*ConnectionConfig, error) {
	ref := strings.TrimSpace(connectionRef)
	if ref == "" {
		return nil, fmt.Errorf("connectionRef must not be empty")
	}

	u, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("parse sankhya connectionRef: %w", err)
	}
	if u.Scheme != "sankhya" {
		return nil, fmt.Errorf("unsupported sankhya connectionRef scheme %q", u.Scheme)
	}

	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("connectionRef host must not be empty")
	}

	port := defaultOraclePort
	if rawPort := strings.TrimSpace(u.Port()); rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil || parsedPort <= 0 {
			return nil, fmt.Errorf("connectionRef port must be a positive integer")
		}
		port = parsedPort
	}

	username, password := parseUserInfo(u)
	if username == "" {
		return nil, fmt.Errorf("connectionRef username must not be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("connectionRef password must not be empty")
	}

	service := strings.TrimSpace(u.Query().Get("service"))
	if service == "" {
		service = strings.TrimSpace(u.Query().Get("serviceName"))
	}

	return &ConnectionConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Service:  service,
	}, nil
}

func parseUserInfo(u *url.URL) (string, string) {
	username := ""
	password := ""

	if u.User != nil {
		username = strings.TrimSpace(u.User.Username())
		if rawPassword, ok := u.User.Password(); ok {
			password = strings.TrimSpace(rawPassword)
		}
	}

	if username == "" {
		username = strings.TrimSpace(u.Query().Get("user"))
	}
	if username == "" {
		username = strings.TrimSpace(u.Query().Get("username"))
	}
	if password == "" {
		password = strings.TrimSpace(u.Query().Get("password"))
	}

	return username, password
}
