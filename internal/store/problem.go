package store

import (
	"context"
	"errors"

	"database/sql"

	"github.com/bwmarrin/snowflake"
	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

// Create the problem from the companion
func (s *Store) CreateProblemWithCompanion(ctx context.Context, p *models.Problem) *appError.Error {
	query := `
		INSERT INTO problems (id, name, difficulty_level, group_name, url, time_limit_ms, memory_limit_mb, test_cases, raw_payload, description, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`
	p.ID = utils.GenerateSnowflakeID()
	p.TestCases = sqlLib.JSONB[[]models.ProblemTestCase]{
		Data:  []models.ProblemTestCase{}, // Initialize empty slice
		Valid: true,
	}
	if err := s.db.GetContext(ctx, p, query,
		p.ID,
		p.Name,
		p.DifficultyLevel,
		p.Group,
		p.URL,
		p.TimeLimit,
		p.MemoryLimit,
		p.TestCases,
		p.RawPayload,
		p.Description,
		p.Notes,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"problem_name": p.Name,
			"group":        p.Group,
		}).WithError(err).Error("Failed to create problem")
		return appError.NewInternal("Failed to create problem")
	}

	return nil
}

// Save problem code
func (s *Store) SaveProblemCode(ctx context.Context, problemId snowflake.ID, langCode string, code string) *appError.Error {
	// Validation: ensure we aren't saving empty data
	if problemId == 0 || langCode == "" {
		logrus.Warn("Received incomplete save code payload")
		return appError.NewBadRequest("Recieve incomplete save code payload")
	}
	// We use the JSONB concatenation operator (||) to update just one key
	query := `
        UPDATE problems
        SET saved_code = COALESCE(saved_code, '{}'::jsonb) || $1::jsonb
        WHERE id = $2
    `

	// This creates our dynamic JSON fragment: {"python": "code..."}
	updateData := sqlLib.NewJSONB(map[string]string{langCode: code})

	_, err := s.db.ExecContext(ctx, query, updateData, problemId)
	if err != nil {
		logrus.WithError(err).Error("Failed to save problem code")
		return appError.NewInternal("Database error")
	}

	return nil
}

// Update problem status
func (s *Store) UpdateProblemStatus(ctx context.Context, problemId snowflake.ID, status string) *appError.Error {
	// Validation: ensure we aren't saving empty data
	if problemId == 0 {
		logrus.Warn("Received incomplete problem status payload")
		return appError.NewBadRequest("Recieve incomplete problem status payload")
	}

	query := `
        UPDATE problems
        SET status=$1
        WHERE id = $2
    `

	_, err := s.db.ExecContext(ctx, query, status, problemId)
	if err != nil {
		logrus.WithError(err).Error("Failed to save problem status")
		return appError.NewInternal("Database error")
	}

	return nil
}

// Update problem difficulty
func (s *Store) UpdateProblemDifficultyLevel(ctx context.Context, problemId snowflake.ID, difficulty_level string) *appError.Error {
	// Validation: ensure we aren't saving empty data
	if problemId == 0 {
		logrus.Warn("Received incomplete problem difficulty level payload")
		return appError.NewBadRequest("Recieve incomplete problem difficulty level payload")
	}

	query := `
        UPDATE problems
        SET difficulty_level=$1
        WHERE id = $2
    `

	_, err := s.db.ExecContext(ctx, query, difficulty_level, problemId)
	if err != nil {
		logrus.WithError(err).Error("Failed to save problem difficulty level")
		return appError.NewInternal("Database error")
	}

	return nil
}

// Update problem descritpion
func (s *Store) UpdateProblemDescription(ctx context.Context, problemId snowflake.ID, description string) *appError.Error {
	// Validation: ensure we aren't saving empty data
	if problemId == 0 {
		logrus.Warn("Received incomplete problem description payload")
		return appError.NewBadRequest("Recieve incomplete problem description payload")
	}

	query := `
        UPDATE problems
        SET description=$1
        WHERE id = $2
    `

	_, err := s.db.ExecContext(ctx, query, description, problemId)
	if err != nil {
		logrus.WithError(err).Error("Failed to save problem description")
		return appError.NewInternal("Database error")
	}

	return nil
}

