-- Migration: Create transactions table
-- Purpose: Immutable audit log for all inventory quantity changes
-- Created: 2026-03-17

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,  -- Prevent item deletion if transactions exist
    user_id BIGINT NOT NULL,  -- Telegram user ID who performed the transaction
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    previous_quantity INTEGER NOT NULL,  -- Quantity before this transaction
    new_quantity INTEGER NOT NULL,  -- Quantity after this transaction
    delta INTEGER NOT NULL,  -- Change amount (positive for additions, negative for reductions)
    note TEXT  -- Optional user-provided explanation for the change
);

-- Create indexes for common queries
CREATE INDEX idx_transactions_item_id ON transactions(item_id, timestamp DESC);  -- Get history for specific item (reverse chronological)
CREATE INDEX idx_transactions_user_id ON transactions(user_id, timestamp DESC);  -- Get user activity history
CREATE INDEX idx_transactions_timestamp ON transactions(timestamp DESC);  -- Global transaction history

-- Add constraint to ensure delta consistency
ALTER TABLE transactions
ADD CONSTRAINT check_delta_consistency
CHECK (delta = new_quantity - previous_quantity);

-- Add comments for documentation
COMMENT ON TABLE transactions IS 'Immutable audit log of all inventory quantity changes';
COMMENT ON COLUMN transactions.id IS 'Unique transaction identifier (UUID)';
COMMENT ON COLUMN transactions.item_id IS 'Reference to the inventory item (foreign key)';
COMMENT ON COLUMN transactions.user_id IS 'Telegram user ID who performed the transaction';
COMMENT ON COLUMN transactions.timestamp IS 'Transaction timestamp (immutable)';
COMMENT ON COLUMN transactions.previous_quantity IS 'Item quantity before this change';
COMMENT ON COLUMN transactions.new_quantity IS 'Item quantity after this change';
COMMENT ON COLUMN transactions.delta IS 'Change amount (+ for additions, - for reductions)';
COMMENT ON COLUMN transactions.note IS 'Optional explanation for the quantity change';

-- Prevent updates and deletes to maintain immutability
CREATE OR REPLACE FUNCTION prevent_transaction_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Transactions are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_transaction_update
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION prevent_transaction_modification();

CREATE TRIGGER prevent_transaction_delete
BEFORE DELETE ON transactions
FOR EACH ROW
EXECUTE FUNCTION prevent_transaction_modification();
