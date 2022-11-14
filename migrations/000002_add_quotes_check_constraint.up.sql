-- Filename: migrations/000002_add_quotes_check_constraint.up.sql

ALTER TABLE quotes ADD CONSTRAINT category_length_check CHECK (array_length(category, 1) BETWEEN 1 AND 10);
