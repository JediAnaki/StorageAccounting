// Package repository handles data access layer for the inventory management system
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"storageaccounting/internal/models"
)

// TransactionRepository handles database operations for inventory transactions
type TransactionRepository struct {
	db *DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// ListByItemID retrieves all transactions for a specific item
// Returns transactions in reverse chronological order (newest first)
// Supports pagination with limit and offset
func (r *TransactionRepository) ListByItemID(ctx context.Context, itemID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, item_id, user_id, timestamp, previous_quantity, new_quantity, delta, note
		FROM transactions
		WHERE item_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, itemID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		txn := &models.Transaction{}
		err := rows.Scan(
			&txn.ID,
			&txn.ItemID,
			&txn.UserID,
			&txn.Timestamp,
			&txn.PreviousQuantity,
			&txn.NewQuantity,
			&txn.Delta,
			&txn.Note,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// ListByUserID retrieves all transactions performed by a specific user
// Returns transactions in reverse chronological order (newest first)
func (r *TransactionRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, item_id, user_id, timestamp, previous_quantity, new_quantity, delta, note
		FROM transactions
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		txn := &models.Transaction{}
		err := rows.Scan(
			&txn.ID,
			&txn.ItemID,
			&txn.UserID,
			&txn.Timestamp,
			&txn.PreviousQuantity,
			&txn.NewQuantity,
			&txn.Delta,
			&txn.Note,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// Create inserts a new transaction into the database
// Transactions are immutable - this is the only write operation allowed
func (r *TransactionRepository) Create(ctx context.Context, txn *models.Transaction) error {
	query := `
		INSERT INTO transactions (item_id, user_id, previous_quantity, new_quantity, delta, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, timestamp
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		txn.ItemID,
		txn.UserID,
		txn.PreviousQuantity,
		txn.NewQuantity,
		txn.Delta,
		txn.Note,
	).Scan(&txn.ID, &txn.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// CountByItemID returns the total number of transactions for a specific item
func (r *TransactionRepository) CountByItemID(ctx context.Context, itemID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE item_id = $1`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, itemID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}

// TransactionFilter represents filtering options for transaction queries
type TransactionFilter struct {
	ItemID     uuid.UUID
	StartDate  *time.Time // Filter transactions >= start date
	EndDate    *time.Time // Filter transactions <= end date
	TxnType    string     // "addition", "reduction", or "" for all
	Limit      int
	Offset     int
}

// ListWithFilter retrieves transactions with optional filtering
// Supports filtering by date range and transaction type (addition/reduction)
func (r *TransactionRepository) ListWithFilter(ctx context.Context, filter TransactionFilter) ([]*models.Transaction, error) {
	// Build dynamic query with filters
	query := `
		SELECT id, item_id, user_id, timestamp, previous_quantity, new_quantity, delta, note
		FROM transactions
		WHERE item_id = $1
	`

	args := []interface{}{filter.ItemID}
	paramCount := 1

	// Add date range filters
	if filter.StartDate != nil {
		paramCount++
		query += fmt.Sprintf(" AND timestamp >= $%d", paramCount)
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		paramCount++
		query += fmt.Sprintf(" AND timestamp <= $%d", paramCount)
		args = append(args, *filter.EndDate)
	}

	// Add transaction type filter
	if filter.TxnType != "" {
		paramCount++
		if strings.ToLower(filter.TxnType) == "addition" {
			query += fmt.Sprintf(" AND delta > 0")
		} else if strings.ToLower(filter.TxnType) == "reduction" {
			query += fmt.Sprintf(" AND delta < 0")
		}
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC"

	paramCount++
	query += fmt.Sprintf(" LIMIT $%d", paramCount)
	args = append(args, filter.Limit)

	paramCount++
	query += fmt.Sprintf(" OFFSET $%d", paramCount)
	args = append(args, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions with filter: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		txn := &models.Transaction{}
		err := rows.Scan(
			&txn.ID,
			&txn.ItemID,
			&txn.UserID,
			&txn.Timestamp,
			&txn.PreviousQuantity,
			&txn.NewQuantity,
			&txn.Delta,
			&txn.Note,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// CountWithFilter returns the total number of transactions matching the filter
func (r *TransactionRepository) CountWithFilter(ctx context.Context, filter TransactionFilter) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE item_id = $1`

	args := []interface{}{filter.ItemID}
	paramCount := 1

	// Add date range filters
	if filter.StartDate != nil {
		paramCount++
		query += fmt.Sprintf(" AND timestamp >= $%d", paramCount)
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		paramCount++
		query += fmt.Sprintf(" AND timestamp <= $%d", paramCount)
		args = append(args, *filter.EndDate)
	}

	// Add transaction type filter
	if filter.TxnType != "" {
		if strings.ToLower(filter.TxnType) == "addition" {
			query += " AND delta > 0"
		} else if strings.ToLower(filter.TxnType) == "reduction" {
			query += " AND delta < 0"
		}
	}

	var count int
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count filtered transactions: %w", err)
	}

	return count, nil
}
