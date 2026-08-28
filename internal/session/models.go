package session

import (
	"context"
	"time"
)

type SessionUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type Session struct {
	TokenHash     string `json:"token_hash"`
	UserID        int64  `json:"user_id"`
	CSRFTokenHash string `json:"csrf_token_hash"`
	// IPAddress     string    `json:"ip_address"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	User SessionUser `json:"user"`
}

func New(tokenHash string, userID int64, csrfTokenHash string, expiresAt time.Time) *Session {
	return &Session{
		TokenHash:     tokenHash,
		UserID:        userID,
		CSRFTokenHash: csrfTokenHash,
		ExpiresAt:     expiresAt,
	}
}

type Repository interface {
	Get(ctx context.Context, tokenHash string) (*Session, error)
	Create(ctx context.Context, session *Session) error
	Delete(ctx context.Context, tokenHash string) error
}

type SessionService struct {
	cache SessionCache
	repo  Repository
}

type SessionCache interface {
	Get(ctx context.Context, tokenHash string) (*Session, error)
	Set(ctx context.Context, session *Session) error
	Delete(ctx context.Context, tokenHash string) error
}
