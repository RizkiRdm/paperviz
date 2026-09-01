-- 015_evidence_graph.sql — Evidence graph for cross-paper claim-evidence relationships

CREATE TABLE IF NOT EXISTS claim_evidence (
    id TEXT PRIMARY KEY,
    claim_id TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    evidence_id TEXT NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL CHECK(relationship_type IN ('supports','contradicts','clarifies')),
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_claim_evidence_claim_id ON claim_evidence(claim_id);
CREATE INDEX IF NOT EXISTS idx_claim_evidence_evidence_id ON claim_evidence(evidence_id);

CREATE TABLE IF NOT EXISTS paper_relationships (
    id TEXT PRIMARY KEY,
    source_paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    target_paper_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL CHECK(relationship_type IN ('supporting','contradicting','citing','similar_methodology','different_findings')),
    evidence_text TEXT,
    created_at INTEGER NOT NULL,
    UNIQUE(source_paper_id, target_paper_id, relationship_type)
);

CREATE INDEX IF NOT EXISTS idx_paper_relationships_source ON paper_relationships(source_paper_id);
CREATE INDEX IF NOT EXISTS idx_paper_relationships_target ON paper_relationships(target_paper_id);
