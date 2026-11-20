
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatusHandler struct {
	DB *pgxpool.Pool
}

func NewStatusHandler(db *pgxpool.Pool) *StatusHandler {
	return &StatusHandler{DB: db}
}

func (h *StatusHandler) SetOnline(c *gin.Context) {
	userID := c.GetString("user_id")

	_, err := h.DB.Exec(
		c.Request.Context(),
		`UPDATE users
		 SET is_online = TRUE,
		     last_seen = NOW()
		 WHERE id = $1`,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_online": true,
	})
}

func (h *StatusHandler) SetOffline(c *gin.Context) {
	userID := c.GetString("user_id")

	_, err := h.DB.Exec(
		c.Request.Context(),
		`UPDATE users
		 SET is_online = FALSE,
		     last_seen = NOW()
		 WHERE id = $1`,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_online": false,
	})
}
