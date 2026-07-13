-- Add interview session scoping to snapshots and timeline

ALTER TABLE code_snapshots
    ADD COLUMN IF NOT EXISTS interview_session_id BIGINT NULL REFERENCES interview_sessions(id) ON DELETE SET NULL;

ALTER TABLE timeline
    ADD COLUMN IF NOT EXISTS interview_session_id BIGINT NULL REFERENCES interview_sessions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_code_snapshots_session ON code_snapshots(problem_id, interview_session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_timeline_session ON timeline(problem_id, interview_session_id, created_at DESC);
