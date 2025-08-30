-- Remove the added columns and constraints
ALTER TABLE wallets
DROP CONSTRAINT IF EXISTS wallets_id_unique;

DROP INDEX IF EXISTS idx_wallets_id;

DROP INDEX IF EXISTS idx_wallets_status;

ALTER TABLE wallets
DROP COLUMN IF EXISTS transactions;

ALTER TABLE wallets
DROP COLUMN IF EXISTS status;

ALTER TABLE wallets
DROP COLUMN IF EXISTS id;