package domain

import "errors"

var (
	ErrTenantIDRequired            = errors.New("pricing tenant id is required")
	ErrPriceIDRequired             = errors.New("pricing price id is required")
	ErrProductIDRequired           = errors.New("pricing product id is required")
	ErrCurrencyCodeRequired        = errors.New("pricing currency code is required")
	ErrInvalidCurrencyCode         = errors.New("pricing currency code is invalid")
	ErrPriceAmountInvalid          = errors.New("pricing price amount must be non-negative")
	ErrCostBasisAmountInvalid      = errors.New("pricing cost basis amount must be non-negative")
	ErrMarginFloorValueInvalid     = errors.New("pricing margin floor value must be non-negative")
	ErrInvalidPricingStatus        = errors.New("pricing status is invalid")
	ErrInvalidOriginType           = errors.New("pricing origin type is invalid")
	ErrReasonCodeRequired          = errors.New("pricing reason code is required")
	ErrUpdatedByRequired           = errors.New("pricing updated by is required")
	ErrEffectiveFromRequired       = errors.New("pricing effective_from is required")
	ErrInvalidEffectiveWindow      = errors.New("pricing effective window is invalid")
	ErrPriceBelowMarginFloor       = errors.New("pricing price is below the governed margin floor")
	ErrManualPriceOverrideDisabled = errors.New("pricing manual price override is disabled by governance")
	ErrProductPriceNotFound        = errors.New("pricing product price not found")
)
