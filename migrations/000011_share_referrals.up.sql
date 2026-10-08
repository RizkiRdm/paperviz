-- 011_share_referrals.sql — Share→visit→analysis funnel counters

-- Visit/conversion counters for the referral loop, keyed by each entity's share token.
ALTER TABLE documents ADD COLUMN share_visits INTEGER NOT NULL DEFAULT 0;
ALTER TABLE documents ADD COLUMN share_conversions INTEGER NOT NULL DEFAULT 0;
ALTER TABLE charts ADD COLUMN share_visits INTEGER NOT NULL DEFAULT 0;
ALTER TABLE charts ADD COLUMN share_conversions INTEGER NOT NULL DEFAULT 0;
