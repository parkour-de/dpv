package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/storage"
	"strings"
	"testing"
)

func setupTestMembershipService(t *testing.T) *Service {
	db, _, err := graph.Init("../../../config.yml", true)
	if err != nil {
		t.Fatalf("could not initialize database: %v", err)
	}
	return NewService(db, storage.NewStorage(""))
}

func TestMembershipWorkflow(t *testing.T) {
	s := setupTestMembershipService(t)
	defer s.DB.Database.Remove(context.Background())
	ctx := context.Background()

	club := &entities.Club{
		Name:      "Membership Club",
		LegalForm: "e.V.",
	}
	userKey := "owner"

	err := s.CreateClub(ctx, club, userKey)
	if err != nil {
		t.Fatalf("CreateClub failed: %v", err)
	}
	key := club.GetKey()

	// Initial status should be inactive
	if club.Membership.Status != "inactive" {
		t.Errorf("Initial status should be inactive, got %s", club.Membership.Status)
	}

	// Apply
	err = s.Apply(ctx, key, &entities.User{Entity: entities.Entity{Key: userKey}})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	updated, _ := s.DB.GetClubByKey(ctx, key)
	if updated.Membership.Status != "requested" {
		t.Errorf("Status after Apply should be requested, got %s", updated.Membership.Status)
	}

	// Approve
	err = s.Approve(ctx, key)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	updated, _ = s.DB.GetClubByKey(ctx, key)
	if updated.Membership.Status != "active" {
		t.Errorf("Status after Approve should be active, got %s", updated.Membership.Status)
	}

	// Cancel
	err = s.Cancel(ctx, key, &entities.User{Entity: entities.Entity{Key: userKey}})
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	updated, _ = s.DB.GetClubByKey(ctx, key)
	if updated.Membership.Status != "cancelled" {
		t.Errorf("Status after Cancel should be cancelled, got %s", updated.Membership.Status)
	}

	// Apply again from cancelled
	err = s.Apply(ctx, key, &entities.User{Entity: entities.Entity{Key: userKey}})
	if err != nil {
		t.Fatalf("Apply after Cancel failed: %v", err)
	}

	updated, _ = s.DB.GetClubByKey(ctx, key)
	if updated.Membership.Status != "requested" {
		t.Errorf("Status after Apply (again) should be requested, got %s", updated.Membership.Status)
	}

	// Deny
	err = s.Deny(ctx, key)
	if err != nil {
		t.Fatalf("Deny failed: %v", err)
	}

	updated, _ = s.DB.GetClubByKey(ctx, key)
	if updated.Membership.Status != "denied" {
		t.Errorf("Status after Deny should be denied, got %s", updated.Membership.Status)
	}
}

func TestMembership_InvalidTransitions(t *testing.T) {
	s := setupTestMembershipService(t)
	defer s.DB.Database.Remove(context.Background())
	ctx := context.Background()

	club := &entities.Club{Name: "Invalid Club", LegalForm: "e.V."}
	userKey := "owner"
	err := s.CreateClub(ctx, club, userKey)
	if err != nil {
		t.Fatalf("CreateClub failed: %v", err)
	}
	key := club.GetKey()

	// Cannot approve if not requested
	err = s.Approve(ctx, key)
	if err == nil || !strings.Contains(err.Error(), "Antrag kann nicht bewilligt werden") {
		t.Errorf("Expected 'cannot approve' error, got %v", err)
	}

	// Cannot deny if not requested
	err = s.Deny(ctx, key)
	if err == nil || !strings.Contains(err.Error(), "Antrag kann nicht abgelehnt werden") {
		t.Errorf("Expected 'cannot deny' error, got %v", err)
	}

	// Mark as active manually
	club.Membership.Status = "active"
	s.DB.UpdateClub(ctx, club)

	// Cannot apply if already active
	err = s.Apply(ctx, key, &entities.User{Entity: entities.Entity{Key: userKey}})
	if err == nil || !strings.Contains(err.Error(), "Antrag kann nicht gestellt werden") {
		t.Errorf("Expected 'cannot apply' error, got %v", err)
	}
}
