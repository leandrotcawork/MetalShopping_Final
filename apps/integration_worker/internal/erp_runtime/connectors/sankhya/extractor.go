package sankhya

import (
	"context"
	"fmt"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

// entityConfig defines the Sankhya service call parameters for an entity.
type entityConfig struct {
	serviceName string
	fields      []string
}

// entityConfigs maps EntityType to the Sankhya configuration.
var entityConfigs = map[erp_runtime.EntityType]entityConfig{
	erp_runtime.EntityTypeProducts: {
		serviceName: "SankhyaW-INF-Produto",
		fields:      []string{"CODPROD", "DESCRPROD", "UNIDADE", "ATIVO", "CODVOL"},
	},
	erp_runtime.EntityTypePrices: {
		serviceName: "SankhyaW-INF-Preco",
		fields:      []string{"NUTAB", "CODPROD", "VLRVENDA", "MOEDA"},
	},
	erp_runtime.EntityTypeCosts: {
		serviceName: "SankhyaW-INF-Produto",
		fields:      []string{"CODPROD", "VLRCUSTO_REP", "VLRCUSTOMEDIO"},
	},
	erp_runtime.EntityTypeInventory: {
		serviceName: "SankhyaW-INF-Estoque",
		fields:      []string{"CODPROD", "CODLOCAL", "ESTOQUE", "RESERVADO"},
	},
	erp_runtime.EntityTypeSales: {
		serviceName: "SankhyaW-INF-CabecalhoNota",
		fields:      []string{"NUNOTA", "CODPARC", "DTNEG", "VLRNOTA", "TIPMOV"},
	},
	erp_runtime.EntityTypePurchases: {
		serviceName: "SankhyaW-INF-CabecalhoNota",
		fields:      []string{"NUNOTA", "CODPARC", "DTNEG", "VLRNOTA", "TIPMOV"},
	},
	erp_runtime.EntityTypeCustomers: {
		serviceName: "SankhyaW-INF-Parceiro",
		fields:      []string{"CODPARC", "NOMEPARC", "CGC_CPF", "EMAIL", "CLIENTE"},
	},
	erp_runtime.EntityTypeSuppliers: {
		serviceName: "SankhyaW-INF-Parceiro",
		fields:      []string{"CODPARC", "NOMEPARC", "CGC_CPF", "EMAIL", "FORNECEDOR"},
	},
}

// Extractor fetches raw records from Sankhya ERP.
type Extractor struct{}

func newExtractor() *Extractor { return &Extractor{} }

func (e *Extractor) Extract(ctx context.Context, req erp_runtime.ExtractRequest) (*erp_runtime.ExtractionResult, error) {
	cfg, ok := entityConfigs[req.Entity]
	if !ok {
		return nil, fmt.Errorf("sankhya: no config for entity %q", req.Entity)
	}
	// v1 stub: return empty result.
	// Real implementation would call Sankhya HTTP API:
	//   POST https://{host}/mge/service.sbr?serviceName={cfg.serviceName}
	//   with authentication token and field list.
	_ = cfg // suppress unused warning
	return &erp_runtime.ExtractionResult{
		Records: []*erp_runtime.RawRecord{},
		HasMore: false,
	}, nil
}
