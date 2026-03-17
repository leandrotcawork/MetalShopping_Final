package application

import (
	"context"
	"strings"
	"time"

	"metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/modules/iam/ports"
)

type UpsertRoleAssignmentCommand struct {
	UserID      string
	DisplayName string
	Role        string
	AssignedBy  string
}

type AdminService struct {
	writer ports.RoleAssignmentWriter
	now    func() time.Time
}

func NewAdminService(writer ports.RoleAssignmentWriter) *AdminService {
	return &AdminService{
		writer: writer,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *AdminService) UpsertRoleAssignment(ctx context.Context, cmd UpsertRoleAssignmentCommand) error {
	role, err := domain.ParseRole(cmd.Role)
	if err != nil {
		return err
	}

	actor := strings.TrimSpace(cmd.AssignedBy)
	if actor == "" {
		return domain.ErrActorRequired
	}

	assignment := domain.RoleAssignment{
		UserID:      strings.TrimSpace(cmd.UserID),
		DisplayName: strings.TrimSpace(cmd.DisplayName),
		Role:        role,
		AssignedBy:  actor,
		AssignedAt:  s.now(),
	}
	if strings.TrimSpace(assignment.DisplayName) == "" {
		assignment.DisplayName = assignment.UserID
	}
	if err := assignment.Validate(); err != nil {
		return err
	}

	return s.writer.UpsertRoleAssignment(ctx, assignment)
}
