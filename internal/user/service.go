package user

import (
	"context"
	"errors"
	"fmt"
	"log"
)

func NewService(cache UserCache, repo Repository) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	usr, err := s.cache.Get(ctx, userID)
	if err == nil {
		return usr, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		log.Printf("user cache error: %v\n", err)
	}

	usr, err = s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", userID, err)
	}

	if err := s.cache.Set(ctx, usr); err != nil {
		log.Printf("could not populate user cache: %v\n", err)
	}

	return usr, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete user %d: %w", userID, err)
	}
	return s.cache.Delete(ctx, userID)
}
