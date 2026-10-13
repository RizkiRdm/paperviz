-- 022_clerk_auth.sql — replace local sessions with Clerk identity.
--
-- Clerk owns authentication now, so a local session token is not a second
-- opinion that anyone should trust: it would be a credential path Clerk cannot
-- revoke. Dropping the table also removes an at-rest bearer-token store.
--
-- users stays as PaperViz's own profile store. documents.user_id, collections,
-- annotations and user_credentials all reference users(id), and those rows must
-- keep working across a Render restart that wipes SQLite.
--
-- users.id now holds the Clerk subject (user_2abc...), so no existing foreign
-- key needed rewriting.
--
-- password_hash stays NOT NULL and is written as '' for Clerk-provisioned
-- rows. The column could be dropped, but keeping it lets a future local
-- break-glass account exist without another table rebuild, and no code reads it
-- any more.

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;