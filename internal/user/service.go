package user

import (
	"context"
	"fmt"
)

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	usr, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", userID, err)
	}
	return usr, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user %d: %w", userID, err)
	}
	return nil
}
