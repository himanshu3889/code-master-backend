package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
	"github.com/himanshu3889/code-master-backend/internal/lib"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/himanshu3889/code-master-backend/internal/store"
	"github.com/sirupsen/logrus"
)

// StartSubmissionRetryWorker starts a background worker that re-executes stale PENDING submissions.
// It returns a stop function that must be called during graceful shutdown; the stop function blocks
// until the current batch finishes.
func StartSubmissionRetryWorker(s *store.Store, interval, minAge time.Duration, batchSize int) func() {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		// Run immediately on boot, then wait for ticker
		processStaleSubmissions(ctx, s, minAge, batchSize)

		for {
			select {
			case <-ticker.C:
				processStaleSubmissions(ctx, s, minAge, batchSize)
			case <-ctx.Done():
				logrus.Info("Submission retry worker shutting down")
				return
			}
		}
	}()

	return func() {
		cancel()
		wg.Wait()
	}
}

func processStaleSubmissions(ctx context.Context, s *store.Store, minAge time.Duration, batchSize int) {
	submissions, appErr := s.GetStalePendingSubmissions(ctx, minAge, batchSize)
	if appErr != nil {
		logrus.WithField("error", appErr.Message).Error("Failed to fetch stale pending submissions")
		return
	}

	if len(submissions) == 0 {
		return
	}

	logrus.WithField("count", len(submissions)).Info("Processing stale pending submissions")

	for _, submission := range submissions {
		select {
		case <-ctx.Done():
			logrus.Info("Submission retry worker cancelled mid-batch")
			return
		default:
		}

		logrus.WithFields(logrus.Fields{
			"submission_id": submission.ID,
			"problem_id":    submission.ProblemID,
			"language":      submission.Language,
		}).Info("Retrying submission execution")

		result := lib.ExecuteSubmissionCode(s, submission)

		if result != nil {
			// Map execution result to submission model (same as API handler)
			status := models.ExecutionStatus(result.Status)
			submission.Status = &status
			submission.ExecutionTimeMs = &result.TimeMs
			memoryKB := result.MemoryBytes / 1000
			submission.MemoryUsedKb = &memoryKB
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
		} else {
			// Execution failed permanently — mark as RE so it's not retried again
			reStatus := models.StatusRuntimeError
			submission.Status = &reStatus
			errMsg := fmt.Sprintf("Background retry failed at %s", time.Now().Format(time.RFC3339))
			submission.Stderr = &errMsg
			submission.ExecutionTimeMs = nil
			submission.MemoryUsedKb = nil
			submission.Stdout = nil
			submission.TestResults = sqlLib.NewJSONB([]*models.TestResult{})

			logrus.WithFields(logrus.Fields{
				"submission_id": submission.ID,
				"problem_id":    submission.ProblemID,
			}).Warn("Submission retry execution returned nil, marking as RE")
		}

		if updateErr := s.UpdateSubmissionExecutionResults(ctx, submission); updateErr != nil {
			logrus.WithFields(logrus.Fields{
				"submission_id": submission.ID,
				"problem_id":    submission.ProblemID,
				"error":         updateErr.Message,
			}).Error("Failed to update submission execution results after retry")
		}
	}
}
