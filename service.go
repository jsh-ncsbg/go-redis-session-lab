package main

import (
	"context"
	"errors"
	"log"
)

type SessionService struct {
	cache SessionCache
	repo  SessionStore
}

func (s *SessionService) Get(ctx context.Context, tokenHash string) (*Session, error) {
	session, err := s.cache.Get(ctx, tokenHash)
	if err == nil {
		log.Println("cache hit")
		return session, nil
	}

	switch {
	case errors.Is(err, ErrCacheMiss):
		log.Println(ErrCacheMiss.Error())
	default:
		log.Printf("cache error: %v\n", err)
	}

	session, err = s.repo.Get(ctx, tokenHash)
	if err == nil {
		log.Println("database hit")
		if err = s.cache.Set(ctx, session); err != nil {
			log.Printf("cache err: could not populate cache, %v\n", err)
		}

		return session, nil
	}

	log.Printf("db error: %v\n", err)
	return nil, err
}

func (s *SessionService) Create(ctx context.Context, session *Session) error {
	if err := s.repo.Put(ctx, session); err != nil {
		return err
	}
	if err := s.cache.Set(ctx, session); err != nil {
		log.Printf("cache err: could not populate cache, %v\n", err)
	}
	return nil
}

func (s *SessionService) Revoke(ctx context.Context, tokenHash string) error {
	if err := s.repo.Delete(ctx, tokenHash); err != nil {
		return err
	}
	return s.cache.Delete(ctx, tokenHash)
}
