package staging

import (
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// ValidationStatus represents whether a staging record passed validation.
type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

// StagingRecord holds a normalized, validated ERP record ready for reconciliation.
type StagingRecord struct {
	StagingID        string
	TenantID         string
	RunID            string
	RawID            string
	EntityType       types.EntityType
	SourceID         string
	BatchOrdinal     int
	NormalizedJSON   []byte
	ValidationStatus ValidationStatus
	ValidationErrors []string // nil or empty if valid
	NormalizedAt     time.Time
}
