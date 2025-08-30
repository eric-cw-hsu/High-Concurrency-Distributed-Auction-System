-- Replace timestamp column with created_at and updated_at in orders table
ALTER TABLE orders
ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW ();

ALTER TABLE orders
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT NOW ();

-- Copy existing timestamp data to created_at
UPDATE orders
SET
  created_at = timestamp,
  updated_at = timestamp;

-- Drop the old timestamp column
ALTER TABLE orders
DROP COLUMN timestamp;

-- Change order_id from TEXT to UUID
ALTER TABLE orders ALTER COLUMN order_id TYPE UUID USING order_id::UUID;

-- Change buyer_id from TEXT to UUID  
ALTER TABLE orders ALTER COLUMN buyer_id TYPE UUID USING buyer_id::UUID;