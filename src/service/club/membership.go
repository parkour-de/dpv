package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
)

// Apply marks a club's membership as requested.
func (s *Service) Apply(ctx context.Context, key string, user *entities.User) error {
	club, err := s.GetClub(ctx, key, user)
	if err != nil {
		return t.Errorf("failed to load club for membership application: %w", err)
	}

	m := club.GetMembership()
	if m.Status != "inactive" && m.Status != "cancelled" {
		return t.Errorf("cannot apply: current status is %s", m.Status)
	}

	m.Status = "requested"
	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership application: %w", err)
	}
	return nil
}

// Approve marks a club's membership as approved.
func (s *Service) Approve(ctx context.Context, key string) error {
	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return t.Errorf("failed to load club for approval: %w", err)
	}

	m := club.GetMembership()
	if m.Status != "requested" {
		return t.Errorf("cannot approve: current status is %s", m.Status)
	}

	m.Status = "active"
	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership approval: %w", err)
	}
	return nil
}

// Deny marks a club's membership as denied.
func (s *Service) Deny(ctx context.Context, key string) error {
	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return t.Errorf("failed to load club for denial: %w", err)
	}

	m := club.GetMembership()
	if m.Status != "requested" {
		return t.Errorf("cannot deny: current status is %s", m.Status)
	}

	m.Status = "denied"
	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership denial: %w", err)
	}
	return nil
}

// Cancel marks a club's membership as cancelled or none.
func (s *Service) Cancel(ctx context.Context, key string, user *entities.User) error {
	club, err := s.GetClub(ctx, key, user)
	if err != nil {
		return t.Errorf("failed to load club for membership cancellation: %w", err)
	}

	m := club.GetMembership()
	if m.Status == "active" {
		m.Status = "cancelled"
	} else {
		m.Status = "inactive"
	}

	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership cancellation: %w", err)
	}
	return nil
}
