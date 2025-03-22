// models/audit_log.go
package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

// AuditLog represents an entry in the auth.audit_log table
type AuditLog struct {
	LogID     uuid.UUID              `json:"log_id"`
	UserID    uuid.UUID              `json:"user_id,omitempty"`
	EventType string                 `json:"event_type"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AuditLogRepository handles database operations for audit logs
type AuditLogRepository struct {
	pool *pgxpool.Pool
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

// Create adds a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	// Generate a new UUID if not provided
	if log.LogID == uuid.Nil {
		log.LogID = uuid.New()
	}

	// Set created_at if not provided
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// SQL query
	query := `
		INSERT INTO auth.audit_log (
			log_id, user_id, event_type, ip_address, user_agent, details
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING log_id, created_at`

	// Execute query
	row := r.pool.QueryRow(ctx, query,
		log.LogID, log.UserID, log.EventType,
		log.IPAddress, log.UserAgent, log.Details,
	)

	// Scan result
	return row.Scan(&log.LogID, &log.CreatedAt)
}

// GetByID retrieves an audit log entry by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, logID uuid.UUID) (*AuditLog, error) {
	query := `
		SELECT 
			log_id, user_id, event_type, ip_address, user_agent, details, created_at
		FROM auth.audit_log
		WHERE log_id = $1`

	var log AuditLog
	err := r.pool.QueryRow(ctx, query, logID).Scan(
		&log.LogID,
		&log.UserID,
		&log.EventType,
		&log.IPAddress,
		&log.UserAgent,
		&log.Details,
		&log.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &log, nil
}

// GetByUserID retrieves audit log entries for a specific user
func (r *AuditLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT 
			log_id, user_id, event_type, ip_address, user_agent, details, created_at
		FROM auth.audit_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.EventType,
			&log.IPAddress,
			&log.UserAgent,
			&log.Details,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetByEventType retrieves audit log entries by event type
func (r *AuditLogRepository) GetByEventType(ctx context.Context, eventType string, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT 
			log_id, user_id, event_type, ip_address, user_agent, details, created_at
		FROM auth.audit_log
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.EventType,
			&log.IPAddress,
			&log.UserAgent,
			&log.Details,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// List retrieves a paginated list of audit log entries
func (r *AuditLogRepository) List(ctx context.Context, limit, offset int) ([]*AuditLog, error) {
	query := `
		SELECT 
			log_id, user_id, event_type, ip_address, user_agent, details, created_at
		FROM auth.audit_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.EventType,
			&log.IPAddress,
			&log.UserAgent,
			&log.Details,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// Count returns the total number of audit log entries
func (r *AuditLogRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM auth.audit_log`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// DeleteOlderThan deletes audit log entries older than the specified time
func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM auth.audit_log WHERE created_at < $1`
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
