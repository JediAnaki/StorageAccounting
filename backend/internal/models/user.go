// Package models contains domain entities for the inventory management system
package models

import "time"

// User represents a Telegram user identity in the system
// User information is extracted from Telegram WebApp initData during authentication
type User struct {
	// TelegramUserID is the unique Telegram user identifier (from initDataUnsafe.user.id)
	TelegramUserID int64 `json:"telegram_user_id"`

	// FirstName is the user's first name from Telegram profile
	FirstName string `json:"first_name"`

	// LastName is the user's last name from Telegram profile (optional)
	LastName string `json:"last_name,omitempty"`

	// Username is the Telegram username (optional, not all users have one)
	Username string `json:"username,omitempty"`

	// PhotoURL is the URL to the user's Telegram profile photo (optional)
	PhotoURL string `json:"photo_url,omitempty"`

	// LanguageCode is the user's language preference from Telegram (e.g., "en", "ru")
	LanguageCode string `json:"language_code,omitempty"`

	// CreatedAt is the timestamp when this user first accessed the system
	CreatedAt time.Time `json:"created_at"`

	// LastAccessAt is the timestamp of the user's most recent activity
	LastAccessAt time.Time `json:"last_access_at"`
}

// DisplayName returns a human-readable display name for the user
// Prefers FirstName + LastName, falls back to Username, then FirstName only
func (u *User) DisplayName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return "Unknown User"
}
