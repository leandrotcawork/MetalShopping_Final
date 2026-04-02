package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
)

type stubAutoPromoGuard struct {
	err error
}

func (g *stubAutoPromoGuard) CheckAutoPromotion(context.Context, string) error {
	return g.err
}

type stubPromotionReader struct {
	staging map[string]*domain.StagingRecord
	err     error
}

func (r *stubPromotionReader) GetStagingRecord(_ context.Context, _, stagingID string) (*domain.StagingRecord, error) {
	if r.err != nil {
		return nil, r.err
	}
	if record, ok := r.staging[stagingID]; ok {
		return record, nil
	}
	return nil, domain.ErrStagingRecordNotFound
}

type stubRunRepo struct {
	runs map[string]*domain.SyncRun
}

func (r *stubRunRepo) Create(_ context.Context, _ *domain.SyncRun) error {
	return nil
}

func (r *stubRunRepo) Get(_ context.Context, _, runID string) (*domain.SyncRun, error) {
	if r == nil {
		return nil, domain.ErrRunNotFound
	}
	if run, ok := r.runs[runID]; ok {
		return run, nil
	}
	return nil, domain.ErrRunNotFound
}

func (r *stubRunRepo) List(_ context.Context, _, _ string, _, _ int) ([]*domain.SyncRun, error) {
	return nil, nil
}

type recordedPromotion struct {
	reconciliationID string
	canonicalID      string
}

type recordedFailure struct {
	reconciliationID string
	reasonCode       string
	warningDetails   *string
}

type stubPromotionRepository struct {
	claimResult bool
	claimErr    error

	claimCalls     []string
	promoted       []recordedPromotion
	failed         []recordedFailure
	reviewRequired []recordedFailure
}

func (r *stubPromotionRepository) ListPromotableResults(context.Context, string, int) ([]*domain.ReconciliationResult, error) {
	return nil, nil
}

func (r *stubPromotionRepository) ListAllPendingPromotion(context.Context, int) ([]*domain.ReconciliationResult, error) {
	return nil, nil
}

func (r *stubPromotionRepository) ClaimForPromotion(_ context.Context, _, reconciliationID string) (bool, error) {
	r.claimCalls = append(r.claimCalls, reconciliationID)
	return r.claimResult, r.claimErr
}

func (r *stubPromotionRepository) MarkPromoted(_ context.Context, _, reconciliationID, canonicalID string) error {
	r.promoted = append(r.promoted, recordedPromotion{reconciliationID: reconciliationID, canonicalID: canonicalID})
	return nil
}

func (r *stubPromotionRepository) MarkPromotionFailed(_ context.Context, _, reconciliationID, reasonCode string, warningDetails *string) error {
	r.failed = append(r.failed, recordedFailure{reconciliationID: reconciliationID, reasonCode: reasonCode, warningDetails: warningDetails})
	return nil
}

func (r *stubPromotionRepository) MarkReviewRequired(_ context.Context, _, reconciliationID, reasonCode string, warningDetails *string) error {
	r.reviewRequired = append(r.reviewRequired, recordedFailure{reconciliationID: reconciliationID, reasonCode: reasonCode, warningDetails: warningDetails})
	return nil
}

type recordingProductWriter struct {
	traceID     string
	result      *domain.ReconciliationResult
	run         *domain.SyncRun
	input       ports.ProductPromotionInput
	canonicalID string
	err         error
}

func (w *recordingProductWriter) PromoteProduct(_ context.Context, traceID string, result *domain.ReconciliationResult, run *domain.SyncRun, input ports.ProductPromotionInput) (string, error) {
	w.traceID = traceID
	w.result = result
	w.run = run
	w.input = input
	if w.err != nil {
		return "", w.err
	}
	if w.canonicalID != "" {
		return w.canonicalID, nil
	}
	return "prd_test", nil
}

