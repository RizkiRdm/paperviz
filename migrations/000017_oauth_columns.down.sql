DROP INDEX IF EXISTS idx_users_oauth;
ALTER TABLE users DROP COLUMN oauth_provider;
ALTER TABLE users DROP COLUMN oauth_id;