// Update problem notes
func (s *Store) UpdateProblemNotes(ctx context.Context, problemId snowflake.ID, notes string) *appError.Error {
	// Validation: ensure we aren't saving empty data
	if problemId == 0 {
		logrus.Warn("Received incomplete problem notes payload")
		return appError.NewBadRequest("Recieve incomplete problem notes payload")
	}

	query := `
        UPDATE problems
        SET notes=$1
        WHERE id = $2
    `

	_, err := s.db.ExecContext(ctx, query, notes, problemId)
	if err != nil {
		logrus.WithError(err).Error("Failed to save problem notes")
		return appError.NewInternal("Database error")
	}

	return nil
}

// CountProblems returns the total number of problems, optionally filtered by status.
func (s *Store) CountProblems(ctx context.Context, status string) (int64, *appError.Error) {
	var count int64
	var err error

	if status == "" {
		err = s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM problems`)
	} else {
		err = s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM problems WHERE status = $1`, status)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to count problems")
		return 0, appError.NewInternal("Failed to count problems")
	}
	return count, nil
}

// GetLatestProblems returns the newest problems first using page-based pagination.
// Ordered by snowflake ID (DESC) which is chronologically sortable.
func (s *Store) GetLatestProblems(ctx context.Context, limit int, offset int, status string) ([]*models.Problem, *appError.Error) {
	problems := []*models.Problem{}
	var err error

	if status == "" {
		query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
		       test_cases, created_at
		FROM problems
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`
		err = s.db.SelectContext(ctx, &problems, query, limit, offset)

	} else {
		query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
		       test_cases, created_at
		FROM problems
		WHERE status=$1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
		`
		err = s.db.SelectContext(ctx, &problems, query, status, limit, offset)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).WithError(err).Error("Failed to get latest problems")
		return nil, appError.NewInternal("Failed to get latest problems")
	}
	return problems, nil
}

// Get the latest problem
func (s *Store) GetLatestProblem(ctx context.Context) (*models.Problem, *appError.Error) {
	var p models.Problem
	query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
		       test_cases, created_at
		FROM problems
		ORDER BY created_at DESC
		LIMIT 1
	`
	if err := s.db.GetContext(ctx, &p, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("No problem found")
		}
		logrus.WithError(err).Error("Failed to get latest problem")
		return nil, appError.NewInternal("Failed to get latest problem")
	}
	return &p, nil
}

// Get problem by id
func (s *Store) GetProblemByID(ctx context.Context, id snowflake.ID) (*models.Problem, *appError.Error) {
	var p models.Problem
	query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, saved_code, status, difficulty_level, description, notes,
		       test_cases, created_at
		FROM problems
		WHERE id = $1
	`
	if err := s.db.GetContext(ctx, &p, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.NewNotFound("Prolem not found")
		}
		logrus.WithField("problem_id", id).WithError(err).Error("Failed to get problem by ID")
		return nil, appError.NewInternal("Failed to find the problem")
	}
	return &p, nil
}

