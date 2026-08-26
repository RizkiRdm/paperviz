-- Chunk 6.1: Usage Measurement
-- Adds processing_time_ms to documents for pipeline timing.
-- Creates analytics_events for comparison and share-referral tracking.

ALTER TABLE documents ADD COLUMN processing_time_ms INTEGER;

CREATE TABLE IF NOT EXISTS analytics_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    entity_id TEXT,
    metadata TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created ON analytics_events(created_at);
