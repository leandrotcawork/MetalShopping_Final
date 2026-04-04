package sankhya

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
)

//go:embed testdata/*.json
var fixtureFS embed.FS

type entityConfig struct {
	serviceName  string
	query        string
	fixtureFile  string
	sourceIDKeys []string
}

var entityConfigs = map[erp_runtime.EntityType]entityConfig{
	erp_runtime.EntityTypeProducts: {
		serviceName:  "SankhyaW-INF-Produto",
		query:        productsSnapshotQuery,
		fixtureFile:  "products_fixture.json",
		sourceIDKeys: []string{"CODPROD"},
	},
	erp_runtime.EntityTypePrices: {
		serviceName:  "SankhyaW-INF-Preco",
		query:        pricesSnapshotQuery,
		fixtureFile:  "prices_fixture.json",
		sourceIDKeys: []string{"NUTAB", "CODPROD", "CODTAB"},
	},
	erp_runtime.EntityTypeInventory: {
		serviceName:  "SankhyaW-INF-Estoque",
		query:        inventorySnapshotQuery,
		fixtureFile:  "inventory_fixture.json",
		sourceIDKeys: []string{"CODPROD", "CODEMP", "CODLOCAL"},
	},
}

const (
	productsSnapshotQuery = `
SELECT
  PRO.CODPROD,
  PRO.DESCRPROD,
  PRO.MARCA,
  PRO.REFERENCIA,
  PRO.REFFORN,
  PRO.ATIVO,
  PRO.CODVOL,
  PRO.CODGRUPOPROD,
  PRO.AD_STATUS,
  PRO.AD_COMPETITIVO
FROM TGFPRO PRO
ORDER BY PRO.CODPROD`

	pricesSnapshotQuery = `
SELECT
  TAB.NUTAB,
  TAB.CODTAB,
  TAB.NOMETAB,
  NTA.DTVIGOR,
  EXC.CODPROD,
  EXC.VLRVENDA
FROM TGFTAB TAB
LEFT JOIN TGFNTA NTA ON NTA.NUTAB = TAB.NUTAB
LEFT JOIN TGFEXC EXC ON EXC.NUTAB = TAB.NUTAB
ORDER BY TAB.NUTAB, NTA.DTVIGOR, EXC.CODPROD`

	inventorySnapshotQuery = `
SELECT
  EST.CODPROD,
  EST.CODEMP,
  EST.CODLOCAL,
  EST.ESTOQUE,
  EST.RESERVADO,
  (NVL(EST.ESTOQUE, 0) - NVL(EST.RESERVADO, 0)) AS RAW_AVAILABLE_POSITION
FROM TGFEST EST
ORDER BY EST.CODPROD, EST.CODEMP, EST.CODLOCAL`

	salesSnapshotQuery = `
SELECT
  CAB.NUNOTA,
  CAB.CODPARC,
  CAB.DTNEG,
  CAB.VLRNOTA,
  CAB.TIPMOV
FROM TGFCAB CAB
ORDER BY CAB.NUNOTA`

	customersSnapshotQuery = `
SELECT
  PAR.CODPARC,
  PAR.NOMEPARC,
  PAR.CGC_CPF,
  PAR.EMAIL,
  PAR.CLIENTE
FROM TGFPAR PAR
ORDER BY PAR.CODPARC`

	suppliersSnapshotQuery = `
SELECT
  PAR.CODPARC,
  PAR.NOMEPARC,
  PAR.CGC_CPF,
  PAR.EMAIL,
  PAR.FORNECEDOR
FROM TGFPAR PAR
ORDER BY PAR.CODPARC`
)

// Extractor fetches raw records from Sankhya ERP.
type Extractor struct{}

func newExtractor() *Extractor { return &Extractor{} }

func (e *Extractor) Extract(ctx context.Context, req erp_runtime.ExtractRequest) (*erp_runtime.ExtractionResult, error) {
	cfg, ok := entityConfigs[req.Entity]
	if !ok {
		return nil, fmt.Errorf("sankhya: no config for entity %q", req.Entity)
	}

	if strings.TrimSpace(req.ConnectionRef) == "" {
		return nil, fmt.Errorf("sankhya: connectionRef must not be empty")
	}

	u, err := url.Parse(req.ConnectionRef)
	if err != nil {
		return nil, fmt.Errorf("sankhya: parse connectionRef: %w", err)
	}

	switch u.Scheme {
	case "fixture":
		rows, err := loadFixtureRows(cfg.fixtureFile)
		if err != nil {
			return nil, err
		}
		return e.extractFromRows(req, rows)
	case "sankhya":
		return nil, fmt.Errorf("sankhya: live extraction is not wired yet; use fixture:// for deterministic connector tests")
	default:
		return nil, fmt.Errorf("sankhya: unsupported connectionRef scheme %q", u.Scheme)
	}
}

func (e *Extractor) extractFromRows(req erp_runtime.ExtractRequest, rows []map[string]any) (*erp_runtime.ExtractionResult, error) {
	cfg, ok := entityConfigs[req.Entity]
	if !ok {
		return nil, fmt.Errorf("sankhya: no config for entity %q", req.Entity)
	}

	records := make([]*erp_runtime.RawRecord, 0, len(rows))
	for _, row := range rows {
		payload, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("sankhya: marshal %s row: %w", req.Entity, err)
		}

		sourceID := sourceIDForRow(row, cfg.sourceIDKeys)
		if sourceID == "" {
			return nil, fmt.Errorf("sankhya: missing source key fields for entity %q", req.Entity)
		}

		hash := sha256.Sum256(payload)
		rec := &erp_runtime.RawRecord{
			SourceID:      sourceID,
			ConnectorType: ConnectorType,
			EntityType:    req.Entity,
			PayloadJSON:   payload,
			PayloadHash:   hex.EncodeToString(hash[:]),
		}
		records = append(records, rec)
	}

	return &erp_runtime.ExtractionResult{
		Records:    records,
		HasMore:    false,
		NextCursor: nil,
	}, nil
}

func loadFixtureRows(fileName string) ([]map[string]any, error) {
	data, err := fixtureFS.ReadFile("testdata/" + fileName)
	if err != nil {
		return nil, fmt.Errorf("sankhya: read fixture %s: %w", fileName, err)
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("sankhya: unmarshal fixture %s: %w", fileName, err)
	}
	return rows, nil
}

func sourceIDForRow(row map[string]any, keys []string) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(fmt.Sprint(row[key]))
		if value == "" || value == "<nil>" {
			return ""
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, ":")
}

func queryForEntity(entity erp_runtime.EntityType) (string, error) {
	cfg, ok := entityConfigs[entity]
	if !ok {
		return "", fmt.Errorf("sankhya: no config for entity %q", entity)
	}
	return cfg.query, nil
}
