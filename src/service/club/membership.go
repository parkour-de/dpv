package club

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/email"
	"dpv/dpv/src/service/membership"
	"fmt"
)

// Apply marks a club's membership as requested.
func (s *Service) Apply(ctx context.Context, key string, user *entities.User, beginDate int64, memType string, fee float64) error {
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

	if memType == "" {
		memType = "ordinary"
	}
	if err := membership.Apply(ctx, club, beginDate, memType, fee, validateFn); err != nil {
		return err
	}

	if err := s.DB.UpdateClub(ctx, club); err != nil {
		return t.Errorf("failed to update club for membership application: %w", err)
	}

	emailService := email.NewService(dpv.ConfigInstance)
	_ = emailService.SendApplicationReceiptEmail(user, club)
	_ = emailService.SendApplicationNoticeEmail(user, club)

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

	if club.Membership.MembershipNumber == "" {
		seq, err := s.DB.GetNextSequence(ctx, "V")
		if err != nil {
			return t.Errorf("failed to generate membership number: %w", err)
		}
		club.Membership.MembershipNumber = fmt.Sprintf("V-%03d-%03d", seq/1000, seq%1000)
	}

	club.Membership.CurrentFee = float64(club.Members) * 1.0
	votes := (club.Members / 100) + 1
	if votes > 5 {
		votes = 5
	}
	club.Membership.CurrentVotes = votes

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
