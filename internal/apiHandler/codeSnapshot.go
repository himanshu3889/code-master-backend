package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/base/utils"
)

// Get the problem code snapshots
func (h *Handler) GetProblemCodeSnapshots(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("problemId"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid problem ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 20
	}

	snapshots, appErr := h.store.GetCodeSnapshotsByProblem(c.Request.Context(), id, limit)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, snapshots)
}

// Get the code snapshot
func (h *Handler) GetCodeSnapshotByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid snapshot ID")
		return
	}

	snapshot, appErr := h.store.GetCodeSnapshotByID(c.Request.Context(), id)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, snapshot)
}
