package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/lib"
	"github.com/himanshu3889/code-master-backend/internal/models"
	appWebsocket "github.com/himanshu3889/code-master-backend/internal/websocket"
	"github.com/sirupsen/logrus"
)

// Log the submission execution
func (h *Handler) SubmitSubmission(c *gin.Context) {
	var submission models.Submission

	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a channel to signal when the DB insert is finished.
	dbReady := make(chan bool, 1)

	// Save in db without it's result
	submission.ID = utils.GenerateSnowflakeID()
	submission.CreatedAt = time.Now()
	go func() {
		err := h.store.CreateSubmissionLog(context.Background(), &submission)
		if err != nil {
			logrus.Error("Unable to log the submission to the database")
			dbReady <- false // Signal failure
			return
		}
		dbReady <- true
	}()

	// submit to the code executor
	// After completed execution update the submission executions
	go func() {
		// First save submission in db then update the submission execution
		submissionExecutionResults := lib.ExecuteSubmissionCode(h.store, &submission)

		if submissionExecutionResults == nil {
			logrus.Warn("Code execution failed or returned nil")
			return
		}

		// send async the submission to the websocket if executed properly
		go func() {
			if submissionExecutionResults != nil {
				appWebsocket.HandleCodeSubmissionResults(&submission)
			}
		}()

		// WAIT for the DB insert to finish to update the submission result
		isReady := <-dbReady
		if !isReady {
			return
		}

		// Update the status
		status := models.ExecutionStatus(submissionExecutionResults.Status)
		submission.Status = &status
		// Update the time
		submission.ExecutionTimeMs = &submissionExecutionResults.TimeMs
		// update the memory
		memoryKB := submissionExecutionResults.MemoryBytes / 1000
		submission.MemoryUsedKb = &memoryKB
		// update the stdout
		submission.Stdout = &submissionExecutionResults.Stdout
		// Update the stderr
		submission.Stderr = &submissionExecutionResults.Stderr
		// Update the testcase results
		testCaseResult := make([]*models.TestResult, 0, len(submissionExecutionResults.TestCasesResult))
		for _, tcResult := range submissionExecutionResults.TestCasesResult {
			status := models.ExecutionStatus(tcResult.Status)
			testCaseResult[tcResult.TestIndex] = &models.TestResult{
				TestIndex:   tcResult.TestIndex,
				Passed:      tcResult.Passed,
				Status:      status,
				TimeMs:      tcResult.TimeMs,
				MemoryBytes: tcResult.MemoryBytes,
				Stderr:      &tcResult.Stderr,
			}
		}

		// update the db
		// what if creation failed but not the update submission result ?
		err := h.store.UpdateSubmissionExecutionResults(context.Background(), &submission)
		if err != nil {
			logrus.Errorf("Submission not %d update successfully", submission.ID)
		}
	}()

	utils.RespondWithSuccess(c, http.StatusAccepted, submission)
}

// Get problem with the submissions
func (h *Handler) GetProblemWithSessionSubmissions(c *gin.Context) {
	idStr := c.Param("problemId")
	id, err := utils.ValidSnowflakeID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	problem, appErr := h.store.GetProblemByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	submissionIdStr := c.DefaultQuery("interview_session", "")
	var submissionId *snowflake.ID
	if submissionIdStr != "" {
		id, err := utils.ValidSnowflakeID(submissionIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		submissionId = &id
	}

	submissions, appErr := h.store.GetProblemSubmissions(c.Request.Context(), id, submissionId)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problem":     problem,
		"submissions": submissions,
	})
}

// Get problem submissions
func (h *Handler) GetProblemSessionSubmissions(c *gin.Context) {
	problemIdStr := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(problemIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	submissionIdStr := c.DefaultQuery("interview_session", "")
	var submissionId *snowflake.ID
	if submissionIdStr != "" {
		id, err := utils.ValidSnowflakeID(submissionIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		submissionId = &id
	}

	submissions, appErr := h.store.GetProblemSubmissions(c.Request.Context(), problemId, submissionId)
	if appErr != nil {
		c.JSON(int(appErr.Code), gin.H{"error": appErr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": submissions,
	})
}
