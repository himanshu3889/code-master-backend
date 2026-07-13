package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/models"
)

// Create a language
func (h *Handler) CreateLanguage(c *gin.Context) {
	var lang models.Language
	if err := c.ShouldBindJSON(&lang); err != nil {
		utils.RespondWithError(c, 400, err.Error())
		return
	}

	if appErr := h.store.CreateLanguage(c.Request.Context(), &lang); appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 201, lang)
}

// Get language by code
func (h *Handler) GetLanguageByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.RespondWithError(c, 400, "Language code required")
		return
	}

	lang, appErr := h.store.GetLanguageByCode(c.Request.Context(), code)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}
	if lang == nil {
		utils.RespondWithError(c, 404, "Language not found")
		return
	}

	utils.RespondWithSuccess(c, 200, lang)
}

// Get all languages
func (h *Handler) GetAllLanguages(c *gin.Context) {
	languages, appErr := h.store.GetAllLanguages(c.Request.Context())
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, languages)
}

// UpdateLanguageTemplate upserts the default code template for a language.
func (h *Handler) UpdateLanguageTemplate(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.RespondWithError(c, 400, "Language code required")
		return
	}

	type UpdateTemplateRequest struct {
		Template string `json:"template" binding:"required"`
	}

	var input UpdateTemplateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid JSON or missing template")
		return
	}

	appErr := h.store.UpsertLanguageTemplate(c.Request.Context(), code, input.Template)
	if appErr != nil {
		utils.RespondWithError(c, int(appErr.Code), appErr.Message)
		return
	}

	utils.RespondWithSuccess(c, 200, gin.H{
		"code":     code,
		"template": input.Template,
	})
}
