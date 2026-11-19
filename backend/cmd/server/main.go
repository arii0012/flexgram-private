
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"nova-messenger/backend/internal/config"
	"nova-messenger/backend/internal/database"
	"nova-messenger/backend/internal/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	db, err := database.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("PostgreSQL error: %v", err)
	}
	defer db.Close()

	redisClient, err := database.OpenRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}
	defer redisClient.Close()

	migrationDir := os.Getenv("MIGRATIONS_DIR")
	if migrationDir == "" {
		migrationDir = "migrations"
	}

	if err := database.RunMigrations(
		ctx,
		db,
		migrationDir,
	); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	router := gin.New()

	router.Use(
		gin.Recovery(),
		handlers.RequestLogger(),
	)

	handlers.RegisterHealthRoutes(
		router,
		db,
		redisClient,
		cfg.AppName,
	)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf(
			"%s listening on port %s",
			cfg.AppName,
			cfg.AppPort,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	shutdownCtx, shutdownCancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
