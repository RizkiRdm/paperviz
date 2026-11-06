-- 023_seed_sample.down.sql — Remove the seeded guest sample (children first,
-- then the document; evidence/claims would also cascade).

DELETE FROM claim_evidence WHERE id = 'SampleClaimEvCell';
DELETE FROM evidence WHERE id = 'SampleEvidenceCell';
DELETE FROM claims WHERE id = 'SampleClaimCells';
DELETE FROM documents WHERE id = 'SamplePaperCells';
