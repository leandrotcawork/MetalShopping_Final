package ports

import (
	"context"
	"time"
)

type Summary struct {
	ProductCount          int64
	ActiveProductCount    int64
	PricedProductCount    int64
	InventoryTrackedCount int64
	LastUpdated           time.Time
}

type SummaryReader interface {
	GetSummary(ctx context.Context, tenantID string) (Summary, error)
}
