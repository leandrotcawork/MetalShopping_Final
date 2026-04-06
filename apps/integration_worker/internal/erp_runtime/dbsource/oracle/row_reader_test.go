package oracle

import "testing"

func TestRowReaderNullString(t *testing.T) {
	t.Parallel()

	reader := rowReader{
		values: map[string]any{
			"CUSTOMER_NAME": nil,
		},
	}

	got, err := reader.NullString("customer_name")
	if err != nil {
		t.Fatalf("NullString returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("NullString() = %v, want nil", got)
	}
}
