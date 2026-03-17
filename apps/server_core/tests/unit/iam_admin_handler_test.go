package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metalshopping/server_core/internal/modules/iam/application"
	"metalshopping/server_core/internal/modules/iam/domain"
	iamhttp "metalshopping/server_core/internal/modules/iam/transport/http"
	platformauth "metalshopping/server_core/internal/platform/auth"
)

type fakeRoleAssignmentReader struct {
	roles []domain.Role
	err   error
}

func (f *fakeRoleAssignmentReader) RolesByUserID(context.Context, string) ([]domain.Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.roles, nil
}

func TestIAMAdminHandlerRequiresAuthenticatedPrincipal(t *testing.T) {
	service := application.NewAdminService(&fakeRoleAssignmentWriter{}, &fakeAdminRoleAssignmentGuard{})
	authz := application.NewAuthorizationService(&fakeRoleAssignmentReader{roles: []domain.Role{domain.RoleAdmin}}, application.NewStaticAuthorizer())
	handler := iamhttp.NewAdminHandler(service, authz)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/user-1/roles", strings.NewReader(`{"display_name":"User","role":"viewer"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestIAMAdminHandlerUpsertsRoleAssignment(t *testing.T) {
	service := application.NewAdminService(&fakeRoleAssignmentWriter{}, &fakeAdminRoleAssignmentGuard{})
	authz := application.NewAuthorizationService(&fakeRoleAssignmentReader{roles: []domain.Role{domain.RoleAdmin}}, application.NewStaticAuthorizer())
	handler := iamhttp.NewAdminHandler(service, authz)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/user-1/roles", strings.NewReader(`{"display_name":"User One","role":"viewer"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "admin-1", TenantID: "tenant-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIAMAdminHandlerRejectsInsufficientPermission(t *testing.T) {
	service := application.NewAdminService(&fakeRoleAssignmentWriter{}, &fakeAdminRoleAssignmentGuard{})
	authz := application.NewAuthorizationService(&fakeRoleAssignmentReader{roles: []domain.Role{domain.RoleViewer}}, application.NewStaticAuthorizer())
	handler := iamhttp.NewAdminHandler(service, authz)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users/user-1/roles", strings.NewReader(`{"display_name":"User One","role":"viewer"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(platformauth.WithPrincipal(req.Context(), platformauth.Principal{SubjectID: "viewer-1"}))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
