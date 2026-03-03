package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/membership"
)

// Apply marks a club's membership as requested.
func (s *Service) Apply(ctx context.Context, key string, user *entities.User, beginDate int64) error {
	club, err := s.GetClub(ctx, key, user)
	if err != nil {
		return t.Errorf("failed to load club for membership application: %w", err)
	}

	validateFn := func() error {
		if club.ParentKey != "" {
			return t.Errorf("club is a subsidiary and cannot hold independent membership")
		}

		hasAuthRep := false
		for _, v := range club.Vorstand {
			if v.AuthorizedRepresentative {
				hasAuthRep = true
				break
			}
		}
		if !hasAuthRep && len(club.Vorstand) > 0 {
			return t.Errorf("at least one manager must be an authorized representative (§26 BGB)")
		}
		return nil
	}

	if err := membership.Apply(ctx, club, beginDate, validateFn); err != nil {
		return err
	}

	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership application: %w", err)
	}
	return nil
}

// Approve marks a club's membership as approved.
func (s *Service) Approve(ctx context.Context, key string, beginDate int64) error {
	club, err := s.DB.GetClubByKey(ctx, key)
	if err != nil {
		return t.Errorf("failed to load club for approval: %w", err)
	}

	if err := membership.Approve(ctx, club, beginDate); err != nil {
		return err
	}

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

	if err := membership.Deny(ctx, club); err != nil {
		return err
	}

	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership denial: %w", err)
	}
	return nil
}

// Cancel marks a club's membership as cancelled or none.
func (s *Service) Cancel(ctx context.Context, key string, user *entities.User, endDate int64) error {
	club, err := s.GetClub(ctx, key, user)
	if err != nil {
		return t.Errorf("failed to load club for membership cancellation: %w", err)
	}

	if err := membership.Cancel(ctx, club, endDate); err != nil {
		return err
	}

	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership cancellation: %w", err)
	}
	return nil
}
