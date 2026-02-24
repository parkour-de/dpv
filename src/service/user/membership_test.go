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

	// 1. Initial State
	if user.Membership.Status != "" {
		t.Fatalf("expected empty membership status, got %s", user.Membership.Status)
	}

	// 2. Apply
	err = service.Apply(ctx, user.Key, 1000)
	if err != nil {
		t.Fatalf("expected successful Apply, got %v", err)
	}
	u, _ := service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "requested" || u.Membership.BeginDate != 1000 {
		t.Fatalf("expected requested status with beginDate 1000, got %s / %d", u.Membership.Status, u.Membership.BeginDate)
	}

	// 3. Approve
	err = service.Approve(ctx, user.Key, 2000)
	if err != nil {
		t.Fatalf("expected successful Approve, got %v", err)
	}
	u, _ = service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "active" || u.Membership.BeginDate != 2000 {
		t.Fatalf("expected active status with beginDate 2000, got %s / %d", u.Membership.Status, u.Membership.BeginDate)
	}

	// 4. Cancel
	err = service.Cancel(ctx, user.Key, 3000)
	if err != nil {
		t.Fatalf("expected successful Cancel, got %v", err)
	}
	u, _ = service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "cancelled" || u.Membership.EndDate != 3000 {
		t.Fatalf("expected cancelled status with endDate 3000, got %s / %d", u.Membership.Status, u.Membership.EndDate)
	}

	// 5. Re-apply
	err = service.Apply(ctx, user.Key, 4000)
	if err != nil {
		t.Fatalf("expected successful Apply from cancelled, got %v", err)
	}
	u, _ = service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "requested" {
		t.Fatalf("expected requested status after re-apply, got %s", u.Membership.Status)
	}

	// 6. Deny
	err = service.Deny(ctx, user.Key)
	if err != nil {
		t.Fatalf("expected successful Deny, got %v", err)
	}
	u, _ = service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "denied" {
		t.Fatalf("expected denied status, got %s", u.Membership.Status)
	}

	// 7. Cancel from denied (moves to inactive)
	err = service.Cancel(ctx, user.Key, 0)
	if err != nil {
		t.Fatalf("expected successful Cancel from denied, got %v", err)
	}
	u, _ = service.DB.Users.Read(user.Key, ctx)
	if u.Membership.Status != "inactive" {
		t.Fatalf("expected inactive status, got %s", u.Membership.Status)
	}
}
