CREATE TABLE evidence (
    id TEXT PRIMARY KEY,
    paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page INTEGER,
    figure_id TEXT,
    table_id TEXT,
    section TEXT,
    source_text TEXT NOT NULL,
    source_reference TEXT NOT NULL
);

CREATE INDEX idx_evidence_paper_id ON evidence(paper_id);
