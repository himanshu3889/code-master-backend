package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

// Create the code snapshot
func (s *Store) CreateCodeSnapshot(ctx context.Context, snapshot *models.CodeSnapshot) *appError.Error {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		return appError.NewInternal("Failed to begin transaction")
	}
	defer tx.Rollback()

	// Generate IDs and timestamps
	snapshot.ID = utils.GenerateSnowflakeID()
	snapshot.CreatedAt = time.Now()

	timeline := &models.Timeline{
		ID:             utils.GenerateSnowflakeID(),
		CreatedAt:      time.Now(),
		ProblemID:      snapshot.ProblemID,
		CodeSnapshotID: &snapshot.ID, // Link to the snapshot we're creating
		SubmissionID:   nil,          // No submission for snapshot-only events
	}

	// 1. Insert code snapshot
	query := `
		INSERT INTO code_snapshots (id, problem_id, language, code, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := tx.GetContext(ctx, snapshot, query,
		snapshot.ID,
		snapshot.ProblemID,
		snapshot.Language,
		snapshot.Code,
		snapshot.CreatedAt,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id": snapshot.ProblemID,
			"language":   snapshot.Language,
		}).WithError(err).Error("Failed to create code snapshot in transaction")
		return appError.NewInternal("Failed to create code snapshot in transaction")
	}

	// 2. Insert timeline entry
	timeQuery := `
		INSERT INTO timeline (id, problem_id, submission_id, code_snapshot_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	if err := tx.GetContext(ctx, timeline, timeQuery,
		timeline.ID,
		timeline.ProblemID,
		timeline.SubmissionID,   // nil
		timeline.CodeSnapshotID, // points to snapshot.ID
		timeline.CreatedAt,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"snapshot_id": snapshot.ID,
			"problem_id":  snapshot.ProblemID,
		}).WithError(err).Error("Failed to create timeline entry for snapshot")
		return appError.NewInternal("Failed to create timeline entry for snapshot")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit snapshot transaction")
		return appError.NewInternal("Failed to commit snapshot transaction")
	}

	return nil
}

// Get problem code snapshots
func (s *Store) GetCodeSnapshotsByProblem(ctx context.Context, problemID int64, limit int) ([]*models.CodeSnapshot, *appError.Error) {
	var snapshots []*models.CodeSnapshot
	query := `
		SELECT id, problem_id, language, code, created_at 
		FROM code_snapshots 
		WHERE problem_id = $1 
		ORDER BY created_at DESC
		LIMIT $2
	`
	err := s.db.SelectContext(ctx, &snapshots, query, problemID, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("Problem code snapshots not found")
		}
		logrus.WithFields(logrus.Fields{
			"problem_id": problemID,
			"limit":      limit,
		}).WithError(err).Error("Failed to get code snapshots")
		return nil, appError.NewInternal("Failed to get code snapshot")
	}
	return snapshots, nil
}

// Get the particular code snapshot
func (s *Store) GetCodeSnapshotByID(ctx context.Context, id int64) (*models.CodeSnapshot, *appError.Error) {
	var snapshot models.CodeSnapshot
	query := `
		SELECT id, problem_id, language, code, created_at 
		FROM code_snapshots 
		WHERE id = $1
	`
	if err := s.db.GetContext(ctx, &snapshot, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("Code snapshot not found")
		}
		logrus.WithField("snapshot_id", id).WithError(err).Error("Failed to get code snapshot")
		return nil, appError.NewInternal("Failed to get code snapshot")
	}
	return &snapshot, nil
}
