-- 016_annotations.sql — User annotations on papers and claims

CREATE TABLE IF NOT EXISTS annotations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    document_id TEXT NOT NULL,
    target_type TEXT NOT NULL CHECK(target_type IN ('paper', 'claim')),
    target_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_annotations_document ON annotations(document_id);
CREATE INDEX IF NOT EXISTS idx_annotations_user ON annotations(user_id);
