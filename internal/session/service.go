package session

import (
	"context"
	"errors"
	"log"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrCacheMiss       = errors.New("cache miss")
)

func NewService(cache SessionCache, repo Repository) *SessionService {
	return &SessionService{cache: cache, repo: repo}
}

func (s *SessionService) Get(ctx context.Context, tokenHash string) (*Session, error) {
	session, err := s.cache.Get(ctx, tokenHash)
	if err == nil {
		if session.ExpiresAt.After(time.Now()) {
			return session, nil
		}

		if err := s.cache.Delete(ctx, tokenHash); err != nil {
			log.Printf("cache delete: %v", err)
		}
		return nil, ErrSessionNotFound
	}

	if !errors.Is(err, ErrCacheMiss) {
		log.Printf("cache error: %v\n", err)
	}

	session, err = s.repo.Get(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, session); err != nil {
		log.Printf("could not populate cache: %v\n", err)
	}

	return session, nil
}

func (s *SessionService) Create(ctx context.Context, session *Session) error {
	if err := s.repo.Create(ctx, session); err != nil {
		return err
	}
	return nil
}

func (s *SessionService) Revoke(ctx context.Context, tokenHash string) error {
	if err := s.repo.Delete(ctx, tokenHash); err != nil {
		return err
	}
	return s.cache.Delete(ctx, tokenHash)
}
