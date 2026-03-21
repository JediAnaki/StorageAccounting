// Package telegram handles Telegram Bot API integration
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Bot represents a Telegram Bot API client
type Bot struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewBot creates a new Telegram bot client
func NewBot(token string) *Bot {
	return &Bot{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
}

// SetWebhook sets the webhook URL for receiving updates
// url: HTTPS URL to send updates to (must be HTTPS)
// Returns error if the webhook setup fails
func (b *Bot) SetWebhook(url string) error {
	endpoint := fmt.Sprintf("%s/setWebhook", b.baseURL)

	payload := map[string]interface{}{
		"url": url,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	resp, err := b.httpClient.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Ok {
		return fmt.Errorf("webhook setup failed: %s", result.Description)
	}

	return nil
}

// GetWebhookInfo retrieves current webhook configuration
func (b *Bot) GetWebhookInfo() (*WebhookInfo, error) {
	endpoint := fmt.Sprintf("%s/getWebhookInfo", b.baseURL)

	resp, err := b.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Ok     bool         `json:"ok"`
		Result WebhookInfo  `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("failed to get webhook info")
	}

	return &result.Result, nil
}

// SendMessage sends a text message to a chat
func (b *Bot) SendMessage(chatID int64, text string) error {
	endpoint := fmt.Sprintf("%s/sendMessage", b.baseURL)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	resp, err := b.httpClient.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Ok {
		return fmt.Errorf("send message failed: %s", result.Description)
	}

	return nil
}

// WebhookInfo represents Telegram webhook configuration
type WebhookInfo struct {
	URL                  string `json:"url"`
	HasCustomCertificate bool   `json:"has_custom_certificate"`
	PendingUpdateCount   int    `json:"pending_update_count"`
	LastErrorDate        int64  `json:"last_error_date"`
	LastErrorMessage     string `json:"last_error_message"`
	MaxConnections       int    `json:"max_connections"`
}
