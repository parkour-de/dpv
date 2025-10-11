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

	token, err := security.GenerateValidationToken(user.Key, expiry, targetEmail, user.PasswordHash, dpv.ConfigInstance.Email.ValidationSecret)
	if err != nil {
		return t.Errorf("could not generate validation token: %w", err)
	}

	// Build validation URL
	baseURL := "http://localhost:8080" // TODO: Make this configurable
	validationURL := fmt.Sprintf("%s/dpv/users/validate-email?key=%s&expiry=%d&email=%s&token=%s",
		baseURL, user.Key, expiry, targetEmail, token)

	// Send email
	emailService := email.NewService(dpv.ConfigInstance)
	return emailService.SendValidationEmail(email.ValidationData{
		User:          user,
		ValidationURL: validationURL,
		ExpiryTime:    time.Unix(expiry, 0),
		NewEmail:      newEmail,
		IsEmailChange: newEmail != "" && newEmail != user.Email,
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
	if !security.ValidateToken(userKey, expiry, email, user.PasswordHash,
		dpv.ConfigInstance.Email.ValidationSecret, token) {
		return t.Errorf("invalid validation token")
	}

	// Update user
	now := time.Now()
	user.EmailVerified = &now
	user.Email = email

	return s.DB.Users.Update(user, ctx)
}
