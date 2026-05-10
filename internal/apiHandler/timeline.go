package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/base/utils"
)

// Get Problem timeline
func (h *Handler) GetProblemTimeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("problemId"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid problem ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 50
	}

	entries, appErr := h.store.GetTimelineByProblem(c.Request.Context(), id, limit)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, entries)
}

// Get problem timeline in detailed, the main story endpoint
func (h *Handler) GetDetailedTimelineByProblem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("problemId"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid problem ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 50
	}

	story, appErr := h.store.GetDetailedTimelineByProblem(c.Request.Context(), id, limit)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, story)
}
