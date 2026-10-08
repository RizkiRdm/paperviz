-- 007_saved_papers.sql — Saved papers feature (Chunk 2.2)

ALTER TABLE documents ADD COLUMN saved INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_documents_saved ON documents(user_id, saved);
