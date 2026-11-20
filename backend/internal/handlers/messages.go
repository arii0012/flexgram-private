package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/models"
	ws "flexgram-private/backend/internal/websocket"
)

type MessageHandler struct {
	DB  *pgxpool.Pool
	Hub *ws.Hub
}

func NewMessageHandler(db *pgxpool.Pool, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{DB: db, Hub: hub}
}

func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	var req struct {
		Content     string  `json:"content"`
		MessageType string  `json:"message_type"`
		ReplyTo     *string `json:"reply_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content required"})
		return
	}
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	var exists bool
	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)
	`, chatID, userID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this chat"})
		return
	}

	var msg models.Message
	err = h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO messages (chat_id, sender_id, content, message_type, reply_to)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, chat_id::text, sender_id::text, content, message_type, created_at
	`, chatID, userID, req.Content, req.MessageType, req.ReplyTo).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.MessageType, &msg.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not send message: " + err.Error()})
		return
	}

	_, _ = h.DB.Exec(c.Request.Context(), `UPDATE chats SET updated_at = NOW() WHERE id = $1`, chatID)

	rows, _ := h.DB.Query(c.Request.Context(), `SELECT user_id::text FROM chat_members WHERE chat_id = $1 AND user_id != $2`, chatID, userID)
	if rows != nil {
		defer rows.Close()
		var ids []string
		for rows.Next() {
			var uid string
			if rows.Scan(&uid) == nil {
				ids = append(ids, uid)
			}
		}
		h.Hub.SendToUsers(ids, "new_message", msg)
	}

	c.JSON(http.StatusCreated, gin.H{"message": msg})
}

func (h *MessageHandler) History(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	var exists bool
	_ = h.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)`, chatID, userID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member"})
		return
	}

	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT m.id::text, m.chat_id::text, COALESCE(m.sender_id::text,''), m.content, m.message_type,
		       m.created_at, COALESCE(u.username,''), COALESCE(u.display_name,''), COALESCE(u.avatar_url,'')
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		WHERE m.chat_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT 100
	`, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		var uname, dname, avatar string
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Content, &m.MessageType, &m.CreatedAt, &uname, &dname, &avatar); err != nil {
			continue
		}
		if uname != "" {
			m.Sender = &models.User{ID: m.SenderID, Username: uname, DisplayName: dname, AvatarURL: avatar}
		}
		messages = append(messages, m)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) Edit(c *gin.Context) {
	userID := c.GetString("user_id")
	msgID := c.Param("messageID")
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content required"})
		return
	}
	tag, err := h.DB.Exec(c.Request.Context(), `
		UPDATE messages SET content = $1, edited_at = NOW()
		WHERE id = $2 AND sender_id = $3 AND deleted_at IS NULL
	`, req.Content, msgID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot edit message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	msgID := c.Param("messageID")
	tag, err := h.DB.Exec(c.Request.Context(), `
		UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND sender_id = $2
	`, msgID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
