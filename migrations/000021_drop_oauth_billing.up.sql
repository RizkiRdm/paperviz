-- 021_drop_oauth_billing.sql — remove the third-party account surface.
--
-- Lands with the commit that deletes the code selecting these columns, so the
-- server is never querying a column that no longer exists.
--
-- PaperViz holds no Google account and no Stripe account. Auth is email and
-- password; model access is BYOK, stored in user_credentials (migration 020).
--
-- Consequence of removing oauth_provider/oauth_id: any account created through
-- Google had password_hash = 'oauth-only' and can no longer log in. Acceptable
-- because the app has never been launched and has no users to strand.

DROP INDEX IF EXISTS idx_users_oauth;
DROP INDEX IF EXISTS idx_users_stripe_customer;

ALTER TABLE users DROP COLUMN oauth_provider;
ALTER TABLE users DROP COLUMN oauth_id;
ALTER TABLE users DROP COLUMN stripe_customer_id;
ALTER TABLE users DROP COLUMN subscription_tier;
ALTER TABLE users DROP COLUMN subscription_status;

-- users.api_key holds a SHA-256 digest, not a key. Rename it so the schema
-- cannot be misread as recoverable credential material.
DROP INDEX IF EXISTS idx_users_api_key;
ALTER TABLE users RENAME COLUMN api_key TO api_key_hash;
CREATE INDEX IF NOT EXISTS idx_users_api_key_hash ON users(api_key_hash);
