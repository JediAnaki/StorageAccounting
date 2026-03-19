// Package models contains domain entities for the inventory management system
package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents an inventory item with photo references and quantity tracking
type Item struct {
	// ID is the unique identifier for this item (UUID)
	ID uuid.UUID `json:"id"`

	// Name is the display name of the item
	Name string `json:"name"`

	// CurrentQuantity is the current stock quantity (can be negative)
	CurrentQuantity int `json:"current_quantity"`

	// PhotoFileID is the Telegram File API file_id for retrieving the photo
	// This is the primary photo storage mechanism
	PhotoFileID *string `json:"photo_file_id,omitempty"`

	// PhotoS3Key is the S3 object key for fallback/archival photo storage
	PhotoS3Key *string `json:"photo_s3_key,omitempty"`

	// CreatedAt is the timestamp when this item was first added
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp of the last modification (auto-updated by database trigger)
	UpdatedAt time.Time `json:"updated_at"`

	// DeletedAt is the soft delete timestamp (NULL = active, non-NULL = deleted)
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// CreatedBy is the Telegram user ID of the user who created this item
	CreatedBy int64 `json:"created_by"`

	// Version is the optimistic locking version (incremented on each update)
	// Used to detect concurrent modifications and prevent lost updates
	Version int `json:"version"`
}

// IsDeleted returns true if the item has been soft-deleted
func (i *Item) IsDeleted() bool {
	return i.DeletedAt != nil
}

// IsOutOfStock returns true if the current quantity is zero or negative
func (i *Item) IsOutOfStock() bool {
	return i.CurrentQuantity <= 0
}

// HasPhoto returns true if the item has at least one photo reference
func (i *Item) HasPhoto() bool {
	return i.PhotoFileID != nil || i.PhotoS3Key != nil
}
