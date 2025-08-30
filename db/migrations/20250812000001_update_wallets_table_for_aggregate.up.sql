-- Add new columns to wallets table to support aggregate functionality
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS id VARCHAR(255);
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS status INTEGER NOT NULL DEFAULT 0;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS transactions JSONB NOT NULL DEFAULT '[]';

-- Create index on status for performance
CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets(status);

-- Create index on id if it's added
CREATE INDEX IF NOT EXISTS idx_wallets_id ON wallets(id) WHERE id IS NOT NULL;

-- Update existing records to have id and proper status
UPDATE wallets SET 
    id = 'wallet_' || user_id::text,
    status = 0
WHERE id IS NULL OR status IS NULL;

-- Make id NOT NULL after updating existing records
ALTER TABLE wallets ALTER COLUMN id SET NOT NULL;

-- Add unique constraint on id
ALTER TABLE wallets ADD CONSTRAINT wallets_id_unique UNIQUE (id);
