-- Revert interview session scoping from snapshots and timeline

DROP INDEX IF EXISTS idx_timeline_session;
DROP INDEX IF EXISTS idx_code_snapshots_session;

ALTER TABLE timeline
    DROP COLUMN IF EXISTS interview_session_id;

ALTER TABLE code_snapshots
    DROP COLUMN IF EXISTS interview_session_id;
