package user

import (
	"context"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/email"
	"dpv/dpv/src/service/membership"
	"fmt"
)

func (s *Service) Apply(ctx context.Context, key string, memType string, fee float64) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for membership application: %w", err)
	}

	if err := membership.Apply(ctx, user, memType, fee, nil); err != nil {
		return err
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership application: %w", err)
	}

	emailService := email.NewService(dpv.ConfigInstance)
	_ = emailService.SendApplicationReceiptEmail(user, nil)
	_ = emailService.SendApplicationNoticeEmail(user, nil)

	return nil
}

func (s *Service) Approve(ctx context.Context, key string, beginDate int64) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for approval: %w", err)
	}

	if err := membership.Approve(ctx, user, beginDate); err != nil {
		return err
	}

	if user.Membership.Type == "supporting" {
		if user.Membership.MembershipNumber == "" {
			seq, err := s.DB.GetNextSequence(ctx, "F")
			if err != nil {
				return t.Errorf("failed to generate membership number: %w", err)
			}
			user.Membership.MembershipNumber = fmt.Sprintf("F-%03d-%03d", seq/1000, seq%1000)
		}
		user.Membership.CurrentFee = user.Membership.Contribution
	} else {
		if user.Membership.MembershipNumber == "" {
			seq, err := s.DB.GetNextSequence(ctx, "A")
			if err != nil {
				return t.Errorf("failed to generate membership number: %w", err)
			}
			user.Membership.MembershipNumber = fmt.Sprintf("A-%03d-%03d", seq/1000, seq%1000)
		}
		user.Membership.CurrentFee = 10.0
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership approval: %w", err)
	}

	emailService := email.NewService(dpv.ConfigInstance)
	_ = emailService.SendApplicationAcceptedEmail(user, nil)

	return nil
}

func (s *Service) Deny(ctx context.Context, key string) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for denial: %w", err)
	}

	if err := membership.Deny(ctx, user); err != nil {
		return err
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership denial: %w", err)
	}
	return nil
}

func (s *Service) Cancel(ctx context.Context, key string) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for membership cancellation: %w", err)
	}

	if err := membership.Cancel(ctx, user); err != nil {
		return err
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership cancellation: %w", err)
	}
	return nil
}
