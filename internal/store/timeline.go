package store

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

// Get the problem timeline
func (s *Store) GetTimelineByProblem(ctx context.Context, problemID int64, interviewSessionID *snowflake.ID, limit int) ([]*models.Timeline, *appError.Error) {
	var entries []*models.Timeline
	var query string
	var args []interface{}

	if interviewSessionID != nil {
		query = `
			SELECT id, problem_id, interview_session_id, submission_id, code_snapshot_id, created_at
			FROM timeline
			WHERE problem_id = $1 AND interview_session_id = $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{problemID, *interviewSessionID, limit}
	} else {
		query = `
			SELECT id, problem_id, interview_session_id, submission_id, code_snapshot_id, created_at
			FROM timeline
			WHERE problem_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{problemID, limit}
	}

	if err := s.db.SelectContext(ctx, &entries, query, args...); err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id": problemID,
			"limit":      limit,
		}).WithError(err).Error("Failed to get timeline")
		return nil, appError.NewInternal("Failed to get the problem timeline")
	}
	return entries, nil
}

// Get the problem detailed timeline
func (s *Store) GetDetailedTimelineByProblem(ctx context.Context, problemID int64, interviewSessionID *snowflake.ID, limit int) ([]*models.DetailedTimeline, *appError.Error) {
	var query string
	var args []interface{}

	if interviewSessionID != nil {
		query = `
			SELECT
				t.id as timeline_id,
				t.created_at as timeline_created_at,
				t.interview_session_id,
				cs.id as code_snapshot_id,
				cs.code as code,
				cs.language as snapshot_language,
				cs.created_at as snapshot_created_at,
				s.id as submission_id,
				s.status as submission_status,
				s.execution_time_ms,
				s.memory_used_kb,
				s.stdin,
				s.stdout,
				s.stderr,
				s.test_cases,
				s.test_results,
				s.created_at as submission_created_at
			FROM timeline t
			LEFT JOIN code_snapshots cs ON t.code_snapshot_id = cs.id
			LEFT JOIN submissions s ON t.submission_id = s.id
			WHERE t.problem_id = $1 AND t.interview_session_id = $2
			ORDER BY t.created_at DESC
			LIMIT $3
		`
		args = []interface{}{problemID, *interviewSessionID, limit}
	} else {
		query = `
			SELECT
				t.id as timeline_id,
				t.created_at as timeline_created_at,
				t.interview_session_id,
				cs.id as code_snapshot_id,
				cs.code as code,
				cs.language as snapshot_language,
				cs.created_at as snapshot_created_at,
				s.id as submission_id,
				s.status as submission_status,
				s.execution_time_ms,
				s.memory_used_kb,
				s.stdin,
				s.stdout,
				s.stderr,
				s.test_cases,
				s.test_results,
				s.created_at as submission_created_at
			FROM timeline t
			LEFT JOIN code_snapshots cs ON t.code_snapshot_id = cs.id
			LEFT JOIN submissions s ON t.submission_id = s.id
			WHERE t.problem_id = $1
			ORDER BY t.created_at DESC
			LIMIT $2
		`
		args = []interface{}{problemID, limit}
	}

	var entries []*models.DetailedTimeline
	if err := s.db.SelectContext(ctx, &entries, query, args...); err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id": problemID,
			"limit":      limit,
		}).WithError(err).Error("Failed to get detailed timeline")
		return nil, appError.NewInternal("Failed to get problem detailed timeline")
	}

	return entries, nil
}
