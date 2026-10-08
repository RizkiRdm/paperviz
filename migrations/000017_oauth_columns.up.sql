-- 017_oauth_columns.sql
-- Add OAuth support columns to users table for Google login

ALTER TABLE users ADD COLUMN oauth_provider TEXT;
ALTER TABLE users ADD COLUMN oauth_id TEXT;

-- Index for OAuth lookup (find user by provider + oauth_id)
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);

-- Existing password-based accounts remain unaffected (nullable columns)
