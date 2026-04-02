package domain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// IntegrationInstance.ValidateForWrite
// ---------------------------------------------------------------------------

func TestIntegrationInstance_ValidateForWrite_Valid(t *testing.T) {
	inst := &IntegrationInstance{
		InstanceID:      "inst-1",
		TenantID:        "tenant-abc",
		ConnectorType:   ConnectorTypeSankhya,
		DisplayName:     "My ERP",
		ConnectionRef:   "conn-ref-1",
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
		EnabledEntities: []EntityType{EntityTypeProducts, EntityType("unknown_entity")},
		Status:          InstanceStatusActive,
	}
	if err := inst.ValidateForWrite(); err != ErrInvalidEntityType {
		t.Fatalf("expected ErrInvalidEntityType, got: %v", err)
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
