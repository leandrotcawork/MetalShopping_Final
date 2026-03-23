package ports

import (
	"context"
	"time"
)

type BlockStatus string

const (
	BlockStatusOK       BlockStatus = "OK"
	BlockStatusNotReady BlockStatus = "NOT_READY"
	BlockStatusError    BlockStatus = "ERROR"
)

type BlockError struct {
	Code    string
	Message string
	Details map[string]any
}

type HomeBlock struct {
	Status BlockStatus
	Data   map[string]any
	Error  *BlockError
}

type HomeSnapshot struct {
	RequestedID string
	ResolvedID  *string
	AsOf        *time.Time
	ServedAt    time.Time
}

type HomeBlocks struct {
	KpisOperational       HomeBlock
	KpisAnalytics         HomeBlock
	KpisProducts          HomeBlock
	ActionsToday          HomeBlock
	AlertsPrioritarios    HomeBlock
	PortfolioDistribution HomeBlock
	Timeline              HomeBlock
}

type Home struct {
	SchemaVersion string
	Snapshot      HomeSnapshot
	Blocks        HomeBlocks
}

type HomeReader interface {
	GetHome(ctx context.Context, tenantID string, requestedSnapshotID string) (Home, error)
}
