-- Filename: migrations/000003_add_quotes_indexes.up.sql

CREATE INDEX IF NOT EXISTS quotes_author_idx ON quotes USING GIN(to_tsvector('simple', author));
CREATE INDEX IF NOT EXISTS quotes_quotestring_idx ON quotes USING GIN(to_tsvector('simple', quote_string));
CREATE INDEX IF NOT EXISTS quotes_category_idx ON quotes USING GIN(category);