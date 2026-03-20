// Package repository handles data access layer for the inventory management system
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"storageaccounting/internal/models"
)

// ItemRepository handles database operations for inventory items
type ItemRepository struct {
	db *DB
}

// NewItemRepository creates a new item repository
func NewItemRepository(db *DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// ListAll retrieves all active (non-deleted) items with pagination support
// Returns items ordered by creation time (newest first)
func (r *ItemRepository) ListAll(ctx context.Context, limit, offset int) ([]*models.Item, error) {
	query := `
		SELECT id, name, current_quantity, photo_file_id, photo_s3_key,
		       created_at, updated_at, deleted_at, created_by, version
		FROM items
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := &models.Item{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.CurrentQuantity,
			&item.PhotoFileID,
			&item.PhotoS3Key,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.CreatedBy,
			&item.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

// GetByID retrieves a single item by its ID
// Returns error if item not found or is deleted
func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	query := `
		SELECT id, name, current_quantity, photo_file_id, photo_s3_key,
		       created_at, updated_at, deleted_at, created_by, version
		FROM items
		WHERE id = $1 AND deleted_at IS NULL
	`

	item := &models.Item{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.CurrentQuantity,
		&item.PhotoFileID,
		&item.PhotoS3Key,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.Version,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return item, nil
}

// Create inserts a new item into the database
func (r *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (name, current_quantity, photo_file_id, photo_s3_key, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at, version
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		item.Name,
		item.CurrentQuantity,
		item.PhotoFileID,
		item.PhotoS3Key,
		item.CreatedBy,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt, &item.Version)

	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// UpdateQuantityWithLock updates an item's quantity using optimistic locking
// Returns error if version mismatch (concurrent modification detected)
func (r *ItemRepository) UpdateQuantityWithLock(ctx context.Context, id uuid.UUID, newQuantity int, expectedVersion int) error {
	query := `
		UPDATE items
		SET current_quantity = $1, version = version + 1
		WHERE id = $2 AND version = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Pool.Exec(ctx, query, newQuantity, id, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("optimistic lock failure: item was modified by another user")
	}

	return nil
}

// SoftDelete marks an item as deleted without removing it from the database
func (r *ItemRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE items
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("item not found or already deleted")
	}

	return nil
}

// Count returns the total number of active (non-deleted) items
func (r *ItemRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM items WHERE deleted_at IS NULL`

	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}

	return count, nil
}
