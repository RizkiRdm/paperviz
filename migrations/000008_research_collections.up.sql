-- 008_research_collections.sql — Research Collections (Chunk 2.3)

CREATE TABLE collections (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_collections_user_id ON collections(user_id);

CREATE TABLE document_collections (
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, document_id)
);

CREATE INDEX idx_document_collections_document ON document_collections(document_id);
