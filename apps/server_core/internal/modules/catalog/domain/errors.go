package domain

import "errors"

var (
	ErrTenantIDRequired        = errors.New("catalog tenant id is required")
	ErrSKURequired             = errors.New("catalog sku is required")
	ErrProductNameRequired     = errors.New("catalog product name is required")
	ErrInvalidProductStatus    = errors.New("catalog product status is invalid")
	ErrProductCreationDisabled = errors.New("catalog product creation is disabled by governance")
)
