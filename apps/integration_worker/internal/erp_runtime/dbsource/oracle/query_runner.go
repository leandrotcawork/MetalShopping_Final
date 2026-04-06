package oracle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
)

// QueryRunner executes Oracle queries through a run-scoped database handle.
type QueryRunner struct {
	db *sql.DB
}

// NewQueryRunner opens and validates an Oracle connection.
func NewQueryRunner(cfg Config) (*QueryRunner, error) {
	dsn, err := cfg.ConnectString()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("godror", dsn)
	if err != nil {
		return nil, fmt.Errorf("open oracle connection: %w", err)
	}

	pingCtx := context.Background()
	cancel := func() {}
	if cfg.ConnectTimeoutSec > 0 {
		pingCtx, cancel = context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSec)*time.Second)
	}
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping oracle connection: %w", err)
	}

	return &QueryRunner{db: db}, nil
}

// Query executes a single SQL statement and streams each row to fn.
func (r *QueryRunner) Query(ctx context.Context, spec dbsource.QuerySpec, fn func(dbsource.RowReader) error) error {
	if strings.TrimSpace(spec.SQL) == "" {
		return errors.New("oracle query runner: SQL must not be empty")
	}
	if fn == nil {
		return errors.New("oracle query runner: callback must not be nil")
	}
	if r == nil || r.db == nil {
		return errors.New("oracle query runner: database is not configured")
	}

	queryCtx := ctx
	cancel := func() {}
	if spec.Timeout > 0 {
		queryCtx, cancel = context.WithTimeout(ctx, spec.Timeout)
	}
	defer cancel()

	rows, err := r.db.QueryContext(queryCtx, spec.SQL, spec.Args...)
	if err != nil {
		return fmt.Errorf("oracle query runner: query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("oracle query runner: columns: %w", err)
	}

	scanTargets := make([]any, len(columns))
	scanValues := make([]any, len(columns))
	for i := range scanValues {
		scanTargets[i] = &scanValues[i]
	}

	for rows.Next() {
		for i := range scanValues {
			scanValues[i] = nil
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("oracle query runner: scan: %w", err)
		}

		reader, err := newRowReader(columns, scanValues)
		if err != nil {
			return err
		}
		if err := fn(reader); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("oracle query runner: rows: %w", err)
	}

	return nil
}

// Close releases the underlying database handle.
func (r *QueryRunner) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

type rowReader struct {
	values map[string]any
}

func newRowReader(columns []string, values []any) (rowReader, error) {
	reader := rowReader{values: make(map[string]any, len(columns))}
	seen := make(map[string]string, len(columns))
	for i, column := range columns {
		columnName := strings.TrimSpace(column)
		if columnName == "" {
			return rowReader{}, fmt.Errorf("oracle row reader: column %d name is empty", i+1)
		}
		folded := strings.ToLower(columnName)
		if prev, ok := seen[folded]; ok {
			return rowReader{}, fmt.Errorf("oracle row reader: duplicate column %q collides with %q", columnName, prev)
		}
		seen[folded] = columnName
		reader.values[columnName] = cloneRowValue(values[i])
	}
	return reader, nil
}

func cloneRowValue(value any) any {
	switch v := value.(type) {
	case []byte:
		if v == nil {
			return nil
		}
		dup := make([]byte, len(v))
		copy(dup, v)
		return dup
	default:
		return value
	}
}

func (r rowReader) lookup(name string) (any, bool) {
	for column, value := range r.values {
		if strings.EqualFold(column, name) {
			return value, true
		}
	}
	return nil, false
}

func (r rowReader) String(name string) (string, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", fmt.Errorf("oracle row reader: column %q is null", name)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	case *time.Time:
		if v == nil {
			return "", fmt.Errorf("oracle row reader: column %q is null", name)
		}
		return v.Format(time.RFC3339Nano), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprint(v), nil
	}
}

func (r rowReader) NullString(name string) (*string, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	got, err := r.String(name)
	if err != nil {
		return nil, err
	}
	return &got, nil
}

func (r rowReader) Float64(name string) (float64, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, fmt.Errorf("oracle row reader: column %q is null", name)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case []byte:
		return parseFloat(name, string(v))
	case string:
		return parseFloat(name, v)
	case fmt.Stringer:
		return parseFloat(name, v.String())
	default:
		return parseFloat(name, fmt.Sprint(v))
	}
}

func (r rowReader) NullFloat64(name string) (*float64, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	got, err := r.Float64(name)
	if err != nil {
		return nil, err
	}
	return &got, nil
}

func (r rowReader) Time(name string) (time.Time, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return time.Time{}, err
	}
	if value == nil {
		return time.Time{}, fmt.Errorf("oracle row reader: column %q is null", name)
	}

	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return time.Time{}, fmt.Errorf("oracle row reader: column %q is null", name)
		}
		return *v, nil
	case string:
		return parseTime(name, v)
	case []byte:
		return parseTime(name, string(v))
	default:
		return parseTime(name, fmt.Sprint(v))
	}
}

func (r rowReader) NullTime(name string) (*time.Time, error) {
	value, err := r.mustValue(name)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	got, err := r.Time(name)
	if err != nil {
		return nil, err
	}
	return &got, nil
}

func (r rowReader) mustValue(name string) (any, error) {
	value, ok := r.lookup(name)
	if !ok {
		return nil, fmt.Errorf("oracle row reader: column %q not found", name)
	}
	return value, nil
}

func parseFloat(name, raw string) (float64, error) {
	got, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("oracle row reader: column %q as float64: %w", name, err)
	}
	return got, nil
}

func parseTime(name, raw string) (time.Time, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return time.Time{}, fmt.Errorf("oracle row reader: column %q as time: empty value", name)
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if got, err := time.Parse(layout, candidate); err == nil {
			return got, nil
		}
	}

	return time.Time{}, fmt.Errorf("oracle row reader: column %q as time: unsupported format %q", name, raw)
}
