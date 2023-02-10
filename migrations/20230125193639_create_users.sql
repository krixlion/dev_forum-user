-- +goose Up
CREATE TABLE IF NOT EXISTS "users" (
    id VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp(),
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp()
);

-- +goose Down
DROP TABLE IF EXISTS "users";
