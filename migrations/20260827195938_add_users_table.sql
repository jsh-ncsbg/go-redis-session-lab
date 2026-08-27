-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,

    username VARCHAR(50) NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_username_lower_idx
    ON users (LOWER(username));
CREATE UNIQUE INDEX users_email_lower_idx
    ON users (LOWER(email));

-- +goose Down
DROP TABLE users;
