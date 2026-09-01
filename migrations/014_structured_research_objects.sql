-- 014_structured_research_objects.sql — Structured extraction for research papers
-- Adds claims, paper_tables, methods, results, citations tables.

CREATE TABLE IF NOT EXISTS claims (
    id TEXT PRIMARY KEY,
    paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    claim_text TEXT NOT NULL,
    claim_type TEXT CHECK(claim_type IN ('hypothesis','finding','conclusion','method','limitation')),
    confidence TEXT CHECK(confidence IN ('high','medium','low')),
    source_page INTEGER,
    source_text TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_claims_paper_id ON claims(paper_id);

CREATE TABLE IF NOT EXISTS paper_tables (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_number INTEGER,
    caption TEXT,
    headers TEXT,
    rows TEXT,
    source_text TEXT,
    display_order INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_paper_tables_document_id ON paper_tables(document_id);

CREATE TABLE IF NOT EXISTS methods (
    id TEXT PRIMARY KEY,
    paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    method_name TEXT NOT NULL,
    description TEXT,
    method_type TEXT CHECK(method_type IN ('experimental','survey','qualitative','quantitative','computational','other')),
    source_page INTEGER,
    source_text TEXT
);

CREATE INDEX IF NOT EXISTS idx_methods_paper_id ON methods(paper_id);

CREATE TABLE IF NOT EXISTS results (
    id TEXT PRIMARY KEY,
    paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    result_text TEXT NOT NULL,
    result_type TEXT CHECK(result_type IN ('primary','secondary','negative')),
    supporting_evidence_id TEXT REFERENCES evidence(id) ON DELETE SET NULL,
    source_page INTEGER,
    source_text TEXT
);

CREATE INDEX IF NOT EXISTS idx_results_paper_id ON results(paper_id);

CREATE TABLE IF NOT EXISTS citations (
    id TEXT PRIMARY KEY,
    paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    cited_paper_id TEXT,
    authors TEXT,
    title TEXT,
    year INTEGER,
    venue TEXT,
    doi TEXT,
    url TEXT,
    source_page INTEGER
);

CREATE INDEX IF NOT EXISTS idx_citations_paper_id ON citations(paper_id);
