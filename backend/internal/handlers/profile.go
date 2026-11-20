
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/models"
)

type ProfileHandler struct {
	DB *pgxpool.Pool
}

func NewProfileHandler(db *pgxpool.Pool) *ProfileHandler {
	return &ProfileHandler{DB: db}
}

func (h *ProfileHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User

	err := h.DB.QueryRow(
		c.Request.Context(),
		`SELECT id::text, username, email,
		 COALESCE(display_name, ''),
		 COALESCE(avatar_url, ''),
		 COALESCE(bio, ''),
		 is_online, last_seen, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.IsOnline,
		&user.LastSeen,
		&user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Bio = strings.TrimSpace(req.Bio)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)

	if len(req.DisplayName) > 64 ||
		len(req.Bio) > 500 ||
		len(req.AvatarURL) > 2048 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Profile field is too long",
		})
		return
	}

	tag, err := h.DB.Exec(
		c.Request.Context(),
		`UPDATE users
		 SET display_name = $1,
		     avatar_url = $2,
		     bio = $3,
		     updated_at = NOW()
		 WHERE id = $4`,
		req.DisplayName,
		req.AvatarURL,
		req.Bio,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profile update failed",
		})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

func (h *ProfileHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))

	if len(query) < 2 || len(query) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query must contain 2-32 characters",
		})
		return
	}

	rows, err := h.DB.Query(
		c.Request.Context(),
		`SELECT id::text, username,
		 COALESCE(display_name, ''),
		 COALESCE(avatar_url, ''),
		 COALESCE(bio, ''),
		 is_online, created_at
		 FROM users
		 WHERE id <> $1
		 AND (
		     username ILIKE $2
		     OR display_name ILIKE $2
		 )
		 ORDER BY username
		 LIMIT 20`,
		c.GetString("user_id"),
		"%"+query+"%",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var user models.User

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Bio,
			&user.IsOnline,
			&user.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not read search results",
			})
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}
