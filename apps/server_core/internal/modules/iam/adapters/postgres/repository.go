package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metalshopping/server_core/internal/modules/iam/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertRoleAssignment(ctx context.Context, assignment domain.RoleAssignment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const upsertUserSQL = `
INSERT INTO iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, is_active = TRUE, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUserSQL, assignment.UserID, assignment.DisplayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	const upsertRoleSQL = `
INSERT INTO iam_user_roles (user_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, role_code)
DO UPDATE SET assigned_at = EXCLUDED.assigned_at, assigned_by = EXCLUDED.assigned_by
`
	if _, err := tx.ExecContext(ctx, upsertRoleSQL, assignment.UserID, string(assignment.Role), assignment.AssignedAt, assignment.AssignedBy); err != nil {
		return fmt.Errorf("upsert iam role assignment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit iam transaction: %w", err)
	}
	return nil
}

func (r *Repository) RolesByUserID(ctx context.Context, userID string) ([]domain.Role, error) {
	const userSQL = `
SELECT is_active
FROM iam_users
WHERE user_id = $1
`

	var isActive bool
	if err := r.db.QueryRowContext(ctx, userSQL, userID).Scan(&isActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("query iam user: %w", err)
	}
	if !isActive {
		return nil, domain.ErrUserNotFound
	}

	const rolesSQL = `
SELECT role_code
FROM iam_user_roles
WHERE user_id = $1
ORDER BY role_code ASC
`
	rows, err := r.db.QueryContext(ctx, rolesSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("query iam roles: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0, 4)
	for rows.Next() {
		var roleCode string
		if err := rows.Scan(&roleCode); err != nil {
			return nil, fmt.Errorf("scan iam role: %w", err)
		}
		roles = append(roles, domain.Role(roleCode))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate iam roles: %w", err)
	}
	if len(roles) == 0 {
		return nil, domain.ErrUserNotFound
	}

	return roles, nil
}
