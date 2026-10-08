-- 018_api_key_column.sql
-- Add API key column to users table

ALTER TABLE users ADD COLUMN api_key TEXT;

-- Index for API key lookup
CREATE INDEX idx_users_api_key ON users(api_key);
