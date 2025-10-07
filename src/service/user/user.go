package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/security"
	"dpv/dpv/src/repository/t"
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
