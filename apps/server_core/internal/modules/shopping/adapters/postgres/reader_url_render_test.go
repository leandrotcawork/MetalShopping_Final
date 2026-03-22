package postgres

import (
	"strings"
	"testing"
)

func TestDeriveSearchURLFromRuntimeConfig_UsesSearchURLTemplate(t *testing.T) {
	raw := `{"searchUrlTemplate":"https://www.condec.com.br/loja/busca.php?palavra_busca={term}"}`
	got := deriveSearchURLFromRuntimeConfig(raw, "CONDEC", "prd_x", "7894200129684")
	want := "https://www.condec.com.br/loja/busca.php?palavra_busca=7894200129684"
	if got != want {
		t.Fatalf("unexpected url. got=%q want=%q", got, want)
	}
}

func TestDeriveSearchURLFromRuntimeConfig_UsesVTEXBaseURLWithParams(t *testing.T) {
	raw := `{
	  "baseUrl":"https://www.telhanorte.com.br/_v/segment/graphql/v1",
	  "operationName":"productSearchV3",
	  "sha256Hash":"31d3fa494df1fc41efef6d16dd96a96e6911b8aed7a037868699a1f3f4d365de",
	  "skusFilter":"ALL_AVAILABLE",
	  "toN":39,
	  "includeVariant":true
	}`
	got := deriveSearchURLFromRuntimeConfig(raw, "TELHA_NORTE", "prd_y", "7894200129684")
	if got == "" {
		t.Fatalf("expected non-empty url")
	}
	if !strings.HasPrefix(got, "https://www.telhanorte.com.br/_v/segment/graphql/v1?") {
		t.Fatalf("unexpected prefix: %q", got)
	}
	if !strings.Contains(got, "operationName=productSearchV3") {
		t.Fatalf("missing operationName in %q", got)
	}
	if !strings.Contains(got, "variables=") {
		t.Fatalf("missing variables in %q", got)
	}
}

func TestRenderSearchURLTemplate_UsesSupplierAndProductPlaceholders(t *testing.T) {
	got := renderSearchURLTemplate(
		"https://example.com/{supplier_code}/{product_id}?q={term}&mode={lookup_mode}",
		"ABC",
		"prd_123",
		"1168.C.LNK",
	)
	want := "https://example.com/ABC/prd_123?q=1168.C.LNK&mode=REFERENCE"
	if got != want {
		t.Fatalf("unexpected rendered template. got=%q want=%q", got, want)
	}
}
