package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"metalshopping/server_core/internal/modules/erp_integrations/application"
	"metalshopping/server_core/internal/modules/erp_integrations/domain"
	"metalshopping/server_core/internal/modules/erp_integrations/ports"
	erphttp "metalshopping/server_core/internal/modules/erp_integrations/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
	"metalshopping/server_core/internal/platform/tenancy_runtime"
)

// ---------------------------------------------------------------------------
// In-memory stubs (minimal duplicates of service_test stubs scoped to this pkg)
// ---------------------------------------------------------------------------

type handlerStubInstanceRepo struct {
	instances  []*domain.IntegrationInstance
	activeFlag bool
	createErr  error
}

func (r *handlerStubInstanceRepo) Create(_ context.Context, instance *domain.IntegrationInstance) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.instances = append(r.instances, instance)
	return nil
}

func (r *handlerStubInstanceRepo) Get(_ context.Context, _, instanceID string) (*domain.IntegrationInstance, error) {
	for _, i := range r.instances {
		if i.InstanceID == instanceID {
			return i, nil
		}
	}
	return nil, domain.ErrInstanceNotFound
}

func (r *handlerStubInstanceRepo) List(_ context.Context, _ string, limit, offset int) ([]*domain.IntegrationInstance, error) {
	end := offset + limit
	if offset >= len(r.instances) {
		return []*domain.IntegrationInstance{}, nil
	}
	if end > len(r.instances) {
		end = len(r.instances)
	}
	return r.instances[offset:end], nil
}

func (r *handlerStubInstanceRepo) HasActiveInstance(_ context.Context, _ string) (bool, error) {
	return r.activeFlag, nil
}

type handlerStubRunRepo struct{}

func (r *handlerStubRunRepo) Create(_ context.Context, _ *domain.SyncRun) error { return nil }
func (r *handlerStubRunRepo) Get(_ context.Context, _, _ string) (*domain.SyncRun, error) {
	return nil, domain.ErrRunNotFound
}
func (r *handlerStubRunRepo) List(_ context.Context, _, _ string, _, _ int) ([]*domain.SyncRun, error) {
	return nil, nil
}

type handlerStubReviewRepo struct {
	items []*domain.ReviewItem
}

func (r *handlerStubReviewRepo) Create(_ context.Context, item *domain.ReviewItem) error {
	r.items = append(r.items, item)
	return nil
}
func (r *handlerStubReviewRepo) Get(_ context.Context, _, reviewID string) (*domain.ReviewItem, error) {
	for _, item := range r.items {
		if item.ReviewID == reviewID {
			return item, nil
		}
	}
	return nil, domain.ErrReviewItemNotFound
}
func (r *handlerStubReviewRepo) List(_ context.Context, _ string, limit, offset int) ([]*domain.ReviewItem, error) {
	end := offset + limit
	if offset >= len(r.items) {
		return []*domain.ReviewItem{}, nil
	}
	if end > len(r.items) {
		end = len(r.items)
	}
	return r.items[offset:end], nil
}
func (r *handlerStubReviewRepo) Resolve(_ context.Context, _, _ string, _ domain.ReviewItemStatus, _ string, _ time.Time) error {
	return nil
}

type handlerStubEnabledGuard struct{ err error }

func (g *handlerStubEnabledGuard) CheckEnabled(_ context.Context, _ string) error { return g.err }

type handlerStubPermChecker struct{ allowed bool }

