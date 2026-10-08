-- 019_billing_columns.sql
-- Add Stripe billing columns to users table

ALTER TABLE users ADD COLUMN stripe_customer_id TEXT;
ALTER TABLE users ADD COLUMN subscription_tier TEXT DEFAULT 'free';
ALTER TABLE users ADD COLUMN subscription_status TEXT DEFAULT 'active';

-- Index for Stripe customer lookup
CREATE INDEX idx_users_stripe_customer ON users(stripe_customer_id);
