package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	T "dpv/dpv/src/repository/t"
	"testing"
)

func setupTestService(t *testing.T) *Service {
	db, _, err := graph.Init("../../../config.yml", true)
	if err != nil {
		t.Fatalf("DB initialization failed: %s", err)
	}
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
		{entities.User{Vorname: "", Name: "N", Email: "e"}, "StrongPass1!", T.T("vorname must not be empty")},
		{entities.User{Vorname: "V", Name: "", Email: "e"}, "StrongPass1!", T.T("name must not be empty")},
		{entities.User{Vorname: "V", Name: "N", Email: ""}, "StrongPass1!", T.T("email must not be empty")},
		{entities.User{Vorname: "V", Name: "N", Email: "e"}, "", T.T("password must not be empty")},
		{entities.User{Vorname: "V", Name: "N", Email: "e"}, "1234567890", T.T("must not be only digits")},
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

	user1 := &entities.User{Vorname: "V", Name: "N", Email: "duplicate@example.com"}
	user2 := &entities.User{Vorname: "V", Name: "N", Email: "duplicate@example.com"}
	password := "StrongPass1!"

	// First creation should succeed
	err := service.CreateUser(ctx, user1, password)
	if err != nil {
		t.Fatalf("First user creation failed: %s", err)
	}

	// Second creation should fail due to duplicate email
	err = service.CreateUser(ctx, user2, password)
	if err == nil || err.Error() != T.T("user with this email already exists") {
		t.Errorf("expected duplicate email error, got '%v'", err)
	}
}
