package user

import (
	"context"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/membership"
)

func (s *Service) Apply(ctx context.Context, key string, beginDate int64) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for membership application: %w", err)
	}

	if err := membership.Apply(ctx, user, beginDate, nil); err != nil {
		return err
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership application: %w", err)
	}
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

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership approval: %w", err)
	}
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

func (s *Service) Cancel(ctx context.Context, key string, endDate int64) error {
	user, err := s.DB.Users.Read(key, ctx)
	if err != nil {
		return t.Errorf("failed to load user for membership cancellation: %w", err)
	}

	if err := membership.Cancel(ctx, user, endDate); err != nil {
		return err
	}

	if err := s.DB.Users.Update(user, ctx); err != nil {
		return t.Errorf("failed to update user for membership cancellation: %w", err)
	}
	return nil
}
