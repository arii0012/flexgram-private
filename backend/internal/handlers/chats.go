package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/models"
	ws "flexgram-private/backend/internal/websocket"
)

type ChatHandler struct {
	DB  *pgxpool.Pool
	Hub *ws.Hub
}

func NewChatHandler(db *pgxpool.Pool, hub *ws.Hub) *ChatHandler {
	return &ChatHandler{DB: db, Hub: hub}
}

func (h *ChatHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Type      string   `json:"type"`
		Username  string   `json:"username"`
		Title     string   `json:"title"`
		MemberIDs []string `json:"member_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if req.Type == "" {
		req.Type = "direct"
	}

	tx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var chatID string
	err = tx.QueryRow(c.Request.Context(), `
		INSERT INTO chats (type, title, created_by)
		VALUES ($1, $2, $3) RETURNING id::text
	`, req.Type, req.Title, userID).Scan(&chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create chat"})
		return
	}

	_, _ = tx.Exec(c.Request.Context(), `INSERT INTO chat_members (chat_id, user_id, role) VALUES ($1, $2, 'owner')`, chatID, userID)

	if req.Type == "direct" && req.Username != "" {
		var peerID string
		err = tx.QueryRow(c.Request.Context(), `SELECT id::text FROM users WHERE username = $1`, strings.ToLower(req.Username)).Scan(&peerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		var existing string
		err = tx.QueryRow(c.Request.Context(), `
			SELECT c.id::text FROM chats c
			JOIN chat_members m1 ON m1.chat_id = c.id AND m1.user_id = $1
			JOIN chat_members m2 ON m2.chat_id = c.id AND m2.user_id = $2
			WHERE c.type = 'direct' LIMIT 1
		`, userID, peerID).Scan(&existing)
		if err == nil && existing != "" {
			c.JSON(http.StatusOK, gin.H{"chat_id": existing, "already_exists": true})
			return
		}
		_, _ = tx.Exec(c.Request.Context(), `INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)`, chatID, peerID)
	}

	for _, mid := range req.MemberIDs {
		if mid != userID {
			_, _ = tx.Exec(c.Request.Context(), `INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, chatID, mid)
		}
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"chat_id": chatID})
}

func (h *ChatHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT c.id::text, c.type, COALESCE(c.title,''), COALESCE(c.description,''), COALESCE(c.avatar_url,''),
		       COALESCE(c.created_by::text,''), c.created_at, c.updated_at
		FROM chats c
		JOIN chat_members cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var ch models.Chat
		if err := rows.Scan(&ch.ID, &ch.Type, &ch.Title, &ch.Description, &ch.AvatarURL, &ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			continue
		}
		if ch.Type == "direct" {
			var peer models.User
			err := h.DB.QueryRow(c.Request.Context(), `
				SELECT u.id::text, u.username, u.display_name, u.avatar_url, u.is_online, u.is_verified, u.is_premium
				FROM users u
				JOIN chat_members cm ON cm.user_id = u.id
				WHERE cm.chat_id = $1 AND u.id::text != $2
				LIMIT 1
			`, ch.ID, userID).Scan(&peer.ID, &peer.Username, &peer.DisplayName, &peer.AvatarURL, &peer.IsOnline, &peer.IsVerified, &peer.IsPremium)
			if err == nil {
				ch.Peer = &peer
				if ch.Title == "" {
					ch.Title = peer.DisplayName
					if ch.Title == "" {
						ch.Title = "@" + peer.Username
					}
				}
			}
		}
		var msg models.Message
		err := h.DB.QueryRow(c.Request.Context(), `
			SELECT id::text, chat_id::text, COALESCE(sender_id::text,''), content, message_type, created_at
			FROM messages WHERE chat_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT 1
		`, ch.ID).Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.MessageType, &msg.CreatedAt)
		if err == nil {
			ch.LastMessage = &msg
		}
		chats = append(chats, ch)
	}
	c.JSON(http.StatusOK, gin.H{"chats": chats})
}
