package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrMissingPostgresDSN = errors.New("postgres DSN is required")

func NewPostgres(ctx context.Context, dsn string, maxOpenConns int) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, ErrMissingPostgresDSN
	}

	client, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	client.SetMaxOpenConns(maxOpenConns)

	if err := client.PingContext(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return client, nil
}
