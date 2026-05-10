package store

import (
	"context"
	"time"

	"github.com/himanshu3889/code-master-backend/base/lib/appError"
	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

// Create the language
func (s *Store) CreateLanguage(ctx context.Context, lang *models.Language) *appError.Error {
	query := `
		INSERT INTO language (id, name, code, extension, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	lang.ID = utils.GenerateSnowflakeID()
	lang.CreatedAt = time.Now()

	if _, err := s.db.ExecContext(ctx, query,
		lang.ID,
		lang.Name,
		lang.Code,
		lang.Extension,
		lang.CreatedAt,
	); err != nil {
		logrus.WithFields(logrus.Fields{
			"name": lang.Name,
			"code": lang.Code,
		}).WithError(err).Error("Failed to create language")
		return appError.NewInternal("Failed to create language")
	}

	return nil
}

// Get the language by codes
func (s *Store) GetLanguageByCode(ctx context.Context, code string) (*models.Language, *appError.Error) {
	var lang models.Language
	query := `
		SELECT id, name, code, extension, created_at 
		FROM language 
		WHERE code = $1
	`
	if err := s.db.GetContext(ctx, &lang, query, code); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, appError.NewNotFound("Language not found")
		}
		logrus.WithField("code", code).WithError(err).Error("Failed to get language by code")
		return nil, appError.NewBadRequest("Failed to get the language")
	}
	return &lang, nil
}

// Get all languages
func (s *Store) GetAllLanguages(ctx context.Context) ([]*models.Language, *appError.Error) {
	var languages []*models.Language
	query := `
		SELECT id, name, code
		FROM language 
		ORDER BY name ASC
	`
	if err := s.db.SelectContext(ctx, &languages, query); err != nil {
		logrus.WithError(err).Error("Failed to get all languages")
		return nil, appError.NewInternal("Failed to get all languages")
	}
	return languages, nil
}
