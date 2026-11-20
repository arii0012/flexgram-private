package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"flexgram-private/backend/internal/config"
	"flexgram-private/backend/internal/database"
	"flexgram-private/backend/internal/handlers"
	"flexgram-private/backend/internal/logger"
	"flexgram-private/backend/internal/middleware"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	logger.Init()

	db := database.ConnectPostgres(cfg.DatabaseURL)
	defer db.Close()

	redisClient := database.ConnectRedis(cfg.RedisURL)
	defer redisClient.Close()

	router := gin.Default()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	profileHandler := handlers.NewProfileHandler(db)
	sessionHandler := handlers.NewSessionHandler(redisClient)
	statusHandler := handlers.NewStatusHandler(db)
	chatHandler := handlers.NewChatHandler(db)
	messageHandler := handlers.NewMessageHandler(db)
	receiptHandler := handlers.NewReceiptHandler(db)

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "flexgram-private",
			"status":  "online",
			"version": "1.0.0",
		})
	})

	router.GET("/api/health", healthHandler.Check)

	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret, redisClient))

	{
		// Profile
		protected.GET("/users/me", profileHandler.Me)
		protected.PUT("/users/me", profileHandler.Update)
		protected.GET("/users/search", profileHandler.Search)

		// User status
		protected.POST("/users/online", statusHandler.SetOnline)
		protected.POST("/users/offline", statusHandler.SetOffline)

		// Sessions
		protected.POST("/auth/logout", sessionHandler.Logout)

		// Chats
		protected.POST("/chats", chatHandler.Create)
		protected.GET("/chats", chatHandler.List)

		// Messages
		protected.POST("/chats/:chatID/messages", messageHandler.Send)
		protected.GET("/chats/:chatID/messages", messageHandler.History)
		protected.PUT("/messages/:messageID", messageHandler.Edit)
		protected.DELETE("/messages/:messageID", messageHandler.Delete)

		// Message receipts
		protected.POST("/chats/:chatID/delivered", receiptHandler.MarkDelivered)
		protected.POST("/chats/:chatID/read", receiptHandler.MarkRead)
	}

	log.Printf("flexgram-private started on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
