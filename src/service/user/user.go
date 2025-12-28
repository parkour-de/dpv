package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/security"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/email"
	"fmt"
	"time"
)

type Service struct {
	DB *graph.Db
}

func NewService(db *graph.Db) *Service {
	return &Service{DB: db}
}

func (s *Service) CreateUser(ctx context.Context, user *entities.User, password string) error {
	if user.Vorname == "" {
		return t.Errorf("vorname must not be empty")
	}
	if user.Name == "" {
		return t.Errorf("name must not be empty")
	}
	if user.Email == "" {
		return t.Errorf("email must not be empty")
	}
	if password == "" {
		return t.Errorf("password must not be empty")
	}
	if ok, err := security.IsStrongPassword(password); !ok {
		return err
	}

	existing, err := s.DB.GetUsersByEmail(ctx, user.Email)
	if err != nil {
		return t.Errorf("could not check for existing user: %w", err)
	}
	if len(existing) > 0 {
		return t.Errorf("user with this email already exists")
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return t.Errorf("could not hash password: %w", err)
	}
	user.PasswordHash = hash
	return s.DB.Users.Create(user, ctx)
}

func (s *Service) UpdateMe(ctx context.Context, newVorname, newName string) error {
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return t.Errorf("user not found in context")
	}
	updated := false
	if newVorname != "" {
		user.Vorname = newVorname
		updated = true
	}
	if newName != "" {
		user.Name = newName
		updated = true
	}
	if !updated {
		return t.Errorf("no fields to update")
	}
	return s.DB.Users.Update(user, ctx)
}

// RequestEmailValidation sends validation email
func (s *Service) RequestEmailValidation(ctx context.Context, newEmail string) error {
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return t.Errorf("user not found in context")
	}

	// Check if new email is already taken (if different from current)
	if newEmail != "" && newEmail != user.Email {
		existing, err := s.DB.GetUsersByEmail(ctx, newEmail)
		if err != nil {
			return t.Errorf("could not check email availability: %w", err)
		}
		if len(existing) > 0 {
			return t.Errorf("email address already in use")
		}
	}

	// Generate validation token
	expiry := time.Now().Add(12 * time.Hour).Unix()
	targetEmail := newEmail
	if targetEmail == "" {
		targetEmail = user.Email
	}

	// Use command and parameter for token
	command := "validate-email"
	parameter := targetEmail

	token, err := security.GenerateValidationToken(command, user.Key, expiry, parameter, user.PasswordHash, dpv.ConfigInstance.Email.ValidationSecret)
	if err != nil {
		return t.Errorf("could not generate validation token: %w", err)
	}

	// Build validation URL
	baseURL := "http://localhost:8080" // TODO: Make this configurable
	validationURL := fmt.Sprintf("%s/dpv/users/validate-email?key=%s&expiry=%d&email=%s&token=%s",
		baseURL, user.Key, expiry, parameter, token)

	// Send email
	emailService := email.NewService(dpv.ConfigInstance)
	return emailService.SendEmailValidationEmail(email.ValidationData{
		User:          user,
		ValidationURL: validationURL,
		ExpiryTime:    time.Unix(expiry, 0),
		NewEmail:      newEmail,
		IsEmailChange: newEmail != "" && newEmail != user.Email,
	})
}

// RequestPasswordReset sends password reset email to user by email
func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) error {
	users, err := s.DB.GetUsersByEmail(ctx, emailAddr)
	if err != nil {
		return t.Errorf("could not find user: %w", err)
	}
	if len(users) == 0 {
		return t.Errorf("user with this email does not exist")
	}
	user := users[0]

	expiry := time.Now().Add(12 * time.Hour).Unix()
	command := "change-password"
	parameter := ""

	token, err := security.GenerateValidationToken(command, user.Key, expiry, parameter, user.PasswordHash, dpv.ConfigInstance.Email.ValidationSecret)
	if err != nil {
		return t.Errorf("could not generate password reset token: %w", err)
	}

	baseURL := "http://localhost:8080" // TODO: Make this configurable
	resetURL := fmt.Sprintf("%s/dpv/users/reset-password?key=%s&expiry=%d&token=%s",
		baseURL, user.Key, expiry, token)

	emailService := email.NewService(dpv.ConfigInstance)
	return emailService.SendPasswordResetEmail(email.PasswordResetData{
		User:       &user,
		ResetURL:   resetURL,
		ExpiryTime: time.Unix(expiry, 0),
	})
}

// ValidateEmail processes email validation from link
func (s *Service) ValidateEmail(ctx context.Context, userKey string, expiry int64, email string, token string) error {
	// Check if expired
	if time.Now().Unix() > expiry {
		return t.Errorf("validation link has expired")
	}

	// Get user by key
	user, err := s.DB.Users.Read(userKey, ctx)
	if err != nil {
		return t.Errorf("user not found: %w", err)
	}

	// Validate token
	if !security.ValidateToken("validate-email", userKey, expiry, email, user.PasswordHash,
		dpv.ConfigInstance.Email.ValidationSecret, token) {
		return t.Errorf("invalid validation token")
	}

	// Update user (for validate-email command)
	now := time.Now()
	user.EmailVerified = &now
	user.Email = email
	return s.DB.Users.Update(user, ctx)
}

// ValidatePasswordReset processes password reset from link
func (s *Service) ValidatePasswordReset(ctx context.Context, userKey string, expiry int64, token string, newPassword string) error {
	if time.Now().Unix() > expiry {
		return t.Errorf("password reset link has expired")
	}

	user, err := s.DB.Users.Read(userKey, ctx)
	if err != nil {
		return t.Errorf("user not found: %w", err)
	}

	if !security.ValidateToken("change-password", userKey, expiry, "", user.PasswordHash,
		dpv.ConfigInstance.Email.ValidationSecret, token) {
		return t.Errorf("invalid password reset token")
	}

	if newPassword == "" {
		return t.Errorf("password must not be empty")
	}
	if ok, err := security.IsStrongPassword(newPassword); !ok {
		return err
	}

	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return t.Errorf("could not hash password: %w", err)
	}
	user.PasswordHash = hash
	return s.DB.Users.Update(user, ctx)
}
func (s *Service) UpdateRoles(ctx context.Context, userKey string, roles []string) error {
	user, err := s.DB.Users.Read(userKey, ctx)
	if err != nil {
		return t.Errorf("user not found: %w", err)
	}

	user.Roles = roles
	return s.DB.Users.Update(user, ctx)
}
