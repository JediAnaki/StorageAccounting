// Package telegram handles Telegram Bot API integration
package telegram

import (
	"fmt"
)

// WebAppConfig contains configuration for the Telegram Mini App
type WebAppConfig struct {
	// BotUsername is the Telegram bot username (e.g., "mybot" from @mybot)
	BotUsername string

	// WebAppURL is the HTTPS URL where the Mini App is hosted
	WebAppURL string

	// ButtonText is the text displayed on the Mini App launch button
	ButtonText string
}

// GenerateWebAppURL generates a deep link URL to launch the Mini App
// This URL can be used in inline keyboard buttons or sent as a message
func GenerateWebAppURL(cfg *WebAppConfig) string {
	// Telegram Mini App URL format: https://t.me/{bot_username}/{app_short_name}
	// For direct web_app launch in inline keyboards, use the web_app parameter
	return fmt.Sprintf("https://t.me/%s?start=webapp", cfg.BotUsername)
}

// GenerateInlineKeyboardButton generates JSON for an inline keyboard button
// that launches the Mini App when clicked
func GenerateInlineKeyboardButton(cfg *WebAppConfig) map[string]interface{} {
	return map[string]interface{}{
		"text": cfg.ButtonText,
		"web_app": map[string]string{
			"url": cfg.WebAppURL,
		},
	}
}

// GenerateMenuButton generates a menu button configuration
// that replaces the default bot menu button with a Mini App launcher
func GenerateMenuButton(cfg *WebAppConfig) map[string]interface{} {
	return map[string]interface{}{
		"type": "web_app",
		"text": cfg.ButtonText,
		"web_app": map[string]string{
			"url": cfg.WebAppURL,
		},
	}
}
