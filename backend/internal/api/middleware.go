// Package api contains HTTP handlers and middleware for the REST API
package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"storageaccounting/internal/models"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the context key for the authenticated user
	UserContextKey contextKey = "user"
)

// AuthMiddleware validates Telegram WebApp initData and extracts user information
// It implements Telegram's authentication specification using HMAC-SHA256 validation
type AuthMiddleware struct {
	botToken string
}

// NewAuthMiddleware creates a new authentication middleware with the bot token
func NewAuthMiddleware(botToken string) *AuthMiddleware {
	return &AuthMiddleware{
		botToken: botToken,
	}
}

// Middleware returns an HTTP middleware function for authentication
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract initData from header
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			respondError(w, http.StatusUnauthorized, "missing authentication header")
			return
		}

		// Validate initData signature
		user, err := am.validateInitData(initData)
		if err != nil {
			respondError(w, http.StatusUnauthorized, fmt.Sprintf("authentication failed: %v", err))
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateInitData validates the Telegram WebApp initData according to Telegram's specification
// Reference: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func (am *AuthMiddleware) validateInitData(initData string) (*models.User, error) {
	// Parse initData as URL query parameters
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	// Extract hash from initData
	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("hash is missing from initData")
	}

	// Remove hash from values for validation
	values.Del("hash")

	// Create data-check-string: key=value pairs sorted alphabetically by key, joined with newlines
	var pairs []string
	for key := range values {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Compute HMAC-SHA256 signature
	// secret_key = HMAC-SHA256(bot_token, "WebAppData")
	secretKey := computeHMAC([]byte("WebAppData"), []byte(am.botToken))

	// signature = HMAC-SHA256(data_check_string, secret_key)
	expectedHash := computeHMAC([]byte(dataCheckString), secretKey)
	expectedHashHex := hex.EncodeToString(expectedHash)

	// Compare hashes in constant time to prevent timing attacks
	if !hmac.Equal([]byte(expectedHashHex), []byte(hash)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Check auth_date to prevent replay attacks (valid for 24 hours)
	authDate := values.Get("auth_date")
	if authDate == "" {
		return nil, fmt.Errorf("auth_date is missing")
	}

	authTimestamp, err := strconv.ParseInt(authDate, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date: %w", err)
	}

	authTime := time.Unix(authTimestamp, 0)
	if time.Since(authTime) > 24*time.Hour {
		return nil, fmt.Errorf("initData expired (older than 24 hours)")
	}

	// Extract user information from initData
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("user data is missing")
	}

	var telegramUser struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string `json:"username"`
		PhotoURL     string `json:"photo_url"`
		LanguageCode string `json:"language_code"`
	}

	if err := json.Unmarshal([]byte(userJSON), &telegramUser); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	// Create User model
	user := &models.User{
		TelegramUserID: telegramUser.ID,
		FirstName:      telegramUser.FirstName,
		LastName:       telegramUser.LastName,
		Username:       telegramUser.Username,
		PhotoURL:       telegramUser.PhotoURL,
		LanguageCode:   telegramUser.LanguageCode,
		LastAccessAt:   time.Now(),
	}

	return user, nil
}

// computeHMAC computes HMAC-SHA256 of data with the given key
func computeHMAC(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// GetUserFromContext extracts the authenticated user from the request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ErrorHandlerMiddleware wraps handlers to catch panics and return structured error responses
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				respondError(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests with method, path, status, and duration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// respondError sends a JSON error response with the given status code and message
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondJSON sends a JSON response with the given status code and data
func RespondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// parseJSON decodes JSON request body into the provided interface
func parseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
