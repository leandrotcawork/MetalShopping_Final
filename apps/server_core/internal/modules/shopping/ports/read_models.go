package ports

import (
	"context"
	"time"
)

type Summary struct {
	TotalRuns     int64
	RunningRuns   int64
	CompletedRuns int64
	FailedRuns    int64
	LastRunAt     *time.Time
}

type Run struct {
	RunID          string
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	ProcessedItems int64
	TotalItems     int64
	Notes          string
}

type RunList struct {
	Rows   []Run
	Offset int64
	Limit  int64
	Total  int64
}

type ProductLatest struct {
	ProductID     string
	RunID         string
	ObservedAt    time.Time
	SellerName    string
	Channel       string
	ObservedPrice float64
	Currency      string
}

type RunListFilter struct {
	Status string
	Limit  int64
	Offset int64
}

type Reader interface {
	GetSummary(ctx context.Context, tenantID string) (Summary, error)
	ListRuns(ctx context.Context, tenantID string, filter RunListFilter) (RunList, error)
	GetRun(ctx context.Context, tenantID, runID string) (Run, error)
	GetProductLatest(ctx context.Context, tenantID, productID string) (ProductLatest, error)
}
