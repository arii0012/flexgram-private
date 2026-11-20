package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/config"
	"flexgram-private/backend/internal/handlers"
	"flexgram-private/backend/internal/middleware"
	ws "flexgram-private/backend/internal/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	cfg := config.Load()

	if len(cfg.JWTSecret) < 16 {
		log.Fatal("JWT_SECRET must be at least 16 characters")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Cannot connect to PostgreSQL:", err)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool); err != nil {
		log.Println("Migration warning:", err)
	}

	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authH := handlers.NewAuthHandler(pool, cfg.JWTSecret)
	chatH := handlers.NewChatHandler(pool, hub)
	msgH := handlers.NewMessageHandler(pool, hub)
	giftsH := handlers.NewGiftsHandler(pool)
	godH := handlers.NewGodHandler(pool)
	profileH := handlers.NewProfileHandler(pool)
	extraH := handlers.NewExtraHandler(pool, hub)

	// Web UI at /app/
	for _, webDir := range []string{"/app/web", "./web", "../web"} {
		if st, err := os.Stat(webDir); err == nil && st.IsDir() {
			r.Static("/app", webDir)
			log.Println("Web UI: /app/ from", webDir)
			break
		}
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"name": "FlexGram", "status": "online", "version": "1.0.0"})
	})

	api := r.Group("/api")
	{
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/users/me", profileH.Me)
			protected.PUT("/users/me", profileH.Update)
			protected.GET("/users/search", profileH.Search)

			protected.POST("/chats", chatH.Create)
			protected.GET("/chats", chatH.List)

			protected.POST("/chats/:chatID/messages", msgH.Send)
			protected.GET("/chats/:chatID/messages", msgH.History)
			protected.PUT("/messages/:messageID", msgH.Edit)
			protected.DELETE("/messages/:messageID", msgH.Delete)

			protected.GET("/gifts/shop", giftsH.ListShop)
			protected.GET("/gifts/my", giftsH.MyGifts)
			protected.POST("/gifts/buy", giftsH.Buy)
			protected.POST("/gifts/transfer", giftsH.Transfer)
			protected.POST("/gifts/:id/equip", giftsH.ToggleEquip)

			protected.POST("/god/stars", godH.AddStars)
			protected.POST("/god/premium", godH.SetPremium)
			protected.POST("/god/gift", godH.GrantGift)
			protected.POST("/god/username", godH.ChangeUsername)

			// Telegram-like extras
			protected.POST("/chats/:chatID/typing", extraH.Typing)
			protected.POST("/chats/:chatID/read", extraH.MarkRead)
			protected.POST("/chats/:chatID/pin", extraH.PinMessage)
			protected.GET("/chats/:chatID/pinned", extraH.GetPinned)
			protected.POST("/messages/:messageID/react", extraH.React)
			protected.POST("/users/online", extraH.SetOnline)
		}
	}

	r.GET("/ws", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		if token == "" {
			c.JSON(401, gin.H{"error": "token required"})
			return
		}

		userID, err := middleware.ParseToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &ws.Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: userID,
		}
		hub.Register(client)

		go client.WritePump()
		client.ReadPump()
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("FlexGram server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	paths := []string{
		"/app/migrations/001_initial_schema.sql",
		"internal/database/migrations/001_initial_schema.sql",
		"migrations/001_initial_schema.sql",
	}

	var content []byte
	for _, p := range paths {
		content, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if content == nil {
		return err
	}

	var applied bool
	_ = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = '001_initial_schema.sql')`).Scan(&applied)
	if applied {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ('001_initial_schema.sql')`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
