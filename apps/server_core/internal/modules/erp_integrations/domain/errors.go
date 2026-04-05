package domain

import "errors"

var (
	ErrInstanceNotFound              = errors.New("erp integration instance not found")
	ErrRunNotFound                   = errors.New("erp sync run not found")
	ErrStagingRecordNotFound         = errors.New("erp staging record not found")
	ErrReviewItemNotFound            = errors.New("erp review item not found")
	ErrActiveInstanceExists          = errors.New("tenant already has an active ERP integration instance")
	ErrIntegrationDisabled           = errors.New("ERP integration is disabled for this tenant")
	ErrInvalidConnectorType          = errors.New("invalid ERP connector type")
	ErrInvalidEntityType             = errors.New("invalid ERP entity type")
	ErrInvalidRunMode                = errors.New("invalid run mode")
	ErrEmptyEnabledEntities          = errors.New("enabled entities must not be empty")
	ErrEmptyEntityScope              = errors.New("entity scope must not be empty")
	ErrEmptyDisplayName              = errors.New("display name must not be empty")
	ErrEmptyTenantID                 = errors.New("tenant ID must not be empty")
	ErrEmptyOracleHost               = errors.New("Oracle host must not be empty")
	ErrEmptyOracleUsername           = errors.New("Oracle username must not be empty")
	ErrEmptyPasswordSecretRef        = errors.New("password secret ref must not be empty")
	ErrInvalidConnectionKind         = errors.New("invalid ERP connection kind")
	ErrInvalidOracleConnectionTarget = errors.New("Oracle connection must set exactly one of service_name or sid")
	ErrInvalidOraclePort             = errors.New("Oracle port must be positive")
	ErrReviewAlreadyResolved         = errors.New("review item is already resolved or dismissed")
	ErrAutoPromotionDisabled         = errors.New("auto-promotion is disabled for this tenant")
	ErrInvalidInstanceStatus         = errors.New("invalid ERP integration instance status")
)
