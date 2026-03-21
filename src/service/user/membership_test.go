package user

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"testing"
)

func TestUserMembershipLifecycle(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	user := &entities.User{FirstName: "Member", LastName: "User", Email: "member@example.com"}
	err := service.CreateUser(ctx, user, "StrongPass1!")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	t.Run("InitialState", func(t *testing.T) {
		if user.Membership.Status != "" {
			t.Fatalf("expected empty membership status, got %s", user.Membership.Status)
		}
	})

	t.Run("Apply", func(t *testing.T) {
		err = service.Apply(ctx, user.Key, "active", 10.0)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "requested" || u.Membership.BeginDate != 0 || u.Membership.ApplicationDate == 0 {
			t.Fatalf("expected requested status with beginDate 0 and ApplicationDate set, got %s / %d / %d", u.Membership.Status, u.Membership.BeginDate, u.Membership.ApplicationDate)
		}
	})

	t.Run("Approve", func(t *testing.T) {
		err = service.Approve(ctx, user.Key, 2000)
		if err != nil {
			t.Fatalf("Approve failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "active" || u.Membership.BeginDate != 2000 {
			t.Fatalf("expected active status with beginDate 2000, got %s / %d", u.Membership.Status, u.Membership.BeginDate)
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		err = service.Cancel(ctx, user.Key)
		if err != nil {
			t.Fatalf("Cancel failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "cancelling" {
			t.Fatalf("expected cancelling status, got %s", u.Membership.Status)
		}

		// Force to cancelled so subsequent tests can re-apply
		u.Membership.Status = "cancelled"
		service.DB.Users.Update(u, ctx)
	})

	t.Run("Reapply", func(t *testing.T) {
		err = service.Apply(ctx, user.Key, "active", 10.0)
		if err != nil {
			t.Fatalf("Re-apply failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "requested" {
			t.Fatalf("expected requested status after re-apply, got %s", u.Membership.Status)
		}
	})

	t.Run("Deny", func(t *testing.T) {
		err = service.Deny(ctx, user.Key)
		if err != nil {
			t.Fatalf("Deny failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "denied" {
			t.Fatalf("expected denied status, got %s", u.Membership.Status)
		}
	})

	t.Run("CancelFromDenied", func(t *testing.T) {
		err = service.Cancel(ctx, user.Key)
		if err != nil {
			t.Fatalf("Cancel from denied failed: %v", err)
		}
		u, _ := service.DB.Users.Read(user.Key, ctx)
		if u.Membership.Status != "cancelled" {
			t.Fatalf("expected cancelled status, got %s", u.Membership.Status)
		}
	})
}
