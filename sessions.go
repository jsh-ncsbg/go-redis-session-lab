package main

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrCacheMiss       = errors.New("cache miss")
)

type Session struct {
	TokenHash     string    `json:"token_hash"`
	UserID        int64     `json:"user_id"`
	CSRFTokenHash string    `json:"csrf_token_hash"`
	Username      string    `json:"username"`
	IsActive      bool      `json:"is_active"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type SessionStore interface {
	Get(ctx context.Context, tokenHash string) (*Session, error)
	Put(ctx context.Context, session *Session) error
	Delete(ctx context.Context, tokenHash string) error
}

type SessionCache interface {
	Get(ctx context.Context, tokenHash string) (*Session, error)
	Set(ctx context.Context, session *Session) error
	Delete(ctx context.Context, tokenHash string) error
}
