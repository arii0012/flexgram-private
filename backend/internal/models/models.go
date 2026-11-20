package models

import "time"

type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email,omitempty"`
	DisplayName    string     `json:"display_name"`
	Bio            string     `json:"bio"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	StatusEmoji    string     `json:"status_emoji,omitempty"`
	ProfileColor   string     `json:"profile_color"`
	ProfilePattern string     `json:"profile_pattern"`
	IsVerified     bool       `json:"is_verified"`
	IsPremium      bool       `json:"is_premium"`
	IsOnline       bool       `json:"is_online"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	Balance        int64      `json:"balance,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Chat struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastMessage *Message  `json:"last_message,omitempty"`
	UnreadCount int       `json:"unread_count,omitempty"`
	Peer        *User     `json:"peer,omitempty"`
}

type Message struct {
	ID          string     `json:"id"`
	ChatID      string     `json:"chat_id"`
	SenderID    string     `json:"sender_id"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	ReplyTo     *string    `json:"reply_to,omitempty"`
	EditedAt    *time.Time `json:"edited_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	Sender      *User      `json:"sender,omitempty"`
}

type Gift struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Rarity      string `json:"rarity"`
	ImageURL    string `json:"image_url"`
	Background  string `json:"background"`
	Pattern     string `json:"pattern"`
	IsActive    bool   `json:"is_active"`
}

type UserGift struct {
	ID           string    `json:"id"`
	GiftID       string    `json:"gift_id"`
	OwnerID      string    `json:"owner_id"`
	SerialNumber *int64    `json:"serial_number,omitempty"`
	IsEquipped   bool      `json:"is_equipped"`
	AcquiredAt   time.Time `json:"acquired_at"`
	Gift         *Gift     `json:"gift,omitempty"`
}
