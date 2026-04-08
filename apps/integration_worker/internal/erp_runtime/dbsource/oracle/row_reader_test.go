package oracle

import (
	"strings"
	"testing"
	"time"
)

func TestRowReaderNullString(t *testing.T) {
	t.Parallel()

	reader := mustRowReader(t, []string{"CUSTOMER_NAME"}, []any{nil})

	got, err := reader.NullString("customer_name")
	if err != nil {
		t.Fatalf("NullString returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("NullString() = %v, want nil", got)
	}
}

func TestRowReaderMissingColumn(t *testing.T) {
	t.Parallel()

	reader := mustRowReader(t, []string{"KNOWN"}, []any{"value"})

	if _, err := reader.String("missing"); err == nil {
		t.Fatal("expected missing column error, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing column error, got %v", err)
	}
}

func TestRowReaderFloat64Conversions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		value    any
		expected float64
	}{
		{name: "float64", value: float64(42.5), expected: 42.5},
		{name: "float32", value: float32(12.25), expected: 12.25},
		{name: "int64", value: int64(7), expected: 7},
		{name: "string", value: "18.75", expected: 18.75},
		{name: "bytes", value: []byte("9.5"), expected: 9.5},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := mustRowReader(t, []string{"AMOUNT"}, []any{tc.value})

			got, err := reader.Float64("amount")
			if err != nil {
				t.Fatalf("Float64 returned error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("Float64() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestRowReaderTimeConversions(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, time.April, 5, 12, 34, 56, 789000000, time.UTC)
	cases := []struct {
		name     string
		value    any
		expected time.Time
	}{
		{name: "time", value: base, expected: base},
		{name: "time-pointer", value: &base, expected: base},
		{name: "rfc3339", value: "2026-04-05T12:34:56.789Z", expected: base},
		{name: "bytes", value: []byte("2026-04-05T12:34:56.789Z"), expected: base},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := mustRowReader(t, []string{"CREATED_AT"}, []any{tc.value})

			got, err := reader.Time("created_at")
			if err != nil {
				t.Fatalf("Time returned error: %v", err)
			}
			if !got.Equal(tc.expected) {
				t.Fatalf("Time() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestRowReaderStringFormatsTime(t *testing.T) {
	t.Parallel()

	value := time.Date(2026, time.April, 5, 12, 34, 56, 789000000, time.UTC)
	reader := mustRowReader(t, []string{"CREATED_AT"}, []any{value})

	got, err := reader.String("created_at")
	if err != nil {
		t.Fatalf("String returned error: %v", err)
	}
	want := value.Format(time.RFC3339Nano)
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestNewRowReaderRejectsDuplicateColumns(t *testing.T) {
	t.Parallel()

	_, err := newRowReader([]string{"ID", "id"}, []any{int64(1), int64(2)})
	if err == nil {
		t.Fatal("expected duplicate column error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate column") {
		t.Fatalf("expected duplicate column error, got %v", err)
	}
}

func mustRowReader(t *testing.T, columns []string, values []any) rowReader {
	t.Helper()

	reader, err := newRowReader(columns, values)
	if err != nil {
		t.Fatalf("newRowReader returned error: %v", err)
	}
	return reader
}
