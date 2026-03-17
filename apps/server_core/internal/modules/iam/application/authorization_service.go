package application

import (
	"context"
	"strings"

	"metalshopping/server_core/internal/modules/iam/domain"
	"metalshopping/server_core/internal/modules/iam/ports"
)

type AuthorizationService struct {
	reader     ports.RoleAssignmentReader
	authorizer domain.Authorizer
}

func NewAuthorizationService(reader ports.RoleAssignmentReader, authorizer domain.Authorizer) *AuthorizationService {
	return &AuthorizationService{
		reader:     reader,
		authorizer: authorizer,
	}
}

func (s *AuthorizationService) HasPermission(ctx context.Context, userID string, permission domain.Permission) (bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return false, domain.ErrUserIDRequired
	}

	roles, err := s.reader.RolesByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return false, nil
		}
		return false, err
	}

	for _, role := range roles {
		if s.authorizer.Can(role, permission) {
			return true, nil
		}
	}

	return false, nil
}
