// Package api contains HTTP handlers and middleware for the REST API
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"storageaccounting/internal/models"
	"storageaccounting/internal/repository"
	"storageaccounting/internal/service"
)

// Server represents the HTTP server with database access and configuration
type Server struct {
	db               *repository.DB
	botToken         string
	inventoryService *service.InventoryService
	photoService     *service.PhotoService
}

// NewServer creates a new API server with the given database and configuration
func NewServer(db *repository.DB, botToken, s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket string) *Server {
	return &Server{
		db:               db,
		botToken:         botToken,
		inventoryService: service.NewInventoryService(db),
		photoService:     service.NewPhotoService(botToken, s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket),
	}
}

// RegisterRoutes registers all HTTP routes with middleware
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", s.handleHealth)

	// API routes (all require authentication)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/items", s.handleItems)
	apiMux.HandleFunc("/api/items/", s.handleItemByID)

	// Apply middleware chain: logging -> error handling -> authentication -> API routes
	authMiddleware := NewAuthMiddleware(s.botToken)
	handler := LoggingMiddleware(
		ErrorHandlerMiddleware(
			authMiddleware.Middleware(apiMux),
		),
	)

	// Combine public and protected routes
	mux.Handle("/api/", handler)

	return LoggingMiddleware(ErrorHandlerMiddleware(mux))
}

// handleHealth handles GET /health for health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check database connectivity
	if err := s.db.Ping(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// handleItems handles GET /api/items (list items)
func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// User is available from context after authentication middleware
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListItems(w, r, user)

	case http.MethodPost:
		s.handleCreateItem(w, r, user)
	}
}

// handleListItems handles GET /api/items
func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request, user *models.User) {
	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get items from service
	items, totalCount, err := s.inventoryService.ListItems(r.Context(), service.ListItemsParams{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	// Return response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       items,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// handleItemByID handles GET /api/items/:id (get item details)
// and PUT /api/items/:id/... (update operations)
func (s *Server) handleItemByID(w http.ResponseWriter, r *http.Request) {
	// User is available from context after authentication middleware
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Extract item ID from path: /api/items/{id} or /api/items/{id}/transactions
	path := strings.TrimPrefix(r.URL.Path, "/api/items/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		respondError(w, http.StatusBadRequest, "item ID required")
		return
	}

	itemID, err := uuid.Parse(parts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	// Route based on path and method
	if len(parts) == 1 {
		// /api/items/:id
		if r.Method == http.MethodGet {
			s.handleGetItemDetails(w, r, itemID, user)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	} else if len(parts) == 2 && parts[1] == "transactions" {
		// /api/items/:id/transactions
		if r.Method == http.MethodGet {
			s.handleGetItemTransactions(w, r, itemID, user)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	} else if len(parts) == 2 && parts[1] == "add" {
		// /api/items/:id/add
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			s.handleAddStock(w, r, itemID, user)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	} else if len(parts) == 2 && parts[1] == "remove" {
		// /api/items/:id/remove
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			s.handleRemoveStock(w, r, itemID, user)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	} else {
		respondError(w, http.StatusNotFound, "endpoint not found")
	}
}

// handleGetItemDetails handles GET /api/items/:id
func (s *Server) handleGetItemDetails(w http.ResponseWriter, r *http.Request, itemID uuid.UUID, user *models.User) {
	// Parse pagination for transactions
	transactionLimit := 50
	transactionOffset := 0

	if limitStr := r.URL.Query().Get("transaction_limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			transactionLimit = l
		}
	}

	if offsetStr := r.URL.Query().Get("transaction_offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			transactionOffset = o
		}
	}

	// Get item details from service
	details, err := s.inventoryService.GetItemDetails(r.Context(), itemID, transactionLimit, transactionOffset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "item not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to get item details")
		}
		return
	}

	RespondJSON(w, http.StatusOK, details)
}

// handleGetItemTransactions handles GET /api/items/:id/transactions
func (s *Server) handleGetItemTransactions(w http.ResponseWriter, r *http.Request, itemID uuid.UUID, user *models.User) {
	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse filter parameters
	var startDate, endDate *time.Time
	txnType := r.URL.Query().Get("type")

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}

	// Get transaction history with filters
	params := service.TransactionHistoryParams{
		ItemID:    itemID,
		StartDate: startDate,
		EndDate:   endDate,
		TxnType:   txnType,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := s.inventoryService.GetTransactionHistory(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "item not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to get transactions")
		}
		return
	}

	// Return transactions with filter metadata
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": result.Transactions,
		"total_count":  result.TotalCount,
		"filtered":     result.Filtered,
		"limit":        limit,
		"offset":       offset,
	})
}

