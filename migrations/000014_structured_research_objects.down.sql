-- Drop in dependency order: claims is referenced by claim_evidence, which
-- migration 015 owns, but drop it last here so this file is independently valid.

DROP TABLE IF EXISTS citations;
DROP TABLE IF EXISTS results;
DROP TABLE IF EXISTS methods;
DROP TABLE IF EXISTS paper_tables;
DROP TABLE IF EXISTS claims;