func (c *handlerStubPermChecker) CanManageIntegrations(_ context.Context, _, _ string) (bool, error) {
	return c.allowed, nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// buildHandler constructs a Handler backed by a real Service with stub repos.
func buildHandler(instanceRepo ports.InstanceRepository, active bool, createErr error) *erphttp.Handler {
	if instanceRepo == nil {
		instanceRepo = &handlerStubInstanceRepo{activeFlag: active, createErr: createErr}
	}
	svc := application.NewService(
		instanceRepo,
		&handlerStubRunRepo{},
		&handlerStubReviewRepo{},
		&handlerStubEnabledGuard{},
		&handlerStubPermChecker{allowed: true},
		nil,
	)
	return erphttp.NewHandler(svc)
}

func stringPtr(v string) *string { return &v }

func connectionPayload() map[string]any {
	return map[string]any{
		"kind":                "oracle",
		"host":                "10.55.10.101",
		"port":                1521,
		"service_name":        "ORCL",
		"username":            "leandroth",
		"password_secret_ref": "erp/sankhya/password",
	}
}

// injectAuth injects a principal and tenant into the request context so that
// the handler's requirePrincipalAndTenant check passes.
func injectAuth(r *http.Request) *http.Request {
	ctx := platformauth.WithPrincipal(r.Context(), platformauth.Principal{SubjectID: "user-1"})
	ctx = tenancy_runtime.WithTenant(ctx, tenancy_runtime.Tenant{ID: "tenant-1"})
	return r.WithContext(ctx)
}

// newMux registers routes from the handler and returns a ready ServeMux.
func newMux(h *erphttp.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateInstance_201(t *testing.T) {
	h := buildHandler(nil, false, nil)
	mux := newMux(h)

	body, _ := json.Marshal(map[string]any{
		"connector_type":   "sankhya",
		"display_name":     "My ERP",
		"connection":       connectionPayload(),
		"enabled_entities": []string{"products"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/erp/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateInstance_409(t *testing.T) {
	// activeFlag=true causes the service to return ErrActiveInstanceExists.
	h := buildHandler(nil, true, nil)
	mux := newMux(h)

	body, _ := json.Marshal(map[string]any{
		"connector_type":   "sankhya",
		"display_name":     "My ERP",
		"connection":       connectionPayload(),
		"enabled_entities": []string{"products"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/erp/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateInstanceHandlerValidatesNestedConnection(t *testing.T) {
	h := buildHandler(nil, false, nil)
	mux := newMux(h)

	connection := connectionPayload()
	delete(connection, "password_secret_ref")

	body, _ := json.Marshal(map[string]any{
		"connector_type":   "sankhya",
		"display_name":     "ERP Oracle",
		"connection":       connection,
		"enabled_entities": []string{"products"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/erp/instances", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d - body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errBody, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope, got %v", resp)
	}
	if got := errBody["code"]; got != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %v", got)
	}
	if got := errBody["message"]; got != domain.ErrEmptyPasswordSecretRef.Error() {
		t.Fatalf("expected password secret ref validation message, got %v", got)
	}
}

func TestListInstances_Paginated(t *testing.T) {
	// Seed 3 instances.
	instances := []*domain.IntegrationInstance{
		{InstanceID: "inst_a", TenantID: "tenant-1", ConnectorType: domain.ConnectorTypeSankhya, DisplayName: "A", Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		}, EnabledEntities: []domain.EntityType{domain.EntityTypeProducts}, Status: domain.InstanceStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{InstanceID: "inst_b", TenantID: "tenant-1", ConnectorType: domain.ConnectorTypeSankhya, DisplayName: "B", Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		}, EnabledEntities: []domain.EntityType{domain.EntityTypeProducts}, Status: domain.InstanceStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{InstanceID: "inst_c", TenantID: "tenant-1", ConnectorType: domain.ConnectorTypeSankhya, DisplayName: "C", Connection: domain.InstanceConnectionConfig{
			Kind:              domain.ConnectionKindOracle,
			Host:              "10.55.10.101",
			Port:              1521,
			ServiceName:       stringPtr("ORCL"),
			Username:          "leandroth",
			PasswordSecretRef: "erp/sankhya/password",
		}, EnabledEntities: []domain.EntityType{domain.EntityTypeProducts}, Status: domain.InstanceStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	repo := &handlerStubInstanceRepo{instances: instances}
	h := buildHandler(repo, false, nil)
	mux := newMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/erp/instances?limit=10&offset=20", nil)
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("expected items array in response")
	}
	// offset=20 is beyond 3 items so the result should be empty.
	if len(items) != 0 {
		t.Errorf("expected 0 items for offset=20 with 3 total, got %d", len(items))
	}
}

func TestListReviewItems_Paginated(t *testing.T) {
	stagingSnapshot := `{"name":"Product A"}`
	reconciliationOutput := `{"classification":"review_required","reason_code":"ERP_PROMOTION_FAILED"}`
	reviewRepo := &handlerStubReviewRepo{
		items: []*domain.ReviewItem{
			{ReviewID: "rev_1", TenantID: "tenant-1", InstanceID: "inst_a", ConnectorType: domain.ConnectorTypeSankhya, EntityType: domain.EntityTypeProducts, Severity: domain.ReviewSeverityWarning, ItemStatus: domain.ReviewItemStatusOpen, StagingSnapshot: &stagingSnapshot, ReconciliationOutput: &reconciliationOutput, CreatedAt: time.Now()},
			{ReviewID: "rev_2", TenantID: "tenant-1", InstanceID: "inst_a", ConnectorType: domain.ConnectorTypeSankhya, EntityType: domain.EntityTypeProducts, Severity: domain.ReviewSeverityWarning, ItemStatus: domain.ReviewItemStatusOpen, CreatedAt: time.Now()},
		},
	}
	instanceRepo := &handlerStubInstanceRepo{}
	svc := application.NewService(
		instanceRepo,
		&handlerStubRunRepo{},
		reviewRepo,
		&handlerStubEnabledGuard{},
		&handlerStubPermChecker{allowed: true},
		nil,
	)
	h := erphttp.NewHandler(svc)
	mux := newMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/erp/review-items?limit=5", nil)
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("expected items array in response")
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatal("expected first item object")
	}
	if _, ok := first["staging_snapshot"].(map[string]any); !ok {
		t.Fatalf("expected staging_snapshot object, got %#v", first["staging_snapshot"])
	}
	if _, ok := first["reconciliation_output"].(map[string]any); !ok {
		t.Fatalf("expected reconciliation_output object, got %#v", first["reconciliation_output"])
	}
}

func TestGetInstance_404(t *testing.T) {
	h := buildHandler(nil, false, nil)
	mux := newMux(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/erp/instances/unknown-id", nil)
	req = injectAuth(req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
	}
}