// handleCreateItem handles POST /api/items
func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request, user *models.User) {
	// Parse multipart form (max 10MB for photo upload)
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	// Get form fields
	name := r.FormValue("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	initialQuantityStr := r.FormValue("initial_quantity")
	initialQuantity := 0
	if initialQuantityStr != "" {
		initialQuantity, err = strconv.Atoi(initialQuantityStr)
		if err != nil || initialQuantity < 0 {
			respondError(w, http.StatusBadRequest, "invalid initial_quantity")
			return
		}
	}

	// Get photo file (optional)
	var photoFileID, photoS3Key *string
	file, fileHeader, err := r.FormFile("photo")
	if err == nil {
		// Photo was provided
		defer file.Close()

		// Process and upload photo
		telegramFileID, s3Key, err := s.photoService.ProcessAndUpload(fileHeader)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("photo upload failed: %v", err))
			return
		}

		photoFileID = &telegramFileID
		if s3Key != "" {
			photoS3Key = &s3Key
		}
	} else if err != http.ErrMissingFile {
		// Error other than missing file
		respondError(w, http.StatusBadRequest, "failed to read photo")
		return
	}

	// Create item using service
	item, err := s.inventoryService.AddItem(r.Context(), name, initialQuantity, photoFileID, photoS3Key, user.TelegramUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create item: %v", err))
		return
	}

	// Return created item
	RespondJSON(w, http.StatusCreated, item)
}

// QuantityChangeRequest represents a request to add or remove stock
type QuantityChangeRequest struct {
	Quantity int     `json:"quantity"`
	Note     *string `json:"note"`
}

// handleAddStock handles PUT /api/items/:id/add
func (s *Server) handleAddStock(w http.ResponseWriter, r *http.Request, itemID uuid.UUID, user *models.User) {
	var req QuantityChangeRequest

	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	// Record addition using service
	err := s.inventoryService.RecordAddition(r.Context(), itemID, req.Quantity, req.Note, user.TelegramUserID)
	if err != nil {
		// Check for optimistic locking error
		if strings.Contains(err.Error(), "optimistic lock failure") {
			respondError(w, http.StatusConflict, "item was modified by another user, please refresh and try again")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "item not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to add stock: %v", err))
		return
	}

	// Get updated item details
	details, err := s.inventoryService.GetItemDetails(r.Context(), itemID, 1, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "stock added but failed to retrieve updated item")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"item":    details.Item,
		"message": "stock added successfully",
	})
}

// handleRemoveStock handles PUT /api/items/:id/remove
func (s *Server) handleRemoveStock(w http.ResponseWriter, r *http.Request, itemID uuid.UUID, user *models.User) {
	var req QuantityChangeRequest

	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	// Record reduction using service
	err := s.inventoryService.RecordReduction(r.Context(), itemID, req.Quantity, req.Note, user.TelegramUserID)
	if err != nil {
		// Check for optimistic locking error
		if strings.Contains(err.Error(), "optimistic lock failure") {
			respondError(w, http.StatusConflict, "item was modified by another user, please refresh and try again")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "item not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to remove stock: %v", err))
		return
	}

	// Get updated item details
	details, err := s.inventoryService.GetItemDetails(r.Context(), itemID, 1, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "stock removed but failed to retrieve updated item")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"item":    details.Item,
		"message": "stock removed successfully",
	})
}
