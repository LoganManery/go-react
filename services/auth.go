// services/auth.go
package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	// "golang.org/x/crypto/bcrypt"

	"github.com/loganmanery/go-react-app/models"
)

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUserLocked            = errors.New("account is locked due to too many failed login attempts")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidToken          = errors.New("invalid or expired token")
)

// AuthService handles authentication-related operations
type AuthService struct {
	pool           *pgxpool.Pool
	userRepo       *models.UserRepository
	sessionRepo    *models.SessionRepository
	jwtSecret      string
	tokenExpiryMin int
}

// NewAuthService creates a new AuthService
func NewAuthService(pool *pgxpool.Pool, jwtSecret string, tokenExpiryMin int) *AuthService {
	return &AuthService{
		pool:           pool,
		userRepo:       models.NewUserRepository(pool),
		sessionRepo:    models.NewSessionRepository(pool),
		jwtSecret:      jwtSecret,
		tokenExpiryMin: tokenExpiryMin,
	}
}

// Login authenticates a user and creates a new session
func (s *AuthService) Login(ctx context.Context, usernameOrEmail, password, ipAddress, userAgent string) (*models.Session, error) {
	// Try to find the user by email first, then by username
	var user *models.User
	var err error

	user, err = s.userRepo.GetByEmail(ctx, usernameOrEmail)
	if user == nil {
		user, err = s.userRepo.GetByUsername(ctx, usernameOrEmail)
	}

	if err != nil {
		return nil, err
	}

	if user == nil {
		// Record failed login attempt but don't indicate whether the user exists
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrUserLocked
	}

	// Verify password
	if !s.userRepo.VerifyPassword(user, password) {
		// Increment failed login attempts
		if err := s.userRepo.IncrementFailedLoginAttempts(ctx, user.UserID); err != nil {
			return nil, err
		}
		return nil, ErrInvalidCredentials
	}

	// Password correct - create session
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Create a new session
	expiryTime := time.Now().Add(time.Duration(s.tokenExpiryMin) * time.Minute)
	session := &models.Session{
		UserID:    user.UserID,
		Token:     token,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiryTime,
		IsValid:   true,
	}

	// Save the session
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Record successful login
	if err := s.userRepo.RecordLogin(ctx, user.UserID); err != nil {
		return nil, err
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.UserID,
		EventType: "login",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   map[string]interface{}{"successful": true},
	}

	if _, err := s.createAuditLog(ctx, auditLog); err != nil {
		// Log the error but don't fail the login
		fmt.Printf("Error creating audit log: %v\n", err)
	}

	return session, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, username, email, password, firstName, lastName string) (*models.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists
	existingUser, err = s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	}

	// Generate verification token
	verificationToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Create new user
	user := &models.User{
		Username:                username,
		Email:                   email,
		FirstName:               firstName,
		LastName:                lastName,
		IsEmailVerified:         false,
		EmailVerificationToken:  &verificationToken,
		EmailVerificationSentAt: &now,
		IsActive:                true,
	}

	// Create the user (will hash the password)
	if err := s.userRepo.Create(ctx, user, password); err != nil {
		return nil, err
	}

	// Send verification email (this would be implemented elsewhere)
	// s.emailService.SendVerificationEmail(user.Email, *user.EmailVerificationToken)

	return user, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Invalidate(ctx, token)
}

// ValidateSession checks if a session is valid
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.Session, *models.User, error) {
	// Find the session
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	if session == nil || !session.IsValid || time.Now().After(session.ExpiresAt) {
		return nil, nil, ErrInvalidToken
	}

	// Get the user associated with the session
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || !user.IsActive {
		return nil, nil, ErrUserNotFound
	}

	// Update the last active time
	if err := s.sessionRepo.UpdateLastActiveAt(ctx, session.SessionID); err != nil {
		// Just log this error, don't fail the validation
		fmt.Printf("Error updating session last active time: %v\n", err)
	}

	return session, user, nil
}

// VerifyEmail verifies a user's email using the verification token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// Find user by verification token
	query := `
		SELECT user_id FROM auth.users 
		WHERE email_verification_token = $1 
		AND is_email_verified = false`

	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		return ErrInvalidToken
	}

	// Update the user to mark email as verified
	updateQuery := `
		UPDATE auth.users SET
			is_email_verified = true,
			email_verification_token = NULL,
			updated_at = NOW()
		WHERE user_id = $1`

	_, err = s.pool.Exec(ctx, updateQuery, userID)
	return err
}

// ForgotPassword initiates the password reset process
func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	// Find the user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		// Don't reveal if the email exists or not
		return "", nil
	}

	// Generate reset token
	resetToken, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	// Set expiry time (e.g., 24 hours from now)
	expiryTime := time.Now().Add(24 * time.Hour)

	// Update user with reset token
	user.PasswordResetToken = &resetToken
	user.PasswordResetExpiresAt = &expiryTime

	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", err
	}

	// Return the token (in a real app, you'd email this to the user)
	return resetToken, nil
}

// ResetPassword resets a user's password using a valid reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find user by reset token
	query := `
		SELECT user_id FROM auth.users 
		WHERE password_reset_token = $1 
		AND password_reset_expires_at > NOW()`

	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		return ErrInvalidToken
	}

	// Update the password
	return s.userRepo.UpdatePassword(ctx, userID, newPassword)
}

// ChangePassword changes a user's password (when they know their current password)
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get the user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if !s.userRepo.VerifyPassword(user, currentPassword) {
		return ErrInvalidCredentials
	}

	// Update to new password
	return s.userRepo.UpdatePassword(ctx, userID, newPassword)
}

// Creates an audit log entry
func (s *AuthService) createAuditLog(ctx context.Context, log *models.AuditLog) (uuid.UUID, error) {
	query := `
		INSERT INTO auth.audit_log (
			user_id, event_type, ip_address, user_agent, details
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING log_id`

	var logID uuid.UUID
	err := s.pool.QueryRow(ctx, query,
		log.UserID, log.EventType, log.IPAddress, log.UserAgent, log.Details,
	).Scan(&logID)

	return logID, err
}

// Helper function to generate a secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
