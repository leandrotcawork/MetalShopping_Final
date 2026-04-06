package dbsource

import "time"

// RowReader provides typed accessors for query result rows.
type RowReader interface {
	String(name string) (string, error)
	NullString(name string) (*string, error)
	Float64(name string) (float64, error)
	NullFloat64(name string) (*float64, error)
	Time(name string) (time.Time, error)
	NullTime(name string) (*time.Time, error)
}
