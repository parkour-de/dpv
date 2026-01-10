package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/security"
	"testing"
	"time"
)

func setupTestService(t *testing.T) *Service {
	db, config, err := graph.Init("../../../config.yml", true)
	if err != nil {
		t.Fatalf("DB initialization failed: %s", err)
	}
	dpv.ConfigInstance = config
	return NewService(db)
}

func TestCreateUser_Validation(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	cases := []struct {
		user     entities.User
		password string
		errMsg   string
	}{
		{entities.User{FirstName: "", LastName: "N", Email: "e"}, "StrongPass1!", "firstname must not be empty"},
		{entities.User{FirstName: "V", LastName: "", Email: "e"}, "StrongPass1!", "lastname must not be empty"},
		{entities.User{FirstName: "V", LastName: "N", Email: ""}, "StrongPass1!", "email must not be empty"},
		{entities.User{FirstName: "V", LastName: "N", Email: "e"}, "", "password must not be empty"},
		{entities.User{FirstName: "V", LastName: "N", Email: "e"}, "1234567890", "must not be only digits"},
	}
	for _, c := range cases {
		err := service.CreateUser(ctx, &c.user, c.password)
		if err == nil {
			t.Errorf("expected error '%s', got nil", c.errMsg)
		} else if !contains(err.Error(), c.errMsg) {
			t.Errorf("expected error containing '%s', got '%v'", c.errMsg, err)
		}
	}
}

func contains(s, substr string) bool {
	return substr == "" || (s != "" && (len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))))
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user1 := &entities.User{FirstName: "V", LastName: "N", Email: "duplicate@example.com"}
	user2 := &entities.User{FirstName: "V", LastName: "N", Email: "duplicate@example.com"}
	password := "StrongPass1!"

	// First creation should succeed
	err := service.CreateUser(ctx, user1, password)
	if err != nil {
		t.Fatalf("First user creation failed: %s", err)
	}

	// Second creation should fail due to duplicate email
	err = service.CreateUser(ctx, user2, password)
	if err == nil || err.Error() != "user with this email already exists" {
		t.Errorf("expected duplicate email error, got '%v'", err)
	}
}

func TestUpdateMe(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user := &entities.User{FirstName: "Old", LastName: "Name", Email: "update@example.com"}
	service.CreateUser(ctx, user, "StrongPass1!")

	// Put user in context
	userCtx := context.WithValue(ctx, "user", user)

	// Success
	err := service.UpdateMe(userCtx, "New", "Name")
	if err != nil {
		t.Fatalf("UpdateMe failed: %v", err)
	}

	if user.FirstName != "New" {
		t.Errorf("Expected FirstName 'New', got '%s'", user.FirstName)
	}

	// No fields to update
	err = service.UpdateMe(userCtx, "", "")
	if err == nil || err.Error() != "no fields to update" {
		t.Errorf("Expected 'no fields to update' error, got '%v'", err)
	}

	// User not in context
	err = service.UpdateMe(ctx, "New", "Name")
	if err == nil || err.Error() != "user not found in context" {
		t.Errorf("Expected 'user not found in context' error, got '%v'", err)
	}
}

func TestUpdateRoles(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user := &entities.User{FirstName: "Role", LastName: "User", Email: "roles@example.com"}
	service.CreateUser(ctx, user, "StrongPass1!")

	newRoles := []string{"admin", "editor"}
	err := service.UpdateRoles(ctx, user.Key, newRoles)
	if err != nil {
		t.Fatalf("UpdateRoles failed: %v", err)
	}

	updatedUser, _ := service.DB.Users.Read(user.Key, ctx)
	if len(updatedUser.Roles) != 2 || updatedUser.Roles[0] != "admin" {
		t.Errorf("Roles not updated correctly: %v", updatedUser.Roles)
	}
}

func TestValidateEmail(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user := &entities.User{FirstName: "V", LastName: "N", Email: "old@example.com"}
	service.CreateUser(ctx, user, "StrongPass1!")

	newEmail := "new@example.com"
	expiry := time.Now().Add(time.Hour).Unix()
	token, _ := security.GenerateValidationToken("validate-email", user.Key, expiry, newEmail, user.PasswordHash, dpv.ConfigInstance.Email.ValidationSecret)

	// Success
	err := service.ValidateEmail(ctx, user.Key, expiry, newEmail, token)
	if err != nil {
		t.Fatalf("ValidateEmail failed: %v", err)
	}

	updatedUser, _ := service.DB.Users.Read(user.Key, ctx)
	if updatedUser.Email != newEmail {
		t.Errorf("Expected email %s, got %s", newEmail, updatedUser.Email)
	}
	if updatedUser.EmailVerified == nil {
		t.Error("EmailVerified should not be nil")
	}

	// Expired
	err = service.ValidateEmail(ctx, user.Key, time.Now().Add(-time.Hour).Unix(), newEmail, token)
	if err == nil || err.Error() != "validation link has expired" {
		t.Errorf("Expected 'expired' error, got '%v'", err)
	}

	// Invalid token
	err = service.ValidateEmail(ctx, user.Key, expiry, newEmail, "invalid-token")
	if err == nil || err.Error() != "invalid validation token" {
		t.Errorf("Expected 'invalid validation token' error, got '%v'", err)
	}
}

func TestValidatePasswordReset(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user := &entities.User{FirstName: "V", LastName: "N", Email: "reset@example.com"}
	err := service.CreateUser(ctx, user, "StrongPass1!")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	expiry := time.Now().Add(time.Hour).Unix()
	token, _ := security.GenerateValidationToken("change-password", user.Key, expiry, "", user.PasswordHash, dpv.ConfigInstance.Email.ValidationSecret)

	newPassword := "NewStrongPass2!"
	err = service.ValidatePasswordReset(ctx, user.Key, expiry, token, newPassword)
	if err != nil {
		t.Fatalf("ValidatePasswordReset failed: %v", err)
	}

	updatedUser, _ := service.DB.Users.Read(user.Key, ctx)
	if !security.CheckPasswordHash(updatedUser.PasswordHash, newPassword) {
		t.Error("Password was not updated correctly")
	}

	// Invalid token (old password hash in token data no longer matches)
	err = service.ValidatePasswordReset(ctx, user.Key, expiry, token, "AnotherPass3!")
	if err == nil || err.Error() != "invalid password reset token" {
		t.Errorf("Expected 'invalid password reset token' error, got '%v'", err)
	}
}
