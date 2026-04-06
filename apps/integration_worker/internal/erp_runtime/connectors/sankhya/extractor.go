package sankhya

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/dbsource"
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
type Extractor struct {
	mapper *Mapper
}

func newExtractor(mapper *Mapper) *Extractor {
	return &Extractor{mapper: mapper}
}

func (e *Extractor) Extract(ctx context.Context, req erp_runtime.ExtractRequest, runner dbsource.QueryRunner) (*erp_runtime.ExtractionResult, error) {
	cfg, ok := entityConfigs[req.Entity]
	if !ok {
		return nil, fmt.Errorf("sankhya: no config for entity %q", req.Entity)
	}

	if isFixtureConnection(req.Connection) {
		rows, err := loadFixtureRows(cfg.fixtureFile)
		if err != nil {
			return nil, err
		}
		return e.extractFromFixtureRows(req, rows)
	}

	if runner == nil {
		return nil, fmt.Errorf("sankhya: live extraction requires a query runner")
	}

	spec := dbsource.QuerySpec{SQL: cfg.query}
	records := make([]*erp_runtime.RawRecord, 0, 64)
	err := runner.Query(ctx, spec, func(row dbsource.RowReader) error {
		payload, sourceID, err := e.mapper.MapRow(req.Entity, row, cfg.sourceIDKeys)
		if err != nil {
			return err
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("sankhya: marshal %s row: %w", req.Entity, err)
		}
		hash := sha256.Sum256(payloadBytes)
		records = append(records, &erp_runtime.RawRecord{
			SourceID:      sourceID,
			ConnectorType: ConnectorType,
			EntityType:    req.Entity,
			PayloadJSON:   payloadBytes,
			PayloadHash:   hex.EncodeToString(hash[:]),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &erp_runtime.ExtractionResult{
		Records:    records,
		HasMore:    false,
		NextCursor: nil,
	}, nil
}

func (e *Extractor) extractFromFixtureRows(req erp_runtime.ExtractRequest, rows []map[string]any) (*erp_runtime.ExtractionResult, error) {
	cfg, ok := entityConfigs[req.Entity]
	if !ok {
		return nil, fmt.Errorf("sankhya: no config for entity %q", req.Entity)
	}

	records := make([]*erp_runtime.RawRecord, 0, len(rows))
	for _, row := range rows {
		payload, sourceID, err := e.mapper.MapRow(req.Entity, newFixtureRowReader(row), cfg.sourceIDKeys)
		if err != nil {
			return nil, err
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("sankhya: marshal %s row: %w", req.Entity, err)
		}

		hash := sha256.Sum256(payloadBytes)
		rec := &erp_runtime.RawRecord{
			SourceID:      sourceID,
			ConnectorType: ConnectorType,
			EntityType:    req.Entity,
			PayloadJSON:   payloadBytes,
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

func queryForEntity(entity erp_runtime.EntityType) (string, error) {
	cfg, ok := entityConfigs[entity]
	if !ok {
		return "", fmt.Errorf("sankhya: no config for entity %q", entity)
	}
	return cfg.query, nil
}

func isFixtureConnection(connection erp_runtime.ExtractConnection) bool {
	if strings.TrimSpace(connection.Host) == "fixture" {
		return true
	}
	if connection.ServiceName != nil && strings.TrimSpace(*connection.ServiceName) == "fixture" {
		return true
	}
	return strings.TrimSpace(connection.Host) == "" &&
		connection.Port == 0 &&
		connection.ServiceName == nil &&
		connection.SID == nil &&
		strings.TrimSpace(connection.Username) == "" &&
		strings.TrimSpace(connection.PasswordSecretRef) == ""
}

type fixtureRowReader struct {
	values map[string]any
}

func newFixtureRowReader(values map[string]any) *fixtureRowReader {
	return &fixtureRowReader{values: values}
}

func (r *fixtureRowReader) String(name string) (string, error) {
	value, ok := r.lookup(name)
	if !ok {
		return "", fmt.Errorf("sankhya fixture row: column %q not found", name)
	}
	if value == nil {
		return "", fmt.Errorf("sankhya fixture row: column %q is null", name)
	}
	return strings.TrimSpace(fmt.Sprint(value)), nil
}

func (r *fixtureRowReader) NullString(name string) (*string, error) {
	value, ok := r.lookup(name)
	if !ok {
		return nil, fmt.Errorf("sankhya fixture row: column %q not found", name)
	}
	if value == nil {
		return nil, nil
	}
	got := strings.TrimSpace(fmt.Sprint(value))
	return &got, nil
}

func (r *fixtureRowReader) Float64(name string) (float64, error) {
	value, ok := r.lookup(name)
	if !ok {
		return 0, fmt.Errorf("sankhya fixture row: column %q not found", name)
	}
	if value == nil {
		return 0, fmt.Errorf("sankhya fixture row: column %q is null", name)
	}
	raw := strings.TrimSpace(fmt.Sprint(value))
	got, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("sankhya fixture row: column %q as float64: %w", name, err)
	}
	return got, nil
}

func (r *fixtureRowReader) NullFloat64(name string) (*float64, error) {
	value, ok := r.lookup(name)
	if !ok {
		return nil, fmt.Errorf("sankhya fixture row: column %q not found", name)
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

func (r *fixtureRowReader) Time(name string) (time.Time, error) {
	value, ok := r.lookup(name)
	if !ok {
		return time.Time{}, fmt.Errorf("sankhya fixture row: column %q not found", name)
	}
	if value == nil {
		return time.Time{}, fmt.Errorf("sankhya fixture row: column %q is null", name)
	}
	raw := strings.TrimSpace(fmt.Sprint(value))
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if got, err := time.Parse(layout, raw); err == nil {
			return got, nil
		}
	}
	return time.Time{}, fmt.Errorf("sankhya fixture row: column %q unsupported time %q", name, raw)
}

func (r *fixtureRowReader) NullTime(name string) (*time.Time, error) {
	value, ok := r.lookup(name)
	if !ok {
		return nil, fmt.Errorf("sankhya fixture row: column %q not found", name)
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

func (r *fixtureRowReader) lookup(name string) (any, bool) {
	for key, value := range r.values {
		if strings.EqualFold(key, name) {
			return value, true
		}
	}
	return nil, false
}
