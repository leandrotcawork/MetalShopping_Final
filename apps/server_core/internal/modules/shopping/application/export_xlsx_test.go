package application

import (
	"path/filepath"
	"testing"
)

func TestResolveExportPathAcceptsDirectoryTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	resolved, err := resolveExportPath(root, root, "shopping_market_report_run-123.xlsx")
	if err != nil {
		t.Fatalf("resolveExportPath returned error: %v", err)
	}

	expected := filepath.Join(root, "shopping_market_report_run-123.xlsx")
	if resolved != expected {
		t.Fatalf("resolved path mismatch: got %q want %q", resolved, expected)
	}
}

func TestResolveExportPathAppendsExtensionForBareFileName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	resolved, err := resolveExportPath(root, "exports/run_123", "ignored.xlsx")
	if err != nil {
		t.Fatalf("resolveExportPath returned error: %v", err)
	}

	expected := filepath.Join(root, "exports", "run_123.xlsx")
	if resolved != expected {
		t.Fatalf("resolved path mismatch: got %q want %q", resolved, expected)
	}
}

func TestResolveExportPathRejectsPathOutsideRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.xlsx")

	_, err := resolveExportPath(root, outside, "ignored.xlsx")
	if err == nil || err.Error() != "outputFilePath must be under export root" {
		t.Fatalf("unexpected error: %v", err)
	}
}
