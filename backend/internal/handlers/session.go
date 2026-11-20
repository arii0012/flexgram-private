
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type SessionHandler struct {
	Redis *redis.Client
}

func NewSessionHandler(client *redis.Client) *SessionHandler {
	return &SessionHandler{
		Redis: client,
	}
}

func (h *SessionHandler) Logout(c *gin.Context) {
	token := c.GetString("token")
	expiresAt, exists := c.Get("token_exp")

	if token == "" || !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid session",
		})
		return
	}

	expiration, ok := expiresAt.(time.Time)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid session expiration",
		})
		return
	}

	ttl := time.Until(expiration)

	if ttl <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Session expired",
		})
		return
	}

	err := h.Redis.Set(
		c.Request.Context(),
		"revoked:"+token,
		"1",
		ttl,
	).Err()

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Could not close session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
