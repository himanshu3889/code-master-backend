package main

import (
	"context"
	"time"

	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/configs"
	"github.com/himanshu3889/code-master-backend/internal/database"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/himanshu3889/code-master-backend/internal/store"
	"github.com/sirupsen/logrus"
)

func AddLanguages() {
	configs.InitializeConfigs()
	utils.InitSnowflake(1)
	database.InitPostgresDB()

	store := store.New(database.PostgresDB)
	languages := []models.Language{
		{
			ID:        utils.GenerateSnowflakeID(),
			Name:      "Python",
			Code:      "python",
			Extension: "py",
			CreatedAt: time.Now(),
		},
		{
			ID:        utils.GenerateSnowflakeID(),
			Name:      "Golang",
			Code:      "go",
			Extension: "go",
			CreatedAt: time.Now(),
		},
	}

	ctx := context.Background()
	for _, lang := range languages {
		if appErr := store.CreateLanguage(ctx, &lang); appErr != nil {
			logrus.Error(appErr.Message)
			return
		}
	}
}

func main() {
	AddLanguages()
}
