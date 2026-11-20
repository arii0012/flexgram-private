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
	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT u.id::text, u.username, COALESCE(u.email,''), u.display_name, u.bio,
		       u.avatar_url, u.status_emoji, u.profile_color, u.profile_pattern,
		       u.is_verified, u.is_premium, u.is_online, u.last_seen, u.created_at,
		       COALESCE(w.balance, 0)
		FROM users u
		LEFT JOIN wallets w ON w.user_id = u.id
		WHERE u.id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.StatusEmoji, &user.ProfileColor, &user.ProfilePattern,
		&user.IsVerified, &user.IsPremium, &user.IsOnline, &user.LastSeen, &user.CreatedAt,
		&user.Balance,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		DisplayName    string `json:"display_name"`
		Bio            string `json:"bio"`
		AvatarURL      string `json:"avatar_url"`
		StatusEmoji    string `json:"status_emoji"`
		ProfileColor   string `json:"profile_color"`
		ProfilePattern string `json:"profile_pattern"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	_, err := h.DB.Exec(c.Request.Context(), `
		UPDATE users SET
			display_name = COALESCE(NULLIF($1,''), display_name),
			bio = COALESCE(NULLIF($2,''), bio),
			avatar_url = COALESCE(NULLIF($3,''), avatar_url),
			status_emoji = COALESCE(NULLIF($4,''), status_emoji),
			profile_color = COALESCE(NULLIF($5,''), profile_color),
			profile_pattern = COALESCE(NULLIF($6,''), profile_pattern),
			updated_at = NOW()
		WHERE id = $7
	`, req.DisplayName, req.Bio, req.AvatarURL, req.StatusEmoji, req.ProfileColor, req.ProfilePattern, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ProfileHandler) Search(c *gin.Context) {
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
		return
	}
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id::text, username, display_name, avatar_url, is_verified, is_premium, is_online
		FROM users
		WHERE username ILIKE $1 OR display_name ILIKE $1
		LIMIT 20
	`, "%"+q+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.IsVerified, &u.IsPremium, &u.IsOnline); err != nil {
			continue
		}
		users = append(users, u)
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}
