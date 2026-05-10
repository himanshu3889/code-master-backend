-- Enable trigram extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Make the index on the problems name
CREATE INDEX idx_problems_trgm 
ON problems 
USING gin(name gin_trgm_ops);

