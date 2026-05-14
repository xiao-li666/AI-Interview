package model

import "time"

type UserAIProviderConfig struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userId"`
	ProviderCode string     `json:"provider"`
	APIKey       string     `json:"-"`
	APIKeyMasked string     `json:"apiKeyMasked"`
	Model        string     `json:"model"`
	BaseURL      string     `json:"baseUrl"`
	IsEnabled    bool       `json:"isEnabled"`
	LastTestOK   bool       `json:"lastTestOk"`
	LastTestedAt *time.Time `json:"lastTestedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
