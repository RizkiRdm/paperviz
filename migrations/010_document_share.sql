-- 010_document_share.sql — Document-level share tokens for full-paper share pages

-- Share token for public paper share pages. NULL = not shared yet (lazy generation).
ALTER TABLE documents ADD COLUMN share_token TEXT;
CREATE UNIQUE INDEX idx_documents_share_token ON documents(share_token) WHERE share_token IS NOT NULL;
