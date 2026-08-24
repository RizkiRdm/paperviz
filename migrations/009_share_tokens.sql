-- 009_share_tokens.sql — Share tokens for individual figure explanations

-- Share token for public figure share pages. NULL = not shared yet (lazy generation).
ALTER TABLE charts ADD COLUMN share_token TEXT;
CREATE UNIQUE INDEX idx_charts_share_token ON charts(share_token) WHERE share_token IS NOT NULL;

-- Document visibility controls who can access share pages.
-- 'private' = only owner (default), 'unlisted' = link-only, 'public' = discoverable.
ALTER TABLE documents ADD COLUMN visibility TEXT DEFAULT 'private' CHECK (visibility IN ('public', 'unlisted', 'private'));
