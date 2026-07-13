package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/himanshu3889/code-master-backend/base/utils"
	"github.com/himanshu3889/code-master-backend/codeRunner"
	"github.com/himanshu3889/code-master-backend/configs"
	handler "github.com/himanshu3889/code-master-backend/internal/apiHandler"
	"github.com/himanshu3889/code-master-backend/internal/database"
	"github.com/himanshu3889/code-master-backend/internal/jobs"
	"github.com/himanshu3889/code-master-backend/internal/store"
	appWebsocket "github.com/himanshu3889/code-master-backend/internal/websocket"
)

func RunServer() {
	configs.InitializeConfigs()
	utils.InitSnowflake(1)
	database.InitPostgresDB()

	// Simple dependency injection
	appWebsocket.InitializeWebsocketStore(database.PostgresDB)

	store := store.New(database.PostgresDB)
	h := handler.New(store)

	// Start background submission retry worker (every 5 min, min age 5 min, batch 20)
	stopRetryWorker := jobs.StartSubmissionRetryWorker(store, 5*time.Minute, 5*time.Minute, 20)

	// Start background session timeout worker (every 30s)
	stopSessionTimeoutWorker := jobs.StartSessionTimeoutWorker(store, 30*time.Second)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	handler.SetupRoutes(r, h)

	// Graceful shutdown
	srv := &http.Server{Addr: ":27122", Handler: r}

	go func() {
		log.Println("Server on :27122")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	stopRetryWorker()                // Stop background retry worker gracefully
	stopSessionTimeoutWorker()       // Stop session timeout worker gracefully
	codeRunner.ShutdownCodeRunners() // Shutdown
}
