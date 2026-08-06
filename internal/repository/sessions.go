package repository

import (
	"fmt"
)

// SessionRepo provides CRUD access to the sessions table.
type SessionRepo struct {
	db dbExecutor
}

func NewSessionRepo(db dbExecutor) *SessionRepo {
	return &SessionRepo{db: db}
}

// Session represents a server-side session.
type Session struct {
	Token     string
	UserID    string
	ExpiresAt int64
}

// Insert creates a new session row.
func (r *SessionRepo) Insert(s Session) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		s.Token, s.UserID, s.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Get retrieves a session by token. Returns sql.ErrNoRows if not found or expired.
func (r *SessionRepo) Get(token string) (*Session, error) {
	row := r.db.QueryRow(
		`SELECT token, user_id, expires_at FROM sessions WHERE token = ? AND expires_at > ?`,
		token, unixNow(),
	)
	var s Session
	err := row.Scan(&s.Token, &s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes a session by token.
func (r *SessionRepo) Delete(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpired removes all sessions past their expiry time.
func (r *SessionRepo) DeleteExpired() (int64, error) {
	res, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, unixNow())
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return res.RowsAffected()
}
