package model

import "time"

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Nickname     string     `json:"nickname"`
	AvatarURL    string     `json:"avatarUrl,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type AuthSession struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
