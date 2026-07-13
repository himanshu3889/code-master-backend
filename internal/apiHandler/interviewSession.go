package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/lib"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

func computeRemainingSeconds(session *models.InterviewSession) int {
	if session.Status != models.InterviewSessionInProgress {
		return 0
	}
	elapsed := time.Since(session.StartedAt)
	remaining := session.TimeLimitSeconds - int(elapsed.Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (h *Handler) CreateInterviewSession(c *gin.Context) {
	var req struct {
		ProblemID        snowflake.ID `json:"problemId"`
		TimeLimitSeconds int          `json:"timeLimitSeconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ProblemID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "problemId is required"})
		return
	}
	if req.TimeLimitSeconds <= 0 {
		req.TimeLimitSeconds = 2700 // Default 45 minutes
	}

	// Return existing active session if one exists
	existing, appErr := h.store.GetActiveSessionByProblem(c.Request.Context(), req.ProblemID)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if existing != nil {
		utils.RespondWithSuccess(c, http.StatusOK, gin.H{
			"session":          existing,
			"remainingSeconds": computeRemainingSeconds(existing),
		})
		return
	}

	session := &models.InterviewSession{
		ProblemID:        req.ProblemID,
		TimeLimitSeconds: req.TimeLimitSeconds,
	}

	if appErr := h.store.CreateInterviewSession(c.Request.Context(), session); appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, gin.H{
		"session":          session,
		"remainingSeconds": computeRemainingSeconds(session),
	})
}

func (h *Handler) GetInterviewSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ValidSnowflakeID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	session, appErr := h.store.GetInterviewSessionByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{
		"session":          session,
		"remainingSeconds": computeRemainingSeconds(session),
	})
}

func (h *Handler) GetActiveInterviewSessionForProblem(c *gin.Context) {
	problemIdStr := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(problemIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}
	session, appErr := h.store.GetActiveSessionByProblem(c.Request.Context(), problemId)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if session == nil {
		utils.RespondWithSuccess(c, http.StatusOK, gin.H{"session": nil})
		return
	}
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{
		"session":          session,
		"remainingSeconds": computeRemainingSeconds(session),
	})
}

func (h *Handler) GetProblemInterviewSessions(c *gin.Context) {
	problemIdStr := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(problemIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}
	sessions, appErr := h.store.GetInterviewSessionsByProblem(c.Request.Context(), problemId)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"sessions": sessions})
}

func (h *Handler) CompleteInterviewSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ValidSnowflakeID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	session, appErr := h.store.GetInterviewSessionByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if session.Status != models.InterviewSessionInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not in progress"})
		return
	}
	if appErr := h.store.UpdateInterviewSessionStatus(c.Request.Context(), id, models.InterviewSessionCompleted); appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Session completed"})
}

func (h *Handler) AbandonInterviewSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ValidSnowflakeID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	session, appErr := h.store.GetInterviewSessionByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if session.Status != models.InterviewSessionInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not in progress"})
		return
	}
	if appErr := h.store.UpdateInterviewSessionStatus(c.Request.Context(), id, models.InterviewSessionAbandoned); appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	utils.RespondWithSuccess(c, http.StatusOK, gin.H{"message": "Session abandoned"})
}

func (h *Handler) TimeoutInterviewSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ValidSnowflakeID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	session, appErr := h.store.GetInterviewSessionByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if session.Status != models.InterviewSessionInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not in progress"})
		return
	}
	// Validate that the session is actually expired (server-side truth)
	elapsed := time.Since(session.StartedAt)
	if elapsed.Seconds() < float64(session.TimeLimitSeconds) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session has not yet expired"})
		return
	}

	// Mark as TIMEOUT and auto_submitted
	if appErr := h.store.UpdateInterviewSessionStatus(c.Request.Context(), id, models.InterviewSessionTimeout); appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}
	if appErr := h.store.SetSessionAutoSubmitted(c.Request.Context(), id); appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}

	// Fetch latest code snapshot for this session
	snapshots, appErr := h.store.GetCodeSnapshotsByProblem(c.Request.Context(), int64(session.ProblemID), &session.ID, 1)
	if appErr != nil {
		logrus.WithField("error", appErr.Message).Warn("No code snapshot found for auto-submit; session timed out without code")
		utils.RespondWithSuccess(c, http.StatusAccepted, gin.H{"message": "Session timed out", "autoSubmitted": false})
		return
	}
	if len(snapshots) == 0 {
		utils.RespondWithSuccess(c, http.StatusAccepted, gin.H{"message": "Session timed out", "autoSubmitted": false})
		return
	}
	snapshot := snapshots[0]

	// Fetch problem to get test cases
	problem, appErr := h.store.GetProblemByID(c.Request.Context(), session.ProblemID)
	if appErr != nil {
		logrus.WithField("error", appErr.Message).Warn("Failed to fetch problem for auto-submit")
		utils.RespondWithSuccess(c, http.StatusAccepted, gin.H{"message": "Session timed out", "autoSubmitted": false})
		return
	}

	// Build test case inputs
	var testInputs []string
	for _, tc := range problem.TestCases.Data {
		testInputs = append(testInputs, tc.Input)
	}

	// Build and save submission
	submission := &models.Submission{
		ID:                 utils.GenerateSnowflakeID(),
		ProblemID:          session.ProblemID,
		InterviewSessionId: &session.ID,
		Language:           snapshot.Language,
		Stdin:              snapshot.Code,
		TestCases:          sqlLib.NewJSONB(testInputs),
		CreatedAt:          time.Now(),
	}

	// Save submission (this creates timeline entry too)
	if appErr := h.store.CreateSubmissionLog(c.Request.Context(), submission); appErr != nil {
		logrus.WithField("error", appErr.Message).Error("Failed to create auto-submit submission log")
		utils.RespondWithSuccess(c, http.StatusAccepted, gin.H{"message": "Session timed out", "autoSubmitted": false})
		return
	}

	// Execute asynchronously
	go func() {
		result := lib.ExecuteSubmissionCode(h.store, submission)
		if result == nil {
			logrus.Warn("Auto-submit code execution returned nil")
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
		if updateErr := h.store.UpdateSubmissionExecutionResults(context.Background(), submission); updateErr != nil {
			logrus.WithField("error", updateErr.Message).Error("Failed to update auto-submit execution results")
		}
	}()

	utils.RespondWithSuccess(c, http.StatusAccepted, gin.H{"message": "Session timed out", "autoSubmitted": true, "submissionId": submission.ID})
}
