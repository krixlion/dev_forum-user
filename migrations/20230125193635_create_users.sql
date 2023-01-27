-- +goose Up
CREATE TABLE IF NOT EXISTS "users" (
    id UUID,
    name VARCHAR,
    email VARCHAR,
    password VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS "users";
