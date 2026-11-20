
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptHandler struct {
	DB *pgxpool.Pool
}

func NewReceiptHandler(db *pgxpool.Pool) *ReceiptHandler {
	return &ReceiptHandler{DB: db}
}

// MarkDelivered marks messages as delivered to the current user.
func (h *ReceiptHandler) MarkDelivered(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	_, err := h.DB.Exec(
		c.Request.Context(),
		`INSERT INTO message_receipts
		 (message_id, user_id, delivered_at)
		 SELECT m.id, $2, NOW()
		 FROM messages m
		 WHERE m.chat_id = $1
		 AND m.sender_id <> $2
		 AND m.deleted_at IS NULL
		 AND EXISTS (
		     SELECT 1 FROM chat_members
		     WHERE chat_id = $1 AND user_id = $2
		 )
		 ON CONFLICT (message_id, user_id)
		 DO UPDATE SET
		     delivered_at = COALESCE(
		         message_receipts.delivered_at,
		         NOW()
		     )`,
		chatID,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update delivery receipts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "delivered",
	})
}

// MarkRead marks messages as read by the current user.
func (h *ReceiptHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatID")

	_, err := h.DB.Exec(
		c.Request.Context(),
		`INSERT INTO message_receipts
		 (message_id, user_id, delivered_at, read_at)
		 SELECT m.id, $2, NOW(), NOW()
		 FROM messages m
		 WHERE m.chat_id = $1
		 AND m.sender_id <> $2
		 AND m.deleted_at IS NULL
		 AND EXISTS (
		     SELECT 1 FROM chat_members
		     WHERE chat_id = $1 AND user_id = $2
		 )
		 ON CONFLICT (message_id, user_id)
		 DO UPDATE SET
		     delivered_at = COALESCE(
		         message_receipts.delivered_at,
		         NOW()
		     ),
		     read_at = COALESCE(
		         message_receipts.read_at,
		         NOW()
		     )`,
		chatID,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update read receipts",
		})
		return
	}

	_, err = h.DB.Exec(
		c.Request.Context(),
		`UPDATE chat_members
		 SET last_read_at = NOW()
		 WHERE chat_id = $1 AND user_id = $2`,
		chatID,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update read position",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "read",
	})
}
