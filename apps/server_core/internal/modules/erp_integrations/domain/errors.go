package domain

import "errors"

var (
	ErrInstanceNotFound      = errors.New("erp integration instance not found")
	ErrRunNotFound           = errors.New("erp sync run not found")
	ErrReviewItemNotFound    = errors.New("erp review item not found")
	ErrActiveInstanceExists  = errors.New("tenant already has an active ERP integration instance")
	ErrIntegrationDisabled   = errors.New("ERP integration is disabled for this tenant")
	ErrInvalidConnectorType  = errors.New("invalid ERP connector type")
	ErrInvalidEntityType     = errors.New("invalid ERP entity type")
	ErrInvalidRunMode        = errors.New("invalid run mode")
	ErrEmptyEnabledEntities  = errors.New("enabled entities must not be empty")
	ErrEmptyEntityScope      = errors.New("entity scope must not be empty")
	ErrEmptyDisplayName      = errors.New("display name must not be empty")
	ErrEmptyTenantID         = errors.New("tenant ID must not be empty")
	ErrReviewAlreadyResolved = errors.New("review item is already resolved or dismissed")
	ErrAutoPromotionDisabled  = errors.New("auto-promotion is disabled for this tenant")
	ErrInvalidInstanceStatus  = errors.New("invalid ERP integration instance status")
)
