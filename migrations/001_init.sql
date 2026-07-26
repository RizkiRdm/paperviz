-- 001_init.sql — PaperViz initial schema (ARCHITECTURE.md Section 3)

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    created_at INTEGER NOT NULL,
    last_accessed_at INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('processing', 'complete', 'failed', 'verification_failed')),
    source_type TEXT NOT NULL CHECK (source_type IN ('pdf', 'pasted_text')),
    reading_level TEXT NOT NULL CHECK (reading_level IN ('simplified', 'eli5')),
    original_text TEXT NOT NULL,
    simplified_text TEXT,
    error_message TEXT,
    chart_extraction_degraded INTEGER NOT NULL DEFAULT 0 CHECK (chart_extraction_degraded IN (0, 1))
);

CREATE INDEX idx_documents_last_accessed_at ON documents(last_accessed_at);

CREATE TABLE charts (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    source_method TEXT NOT NULL CHECK (source_method IN ('data_extracted', 'image_fallback', 'omitted')),
    chart_data TEXT,
    image_blob BLOB,
    annotation TEXT,
    page_number INTEGER,
    display_order INTEGER NOT NULL
);

CREATE INDEX idx_charts_document_id ON charts(document_id);

CREATE TABLE claim_diffs (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    original_claims TEXT NOT NULL,
    simplified_claims TEXT NOT NULL,
    mismatch_detected INTEGER NOT NULL CHECK (mismatch_detected IN (0, 1)),
    mismatch_detail TEXT
);

CREATE INDEX idx_claim_diffs_document_id ON claim_diffs(document_id);
