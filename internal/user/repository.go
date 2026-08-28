package user

import (
	"context"
	"database/sql"
	"errors"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password_hash = $3, is_active = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	args := []any{u.Username, u.Email, u.PasswordHash, u.IsActive, u.ID}
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *repository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	args := []any{u.Username, u.Email, u.PasswordHash, u.IsActive}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
