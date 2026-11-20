
package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"nova-messenger/backend/internal/config"
	"nova-messenger/backend/internal/database"
	"nova-messenger/backend/internal/handlers"
	"nova-messenger/backend/internal/logger"
)

func main() {
	cfg := config.Load()

	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}

	logger.Init()

	db := database.ConnectPostgres(cfg.DatabaseURL)
	defer db.Close()

	redisClient := database.ConnectRedis(cfg.RedisURL)
	defer redisClient.Close()

	router := gin.Default()

	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)

	healthHandler := handlers.NewHealthHandler(db, redisClient)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "NovaMessenger",
			"status":  "online",
			"version": "1.0.0",
		})
	})

	router.GET("/api/health", healthHandler.Check)

	api := router.Group("/api/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
	}

	log.Printf("NovaMessenger running on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
