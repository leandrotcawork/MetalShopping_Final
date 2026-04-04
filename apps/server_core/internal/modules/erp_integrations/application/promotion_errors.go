package application

import (
	"errors"

	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

var (
	ErrRelatedProductNotPromoted       = errors.New("related product promotion is not resolved")
	ErrPriceContextMappingNotFound     = ports.ErrPriceContextMappingNotFound
	ErrPricePromotionNotConfigured     = errors.New("price promotion is not configured")
	ErrInventoryPromotionNotConfigured = errors.New("inventory promotion is not configured")
)

const (
	relatedProductBlockedReasonCode = "ERP_RELATED_PRODUCT_NOT_PROMOTED"
	relatedProductBlockedSummary    = "related product promotion is unresolved or blocked"
	relatedProductBlockedAction     = "promote the product first or resolve the product review before rerunning promotion"

	priceContextMappingMissingReasonCode = "ERP_PRICE_CONTEXT_MAPPING_MISSING"
	priceContextMappingMissingSummary    = "price context mapping is missing"
	priceContextMappingMissingAction     = "configure a source price-table mapping before rerunning promotion"
)
