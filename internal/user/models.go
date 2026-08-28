package user

import (
	"context"
	"fmt"
	"time"
)

var (
	ErrUserNotFound = fmt.Errorf("user not found")
)

type User struct {
	ID int64 `json:"id"`

	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`

	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New(username, email, passwordHash string, isActive bool) *User {
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		IsActive:     isActive,
	}
}

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	Update(ctx context.Context, u *User) error
	Create(ctx context.Context, u *User) error
	Delete(ctx context.Context, userID int64) error
}

type Service struct {
	repo Repository
}
