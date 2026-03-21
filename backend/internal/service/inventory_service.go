// Package service contains business logic for the inventory management system
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"storageaccounting/internal/models"
	"storageaccounting/internal/repository"
)

// InventoryService handles business logic for inventory operations
type InventoryService struct {
	itemRepo        *repository.ItemRepository
	transactionRepo *repository.TransactionRepository
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *repository.DB) *InventoryService {
	return &InventoryService{
		itemRepo:        repository.NewItemRepository(db),
		transactionRepo: repository.NewTransactionRepository(db),
	}
}

// ItemWithStats represents an item with additional statistics
type ItemWithStats struct {
	*models.Item
	TransactionCount int `json:"transaction_count"`
}

// ItemDetails represents detailed item information including transaction history
type ItemDetails struct {
	Item         *models.Item           `json:"item"`
	Transactions []*models.Transaction  `json:"transactions"`
	TotalCount   int                    `json:"total_count"`
}

// ListItemsParams contains parameters for listing items
type ListItemsParams struct {
	Limit  int
	Offset int
}

// ListItems retrieves a paginated list of active items
// Returns items ordered by creation time (newest first)
func (s *InventoryService) ListItems(ctx context.Context, params ListItemsParams) ([]*models.Item, int, error) {
	// Validate pagination parameters
	if params.Limit <= 0 {
		params.Limit = 50 // Default limit
	}
	if params.Limit > 100 {
		params.Limit = 100 // Max limit to prevent excessive queries
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Get items from repository
	items, err := s.itemRepo.ListAll(ctx, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}

	// Get total count for pagination
	totalCount, err := s.itemRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	return items, totalCount, nil
}

// GetItemDetails retrieves detailed information about a specific item
// Includes item data and paginated transaction history
func (s *InventoryService) GetItemDetails(ctx context.Context, itemID uuid.UUID, transactionLimit, transactionOffset int) (*ItemDetails, error) {
	// Validate pagination parameters
	if transactionLimit <= 0 {
		transactionLimit = 50
	}
	if transactionLimit > 100 {
		transactionLimit = 100
	}
	if transactionOffset < 0 {
		transactionOffset = 0
	}

	// Get item
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Get transactions for this item
	transactions, err := s.transactionRepo.ListByItemID(ctx, itemID, transactionLimit, transactionOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Get total transaction count
	totalCount, err := s.transactionRepo.CountByItemID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	return &ItemDetails{
		Item:         item,
		Transactions: transactions,
		TotalCount:   totalCount,
	}, nil
}

// AddItem creates a new inventory item with an initial quantity
// Records the creation as an initial transaction
func (s *InventoryService) AddItem(ctx context.Context, name string, initialQuantity int, photoFileID, photoS3Key *string, userID int64) (*models.Item, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("item name cannot be empty")
	}
	if initialQuantity < 0 {
		return nil, fmt.Errorf("initial quantity cannot be negative")
	}

	// Create item
	item := &models.Item{
		Name:            name,
		CurrentQuantity: initialQuantity,
		PhotoFileID:     photoFileID,
		PhotoS3Key:      photoS3Key,
		CreatedBy:       userID,
	}

	err := s.itemRepo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Create initial transaction record if initial quantity > 0
	if initialQuantity > 0 {
		initialNote := "Initial stock"
		transaction := &models.Transaction{
			ItemID:           item.ID,
			UserID:           userID,
			PreviousQuantity: 0,
			NewQuantity:      initialQuantity,
			Delta:            initialQuantity,
			Note:             &initialNote,
		}

		err = s.transactionRepo.Create(ctx, transaction)
		if err != nil {
			// Item was created but transaction failed - log this but don't fail the operation
			// In production, you might want to use a distributed transaction or saga pattern
			return item, fmt.Errorf("item created but failed to record initial transaction: %w", err)
		}
	}

	return item, nil
}

// RecordAddition records a quantity addition to an item
// Creates a transaction record with positive delta
func (s *InventoryService) RecordAddition(ctx context.Context, itemID uuid.UUID, quantity int, note *string, userID int64) error {
	if quantity <= 0 {
		return fmt.Errorf("addition quantity must be positive")
	}

	// Get current item state
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	previousQuantity := item.CurrentQuantity
	newQuantity := previousQuantity + quantity

	// Update item quantity with optimistic locking
	err = s.itemRepo.UpdateQuantityWithLock(ctx, itemID, newQuantity, item.Version)
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		ItemID:           itemID,
		UserID:           userID,
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		Delta:            quantity,
		Note:             note,
	}

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		// Quantity was updated but transaction failed - this is a serious consistency issue
		// In production, wrap this in a database transaction
		return fmt.Errorf("quantity updated but failed to record transaction: %w", err)
	}

	return nil
}

// RecordReduction records a quantity reduction from an item
// Creates a transaction record with negative delta
// Allows negative inventory with a warning
func (s *InventoryService) RecordReduction(ctx context.Context, itemID uuid.UUID, quantity int, note *string, userID int64) error {
	if quantity <= 0 {
		return fmt.Errorf("reduction quantity must be positive")
	}

	// Get current item state
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	previousQuantity := item.CurrentQuantity
	newQuantity := previousQuantity - quantity

	// Update item quantity with optimistic locking
	err = s.itemRepo.UpdateQuantityWithLock(ctx, itemID, newQuantity, item.Version)
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		ItemID:           itemID,
		UserID:           userID,
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		Delta:            -quantity,
		Note:             note,
	}

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		// Quantity was updated but transaction failed - this is a serious consistency issue
		return fmt.Errorf("quantity updated but failed to record transaction: %w", err)
	}

	return nil
}

// TransactionHistoryParams contains parameters for filtering transaction history
type TransactionHistoryParams struct {
	ItemID    uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	TxnType   string // "addition", "reduction", or "" for all
	Limit     int
	Offset    int
}

// TransactionHistoryResult contains filtered transaction history with metadata
type TransactionHistoryResult struct {
	Transactions []*models.Transaction `json:"transactions"`
	TotalCount   int                   `json:"total_count"`
	Filtered     bool                  `json:"filtered"` // True if any filters were applied
}

// GetTransactionHistory retrieves transaction history with optional filtering
// Supports filtering by date range, transaction type, and pagination
func (s *InventoryService) GetTransactionHistory(ctx context.Context, params TransactionHistoryParams) (*TransactionHistoryResult, error) {
	// Validate pagination parameters
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// Check if filters are applied
	hasFilters := params.StartDate != nil || params.EndDate != nil || params.TxnType != ""

	// Build filter for repository
	filter := repository.TransactionFilter{
		ItemID:    params.ItemID,
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		TxnType:   params.TxnType,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	// Get transactions with filter
	transactions, err := s.transactionRepo.ListWithFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	// Get total count with same filters
	totalCount, err := s.transactionRepo.CountWithFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	return &TransactionHistoryResult{
		Transactions: transactions,
		TotalCount:   totalCount,
		Filtered:     hasFilters,
	}, nil
}
