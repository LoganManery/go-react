package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user from the auth.users table
type User struct {
	UserID                  uuid.UUID  `json:"user_id"`
	Username                string     `json:"username"`
	Email                   string     `json:"email"`
	PasswordHash            string     `json:"-"` // Never expose password hash in JSON
	FirstName               string     `json:"first_name,omitempty"`
	LastName                string     `json:"last_name,omitempty"`
	IsEmailVerified         bool       `json:"is_email_verified"`
	EmailVerificationToken  *string    `json:"-"`
	EmailVerificationSentAt *time.Time `json:"-"`
	PasswordResetToken      *string    `json:"-"`
	PasswordResetExpiresAt  *time.Time `json:"-"`
	FailedLoginAttempts     int        `json:"-"`
	LockedUntil             *time.Time `json:"-"`
	LastLoginAt             *time.Time `json:"last_login_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	IsActive                bool       `json:"is_active"`
}

// UserRepository handles database operations for users
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create adds a new user to the database
func (r *UserRepository) Create(ctx context.Context, user *User, password string) error {
	// Generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate a new UUID if not provided
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// SQL query
	query := `
		INSERT INTO auth.users (
			user_id, username, email, password_hash, first_name, last_name,
			is_email_verified, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING user_id, created_at, updated_at`

	// Execute query
	row := r.pool.QueryRow(ctx, query,
		user.UserID, user.Username, user.Email, string(hashedPassword),
		user.FirstName, user.LastName, user.IsEmailVerified, user.IsActive,
		user.CreatedAt, user.UpdatedAt,
	)

	// Scan result
	return row.Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	query := `
		SELECT 
			user_id, username, email, password_hash, first_name, last_name,
			is_email_verified, email_verification_token, email_verification_sent_at,
			password_reset_token, password_reset_expires_at, failed_login_attempts,
			locked_until, last_login_at, created_at, updated_at, is_active
		FROM auth.users
		WHERE user_id = $1`

	row := r.pool.QueryRow(ctx, query, userID)

	var user User
	err := scanUser(row, &user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT 
			user_id, username, email, password_hash, first_name, last_name,
			is_email_verified, email_verification_token, email_verification_sent_at,
			password_reset_token, password_reset_expires_at, failed_login_attempts,
			locked_until, last_login_at, created_at, updated_at, is_active
		FROM auth.users
		WHERE email = $1`

	row := r.pool.QueryRow(ctx, query, email)

	var user User
	err := scanUser(row, &user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by their username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT 
			user_id, username, email, password_hash, first_name, last_name,
			is_email_verified, email_verification_token, email_verification_sent_at,
			password_reset_token, password_reset_expires_at, failed_login_attempts,
			locked_until, last_login_at, created_at, updated_at, is_active
		FROM auth.users
		WHERE username = $1`

	row := r.pool.QueryRow(ctx, query, username)

	var user User
	err := scanUser(row, &user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

// Update updates a user's information
func (r *UserRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE auth.users SET
			username = $1,
			email = $2,
			first_name = $3,
			last_name = $4,
			is_email_verified = $5,
			email_verification_token = $6,
			email_verification_sent_at = $7,
			password_reset_token = $8,
			password_reset_expires_at = $9,
			failed_login_attempts = $10,
			locked_until = $11,
			last_login_at = $12,
			updated_at = $13,
			is_active = $14
		WHERE user_id = $15
		RETURNING updated_at`

	row := r.pool.QueryRow(ctx, query,
		user.Username, user.Email, user.FirstName, user.LastName,
		user.IsEmailVerified, user.EmailVerificationToken, user.EmailVerificationSentAt,
		user.PasswordResetToken, user.PasswordResetExpiresAt,
		user.FailedLoginAttempts, user.LockedUntil, user.LastLoginAt,
		user.UpdatedAt, user.IsActive, user.UserID,
	)

	return row.Scan(&user.UpdatedAt)
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Generate new password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update the password hash and reset any password reset fields
	query := `
		UPDATE auth.users SET
			password_hash = $1,
			password_reset_token = NULL,
			password_reset_expires_at = NULL,
			updated_at = NOW()
		WHERE user_id = $2`

	_, err = r.pool.Exec(ctx, query, string(hashedPassword), userID)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM auth.users WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// List retrieves a paginated list of users
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*User, error) {
	query := `
		SELECT 
			user_id, username, email, password_hash, first_name, last_name,
			is_email_verified, email_verification_token, email_verification_sent_at,
			password_reset_token, password_reset_expires_at, failed_login_attempts,
			locked_until, last_login_at, created_at, updated_at, is_active
		FROM auth.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := scanUserFromRows(rows, &user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM auth.users`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// VerifyPassword checks if the provided password matches the stored hash
func (r *UserRepository) VerifyPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// RecordLogin updates the last login time and resets failed login attempts
func (r *UserRepository) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE auth.users SET
			last_login_at = $1,
			failed_login_attempts = 0,
			locked_until = NULL,
			updated_at = $1
		WHERE user_id = $2`

	_, err := r.pool.Exec(ctx, query, now, userID)
	return err
}

// IncrementFailedLoginAttempts increments the failed login attempts counter
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth.users SET
			failed_login_attempts = failed_login_attempts + 1,
			updated_at = NOW()
		WHERE user_id = $1
		RETURNING failed_login_attempts`

	var attempts int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&attempts)
	if err != nil {
		return err
	}

	// Lock the account after 5 failed attempts
	if attempts >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		query := `
			UPDATE auth.users SET
				locked_until = $1,
				updated_at = NOW()
			WHERE user_id = $2`
		_, err = r.pool.Exec(ctx, query, lockTime, userID)
	}

	return err
}

// Helper function to scan a user from a row
func scanUser(row pgx.Row, user *User) error {
	return row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsEmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationSentAt,
		&user.PasswordResetToken,
		&user.PasswordResetExpiresAt,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)
}

// Helper function to scan a user from rows
func scanUserFromRows(rows pgx.Rows, user *User) error {
	return rows.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsEmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationSentAt,
		&user.PasswordResetToken,
		&user.PasswordResetExpiresAt,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)
}
