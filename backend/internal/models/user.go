
package models

import "time"

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Bio         string    `json:"bio"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
