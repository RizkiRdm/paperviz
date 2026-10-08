DROP INDEX IF EXISTS idx_charts_share_token;
ALTER TABLE charts DROP COLUMN share_token;
ALTER TABLE documents DROP COLUMN visibility;
