-- 020_user_credentials.sql — BYOK credential storage.
--
-- Additive only. The OAuth, Stripe, and api_key columns are dropped in
-- 021_drop_oauth_billing.sql, which must land in the same commit as the code
-- that stops selecting them. Dropping them here would leave apikey.go,
-- users.go, and account.go querying columns that no longer exist.
--
-- PaperViz holds no AI vendor credential of its own. Users supply their own
-- model key (BYOK), so this table is the highest-value data in the database:
-- a stolen dump is inert without CREDENTIAL_ENCRYPTION_KEY, which lives in the
-- process environment and is never written here.

CREATE TABLE IF NOT EXISTS user_credentials (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,
    model      TEXT NOT NULL,
    -- AES-256-GCM output. AAD binds the blob to
    -- "user_id|provider|model", so a row cannot be moved between accounts.
    ciphertext BLOB NOT NULL,
    nonce      BLOB NOT NULL,
    -- Last 4 characters, for display. Never a prefix, never the full key.
    key_hint   TEXT NOT NULL,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- At most one default credential per user, enforced by the database rather
-- than by application code, so a race cannot produce two defaults.
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_credentials_one_default
    ON user_credentials(user_id) WHERE is_default = 1;

CREATE INDEX IF NOT EXISTS idx_user_credentials_user
    ON user_credentials(user_id);