func TestPromotionConsumerPromotesProductSuccessfully(t *testing.T) {
	reconRepo := &stubPromotionRepository{claimResult: true}
	stagingRepo := &stubPromotionReader{
		staging: map[string]*domain.StagingRecord{
			"stg_1": {
				StagingID:        "stg_1",
				TenantID:         "tenant-1",
				RunID:            "run_1",
				RawID:            "raw_1",
				EntityType:       domain.EntityTypeProducts,
				SourceID:         "src_1",
				NormalizedJSON:   []byte(`{"pn_interno":"PN-001","descricao":"Galvanized steel sheet","marca":"Acme","tipo_estoque":"standard","taxonomy_node_id":"txn_leaf_1","ativo":true,"reference":"REF-001","ean":"789000000001","identifiers":[{"identifier_type":"reference","identifier_value":"REF-001","source_system":"erp","is_primary":false}]}`),
				ValidationStatus: "valid",
				NormalizedAt:     time.Now().UTC(),
			},
		},
	}
	runRepo := &stubRunRepo{
		runs: map[string]*domain.SyncRun{
			"run_1": {
				RunID:         "run_1",
				TenantID:      "tenant-1",
				InstanceID:    "inst_1",
				ConnectorType: domain.ConnectorTypeSankhya,
			},
		},
	}
	writer := &recordingProductWriter{canonicalID: "prd_123"}
	consumer := NewPromotionConsumer(reconRepo, &stubAutoPromoGuard{}, NewProductPromotion(stagingRepo, runRepo, writer))

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		RunID:            "run_1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypeProducts,
		SourceID:         "src_1",
		ReconciledAt:     time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC),
		PromotionStatus:  domain.PromotionStatusPending,
	}

	consumer.promoteOne(context.Background(), result)

	if got := len(reconRepo.claimCalls); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}
	if got := len(reconRepo.promoted); got != 1 {
		t.Fatalf("expected 1 promoted call, got %d", got)
	}
	if reconRepo.promoted[0].canonicalID != "prd_123" {
		t.Fatalf("expected canonical id prd_123, got %s", reconRepo.promoted[0].canonicalID)
	}
	if got := len(reconRepo.failed); got != 0 {
		t.Fatalf("expected no failed rows, got %d", got)
	}
	if got := len(reconRepo.reviewRequired); got != 0 {
		t.Fatalf("expected no review-required rows, got %d", got)
	}
	if writer.run == nil || writer.run.RunID != "run_1" {
		t.Fatalf("expected run lookup to be passed to writer, got %#v", writer.run)
	}
	if writer.traceID != "erp-promotion:rec_1" {
		t.Fatalf("expected promotion trace id, got %s", writer.traceID)
	}
	if writer.input.SKU != "PN-001" {
		t.Fatalf("expected sku PN-001, got %s", writer.input.SKU)
	}
	if writer.input.Name != "Galvanized steel sheet" {
		t.Fatalf("expected translated name from descricao, got %s", writer.input.Name)
	}
	if writer.input.Description != "Galvanized steel sheet" {
		t.Fatalf("expected translated description, got %s", writer.input.Description)
	}
	if writer.input.BrandName != "Acme" {
		t.Fatalf("expected translated brand, got %s", writer.input.BrandName)
	}
	if writer.input.StockProfileCode != "standard" {
		t.Fatalf("expected translated stock profile, got %s", writer.input.StockProfileCode)
	}
	if writer.input.PrimaryTaxonomyNodeID != "txn_leaf_1" {
		t.Fatalf("expected translated taxonomy node, got %s", writer.input.PrimaryTaxonomyNodeID)
	}
	if writer.input.Status != "active" {
		t.Fatalf("expected active status, got %s", writer.input.Status)
	}
	if len(writer.input.Identifiers) < 3 {
		t.Fatalf("expected at least 3 identifiers, got %+v", writer.input.Identifiers)
	}
}

func TestPromotionConsumerSkipsDuplicateClaim(t *testing.T) {
	reconRepo := &stubPromotionRepository{claimResult: false}
	consumer := NewPromotionConsumer(reconRepo, &stubAutoPromoGuard{}, NewProductPromotion(&stubPromotionReader{}, &stubRunRepo{}, &recordingProductWriter{}))

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypeProducts,
		PromotionStatus:  domain.PromotionStatusPending,
	}

	consumer.promoteOne(context.Background(), result)

	if got := len(reconRepo.claimCalls); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}
	if got := len(reconRepo.promoted); got != 0 {
		t.Fatalf("expected no promotion after duplicate claim, got %d", got)
	}
	if got := len(reconRepo.failed); got != 0 {
		t.Fatalf("expected no failed rows, got %d", got)
	}
	if got := len(reconRepo.reviewRequired); got != 0 {
		t.Fatalf("expected no review-required rows, got %d", got)
	}
}

func TestPromotionConsumerSkipsWhenAutoPromotionDisabled(t *testing.T) {
	reconRepo := &stubPromotionRepository{claimResult: true}
	guard := &stubAutoPromoGuard{err: domain.ErrAutoPromotionDisabled}
	consumer := NewPromotionConsumer(reconRepo, guard, NewProductPromotion(&stubPromotionReader{}, &stubRunRepo{}, &recordingProductWriter{}))

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypeProducts,
		PromotionStatus:  domain.PromotionStatusPending,
	}

	consumer.promoteOne(context.Background(), result)

	if got := len(reconRepo.claimCalls); got != 0 {
		t.Fatalf("expected no claim when auto-promotion is disabled, got %d", got)
	}
}

