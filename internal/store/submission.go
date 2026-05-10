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

// Create submission log
func (s *Store) CreateSubmissionLog(ctx context.Context, submission *models.Submission) *appError.Error {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"problem_id": submission.ProblemID, "submission_id": submission.ID}).Infof("Failed to begin transaction")
		return appError.NewInternal("Failed to create submission log")
	}
	// Rollback if anything fails (safe no-op if already committed)
	defer tx.Rollback()

	// Generate IDs and timestamps
	if submission.ID == 0 {
		submission.ID = utils.GenerateSnowflakeID()
		submission.CreatedAt = time.Now()
	}

	timeline := &models.Timeline{
		ID:           submission.ID,
		CreatedAt:    submission.CreatedAt,
		ProblemID:    submission.ProblemID,
		SubmissionID: &submission.ID,
	}

	// Insert submission
	subQuery := `
        INSERT INTO submissions 
        (id, problem_id, stdin, language, test_cases, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	if err := tx.GetContext(ctx, submission, subQuery,
		submission.ID,
		submission.ProblemID,
		submission.Stdin, // Note: You might want to check if this should be log.CodeSnapshot
		submission.Language,
		submission.TestCases,
		submission.CreatedAt,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id":    submission.ProblemID,
			"submission_id": submission.ID,
			"language":      submission.Language,
		}).WithError(err).Error("Failed to create submission in transaction")
		return appError.NewInternal("Failed to create submission log")
	}

	// Insert timeline entry
	timeQuery := `
        INSERT INTO timeline (id, problem_id, submission_id, code_snapshot_id, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	if err := tx.GetContext(ctx, timeline, timeQuery,
		timeline.ID,
		timeline.ProblemID,
		timeline.SubmissionID,
		timeline.CodeSnapshotID,
		timeline.CreatedAt,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"submission_id":    submission.ID,
			"problem_id":       submission.ProblemID,
			"code_snapshot_id": timeline.CodeSnapshotID,
		}).WithError(err).Error("Failed to create timeline entry in transaction")
		return appError.NewInternal("Failed to create submission log")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"submission_id":    submission.ID,
			"problem_id":       submission.ProblemID,
			"code_snapshot_id": timeline.CodeSnapshotID,
		}).Error("Failed to commit submission transaction")
		return appError.NewInternal("Failed to create submission log")
	}

	return nil
}

// Update submission log
func (s *Store) UpdateSubmissionExecutionResults(ctx context.Context, submission *models.Submission) *appError.Error {

	// Insert submission
	status := submission.Status
	executionTimeMs := submission.ExecutionTimeMs
	memoryUsedKB := submission.MemoryUsedKb
	stdout := submission.Stdout
	stderr := submission.Stderr
	testResults := submission.TestResults

	query := `
        UPDATE submissions 
        SET status=$2, execution_time_ms=$3, memory_used_kb=$4, stdout=$5, stderr=$6, test_results=$7
        WHERE id = $1
    `
	_, err := s.db.ExecContext(ctx, query,
		submission.ID,
		status,
		executionTimeMs,
		memoryUsedKB,
		stdout,
		stderr,
		testResults,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id": submission.ProblemID,
			"language":   submission.Language,
		}).WithError(err).Error("Failed to update submission execution result")
		return appError.NewInternal("Failed to update the submission execution result")
	}

	return nil
}

// Get the problems submissions
func (s *Store) GetProblemSubmissions(ctx context.Context, problemID snowflake.ID, interviewSessionId *snowflake.ID) ([]*models.Submission, *appError.Error) {
	var logs []*models.Submission
	query := `
		SELECT id, problem_id, interview_session_id, language, status, 
		       execution_time_ms, memory_used_kb, stdin, stdout, stderr, test_cases, test_results, created_at 
		FROM submissions 
		WHERE problem_id = $1 AND interview_session_id IS NULL
		ORDER BY created_at DESC
	`
	args := []interface{}{problemID}
	if interviewSessionId != nil {
		query = `
			SELECT id, problem_id, interview_session_id, language, status, 
				execution_time_ms, memory_used_kb, stdin, stdout, stderr, test_cases, test_results, created_at 
			FROM submissions 
			WHERE problem_id = $1 AND interview_session_id=$2
			ORDER BY created_at DESC
		`
		args = append(args, *interviewSessionId)
	}
	if err := s.db.SelectContext(ctx, &logs, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("Problem submissions not found")
		}
		logrus.WithField("problem_id", problemID).WithError(err).Error("Failed to get submissions")
		return nil, appError.NewInternal("Failed to get problem submissions")
	}
	return logs, nil
}
