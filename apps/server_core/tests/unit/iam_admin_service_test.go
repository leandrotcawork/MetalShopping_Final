package unit

import (
	"context"
	"errors"
	"testing"

	"metalshopping/server_core/internal/modules/iam/application"
	"metalshopping/server_core/internal/modules/iam/domain"
)

type fakeRoleAssignmentWriter struct {
	assignment domain.RoleAssignment
	err        error
}

func (f *fakeRoleAssignmentWriter) UpsertRoleAssignment(_ context.Context, assignment domain.RoleAssignment) error {
	if f.err != nil {
		return f.err
	}
	f.assignment = assignment
	return nil
}

func TestAdminServiceUpsertsAssignment(t *testing.T) {
	writer := &fakeRoleAssignmentWriter{}
	service := application.NewAdminService(writer)

	err := service.UpsertRoleAssignment(context.Background(), application.UpsertRoleAssignmentCommand{
		UserID:      "user-1",
		DisplayName: "User One",
		Role:        "viewer",
		AssignedBy:  "admin-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if writer.assignment.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", writer.assignment.UserID)
	}
	if writer.assignment.Role != domain.RoleViewer {
		t.Fatalf("expected viewer role, got %q", writer.assignment.Role)
	}
}

func TestAdminServiceRejectsInvalidRole(t *testing.T) {
	writer := &fakeRoleAssignmentWriter{}
	service := application.NewAdminService(writer)

	err := service.UpsertRoleAssignment(context.Background(), application.UpsertRoleAssignmentCommand{
		UserID:     "user-1",
		Role:       "super-admin",
		AssignedBy: "admin-1",
	})
	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestAdminServiceRequiresActor(t *testing.T) {
	writer := &fakeRoleAssignmentWriter{}
	service := application.NewAdminService(writer)

	err := service.UpsertRoleAssignment(context.Background(), application.UpsertRoleAssignmentCommand{
		UserID: "user-1",
		Role:   "viewer",
	})
	if !errors.Is(err, domain.ErrActorRequired) {
		t.Fatalf("expected actor required error, got %v", err)
	}
}
