package handlers

import (
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"flexgram-private/backend/internal/models"
)

type GiftsHandler struct {
	DB *pgxpool.Pool
}

func NewGiftsHandler(db *pgxpool.Pool) *GiftsHandler {
	return &GiftsHandler{DB: db}
}

func (h *GiftsHandler) ListShop(c *gin.Context) {
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id::text, name, description, price, rarity, image_url, background, pattern, is_active
		FROM gifts WHERE is_active = true ORDER BY price
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var gifts []models.Gift
	for rows.Next() {
		var g models.Gift
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Price, &g.Rarity, &g.ImageURL, &g.Background, &g.Pattern, &g.IsActive); err != nil {
			continue
		}
		gifts = append(gifts, g)
	}
	c.JSON(http.StatusOK, gin.H{"gifts": gifts})
}

func (h *GiftsHandler) MyGifts(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT ug.id::text, ug.gift_id::text, ug.owner_id::text, ug.serial_number, ug.is_equipped, ug.acquired_at,
		       g.name, g.description, g.price, g.rarity, g.image_url, g.background, g.pattern
		FROM user_gifts ug
		JOIN gifts g ON g.id = ug.gift_id
		WHERE ug.owner_id = $1
		ORDER BY ug.acquired_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var list []models.UserGift
	for rows.Next() {
		var ug models.UserGift
		var g models.Gift
		if err := rows.Scan(&ug.ID, &ug.GiftID, &ug.OwnerID, &ug.SerialNumber, &ug.IsEquipped, &ug.AcquiredAt,
			&g.Name, &g.Description, &g.Price, &g.Rarity, &g.ImageURL, &g.Background, &g.Pattern); err != nil {
			continue
		}
		g.ID = ug.GiftID
		ug.Gift = &g
		list = append(list, ug)
	}
	c.JSON(http.StatusOK, gin.H{"gifts": list})
}

func (h *GiftsHandler) Buy(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		GiftID string `json:"gift_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.GiftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gift_id required"})
		return
	}

	tx, err := h.DB.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	var price int64
	err = tx.QueryRow(c.Request.Context(), `SELECT price FROM gifts WHERE id = $1 AND is_active`, req.GiftID).Scan(&price)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift not found"})
		return
	}

	var balance int64
	err = tx.QueryRow(c.Request.Context(), `SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE`, userID).Scan(&balance)
	if err != nil || balance < price {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	_, err = tx.Exec(c.Request.Context(), `UPDATE wallets SET balance = balance - $1, updated_at = NOW() WHERE user_id = $2`, price, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment failed"})
		return
	}

	var serial int64
	_ = tx.QueryRow(c.Request.Context(), `SELECT COALESCE(MAX(serial_number),0)+1 FROM user_gifts WHERE gift_id = $1`, req.GiftID).Scan(&serial)

	var ugID string
	err = tx.QueryRow(c.Request.Context(), `
		INSERT INTO user_gifts (gift_id, owner_id, serial_number)
		VALUES ($1, $2, $3) RETURNING id::text
	`, req.GiftID, userID, serial).Scan(&ugID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create gift"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "user_gift_id": ugID, "serial_number": serial, "new_balance": balance - price})
}

func (h *GiftsHandler) Transfer(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		UserGiftID string `json:"user_gift_id"`
		ToUsername string `json:"to_username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var toID string
	err := h.DB.QueryRow(c.Request.Context(), `SELECT id::text FROM users WHERE username = $1`, strings.ToLower(req.ToUsername)).Scan(&toID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	tag, err := h.DB.Exec(c.Request.Context(), `
		UPDATE user_gifts SET owner_id = $1, is_equipped = false
		WHERE id = $2 AND owner_id = $3
	`, toID, req.UserGiftID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer this gift"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *GiftsHandler) ToggleEquip(c *gin.Context) {
	userID := c.GetString("user_id")
	ugID := c.Param("id")

	var equipped bool
	err := h.DB.QueryRow(c.Request.Context(), `
		UPDATE user_gifts SET is_equipped = NOT is_equipped
		WHERE id = $1 AND owner_id = $2
		RETURNING is_equipped
	`, ugID, userID).Scan(&equipped)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gift not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_equipped": equipped})
}

// ---- God Mode / Developer Panel ----

type GodHandler struct {
	DB *pgxpool.Pool
}

func NewGodHandler(db *pgxpool.Pool) *GodHandler {
	return &GodHandler{DB: db}
}

func (h *GodHandler) AddStars(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount required"})
		return
	}
	if req.Amount > 1_000_000 {
		req.Amount = 1_000_000
	}

	var newBalance int64
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO wallets (user_id, balance) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET balance = wallets.balance + $2, updated_at = NOW()
		RETURNING balance
	`, userID, req.Amount).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "balance": newBalance})
}

func (h *GodHandler) SetPremium(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Premium  bool   `json:"premium"`
		Verified bool   `json:"verified"`
		Color    string `json:"color"`
		Pattern  string `json:"pattern"`
	}
	_ = c.ShouldBindJSON(&req)

	_, err := h.DB.Exec(c.Request.Context(), `
		UPDATE users SET
			is_premium = $1,
			is_verified = $2,
			profile_color = COALESCE(NULLIF($3,''), profile_color),
			profile_pattern = COALESCE(NULLIF($4,''), profile_pattern),
			updated_at = NOW()
		WHERE id = $5
	`, req.Premium, req.Verified, req.Color, req.Pattern, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *GodHandler) GrantGift(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		GiftID       string `json:"gift_id"`
		SerialNumber int64  `json:"serial_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.GiftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gift_id required"})
		return
	}

	serial := req.SerialNumber
	if serial <= 0 {
		_ = h.DB.QueryRow(c.Request.Context(), `SELECT COALESCE(MAX(serial_number),0)+1 FROM user_gifts WHERE gift_id = $1`, req.GiftID).Scan(&serial)
	}

	var ugID string
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO user_gifts (gift_id, owner_id, serial_number)
		VALUES ($1, $2, $3) RETURNING id::text
	`, req.GiftID, userID, serial).Scan(&ugID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant gift"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "user_gift_id": ugID, "serial_number": serial})
}

func (h *GodHandler) ChangeUsername(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if len(req.Username) < 3 || len(req.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username 3-32 chars"})
		return
	}

	tag, err := h.DB.Exec(c.Request.Context(), `UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2`, req.Username, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username taken"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "username": req.Username})
}

