package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

func (s *Store) CreateInterviewSession(ctx context.Context, session *models.InterviewSession) *appError.Error {
	if session.ID == 0 {
		session.ID = utils.GenerateSnowflakeID()
	}
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	session.StartedAt = now
	session.Status = models.InterviewSessionInProgress

	query := `
		INSERT INTO interview_sessions (id, problem_id, status, time_limit_seconds, started_at, auto_submitted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	if err := s.db.GetContext(ctx, session, query,
		session.ID,
		session.ProblemID,
		session.Status,
		session.TimeLimitSeconds,
		session.StartedAt,
		session.AutoSubmitted,
		session.CreatedAt,
		session.UpdatedAt,
	); err != nil {
		logrus.WithError(err).Error("Failed to create interview session")
		return appError.NewInternal("Failed to create interview session")
	}
	return nil
}

func (s *Store) GetInterviewSessionByID(ctx context.Context, id snowflake.ID) (*models.InterviewSession, *appError.Error) {
	var session models.InterviewSession
	query := `
		SELECT id, problem_id, status, time_limit_seconds, started_at, ended_at, auto_submitted, created_at, updated_at
		FROM interview_sessions
		WHERE id = $1
	`
	if err := s.db.GetContext(ctx, &session, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("Interview session not found")
		}
		logrus.WithError(err).Error("Failed to get interview session")
		return nil, appError.NewInternal("Failed to get interview session")
	}
	return &session, nil
}

func (s *Store) GetActiveSessionByProblem(ctx context.Context, problemID snowflake.ID) (*models.InterviewSession, *appError.Error) {
	var session models.InterviewSession
	query := `
		SELECT id, problem_id, status, time_limit_seconds, started_at, ended_at, auto_submitted, created_at, updated_at
		FROM interview_sessions
		WHERE problem_id = $1 AND status = 'IN_PROGRESS'
		ORDER BY started_at DESC
		LIMIT 1
	`
	if err := s.db.GetContext(ctx, &session, query, problemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logrus.WithError(err).Error("Failed to get active interview session")
		return nil, appError.NewInternal("Failed to get active interview session")
	}
	return &session, nil
}

func (s *Store) UpdateInterviewSessionStatus(ctx context.Context, id snowflake.ID, status models.InterviewSessionStatus) *appError.Error {
	now := time.Now()
	var query string
	var args []interface{}
	if status == models.InterviewSessionInProgress {
		query = `UPDATE interview_sessions SET status=$1, updated_at=$2 WHERE id=$3`
		args = []interface{}{status, now, id}
	} else {
		query = `UPDATE interview_sessions SET status=$1, ended_at=$2, updated_at=$3 WHERE id=$4`
		args = []interface{}{status, now, now, id}
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to update interview session status")
		return appError.NewInternal("Failed to update interview session status")
	}
	return nil
}

func (s *Store) SetSessionAutoSubmitted(ctx context.Context, id snowflake.ID) *appError.Error {
	query := `UPDATE interview_sessions SET auto_submitted = true, updated_at = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to set auto submitted flag")
		return appError.NewInternal("Failed to set auto submitted flag")
	}
	return nil
}

func (s *Store) GetInterviewSessionsByProblem(ctx context.Context, problemID snowflake.ID) ([]*models.InterviewSession, *appError.Error) {
	var sessions []*models.InterviewSession
	query := `
		SELECT id, problem_id, status, time_limit_seconds, started_at, ended_at, auto_submitted, created_at, updated_at
		FROM interview_sessions
		WHERE problem_id = $1
		ORDER BY created_at DESC
	`
	if err := s.db.SelectContext(ctx, &sessions, query, problemID); err != nil {
		logrus.WithError(err).Error("Failed to get interview sessions by problem")
		return nil, appError.NewInternal("Failed to get interview sessions")
	}
	return sessions, nil
}

func (s *Store) GetExpiredSessions(ctx context.Context) ([]*models.InterviewSession, *appError.Error) {
	var sessions []*models.InterviewSession
	query := `
		SELECT id, problem_id, status, time_limit_seconds, started_at, ended_at, auto_submitted, created_at, updated_at
		FROM interview_sessions
		WHERE status = 'IN_PROGRESS'
		  AND started_at + INTERVAL '1 second' * time_limit_seconds < NOW()
	`
	if err := s.db.SelectContext(ctx, &sessions, query); err != nil {
		logrus.WithError(err).Error("Failed to get expired interview sessions")
		return nil, appError.NewInternal("Failed to get expired interview sessions")
	}
	return sessions, nil
}
