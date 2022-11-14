-- Filename: migrations/000001_create_quotes_table.up.sql

CREATE TABLE IF NOT EXISTS quotes (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    author text NOT NULL,
    quote_string text NOT NULL,
    category text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);