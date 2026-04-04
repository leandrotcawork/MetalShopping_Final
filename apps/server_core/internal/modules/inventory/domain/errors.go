package domain

import "errors"

var (
	ErrPositionIDRequired      = errors.New("inventory position id is required")
	ErrTenantIDRequired        = errors.New("inventory tenant id is required")
	ErrProductIDRequired       = errors.New("inventory product id is required")
	ErrOnHandQuantityInvalid   = errors.New("inventory on_hand_quantity is invalid")
	ErrInvalidPositionStatus   = errors.New("inventory position status is invalid")
	ErrEffectiveFromRequired   = errors.New("inventory effective_from is required")
	ErrInvalidEffectiveWindow  = errors.New("inventory effective_to must be after effective_from")
	ErrInvalidOriginType       = errors.New("inventory origin type is invalid")
	ErrReasonCodeRequired      = errors.New("inventory reason code is required")
	ErrUpdatedByRequired       = errors.New("inventory updated_by is required")
	ErrProductPositionNotFound = errors.New("inventory product position not found")
)
