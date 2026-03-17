package domain

import "errors"

var (
	ErrTenantIDRequired           = errors.New("catalog tenant id is required")
	ErrSKURequired                = errors.New("catalog sku is required")
	ErrProductNameRequired        = errors.New("catalog product name is required")
	ErrInvalidProductStatus       = errors.New("catalog product status is invalid")
	ErrTaxonomyNodeIDRequired     = errors.New("catalog taxonomy node id is required")
	ErrTaxonomyNodeNameRequired   = errors.New("catalog taxonomy node name is required")
	ErrTaxonomyLevelLabelRequired = errors.New("catalog taxonomy level label is required")
	ErrInvalidTaxonomyLevel       = errors.New("catalog taxonomy level is invalid")
	ErrProductCreationDisabled    = errors.New("catalog product creation is disabled by governance")
)