func TestPromotionConsumerMarksFailedOnDomainWriteFailure(t *testing.T) {
	reconRepo := &stubPromotionRepository{claimResult: true}
	stagingRepo := &stubPromotionReader{
		staging: map[string]*domain.StagingRecord{
			"stg_1": {
				StagingID:        "stg_1",
				TenantID:         "tenant-1",
				EntityType:       domain.EntityTypeProducts,
				NormalizedJSON:   []byte(`{"pn_interno":"PN-001","descricao":"Galvanized steel sheet","ativo":true}`),
				ValidationStatus: "valid",
			},
		},
	}
	runRepo := &stubRunRepo{
		runs: map[string]*domain.SyncRun{
			"run_1": {
				RunID:         "run_1",
				TenantID:      "tenant-1",
				InstanceID:    "inst_1",
				ConnectorType: domain.ConnectorTypeSankhya,
			},
		},
	}
	writer := &recordingProductWriter{err: errors.New("catalog write failed")}
	consumer := NewPromotionConsumer(reconRepo, &stubAutoPromoGuard{}, NewProductPromotion(stagingRepo, runRepo, writer))

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypeProducts,
		PromotionStatus:  domain.PromotionStatusPending,
	}

	consumer.promoteOne(context.Background(), result)

	if got := len(reconRepo.claimCalls); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}
	if got := len(reconRepo.promoted); got != 0 {
		t.Fatalf("expected no promoted rows, got %d", got)
	}
	if got := len(reconRepo.failed); got != 1 {
		t.Fatalf("expected 1 failed row, got %d", got)
	}
	if reconRepo.failed[0].reasonCode != promotionFailureReasonCode {
		t.Fatalf("expected reason code %s, got %s", promotionFailureReasonCode, reconRepo.failed[0].reasonCode)
	}
	if reconRepo.failed[0].warningDetails == nil || !strings.Contains(*reconRepo.failed[0].warningDetails, "catalog promotion failed") {
		t.Fatalf("expected structured warning details, got %#v", reconRepo.failed[0].warningDetails)
	}
	if reconRepo.failed[0].warningDetails == nil || !strings.Contains(*reconRepo.failed[0].warningDetails, `"promotion_status":"promoting"`) {
		t.Fatalf("expected promoting lifecycle state in failure payload, got %#v", reconRepo.failed[0].warningDetails)
	}
}

func TestPromotionConsumerRoutesUnsupportedEntityTypeToReviewRequired(t *testing.T) {
	reconRepo := &stubPromotionRepository{claimResult: true}
	consumer := NewPromotionConsumer(reconRepo, &stubAutoPromoGuard{}, NewProductPromotion(&stubPromotionReader{}, &stubRunRepo{}, &recordingProductWriter{}))

	result := &domain.ReconciliationResult{
		ReconciliationID: "rec_1",
		TenantID:         "tenant-1",
		StagingID:        "stg_1",
		EntityType:       domain.EntityTypePrices,
		PromotionStatus:  domain.PromotionStatusPending,
	}

	consumer.promoteOne(context.Background(), result)

	if got := len(reconRepo.claimCalls); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}
	if got := len(reconRepo.reviewRequired); got != 1 {
		t.Fatalf("expected 1 review-required row, got %d", got)
	}
	if got := len(reconRepo.promoted); got != 0 {
		t.Fatalf("expected no promoted rows, got %d", got)
	}
	if got := len(reconRepo.failed); got != 0 {
		t.Fatalf("expected no failed rows in normal review-required path, got %d", got)
	}
	if reconRepo.reviewRequired[0].reasonCode != unsupportedPromotionEntityReasonCode {
		t.Fatalf("expected reason code %s, got %s", unsupportedPromotionEntityReasonCode, reconRepo.reviewRequired[0].reasonCode)
	}
	if reconRepo.reviewRequired[0].warningDetails == nil || !strings.Contains(*reconRepo.reviewRequired[0].warningDetails, unsupportedPromotionEntityReasonCode) {
		t.Fatalf("expected review warning details, got %#v", reconRepo.reviewRequired[0].warningDetails)
	}
	if reconRepo.reviewRequired[0].warningDetails == nil || !strings.Contains(*reconRepo.reviewRequired[0].warningDetails, `"promotion_status":"promoting"`) {
		t.Fatalf("expected promoting lifecycle state in review payload, got %#v", reconRepo.reviewRequired[0].warningDetails)
	}
}

func TestBuildProductPromotionInputRejectsInvalidStatus(t *testing.T) {
	_, err := buildProductPromotionInput(&domain.StagingRecord{
		StagingID:        "stg_1",
		NormalizedJSON:   []byte(`{"pn_interno":"PN-001","descricao":"Steel","status":"broken"}`),
		ValidationStatus: "valid",
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
	if !strings.Contains(err.Error(), "invalid product status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}
