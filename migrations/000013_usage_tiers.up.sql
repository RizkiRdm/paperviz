-- Chunk 6.2: Usage Limits
-- Adds user_tiers table for tier-based usage tracking.
-- Tracks papers_used per fingerprint with monthly reset.

CREATE TABLE IF NOT EXISTS user_tiers (
    fingerprint TEXT PRIMARY KEY,
    tier TEXT NOT NULL DEFAULT 'free',
    papers_used INTEGER NOT NULL DEFAULT 0,
    last_reset INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_tiers_tier ON user_tiers(tier);
