package dbsource

import "time"

// RowReader provides typed accessors for query result rows.
type RowReader interface {
	String(idx int) string
	NullString(idx int) (string, bool)
	Float64(idx int) float64
	NullFloat64(idx int) (float64, bool)
	Time(idx int) time.Time
	NullTime(idx int) (time.Time, bool)
}
