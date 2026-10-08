DROP INDEX IF EXISTS idx_users_stripe_customer;
ALTER TABLE users DROP COLUMN stripe_customer_id;
ALTER TABLE users DROP COLUMN subscription_tier;
ALTER TABLE users DROP COLUMN subscription_status;
