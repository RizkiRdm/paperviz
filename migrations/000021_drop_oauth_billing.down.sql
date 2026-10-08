-- Restores the pre-021 shape. Reverses the api_key rename first so the column
-- name matches the index created by migration 018.

DROP INDEX IF EXISTS idx_users_api_key_hash;
ALTER TABLE users RENAME COLUMN api_key_hash TO api_key;
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

ALTER TABLE users ADD COLUMN oauth_provider TEXT;
ALTER TABLE users ADD COLUMN oauth_id TEXT;
CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id);

ALTER TABLE users ADD COLUMN stripe_customer_id TEXT;
ALTER TABLE users ADD COLUMN subscription_tier TEXT DEFAULT 'free';
ALTER TABLE users ADD COLUMN subscription_status TEXT DEFAULT 'active';
CREATE INDEX IF NOT EXISTS idx_users_stripe_customer ON users(stripe_customer_id);
