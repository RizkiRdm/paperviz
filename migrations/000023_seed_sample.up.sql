-- 023_seed_sample.sql — Seed the guest sample document
-- A real, complete document with user_id NULL so anonymous visitors can read
-- it via GET /api/documents/:id (DocumentAccessAllowed allows NULL owners).
-- created_at/last_accessed_at use 'now' so the 7-day expiry (refreshed by
-- TouchLastAccessed on every read) matches a freshly processed document.
-- The ID must match SAMPLE_DOCUMENT_ID in frontend/src/lib/sample.js and
-- must satisfy the result page's 16-char ID shape check.

INSERT INTO documents
    (id, created_at, last_accessed_at, status, source_type, reading_level, title, original_text, simplified_text, chart_extraction_degraded, user_id, saved, visibility)
VALUES
    ('SamplePaperCells',
     CAST(strftime('%s', 'now') AS INTEGER),
     CAST(strftime('%s', 'now') AS INTEGER),
     'complete',
     'pasted_text',
     'simplified',
     'A practical framework for evaluating retrieval-augmented generation systems',
     'Retrieval-augmented generation (RAG) systems answer questions by retrieving passages from a corpus and conditioning a language model on them. We evaluate such systems along three axes: retrieval quality, faithfulness, and answer relevance. Across four benchmarks, retrieval recall is the strongest predictor of downstream answer reliability: when the top retrieved passages cover the answer-bearing evidence, faithfulness exceeds 0.9; when recall drops below 0.6, models increasingly answer from parametric memory and hallucinate. We recommend evaluating retrieval separately from generation, because tuning the generator alone cannot recover evidence the retriever never surfaced.',
     'This paper asks: how do we know a retrieval-augmented generation (RAG) system is good? The authors split quality into three parts. First, retrieval quality: did the system find the right passages? Second, faithfulness: does the answer stick to what those passages say? Third, answer relevance: does the answer actually address the question? Their main finding is that finding the right passages matters most. When the retrieved passages contain the evidence needed for the answer, the system almost always answers faithfully (faithfulness above 0.9). When retrieval misses that evidence, the model fills the gap from memory and starts making things up. The practical lesson: measure and improve retrieval on its own, before tuning the language model, because a better generator cannot recover evidence that was never retrieved.',
     0, NULL, 0, 'private');

INSERT INTO claims
    (id, paper_id, claim_text, claim_type, confidence, source_page, source_text, created_at)
VALUES
    ('SampleClaimCells',
     'SamplePaperCells',
     'Retrieval quality changes the reliability of downstream answers.',
     'finding', 'high', 7,
     'Retrieval recall is the strongest predictor of downstream answer reliability.',
     CAST(strftime('%s', 'now') AS INTEGER));

INSERT INTO evidence
    (id, paper_id, page, section, source_text, source_reference)
VALUES
    ('SampleEvidenceCell',
     'SamplePaperCells', 7, 'Results',
     'When the top retrieved passages cover the answer-bearing evidence, faithfulness exceeds 0.9; when recall drops below 0.6, models increasingly answer from parametric memory and hallucinate.',
     'Results · p. 7');

INSERT INTO claim_evidence
    (id, claim_id, evidence_id, relationship_type, created_at)
VALUES
    ('SampleClaimEvCell',
     'SampleClaimCells', 'SampleEvidenceCell', 'supports',
     CAST(strftime('%s', 'now') AS INTEGER));
