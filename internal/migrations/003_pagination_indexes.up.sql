-- Composite index for efficient status-filtered pagination ordered by snowflake ID
DROP INDEX IF EXISTS CREATE INDEX idx_problems_status;
CREATE INDEX IF NOT EXISTS idx_problems_status_id ON problems(status, id DESC);
