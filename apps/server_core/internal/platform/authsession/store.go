package authsession

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type LoginState struct {
	LoginStateID string
	CodeVerifier string
	ReturnTo     string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type Session struct {
	SessionID                string
	SubjectID                string
	TenantID                 string
	Email                    string
	DisplayName              string
	IssuedAt                 time.Time
	LastSeenAt               time.Time
	IdleTimeoutExpiresAt     time.Time
	AbsoluteTimeoutExpiresAt time.Time
	InvalidatedAt            *time.Time
}

type Store struct {
	db *sql.DB
}

type SessionStorage interface {
	CreateLoginState(ctx context.Context, loginState LoginState) error
	GetLoginState(ctx context.Context, loginStateID string) (LoginState, error)
	DeleteLoginState(ctx context.Context, loginStateID string) error
	CreateSession(ctx context.Context, session Session) error
	GetActiveSession(ctx context.Context, sessionID string, now time.Time) (Session, error)
	RotateSession(ctx context.Context, currentSessionID string, next Session, invalidatedAt time.Time) error
	InvalidateSession(ctx context.Context, sessionID string, reason string, invalidatedAt time.Time) error
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateLoginState(ctx context.Context, loginState LoginState) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_web_login_states (
			login_state_id,
			code_verifier,
			return_to,
			expires_at
		) VALUES ($1, $2, $3, $4)
	`, loginState.LoginStateID, loginState.CodeVerifier, nullableText(loginState.ReturnTo), loginState.ExpiresAt.UTC())
	if err != nil {
		return fmt.Errorf("insert auth login state: %w", err)
	}
	return nil
}

func (s *Store) GetLoginState(ctx context.Context, loginStateID string) (LoginState, error) {
	var state LoginState
	err := s.db.QueryRowContext(ctx, `
		SELECT login_state_id, code_verifier, COALESCE(return_to, ''), expires_at, created_at
		FROM auth_web_login_states
		WHERE login_state_id = $1
	`, strings.TrimSpace(loginStateID)).Scan(
		&state.LoginStateID,
		&state.CodeVerifier,
		&state.ReturnTo,
		&state.ExpiresAt,
		&state.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginState{}, ErrLoginStateNotFound
		}
		return LoginState{}, fmt.Errorf("query auth login state: %w", err)
	}
	return state, nil
}

func (s *Store) DeleteLoginState(ctx context.Context, loginStateID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_web_login_states WHERE login_state_id = $1`, strings.TrimSpace(loginStateID))
	if err != nil {
		return fmt.Errorf("delete auth login state: %w", err)
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context, session Session) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auth_web_sessions (
			session_id,
			subject_id,
			tenant_id,
			email,
			display_name,
			issued_at,
			last_seen_at,
			idle_timeout_expires_at,
			absolute_timeout_expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		session.SessionID,
		session.SubjectID,
		session.TenantID,
		nullableText(session.Email),
		nullableText(session.DisplayName),
		session.IssuedAt.UTC(),
		session.LastSeenAt.UTC(),
		session.IdleTimeoutExpiresAt.UTC(),
		session.AbsoluteTimeoutExpiresAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert auth web session: %w", err)
	}
	return nil
}

func (s *Store) GetActiveSession(ctx context.Context, sessionID string, now time.Time) (Session, error) {
	var session Session
	var invalidatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT
			session_id,
			subject_id,
			tenant_id,
			COALESCE(email, ''),
			COALESCE(display_name, ''),
			issued_at,
			last_seen_at,
			idle_timeout_expires_at,
			absolute_timeout_expires_at,
			invalidated_at
		FROM auth_web_sessions
		WHERE session_id = $1
	`, strings.TrimSpace(sessionID)).Scan(
		&session.SessionID,
		&session.SubjectID,
		&session.TenantID,
		&session.Email,
		&session.DisplayName,
		&session.IssuedAt,
		&session.LastSeenAt,
		&session.IdleTimeoutExpiresAt,
		&session.AbsoluteTimeoutExpiresAt,
		&invalidatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("query auth web session: %w", err)
	}
	if invalidatedAt.Valid {
		value := invalidatedAt.Time.UTC()
		session.InvalidatedAt = &value
		return Session{}, ErrSessionInvalidated
	}
	now = now.UTC()
	if now.After(session.IdleTimeoutExpiresAt.UTC()) || now.After(session.AbsoluteTimeoutExpiresAt.UTC()) {
		return Session{}, ErrSessionExpired
	}
	return session, nil
}

func (s *Store) RotateSession(ctx context.Context, currentSessionID string, next Session, invalidatedAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin auth session rotation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		UPDATE auth_web_sessions
		SET invalidated_at = $2, invalidation_reason = 'rotated', updated_at = NOW()
		WHERE session_id = $1
	`, strings.TrimSpace(currentSessionID), invalidatedAt.UTC()); err != nil {
		return fmt.Errorf("invalidate auth web session during rotation: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO auth_web_sessions (
			session_id,
			subject_id,
			tenant_id,
			email,
			display_name,
			issued_at,
			last_seen_at,
			idle_timeout_expires_at,
			absolute_timeout_expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		next.SessionID,
		next.SubjectID,
		next.TenantID,
		nullableText(next.Email),
		nullableText(next.DisplayName),
		next.IssuedAt.UTC(),
		next.LastSeenAt.UTC(),
		next.IdleTimeoutExpiresAt.UTC(),
		next.AbsoluteTimeoutExpiresAt.UTC(),
	); err != nil {
		return fmt.Errorf("insert rotated auth web session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit auth session rotation: %w", err)
	}
	return nil
}

func (s *Store) InvalidateSession(ctx context.Context, sessionID string, reason string, invalidatedAt time.Time) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE auth_web_sessions
		SET invalidated_at = $2, invalidation_reason = $3, updated_at = NOW()
		WHERE session_id = $1 AND invalidated_at IS NULL
	`, strings.TrimSpace(sessionID), invalidatedAt.UTC(), strings.TrimSpace(reason))
	if err != nil {
		return fmt.Errorf("invalidate auth web session: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count invalidated auth web sessions: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