// Get the problems after ID
func (s *Store) GetProblemsAfterID(ctx context.Context, afterID int64, size int, status string) ([]*models.Problem, *appError.Error) {
	problems := []*models.Problem{}
	var err error

	if status == "" {
		query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
		       test_cases, created_at
		FROM problems
		WHERE id > $1
		ORDER BY id ASC
		LIMIT $2
	`
		err = s.db.SelectContext(ctx, &problems, query, afterID, size)
	} else {
		query := `
		SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
		       test_cases, created_at
		FROM problems
		WHERE id > $1 AND status=$2
		ORDER BY id ASC
		LIMIT $3
		`
		err = s.db.SelectContext(ctx, &problems, query, afterID, status, size)
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"after_id": afterID,
			"size":     size,
		}).WithError(err).Error("Failed to get problems after ID")
		return nil, appError.NewBadRequest("Failed to problems after ID")
	}
	return problems, nil
}

// Get the problems before the ID
func (s *Store) GetProblemsBeforeID(ctx context.Context, beforeID int64, size int, status string) ([]*models.Problem, *appError.Error) {
	problems := []*models.Problem{}
	var err error
	if status == "" {
		query := `
			SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
				   test_cases, created_at
			FROM problems
			WHERE id < $1
			ORDER BY id DESC
			LIMIT $2
		`
		err = s.db.SelectContext(ctx, &problems, query, beforeID, size)
	} else {
		query := `
			SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
				   test_cases, created_at
			FROM problems
			WHERE id < $1 AND status=$2
			ORDER BY id DESC
			LIMIT $3
		`
		err = s.db.SelectContext(ctx, &problems, query, beforeID, status, size)
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"before_id": beforeID,
			"size":      size,
		}).WithError(err).Error("Failed to get problems before ID")
		return nil, appError.NewInternal("Failed to problems before ID")
	}
	return problems, nil
}

// Fuzzy only search
func (s *Store) SearchProblemsFuzzy(ctx context.Context, query string, limit int) ([]*models.Problem, *appError.Error) {
	var problems []*models.Problem

	// Previously headache for the transacions
	// setQuery := fmt.Sprintf("SET LOCAL pg_trgm.similarity_threshold = %f", threshold)
	// _, err = tx.ExecContext(ctx, setQuery)
	// if err != nil {
	//     return nil, err
	// }

	// Convert similarity threshold to a max distance (1.0 - threshold) so (lowest distance = highest similarity)
	maxDistance := 0.99
	if len(query) <= 3 {
		maxDistance = 0.9999
	}

	// Use <-> for distance. Order by distance ASC (lowest distance = highest similarity)
	searchQuery := `
        SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
            created_at, status
        FROM problems
        WHERE name <-> $1 <= $2
        ORDER BY name <-> $1 ASC
        LIMIT $3
    `

	// No transaction needed! Just run it directly.
	if err := s.db.SelectContext(ctx, &problems, searchQuery, query, maxDistance, limit); err != nil {
		logrus.WithError(err).Error("Failed fuzzy search")
		return nil, appError.NewInternal("Failed to search problems")
	}

	return problems, nil
}

// Fuzzy on name + exact status (bypasses SET LOCAL by using distance operator)
func (s *Store) SearchProblemsFuzzyWithStatus(ctx context.Context, query string, status string, limit int) ([]*models.Problem, *appError.Error) {
	var problems []*models.Problem

	// Convert similarity threshold to a max distance (1.0 - threshold)
	maxDistance := 0.90 // Equivalent to similarity >= 0.1
	if len(query) <= 3 {
		maxDistance = 0.999 // Equivalent to similarity >= 0.001
	}

	// Exact status match + true fuzzy name using distance (<->)
	searchQuery := `
        SELECT id, name, group_name, url, time_limit_ms, memory_limit_mb, status, difficulty_level,
               created_at, status
        FROM problems
        WHERE name <-> $1 <= $2
          AND status = $3
        ORDER BY name <-> $1 ASC
        LIMIT $4
    `

	// Notice the updated parameter order: query ($1), maxDistance ($2), status ($3), limit ($4)
	if err := s.db.SelectContext(ctx, &problems, searchQuery, query, maxDistance, status, limit); err != nil {
		logrus.WithFields(logrus.Fields{
			"query":  query,
			"status": status,
			"limit":  limit,
		}).WithError(err).Error("Failed fuzzy search with status")
		return nil, appError.NewInternal("Failed fuzzy search with status")
	}

	return problems, nil
}
