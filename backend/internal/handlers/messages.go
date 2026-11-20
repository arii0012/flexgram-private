package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageHandler struct {
	DB *pgxpool.Pool
}

func NewMessageHandler(db *pgxpool.Pool) *MessageHandler {
	return &MessageHandler{
		DB: db,
	}
}

type Message struct {
	ID          string     `json:"id"`
	ChatID      string     `json:"chat_id"`
	SenderID    *string    `json:"sender_id"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	ReplyTo     *string    `json:"reply_to"`
	EditedAt    *time.Time `json:"edited_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Send отправляет сообщение в чат.
func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if _, err := uuid.Parse(chatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	var request struct {
		Content     string  `json:"content"`
		MessageType string  `json:"message_type"`
		ReplyTo     *string `json:"reply_to"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	request.Content = strings.TrimSpace(request.Content)
	request.MessageType = strings.TrimSpace(request.MessageType)

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message content cannot be empty",
		})
		return
	}

	if len(request.Content) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message is too long",
		})
		return
	}

	if request.MessageType == "" {
		request.MessageType = "text"
	}

	if request.MessageType != "text" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported message type",
		})
		return
	}

	if request.ReplyTo != nil {
		if _, err := uuid.Parse(*request.ReplyTo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid reply_to ID",
			})
			return
		}
	}

	// Проверяем, состоит ли пользователь в чате.
	var isMember bool

	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM chat_members
			WHERE chat_id = $1
			  AND user_id = $2
		)
	`, chatID, userID).Scan(&isMember)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check chat membership",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not a member of this chat",
		})
		return
	}

	// Если это ответ на другое сообщение, проверяем его наличие в этом чате.
	if request.ReplyTo != nil {
		var replyExists bool

		err = h.DB.QueryRow(c.Request.Context(), `
			SELECT EXISTS (
				SELECT 1
				FROM messages
				WHERE id = $1
				  AND chat_id = $2
				  AND deleted_at IS NULL
			)
		`, *request.ReplyTo, chatID).Scan(&replyExists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check reply message",
			})
			return
		}

		if !replyExists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Reply message not found",
			})
			return
		}
	}

	var message Message

	err = h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO messages (
			chat_id,
			sender_id,
			content,
			message_type,
			reply_to
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			chat_id,
			sender_id,
			content,
			message_type,
			reply_to,
			edited_at,
			deleted_at,
			created_at
	`,
		chatID,
		userID,
		request.Content,
		request.MessageType,
		request.ReplyTo,
	).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.Content,
		&message.MessageType,
		&message.ReplyTo,
		&message.EditedAt,
		&message.DeletedAt,
		&message.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send message",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
	})
}

// History возвращает историю сообщений чата.
func (h *MessageHandler) History(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if _, err := uuid.Parse(chatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	limit := 50

	if limitValue := c.Query("limit"); limitValue != "" {
		parsedLimit, err := strconv.Atoi(limitValue)

		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Limit must be between 1 and 100",
			})
			return
		}

		limit = parsedLimit
	}

	var isMember bool

	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM chat_members
			WHERE chat_id = $1
			  AND user_id = $2
		)
	`, chatID, userID).Scan(&isMember)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check chat membership",
		})
		return
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not a member of this chat",
		})
		return
	}

	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT
			id,
			chat_id,
			sender_id,
			CASE
				WHEN deleted_at IS NOT NULL THEN ''
				ELSE content
			END AS content,
			message_type,
			reply_to,
			edited_at,
			deleted_at,
			created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, chatID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load messages",
		})
		return
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.MessageType,
			&message.ReplyTo,
			&message.EditedAt,
			&message.DeletedAt,
			&message.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read messages",
			})
			return
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read messages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// Edit редактирует сообщение.
func (h *MessageHandler) Edit(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("messageID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if _, err := uuid.Parse(messageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message ID",
		})
		return
	}

	var request struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	request.Content = strings.TrimSpace(request.Content)

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message content cannot be empty",
		})
		return
	}

	if len(request.Content) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Message is too long",
		})
		return
	}

	var message Message

	err := h.DB.QueryRow(c.Request.Context(), `
		UPDATE messages
		SET content = $1,
		    edited_at = NOW()
		WHERE id = $2
		  AND sender_id = $3
		  AND deleted_at IS NULL
		  AND EXISTS (
				SELECT 1
				FROM chat_members
				WHERE chat_members.chat_id = messages.chat_id
				  AND chat_members.user_id = $3
		  )
		RETURNING
			id,
			chat_id,
			sender_id,
			content,
			message_type,
			reply_to,
			edited_at,
			deleted_at,
			created_at
	`, request.Content, messageID, userID).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.Content,
		&message.MessageType,
		&message.ReplyTo,
		&message.EditedAt,
		&message.DeletedAt,
		&message.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Message not found or you cannot edit it",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to edit message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// Delete удаляет сообщение.
func (h *MessageHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("messageID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if _, err := uuid.Parse(messageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message ID",
		})
		return
	}

	var deletedMessageID string

	err := h.DB.QueryRow(c.Request.Context(), `
		UPDATE messages
		SET content = '',
		    deleted_at = NOW()
		WHERE id = $1
		  AND sender_id = $2
		  AND deleted_at IS NULL
		  AND EXISTS (
				SELECT 1
				FROM chat_members
				WHERE chat_members.chat_id = messages.chat_id
				  AND chat_members.user_id = $2
		  )
		RETURNING id
	`, messageID, userID).Scan(&deletedMessageID)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Message not found or you cannot delete it",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"id":     deletedMessageID,
	})
}
