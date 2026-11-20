
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/models"
)

type ChatHandler struct {
	DB *pgxpool.Pool
}

func NewChatHandler(db *pgxpool.Pool) *ChatHandler {
	return &ChatHandler{DB: db}
}

type CreateChatRequest struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	AvatarURL   string   `json:"avatar_url"`
	UserIDs     []string `json:"user_ids"`
}

func (h *ChatHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Type != "direct" && req.Type != "group" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Chat type must be direct or group",
		})
		return
	}

	if len(req.UserIDs) == 0 || len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid participants count",
		})
		return
	}

	if req.Type == "group" && req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group title is required",
		})
		return
	}

	if len(req.Title) > 100 ||
		len(req.Description) > 500 ||
		len(req.AvatarURL) > 2048 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Chat information is too long",
		})
		return
	}

	participants := make(map[string]bool)
	participants[userID] = true

	for _, id := range req.UserIDs {
		id = strings.TrimSpace(id)

		if id != "" {
			participants[id] = true
		}
	}

	if req.Type == "direct" && len(participants) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A direct chat requires exactly two users",
		})
		return
	}

	if req.Type == "group" && len(participants) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A group requires at least two users",
		})
		return
	}

	ids := make([]string, 0, len(participants))

	for id := range participants {
		ids = append(ids, id)
	}

	tx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check that all participants exist.
	var count int

	err = tx.QueryRow(
		c.Request.Context(),
		`SELECT COUNT(*) FROM users WHERE id = ANY($1::uuid[])`,
		ids,
	).Scan(&count)

	if err != nil || count != len(ids) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "One or more users do not exist",
		})
		return
	}

	// Reuse an existing direct chat if one exists.
	if req.Type == "direct" {
		var existingID string

		err = tx.QueryRow(
			c.Request.Context(),
			`SELECT c.id::text
			 FROM chats c
			 JOIN chat_members m ON m.chat_id = c.id
			 WHERE c.type = 'direct'
			 AND m.user_id = ANY($1::uuid[])
			 GROUP BY c.id
			 HAVING COUNT(DISTINCT m.user_id) = 2
			 AND (
			     SELECT COUNT(*)
			     FROM chat_members all_members
			     WHERE all_members.chat_id = c.id
			 ) = 2
			 LIMIT 1`,
			ids,
		).Scan(&existingID)

		if err == nil {
			if err := tx.Commit(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Could not complete transaction",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"chat_id": existingID,
				"message": "Existing chat found",
			})
			return
		}

		if err != pgx.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not find existing chat",
			})
			return
		}
	}

	var chat models.Chat

	err = tx.QueryRow(
		c.Request.Context(),
		`INSERT INTO chats
		 (type, title, description, avatar_url, created_by)
		 VALUES ($1, NULLIF($2, ''), NULLIF($3, ''),
		         NULLIF($4, ''), $5)
		 RETURNING id::text, type,
		 COALESCE(title, ''),
		 COALESCE(description, ''),
		 COALESCE(avatar_url, ''),
		 created_by::text, created_at, updated_at`,
		req.Type,
		req.Title,
		req.Description,
		req.AvatarURL,
		userID,
	).Scan(
		&chat.ID,
		&chat.Type,
		&chat.Title,
		&chat.Description,
		&chat.AvatarURL,
		&chat.CreatedBy,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create chat",
		})
		return
	}

	for _, id := range ids {
		role := "member"

		if id == userID {
			role = "owner"
		}

		_, err = tx.Exec(
			c.Request.Context(),
			`INSERT INTO chat_members
			 (chat_id, user_id, role)
			 VALUES ($1, $2, $3)`,
			chat.ID,
			id,
			role,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not add chat participants",
			})
			return
		}
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not save chat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"chat": chat,
	})
}

func (h *ChatHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.DB.Query(
		c.Request.Context(),
		`SELECT
		    ch.id::text,
		    ch.type,
		    COALESCE(ch.title, ''),
		    COALESCE(ch.description, ''),
		    COALESCE(ch.avatar_url, ''),
		    ch.created_at,
		    ch.updated_at,
		    (
		        SELECT COUNT(*)
		        FROM chat_members cm
		        WHERE cm.chat_id = ch.id
		    )
		 FROM chats ch
		 JOIN chat_members mine ON mine.chat_id = ch.id
		 WHERE mine.user_id = $1
		 ORDER BY ch.updated_at DESC`,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not load chats",
		})
		return
	}
	defer rows.Close()

	chats := make([]gin.H, 0)

	for rows.Next() {
		var chat models.Chat
		var members int

		err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.Title,
			&chat.Description,
			&chat.AvatarURL,
			&chat.CreatedAt,
			&chat.UpdatedAt,
			&members,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not read chats",
			})
			return
		}

		chats = append(chats, gin.H{
			"chat":          chat,
			"members_count": members,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not read chats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
		"count": len(chats),
	})
}
