package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	ws "flexgram-private/backend/internal/websocket"
)

type ExtraHandler struct {
	DB  *pgxpool.Pool
	Hub *ws.Hub
}

func NewExtraHandler(db *pgxpool.Pool, hub *ws.Hub) *ExtraHandler {
	return &ExtraHandler{DB: db, Hub: hub}
}

// POST /api/chats/:chatID/typing
func (h *ExtraHandler) Typing(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	var req struct {
		IsTyping bool `json:"is_typing"`
	}
	_ = c.ShouldBindJSON(&req)

	// notify other members
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT user_id::text FROM chat_members WHERE chat_id = $1 AND user_id != $2
	`, chatID, userID)
	if err == nil {
		defer rows.Close()
		var ids []string
		for rows.Next() {
			var uid string
			if rows.Scan(&uid) == nil {
				ids = append(ids, uid)
			}
		}
		h.Hub.SendToUsers(ids, "typing", map[string]interface{}{
			"chat_id":   chatID,
			"user_id":   userID,
			"is_typing": req.IsTyping,
		})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/messages/:messageID/react
func (h *ExtraHandler) React(c *gin.Context) {
	userID := c.GetString("user_id")
	msgID := c.Param("messageID")

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Emoji == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "emoji required"})
		return
	}

	// toggle reaction
	var exists bool
	_ = h.DB.QueryRow(c.Request.Context(), `
		SELECT EXISTS(SELECT 1 FROM message_reactions WHERE message_id=$1 AND user_id=$2 AND emoji=$3)
	`, msgID, userID, req.Emoji).Scan(&exists)

	if exists {
		_, _ = h.DB.Exec(c.Request.Context(), `
			DELETE FROM message_reactions WHERE message_id=$1 AND user_id=$2 AND emoji=$3
		`, msgID, userID, req.Emoji)
	} else {
		_, _ = h.DB.Exec(c.Request.Context(), `
			INSERT INTO message_reactions (message_id, user_id, emoji) VALUES ($1,$2,$3)
			ON CONFLICT DO NOTHING
		`, msgID, userID, req.Emoji)
	}

	// load all reactions for this message
	rows, _ := h.DB.Query(c.Request.Context(), `
		SELECT emoji, COUNT(*) FROM message_reactions WHERE message_id=$1 GROUP BY emoji
	`, msgID)
	reactions := map[string]int{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var e string
			var cnt int
			if rows.Scan(&e, &cnt) == nil {
				reactions[e] = cnt
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"reactions": reactions, "toggled": !exists})
}

// POST /api/chats/:chatID/read
func (h *ExtraHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	_, err := h.DB.Exec(c.Request.Context(), `
		UPDATE chat_members SET last_read_at = NOW() WHERE chat_id=$1 AND user_id=$2
	`, chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	// notify peer that messages were read
	rows, _ := h.DB.Query(c.Request.Context(), `
		SELECT user_id::text FROM chat_members WHERE chat_id=$1 AND user_id != $2
	`, chatID, userID)
	if rows != nil {
		defer rows.Close()
		var ids []string
		for rows.Next() {
			var uid string
			if rows.Scan(&uid) == nil {
				ids = append(ids, uid)
			}
		}
		h.Hub.SendToUsers(ids, "messages_read", map[string]interface{}{
			"chat_id": chatID,
			"user_id": userID,
			"read_at": time.Now().UTC().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/chats/:chatID/pin
func (h *ExtraHandler) PinMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")
	var req struct {
		MessageID string `json:"message_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.MessageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id required"})
		return
	}

	_, err := h.DB.Exec(c.Request.Context(), `
		INSERT INTO pinned_messages (chat_id, message_id, pinned_by)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, chatID, req.MessageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pin failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/chats/:chatID/pinned
func (h *ExtraHandler) GetPinned(c *gin.Context) {
	chatID := c.Param("chatID")
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT m.id::text, m.content, m.created_at, COALESCE(u.username,'')
		FROM pinned_messages p
		JOIN messages m ON m.id = p.message_id
		LEFT JOIN users u ON u.id = m.sender_id
		WHERE p.chat_id = $1
		ORDER BY p.pinned_at DESC
	`, chatID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"pinned": []interface{}{}})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, content, username string
		var created time.Time
		if rows.Scan(&id, &content, &created, &username) == nil {
			list = append(list, map[string]interface{}{
				"id": id, "content": content, "created_at": created, "username": username,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pinned": list})
}

// POST /api/users/online
func (h *ExtraHandler) SetOnline(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Online bool `json:"online"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.Online {
		_, _ = h.DB.Exec(c.Request.Context(), `UPDATE users SET is_online=true, last_seen=NOW() WHERE id=$1`, userID)
	} else {
		_, _ = h.DB.Exec(c.Request.Context(), `UPDATE users SET is_online=false, last_seen=NOW() WHERE id=$1`, userID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
