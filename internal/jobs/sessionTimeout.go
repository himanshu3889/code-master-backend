package jobs

import (
	"context"
	"sync"
	"time"

	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/lib"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/himanshu3889/code-master-backend/internal/store"
	"github.com/sirupsen/logrus"
)

// StartSessionTimeoutWorker starts a background worker that checks for expired interview sessions
// and auto-submits the latest code snapshot. It returns a stop function that must be called
// during graceful shutdown; the stop function blocks until the current batch finishes.
func StartSessionTimeoutWorker(s *store.Store, interval time.Duration) func() {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		// Run immediately on boot, then wait for ticker
		processExpiredSessions(ctx, s)

		for {
			select {
			case <-ticker.C:
				processExpiredSessions(ctx, s)
			case <-ctx.Done():
				logrus.Info("Session timeout worker shutting down")
				return
			}
		}
	}()

	return func() {
		cancel()
		wg.Wait()
	}
}

func processExpiredSessions(ctx context.Context, s *store.Store) {
	sessions, appErr := s.GetExpiredSessions(ctx)
	if appErr != nil {
		logrus.WithField("error", appErr.Message).Error("Failed to fetch expired interview sessions")
		return
	}

	if len(sessions) == 0 {
		return
	}

	logrus.WithField("count", len(sessions)).Info("Processing expired interview sessions")

	for _, session := range sessions {
		select {
		case <-ctx.Done():
			logrus.Info("Session timeout worker cancelled mid-batch")
			return
		default:
		}

		logrus.WithFields(logrus.Fields{
			"session_id": session.ID,
			"problem_id": session.ProblemID,
		}).Info("Auto-submitting expired session")

		processSessionTimeout(ctx, s, session)
	}
}

func processSessionTimeout(ctx context.Context, s *store.Store, session *models.InterviewSession) {
	// Mark session as TIMEOUT
	if err := s.UpdateInterviewSessionStatus(ctx, session.ID, models.InterviewSessionTimeout); err != nil {
		logrus.WithField("error", err.Message).Error("Failed to update expired session status")
		return
	}

	// Re-fetch to check if another process already handled auto-submit
	freshSession, err := s.GetInterviewSessionByID(ctx, session.ID)
	if err != nil {
		logrus.WithField("error", err.Message).Error("Failed to re-fetch session after timeout update")
		return
	}
	if freshSession.AutoSubmitted {
		logrus.WithField("session_id", session.ID).Info("Session already auto-submitted by another process; skipping duplicate")
		return
	}

	// Set auto_submitted flag
	if err := s.SetSessionAutoSubmitted(ctx, session.ID); err != nil {
		logrus.WithField("error", err.Message).Error("Failed to set auto_submitted flag")
		return
	}

	// Fetch latest code snapshot for this session
	snapshots, err := s.GetCodeSnapshotsByProblem(ctx, int64(session.ProblemID), &session.ID, 1)
	if err != nil {
		logrus.WithField("error", err.Message).Warn("Failed to fetch snapshots for auto-submit")
		return
	}
	if len(snapshots) == 0 {
		logrus.Warn("No code snapshots available for auto-submit")
		return
	}
	snapshot := snapshots[0]

	// Fetch problem to get test cases
	problem, err := s.GetProblemByID(ctx, session.ProblemID)
	if err != nil {
		logrus.WithField("error", err.Message).Warn("Failed to fetch problem for auto-submit")
		return
	}

	// Build test case inputs
	var testInputs []string
	for _, tc := range problem.TestCases.Data {
		testInputs = append(testInputs, tc.Input)
	}

	// Build submission
	submission := &models.Submission{
		ID:                 utils.GenerateSnowflakeID(),
		ProblemID:          session.ProblemID,
		InterviewSessionId: &session.ID,
		Language:           snapshot.Language,
		Stdin:              snapshot.Code,
		TestCases:          sqlLib.NewJSONB(testInputs),
		CreatedAt:          time.Now(),
	}

	// Save submission (creates timeline entry too)
	if err := s.CreateSubmissionLog(ctx, submission); err != nil {
		logrus.WithField("error", err.Message).Error("Failed to create auto-submit submission log")
		return
	}

	// Execute
	result := lib.ExecuteSubmissionCode(s, submission)
	if result == nil {
		logrus.Warn("Auto-submit execution returned nil")
		reStatus := models.StatusRuntimeError
		submission.Status = &reStatus
		errMsg := "Auto-submit execution failed"
		submission.Stderr = &errMsg
		submission.TestResults = sqlLib.NewJSONB([]*models.TestResult{})
	} else {
		status := models.ExecutionStatus(result.Status)
		submission.Status = &status
		submission.ExecutionTimeMs = &result.TimeMs
		memKB := result.MemoryBytes / 1000
		submission.MemoryUsedKb = &memKB
		submission.Stdout = &result.Stdout
		submission.Stderr = &result.Stderr

		testCaseResult := make([]*models.TestResult, len(result.TestCasesResult))
		for _, tcResult := range result.TestCasesResult {
			status := models.ExecutionStatus(tcResult.Status)
			testCaseResult[tcResult.TestIndex] = &models.TestResult{
				TestIndex:   tcResult.TestIndex,
				Passed:      tcResult.Passed,
				Status:      status,
				TimeMs:      tcResult.TimeMs,
				MemoryBytes: tcResult.MemoryBytes,
				Stderr:      &tcResult.Stderr,
				Stdout:      &tcResult.ActualOutput,
			}
		}
		submission.TestResults = sqlLib.NewJSONB(testCaseResult)
	}

	// Update results
	if updateErr := s.UpdateSubmissionExecutionResults(ctx, submission); updateErr != nil {
		logrus.WithField("error", updateErr.Message).Error("Failed to update auto-submit execution results")
	}
}
