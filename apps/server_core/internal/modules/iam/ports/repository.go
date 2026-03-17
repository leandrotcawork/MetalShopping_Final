package ports

import (
	"context"

	"metalshopping/server_core/internal/modules/iam/domain"
)

type RoleAssignmentWriter interface {
	UpsertRoleAssignment(ctx context.Context, assignment domain.RoleAssignment) error
}

type RoleAssignmentReader interface {
	RolesByUserID(ctx context.Context, userID string) ([]domain.Role, error)
}
