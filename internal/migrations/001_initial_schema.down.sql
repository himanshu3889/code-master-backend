-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS timeline CASCADE;
DROP TABLE IF EXISTS submissions CASCADE;
DROP TABLE IF EXISTS code_snapshots CASCADE;
DROP TABLE IF EXISTS problems CASCADE;
DROP TABLE IF EXISTS language CASCADE;
DROP TABLE IF EXISTS interview_sessions;

-- Drop custom ENUM types
DROP TYPE IF EXISTS submission_status CASCADE;
DROP TYPE IF EXISTS problem_status CASCADE;
DROP TYPE IF EXISTS problem_difficulty_level CASCADE;
DROP TYPE IF EXISTS interview_session_status;