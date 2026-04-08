package domain

import (
	"testing"
)

func stringPtr(v string) *string { return &v }

func validConnectionConfig() InstanceConnectionConfig {
	return InstanceConnectionConfig{
		Kind:              ConnectionKindOracle,
		Host:              "10.55.10.101",
		Port:              1521,
		ServiceName:       stringPtr("ORCL"),
		Username:          "leandroth",
		PasswordSecretRef: "erp/sankhya/password",
	}
}

// ---------------------------------------------------------------------------
// IntegrationInstance.ValidateForWrite
// ---------------------------------------------------------------------------

func TestIntegrationInstance_ValidateForWrite_Valid(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection:      validConnectionConfig(),
		EnabledEntities: []EntityType{EntityTypeProducts, EntityTypePrices},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_MissingTenantID(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection:      validConnectionConfig(),
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrEmptyTenantID {
		t.Fatalf("expected ErrEmptyTenantID, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_InvalidConnectorType(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorType("unknown"),
		DisplayName:     "My ERP",
		Connection:      validConnectionConfig(),
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidConnectorType {
		t.Fatalf("expected ErrInvalidConnectorType, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_EmptyEntities(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection:      validConnectionConfig(),
		EnabledEntities: []EntityType{},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrEmptyEnabledEntities {
		t.Fatalf("expected ErrEmptyEnabledEntities, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_InvalidEntity(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection:      validConnectionConfig(),
		EnabledEntities: []EntityType{EntityTypeProducts, EntityType("unknown_entity")},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidEntityType {
		t.Fatalf("expected ErrInvalidEntityType, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_MissingOracleTarget(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection: InstanceConnectionConfig{
			Kind:              ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidOracleConnectionTarget {
		t.Fatalf("expected ErrInvalidOracleConnectionTarget, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_RejectsBothOracleTargets(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection: InstanceConnectionConfig{
			Kind:              ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			SID:               stringPtr("ORCLSID"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidOracleConnectionTarget {
		t.Fatalf("expected ErrInvalidOracleConnectionTarget, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_InvalidConnectionKind(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection: InstanceConnectionConfig{
			Kind:              ConnectionKind("unknown"),
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidConnectionKind {
		t.Fatalf("expected ErrInvalidConnectionKind, got: %v", err)
	}
}

func TestIntegrationInstance_ValidateForWrite_InvalidOraclePort(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		Connection: InstanceConnectionConfig{
			Kind:              ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              0,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		},
		EnabledEntities: []EntityType{EntityTypeProducts},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidOraclePort {
		t.Fatalf("expected ErrInvalidOraclePort, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SyncRun.ValidateForCreate
// ---------------------------------------------------------------------------

func TestSyncRun_ValidateForCreate_Valid(t *testing.T) {
	run := &SyncRun{
		RunID:         "run-1",
		TenantID:      "tenant-abc",
		InstanceID:    "inst-1",
		ConnectorType: ConnectorTypeSankhya,
		RunMode:       RunModeBulk,
		EntityScope:   []EntityType{EntityTypeProducts, EntityTypeInventory},
		Status:        RunStatusPending,
	}
	if err := run.ValidateForCreate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSyncRun_ValidateForCreate_InvalidRunMode(t *testing.T) {
	run := &SyncRun{
		RunID:         "run-1",
		TenantID:      "tenant-abc",
		InstanceID:    "inst-1",
		ConnectorType: ConnectorTypeSankhya,
		RunMode:       RunMode("bad_mode"),
		EntityScope:   []EntityType{EntityTypeProducts},
		Status:        RunStatusPending,
	}
	if err := run.ValidateForCreate(); err != ErrInvalidRunMode {
		t.Fatalf("expected ErrInvalidRunMode, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ConnectorType.IsValid
// ---------------------------------------------------------------------------

func TestConnectorType_IsValid(t *testing.T) {
	tests := []struct {
		value ConnectorType
		want  bool
	}{
		{ConnectorTypeSankhya, true},
		{ConnectorType("unknown"), false},
		{ConnectorType(""), false},
		{ConnectorType("SANKHYA"), false},
	}
	for _, tc := range tests {
		got := tc.value.IsValid()
		if got != tc.want {
			t.Errorf("ConnectorType(%q).IsValid() = %v, want %v", tc.value, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// EntityType.IsValid
// ---------------------------------------------------------------------------

func TestEntityType_IsValid(t *testing.T) {
	valid := []EntityType{
		EntityTypeProducts,
		EntityTypePrices,
		EntityTypeCosts,
		EntityTypeInventory,
		EntityTypeSales,
		EntityTypePurchases,
		EntityTypeCustomers,
		EntityTypeSuppliers,
	}
	for _, e := range valid {
		if !e.IsValid() {
			t.Errorf("EntityType(%q).IsValid() = false, want true", e)
		}
	}

	invalid := []EntityType{
		EntityType("unknown"),
		EntityType(""),
		EntityType("PRODUCTS"),
	}
	for _, e := range invalid {
		if e.IsValid() {
			t.Errorf("EntityType(%q).IsValid() = true, want false", e)
		}
	}
}

func TestReviewBlockScopeProductPricesInventory(t *testing.T) {
	got := ReviewBlockScopeProductPricesInventory.BlockedEntities()
	want := []EntityType{EntityTypeProducts, EntityTypePrices, EntityTypeInventory}
	if len(got) != len(want) {
		t.Fatalf("expected %d blocked entities, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("blocked entity %d = %q, want %q", i, got[i], want[i])
		}
	}
}
