package dbsource

import (
	"context"
	"time"
)

// QuerySpec describes a single query to execute against a source database.
type QuerySpec struct {
	SQL     string
	Args    []any
	Timeout time.Duration
}

// QueryRunner executes a query and streams each row through the supplied callback.
type QueryRunner interface {
	Query(ctx context.Context, spec QuerySpec, fn func(RowReader) error) error
	Close() error
}
