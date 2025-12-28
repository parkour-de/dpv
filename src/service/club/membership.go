package club

import (
	"context"
	"dpv/dpv/src/repository/t"
)

// Apply marks a club's membership as requested.
func (s *Service) Apply(ctx context.Context, key string, userKey string) error {
	club, err := s.GetClub(ctx, key, userKey)
	if err != nil {
		return err
	}

	m := club.GetMembership()
	if m.Mitgliedsstatus != "none" && m.Mitgliedsstatus != "cancelled" {
		return t.Errorf("cannot apply: current status is %s", m.Mitgliedsstatus)
	}

	m.Mitgliedsstatus = "requested"
	return s.DB.UpdateClub(ctx, club)
}

// Approve marks a club's membership as approved.
func (s *Service) Approve(ctx context.Context, key string) error {
	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return err
	}

	m := club.GetMembership()
	if m.Mitgliedsstatus != "requested" {
		return t.Errorf("cannot approve: current status is %s", m.Mitgliedsstatus)
	}

	m.Mitgliedsstatus = "approved"
	return s.DB.UpdateClub(ctx, club)
}

// Deny marks a club's membership as denied.
func (s *Service) Deny(ctx context.Context, key string) error {
	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return err
	}

	m := club.GetMembership()
	if m.Mitgliedsstatus != "requested" {
		return t.Errorf("cannot deny: current status is %s", m.Mitgliedsstatus)
	}

	m.Mitgliedsstatus = "denied"
	return s.DB.UpdateClub(ctx, club)
}

// Cancel marks a club's membership as cancelled or none.
func (s *Service) Cancel(ctx context.Context, key string, userKey string) error {
	club, err := s.GetClub(ctx, key, userKey)
	if err != nil {
		return err
	}

	m := club.GetMembership()
	if m.Mitgliedsstatus == "approved" {
		m.Mitgliedsstatus = "cancelled"
	} else {
		m.Mitgliedsstatus = "none"
	}

	return s.DB.UpdateClub(ctx, club)
}
