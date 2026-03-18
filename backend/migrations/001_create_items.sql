-- Migration: Create items table
-- Purpose: Store inventory items with photo references and optimistic locking
-- Created: 2026-03-17

-- Create items table
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    current_quantity INTEGER NOT NULL DEFAULT 0,
    photo_file_id VARCHAR(255),  -- Telegram File API file_id (primary storage)
    photo_s3_key VARCHAR(500),   -- S3 object key (fallback/archival storage)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,  -- Soft delete support
    created_by BIGINT NOT NULL,  -- Telegram user ID who created this item
    version INTEGER NOT NULL DEFAULT 1  -- Optimistic locking version
);

-- Create indexes for common queries
CREATE INDEX idx_items_deleted_at ON items(deleted_at) WHERE deleted_at IS NULL;  -- Filter active items efficiently
CREATE INDEX idx_items_name ON items(name) WHERE deleted_at IS NULL;  -- Search by name
CREATE INDEX idx_items_created_at ON items(created_at DESC) WHERE deleted_at IS NULL;  -- Sort by creation time

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_updated_at_trigger
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION update_items_updated_at();

-- Add comments for documentation
COMMENT ON TABLE items IS 'Inventory items with photo references and quantity tracking';
COMMENT ON COLUMN items.id IS 'Unique identifier (UUID)';
COMMENT ON COLUMN items.name IS 'Item display name';
COMMENT ON COLUMN items.current_quantity IS 'Current stock quantity (can be negative)';
COMMENT ON COLUMN items.photo_file_id IS 'Telegram File API file_id for photo retrieval';
COMMENT ON COLUMN items.photo_s3_key IS 'S3 object key for fallback photo storage';
COMMENT ON COLUMN items.created_at IS 'Item creation timestamp';
COMMENT ON COLUMN items.updated_at IS 'Last modification timestamp (auto-updated)';
COMMENT ON COLUMN items.deleted_at IS 'Soft delete timestamp (NULL = active)';
COMMENT ON COLUMN items.created_by IS 'Telegram user ID of creator';
COMMENT ON COLUMN items.version IS 'Optimistic locking version (incremented on each update)';
