CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    lichess_username VARCHAR(100) NOT NULL,
    password VARCHAR(60) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);