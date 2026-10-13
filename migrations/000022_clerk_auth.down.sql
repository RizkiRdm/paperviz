-- Restores the sessions table dropped by 022 so the migration chain can roll
-- back to 021. Existing rows are not repopulated: every session is invalidated
-- by the reversal, which is the correct outcome when restoring a credential
-- store that Clerk has since replaced.

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);