-- Revert: Replace created_at and updated_at with timestamp in orders table
-- First revert UUID changes back to TEXT
ALTER TABLE orders ALTER COLUMN order_id TYPE TEXT USING order_id::TEXT;
ALTER TABLE orders ALTER COLUMN buyer_id TYPE TEXT USING buyer_id::TEXT;

-- Add back timestamp column
ALTER TABLE orders
ADD COLUMN timestamp TIMESTAMP NOT NULL DEFAULT NOW ();

-- Copy created_at data back to timestamp
UPDATE orders
SET
  timestamp = created_at;

-- Drop the new columns
ALTER TABLE orders
DROP COLUMN created_at;

ALTER TABLE orders
DROP COLUMN updated_at;