package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/internal/store"
	appWebsocket "github.com/himanshu3889/code-master-backend/internal/websocket"
)

type Handler struct {
	store *store.Store
}

func New(store *store.Store) *Handler {
	return &Handler{store: store}
}

func SetupRoutes(r *gin.Engine, h *Handler) {
	// Competitive Companion endpoint (no /api prefix)
	r.POST("/", h.ReceiveProblem)

	// Websocket handle
	r.GET("/ws", appWebsocket.WsHandler)

	api := r.Group("/api")
	{
		// Problems
		api.GET("/problems", h.GetLatestProblems)
		api.GET("/problems/:problemId", h.GetProblem)
		api.GET("problems/latest", h.GetLatestProblem)
		api.GET("/problems/after/:afterId", h.GetProblemsAfterID)
		api.GET("/problems/before/:beforeId", h.GetProblemsBeforeID)
		api.GET("/problems/search", h.SearchProblemsFuzzy)
		api.PATCH("/problems/search", h.SearchProblemsFuzzy)
		api.PATCH("/problems/:problemId/status", h.UpdateProblemStatus)
		api.PATCH("/problems/:problemId/difficulty", h.UpdateProblemDifficultyLevel)
		api.PATCH("/problems/:problemId/notes", h.UpdateProblemNotes)
		api.PATCH("/problems/:problemId/description", h.UpdateProblemDescription)

		// Languages
		api.POST("/languages", h.CreateLanguage)
		api.GET("/languages", h.GetAllLanguages)
		api.GET("/languages/:code", h.GetLanguageByCode)
		api.PATCH("/languages/:code/template", h.UpdateLanguageTemplate)

		// Code submission
		api.POST("/problems/:problemId/submissions", h.SubmitSubmission)
		api.GET("/problems/:problemId/with-submissions", h.GetProblemWithSessionSubmissions)
		api.GET("/problems/:problemId/submissions", h.GetProblemSessionSubmissions)

		// Code Snapshots
		api.GET("/snapshots/:snapshotId", h.GetCodeSnapshotByID)
		api.GET("/problems/:problemId/snapshots", h.GetProblemCodeSnapshots)

		// Timeline (The Story)
		api.GET("/problems/:problemId/timeline", h.GetProblemTimeline)
		api.GET("/problems/:problemId/story", h.GetDetailedTimelineByProblem)

		// Interview Sessions
		api.POST("/interview-sessions", h.CreateInterviewSession)
		api.GET("/interview-sessions/:id", h.GetInterviewSession)
		api.GET("/problems/:problemId/interview-sessions", h.GetProblemInterviewSessions)
		api.PATCH("/interview-sessions/:id/complete", h.CompleteInterviewSession)
		api.PATCH("/interview-sessions/:id/abandon", h.AbandonInterviewSession)
		api.POST("/interview-sessions/:id/timeout", h.TimeoutInterviewSession)
		api.GET("/problems/:problemId/active-session", h.GetActiveInterviewSessionForProblem)
	}

}
