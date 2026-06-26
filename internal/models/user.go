package models

import "time"

type User struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	FamilyID          *int      `json:"family_id,omitempty"`
	TelegramChatID    *int64    `json:"telegram_chat_id,omitempty"`
	TelegramAuthToken *string   `json:"telegram_auth_token,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
