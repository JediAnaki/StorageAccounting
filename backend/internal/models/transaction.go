// Package models contains domain entities for the inventory management system
package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents an immutable audit log entry for inventory quantity changes
// Transactions are append-only and cannot be modified or deleted to maintain audit integrity
type Transaction struct {
	// ID is the unique identifier for this transaction (UUID)
	ID uuid.UUID `json:"id"`

	// ItemID is the reference to the inventory item that was modified
	ItemID uuid.UUID `json:"item_id"`

	// UserID is the Telegram user ID who performed this transaction
	UserID int64 `json:"user_id"`

	// Timestamp is when this transaction occurred (immutable)
	Timestamp time.Time `json:"timestamp"`

	// PreviousQuantity is the item quantity before this transaction
	PreviousQuantity int `json:"previous_quantity"`

	// NewQuantity is the item quantity after this transaction
	NewQuantity int `json:"new_quantity"`

	// Delta is the change amount (positive for additions, negative for reductions)
	// Delta = NewQuantity - PreviousQuantity
	Delta int `json:"delta"`

	// Note is an optional user-provided explanation for this quantity change
	Note *string `json:"note,omitempty"`
}

// IsAddition returns true if this transaction represents a stock addition (positive delta)
func (t *Transaction) IsAddition() bool {
	return t.Delta > 0
}

// IsReduction returns true if this transaction represents a stock reduction (negative delta)
func (t *Transaction) IsReduction() bool {
	return t.Delta < 0
}

// IsNeutral returns true if this transaction has no quantity change (delta = 0)
// This can happen for administrative corrections or notes
func (t *Transaction) IsNeutral() bool {
	return t.Delta == 0
}

// AbsDelta returns the absolute value of the delta (magnitude of change)
func (t *Transaction) AbsDelta() int {
	if t.Delta < 0 {
		return -t.Delta
	}
	return t.Delta
}
