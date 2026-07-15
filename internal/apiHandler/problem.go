package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	"github.com/himanshu3889/code-master-backend/base/lib/pagination"
	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/lib/tagging"
	"github.com/himanshu3889/code-master-backend/internal/models"
	appWebsocket "github.com/himanshu3889/code-master-backend/internal/websocket"
	"github.com/sirupsen/logrus"
)

// Receive problem from the competitive companion
func (h *Handler) ReceiveProblem(c *gin.Context) {

	var problem models.Problem

	// Direct binding - much cleaner!
	if err := c.ShouldBindJSON(&problem); err != nil {
		utils.RespondWithError(c, 400, err.Error())
		return
	}

	// Read the raw bytes from the request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.RespondWithError(c, 400, "failed to read request body")
		return
	}

	problem.RawPayload = string(bodyBytes)
	problem.Status = "TODO"
	if !problem.Tags.Valid || len(problem.Tags.Data) == 0 {
		problem.Tags = sqlLib.NewJSONB(tagging.DetectTags(problem.Name, problem.Description))
	}

	if appErr := h.store.CreateProblemWithCompanion(c.Request.Context(), &problem); appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	go func() {
		appWebsocket.HandleNewProblem(&problem)
	}()

	utils.RespondWithSuccess(c, 200, gin.H{"id": problem.ID, "name": problem.Name})
}

// Update the problem status
func (h *Handler) UpdateProblemStatus(c *gin.Context) {
	idString := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	type UpdateStatusRequest struct {
		Status string `json:"status" binding:"required"`
	}

	var input UpdateStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing status")
		return
	}

	status := input.Status

	appErr := h.store.UpdateProblemStatus(c.Request.Context(), problemId, status)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	utils.RespondWithSuccess(c, 201, gin.H{"message": "Problem status updated"})
}

// Update the problem difficulty level
func (h *Handler) UpdateProblemDifficultyLevel(c *gin.Context) {
	idString := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	type UpdateDifficultyLevelRequest struct {
		DifficultyLevel string `json:"difficulty_level" binding:"required"`
	}

	var input UpdateDifficultyLevelRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing difficulty level")
		return
	}

	difficulty_level := input.DifficultyLevel

	appErr := h.store.UpdateProblemDifficultyLevel(c.Request.Context(), problemId, difficulty_level)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	utils.RespondWithSuccess(c, 201, gin.H{"message": "Problem difficulty level updated"})
}

// Update problem description
func (h *Handler) UpdateProblemDescription(c *gin.Context) {
	idString := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	type UpdateDescriptionRequest struct {
		Description string `json:"description" binding:"required"`
	}

	var input UpdateDescriptionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing description")
		return
	}

	description := input.Description

	appErr := h.store.UpdateProblemDescription(c.Request.Context(), problemId, description)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	utils.RespondWithSuccess(c, 201, gin.H{"message": "Problem description updated"})
}

// Update problem note
func (h *Handler) UpdateProblemNotes(c *gin.Context) {
	idString := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	type UpdateNotesRequest struct {
		Notes string `json:"notes" binding:"required"`
	}

	var input UpdateNotesRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing notes")
		return
	}

	notes := input.Notes

	appErr := h.store.UpdateProblemNotes(c.Request.Context(), problemId, notes)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	utils.RespondWithSuccess(c, 201, gin.H{"message": "Problem note updated"})
}

// Get the the latest problems with page-based pagination
func (h *Handler) GetLatestProblems(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	status := c.DefaultQuery("status", "")

	// Calculate offset from page
	offset := (page - 1) * limit

	// Get total count for pagination metadata
	totalCount, appErr := h.store.CountProblems(c.Request.Context(), status)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	problems, appErr := h.store.GetLatestProblems(c.Request.Context(), limit, offset, status)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	// Compute pagination metadata
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	var nextPage *int
	if hasNext {
		np := page + 1
		nextPage = &np
	}

	var prevPage *int
	if hasPrev {
		pp := page - 1
		prevPage = &pp
	}

	response := pagination.PaginatedResponse[[]*models.Problem]{
		Data:       problems,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
		NextPage:   nextPage,
		PrevPage:   prevPage,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	utils.RespondWithSuccess(c, 200, response)
}

// Get the the problem
func (h *Handler) GetLatestProblem(c *gin.Context) {
	problem, appErr := h.store.GetLatestProblem(c.Request.Context())
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	utils.RespondWithSuccess(c, 200, problem)
}

// Get problem with the submissions
func (h *Handler) GetProblem(c *gin.Context) {
	idString := c.Param("problemId")
	id, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	problem, appErr := h.store.GetProblemByID(c.Request.Context(), id)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, problem)
}

// Get the problems after ID
func (h *Handler) GetProblemsAfterID(c *gin.Context) {
	afterID, err := strconv.ParseInt(c.Param("afterId"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid after ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 50
	}

	status := c.DefaultQuery("status", "")

	problems, appErr := h.store.GetProblemsAfterID(c.Request.Context(), afterID, limit, status)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, problems)
}

// Get the problems before the ID
func (h *Handler) GetProblemsBeforeID(c *gin.Context) {
	beforeID, err := strconv.ParseInt(c.Param("beforeId"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid before ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 50
	}

	status := c.DefaultQuery("status", "")

	problems, appErr := h.store.GetProblemsBeforeID(c.Request.Context(), beforeID, limit, status)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, problems)
}

// Fuzzy search of problem
func (h *Handler) SearchProblemsFuzzy(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.RespondWithError(c, 400, "Search query required")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 50 {
		limit = 50
	}

	status := c.DefaultQuery("status", "")

	var problems []*models.Problem

	var appErr *appError.Error
	if status != "" {
		logrus.Info("With status is running...")
		problems, appErr = h.store.SearchProblemsFuzzyWithStatus(c.Request.Context(), query, status, limit)
	} else {
		problems, appErr = h.store.SearchProblemsFuzzy(c.Request.Context(), query, limit)
	}

	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	response := gin.H{
		"query":   query,
		"count":   len(problems),
		"results": problems,
	}

	if status != "" {
		response["status_filter"] = status
	}

	utils.RespondWithSuccess(c, 200, response)
}

// UpdateProblemTags updates the tags of a problem
func (h *Handler) UpdateProblemTags(c *gin.Context) {
	idString := c.Param("problemId")
	problemId, err := utils.ValidSnowflakeID(idString)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid ID")
		return
	}

	type UpdateTagsRequest struct {
		Tags []string `json:"tags" binding:"required"`
	}

	var input UpdateTagsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing tags")
		return
	}

	// Deduplicate tags
	seen := make(map[string]struct{})
	uniqueTags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		if _, ok := seen[tag]; !ok {
			seen[tag] = struct{}{}
			uniqueTags = append(uniqueTags, tag)
		}
	}

	appErr := h.store.UpdateProblemTags(c.Request.Context(), problemId, uniqueTags)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 201, gin.H{"message": "Problem tags updated"})
}

// GetAllTags returns the list of all available pattern tags
func (h *Handler) GetAllTags(c *gin.Context) {
	tags := tagging.GetAllTags()
	utils.RespondWithSuccess(c, 200, gin.H{"tags": tags})
}
