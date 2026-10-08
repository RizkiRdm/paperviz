-- documents.user_id must go before users, and its index must go before the column.

DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS idx_documents_user_id;
ALTER TABLE documents DROP COLUMN user_id;
DROP TABLE IF EXISTS users;
