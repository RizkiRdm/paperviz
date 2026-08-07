-- 003_chapters.sql — persist chapter structure from DetectChapters()

CREATE TABLE chapters (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    excerpt TEXT NOT NULL,
    display_order INTEGER NOT NULL
);

CREATE INDEX idx_chapters_document_id ON chapters(document_id);
