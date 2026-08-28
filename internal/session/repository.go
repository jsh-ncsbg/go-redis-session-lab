package session

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

func (r *repository) Get(ctx context.Context, tokenHash string) (*Session, error) {
	query := `
		SELECT
			s.token_hash,
			s.user_id,
			s.csrf_token_hash,
			s.created_at,
			s.expires_at,
			u.id,
			u.username,
			u.email,
			u.is_active
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
			AND s.expires_at > NOW()
			AND u.is_active = TRUE
	`

	sess := &Session{}
	if err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&sess.TokenHash,
		&sess.UserID,
		&sess.CSRFTokenHash,
		&sess.CreatedAt,
		&sess.ExpiresAt,
		&sess.User.ID,
		&sess.User.Username,
		&sess.User.Email,
		&sess.User.IsActive,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return sess, nil
}

func (r *repository) Create(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (token_hash, user_id, csrf_token_hash, created_at, expires_at)
		VALUES ($1, $2, $3, NOW(), $4)
		RETURNING created_at
	`
	if err := r.db.QueryRowContext(
		ctx, query,
		session.TokenHash, session.UserID,
		session.CSRFTokenHash, session.ExpiresAt,
	).Scan(&session.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, tokenHash string) error {
	query := `
		DELETE FROM sessions WHERE token_hash = $1
	`
	if _, err := r.db.ExecContext(ctx, query, tokenHash); err != nil {
		return err
	}
	return nil
}
