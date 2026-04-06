package oracle

import (
	"context"
	"strings"
	"testing"

	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
)

func TestQueryRunnerRejectsEmptySQL(t *testing.T) {
	t.Parallel()

	var runner QueryRunner

	err := runner.Query(context.Background(), dbsource.QuerySpec{}, func(dbsource.RowReader) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "SQL must not be empty") {
		t.Fatalf("expected empty SQL validation error, got %v", err)
	}
}
