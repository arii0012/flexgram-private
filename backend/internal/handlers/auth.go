package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"flexgram-private/backend/internal/models"
)

type AuthHandler struct {
	DB        *pgxpool.Pool
	JWTSecret []byte
}

func NewAuthHandler(db *pgxpool.Pool, secret string) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: []byte(secret)}
}

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
}

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func (h *AuthHandler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !usernamePattern.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be 3-32 chars (letters, numbers, underscore)"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process password"})
		return
	}

	tx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var user models.User
	err = tx.QueryRow(c.Request.Context(), `
		INSERT INTO users (username, email, password_hash, display_name)
		VALUES ($1, NULLIF($2,''), $3, $1)
		RETURNING id::text, username, COALESCE(email,''), display_name, bio,
		          avatar_url, status_emoji, profile_color, profile_pattern,
		          is_verified, is_premium, is_online, last_seen, created_at
	`, req.Username, req.Email, string(hash)).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.StatusEmoji, &user.ProfileColor, &user.ProfilePattern,
		&user.IsVerified, &user.IsPremium, &user.IsOnline, &user.LastSeen, &user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}

	_, _ = tx.Exec(c.Request.Context(), `INSERT INTO wallets (user_id, balance) VALUES ($1, 100)`, user.ID)

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	user.Balance = 100
	token, _ := h.generateToken(user.ID)

	c.JSON(http.StatusCreated, AuthResponse{
		User: user, AccessToken: token, TokenType: "Bearer", ExpiresIn: 86400 * 7,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	login := strings.ToLower(strings.TrimSpace(req.Username))
	if login == "" {
		login = strings.ToLower(strings.TrimSpace(req.Email))
	}

	var user models.User
	var passwordHash string

	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT u.id::text, u.username, COALESCE(u.email,''), u.display_name, u.bio,
		       u.avatar_url, u.status_emoji, u.profile_color, u.profile_pattern,
		       u.is_verified, u.is_premium, u.is_online, u.last_seen, u.created_at,
		       u.password_hash, COALESCE(w.balance, 0)
		FROM users u
		LEFT JOIN wallets w ON w.user_id = u.id
		WHERE u.username = $1 OR u.email = $1
	`, login).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.StatusEmoji, &user.ProfileColor, &user.ProfilePattern,
		&user.IsVerified, &user.IsPremium, &user.IsOnline, &user.LastSeen, &user.CreatedAt,
		&passwordHash, &user.Balance,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := h.generateToken(user.ID)
	c.JSON(http.StatusOK, AuthResponse{
		User: user, AccessToken: token, TokenType: "Bearer", ExpiresIn: 86400 * 7,
	})
}

func (h *AuthHandler) generateToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "flexgram",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.JWTSecret)
}
