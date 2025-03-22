// models/session.go
package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Session represents a user session from the auth.sessions table
type Session struct {
	SessionID    uuid.UUID `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	Token        string    `json:"token"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsValid      bool      `json:"is_valid"`
}

// SessionRepository handles database operations for sessions
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// Create adds a new session to the database
func (r *SessionRepository) Create(ctx context.Context, session *Session) error {
	// Generate a new UUID if not provided
	if session.SessionID == uuid.Nil {
		session.SessionID = uuid.New()
	}

	// Set timestamps if not provided
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.LastActiveAt.IsZero() {
		session.LastActiveAt = now
	}

	// SQL query
	query := `
		INSERT INTO auth.sessions (
			session_id, user_id, token, ip_address, user_agent,
			expires_at, created_at, last_active_at, is_valid
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING session_id, created_at`

	// Execute query
	row := r.pool.QueryRow(ctx, query,
		session.SessionID, session.UserID, session.Token,
		session.IPAddress, session.UserAgent, session.ExpiresAt,
		session.CreatedAt, session.LastActiveAt, session.IsValid,
	)

	// Scan result
	return row.Scan(&session.SessionID, &session.CreatedAt)
}

// GetByID retrieves a session by ID
func (r *SessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	query := `
		SELECT 
			session_id, user_id, token, ip_address, user_agent,
			expires_at, created_at, last_active_at, is_valid
		FROM auth.sessions
		WHERE session_id = $1`

	row := r.pool.QueryRow(ctx, query, sessionID)

	var session Session
	err := scanSession(row, &session)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Session not found
		}
		return nil, err
	}

	return &session, nil
}

// GetByToken retrieves a session by token
func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT 
			session_id, user_id, token, ip_address, user_agent,
			expires_at, created_at, last_active_at, is_valid
		FROM auth.sessions
		WHERE token = $1`

	row := r.pool.QueryRow(ctx, query, token)

	var session Session
	err := scanSession(row, &session)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Session not found
		}
		return nil, err
	}

	return &session, nil
}

// GetAllByUserID retrieves all sessions for a user
func (r *SessionRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	query := `
		SELECT 
			session_id, user_id, token, ip_address, user_agent,
			expires_at, created_at, last_active_at, is_valid
		FROM auth.sessions
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		if err := scanSessionFromRows(rows, &session); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Invalidate marks a session as invalid
func (r *SessionRepository) Invalidate(ctx context.Context, token string) error {
	query := `
		UPDATE auth.sessions SET
			is_valid = false,
			last_active_at = NOW()
		WHERE token = $1`

	_, err := r.pool.Exec(ctx, query, token)
	return err
}

// InvalidateAllForUser invalidates all sessions for a user
func (r *SessionRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth.sessions SET
			is_valid = false,
			last_active_at = NOW()
		WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// UpdateLastActiveAt updates the last_active_at timestamp
func (r *SessionRepository) UpdateLastActiveAt(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE auth.sessions SET
			last_active_at = NOW()
		WHERE session_id = $1`

	_, err := r.pool.Exec(ctx, query, sessionID)
	return err
}

// DeleteExpiredSessions deletes all expired sessions
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM auth.sessions WHERE expires_at < NOW()`
	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// DeleteByID deletes a session by ID
func (r *SessionRepository) DeleteByID(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM auth.sessions WHERE session_id = $1`
	_, err := r.pool.Exec(ctx, query, sessionID)
	return err
}

// Helper function to scan a session from a row
func scanSession(row pgx.Row, session *Session) error {
	return row.Scan(
		&session.SessionID,
		&session.UserID,
		&session.Token,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActiveAt,
		&session.IsValid,
	)
}

// Helper function to scan a session from rows
func scanSessionFromRows(rows pgx.Rows, session *Session) error {
	return rows.Scan(
		&session.SessionID,
		&session.UserID,
		&session.Token,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActiveAt,
		&session.IsValid,
	)
}
