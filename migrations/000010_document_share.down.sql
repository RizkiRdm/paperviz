DROP INDEX IF EXISTS idx_documents_share_token;
ALTER TABLE documents DROP COLUMN share_token;
