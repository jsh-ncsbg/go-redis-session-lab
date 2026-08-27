package main

import (
	"context"
	"sync"
	"time"
)

type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: map[string]Session{},
	}
}

func (s *MemorySessionStore) Get(ctx context.Context, tokenHash string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[tokenHash]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionNotFound
	}

	sessionClone := session
	return &sessionClone, nil
}

func (s *MemorySessionStore) Put(ctx context.Context, session *Session) error {
	s.mu.Lock()
	s.sessions[session.TokenHash] = *session
	s.mu.Unlock()
	return nil
}

func (s *MemorySessionStore) Delete(ctx context.Context, tokenHash string) error {
	s.mu.Lock()
	delete(s.sessions, tokenHash)
	s.mu.Unlock()
	return nil
}
