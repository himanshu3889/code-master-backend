CREATE TABLE IF NOT EXISTS language (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(30) NOT NULL UNIQUE,
    extension VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ENUMs for the problem
CREATE TYPE problem_status AS ENUM (
    'TODO',           -- Not started
    'IN_PROGRESS',    -- Currently coding
    'SOLVED',         -- Success (Independent)
    'REVIEW_NEEDED',  -- Solved, but needs optimization/cleanup
    'STUCK',          -- Attempted, but hit a wall
    'SKIPPED',        -- Decided not to do it
    'REDO_LIST',      -- Solved, but marked for a mandatory retry later
    'ARCHIVED'        -- Old or irrelevant
);
CREATE TYPE problem_difficulty_level AS ENUM ('EASY', 'MEDIUM', 'HARD');

CREATE TABLE IF NOT EXISTS problems (
    id BIGINT PRIMARY KEY,
    name VARCHAR(512) NOT NULL,
    group_name VARCHAR(512), -- Platform name
    description TEXT NOT NULL DEFAULT '',
    difficulty_level problem_difficulty_level DEFAULT NULL, 
    rating INTEGER DEFAULT NULL, -- difficulty rating
    time_limit_ms INTEGER DEFAULT NULL,
    memory_limit_mb INTEGER DEFAULT NULL,
    test_cases JSONB NOT NULL DEFAULT '[]',
    raw_payload TEXT NOT NULL, -- Original problem description/metadata
    url TEXT, 
    saved_code JSONB DEFAULT '{}', -- eg. {"python": "...", "cpp": "..."}
    status problem_status DEFAULT 'TODO',
    notes TEXT DEFAULT '',
    attempt_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- What happened in the end?
CREATE TYPE interview_session_status AS ENUM (
    'IN_PROGRESS',
    'COMPLETED',      -- User hit submit before time ran out
    'TIMEOUT',        -- Mock mode: 45 min hit, editor locked, auto-submitted
    'ABANDONED'       -- User closed the tab without submitting
);

CREATE TABLE IF NOT EXISTS interview_sessions (
    id BIGINT PRIMARY KEY,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    
    status interview_session_status NOT NULL DEFAULT 'IN_PROGRESS',
    
    time_limit_seconds INTEGER NOT NULL DEFAULT 0, -- 2700 for mock, 0 or NULL for casual
    
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,

    auto_submitted BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- ENUM for submission status 
CREATE TYPE submission_status AS ENUM (
    'AC',   -- Accepted
    'WA',   -- Wrong Answer
    'TLE',  -- Time Limit Exceeded
    'MLE',  -- Memory Limit Exceeded
    'RE',   -- Runtime Error
    'CE',   -- Compilation Error
    'SKIPPED',
    'PENDING'
);

CREATE TABLE IF NOT EXISTS submissions (
    id BIGINT PRIMARY KEY,
    interview_session_id BIGINT NULL REFERENCES interview_sessions(id) ON DELETE SET NULL,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    language VARCHAR(15) REFERENCES language(code) ON DELETE CASCADE,
    status submission_status DEFAULT 'PENDING',
    execution_time_ms BIGINT NULL,
    memory_used_kb BIGINT NULL,
    stdin TEXT NOT NULL,
    stdout TEXT NULL,
    stderr TEXT NULL,
    test_cases JSONB,
    test_results JSONB, -- Array of {test_index, passed, actual_output, expected_output}
    created_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS code_snapshots (
    id BIGINT PRIMARY KEY,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    language VARCHAR(15) REFERENCES language(code) ON DELETE CASCADE,
    code TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS timeline (
    id BIGINT PRIMARY KEY,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    submission_id BIGINT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    code_snapshot_id BIGINT NULL REFERENCES code_snapshots(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_language_code ON language(code);
CREATE INDEX idx_problems_created_at ON problems(created_at DESC);
CREATE INDEX idx_problems_status ON problems(status);
CREATE INDEX idx_interview_sessions_problem_id ON interview_sessions(problem_id, created_at DESC);
CREATE INDEX idx_submissions_problem_session_id_created_at ON submissions(problem_id, interview_session_id, created_at DESC);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_code_snapshots_id_created_at ON code_snapshots(problem_id, created_at DESC);
CREATE INDEX idx_timeline_problem_id_created_at ON timeline(problem_id, created_at DESC);


