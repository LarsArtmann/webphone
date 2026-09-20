// SQLite-backed session store: sessions survive service restarts, so a
// deploy no longer signs every tab out (the "Tab session ended" toast
// becomes the rare path it was always meant to be). Design + threat
// model: docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/larsartmann/webphone/internal/domain"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// SQLiteStore persists sessions in the same webphone.db the rest of the
// app uses. TTL semantics match MemStore exactly: absolute ExpiresAt
// (no idle extension), expiry enforced on Get, sweep on Create.
type SQLiteStore struct {
	db  *sql.DB
	ttl time.Duration
}

const sessionsSchema = `CREATE TABLE IF NOT EXISTS sessions (
	token      TEXT PRIMARY KEY,
	extension  TEXT NOT NULL,
	password   TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
)`

const sessionsExpiryIndex = `CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expires_at)`

// Stamps are unix MILLIS: second-truncation would make sub-second TTLs
// untestable and blur expiry at the boundary.

// NewSQLiteStore applies the schema (idempotent) and returns the store.
// An empty table on first boot means existing users re-sign-in once —
// the honest migration, nothing is resurrected.
func NewSQLiteStore(db *sql.DB, ttl time.Duration) (*SQLiteStore, error) {
	if db == nil {
		return nil, fmt.Errorf("session store: nil db")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("session store: ttl must be positive, got %s", ttl)
	}
	for _, stmt := range []string{sessionsSchema, sessionsExpiryIndex} {
		if _, err := db.Exec(stmt); err != nil {
			return nil, fmt.Errorf("migrate sessions: %w", err)
		}
	}
	return &SQLiteStore{db: db, ttl: ttl}, nil
}

// Create mints a session row and returns its token. Expired rows are
// swept alongside, so restarts across long downtimes cannot accumulate
// corpses.
func (s *SQLiteStore) Create(extension domain.Extension, password string) (string, error) {
	token, err := mintToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	if _, err := s.db.Exec(
		`DELETE FROM sessions WHERE expires_at < ?`, now.UnixMilli(),
	); err != nil {
		return "", fmt.Errorf("sweep sessions: %w", err)
	}
	if _, err := s.db.Exec(
		`INSERT INTO sessions (token, extension, password, created_at, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		token, extension.String(), password, now.UnixMilli(), now.Add(s.ttl).UnixMilli(),
	); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

// Get returns a live session by token; an expired row is deleted on
// sight (a session that died while the server was down must not
// resurrect on the first read after boot).
func (s *SQLiteStore) Get(token string) (Session, bool) {
	var (
		extension, password string
		created, expires    int64
	)
	err := s.db.QueryRow(
		`SELECT extension, password, created_at, expires_at
		 FROM sessions WHERE token = ?`, token,
	).Scan(&extension, &password, &created, &expires)
	if err != nil {
		return Session{}, false
	}
	sess := Session{
		Extension: domain.MustParseExtension(extension),
		Password:  password,
		CreatedAt: time.UnixMilli(created),
		ExpiresAt: time.UnixMilli(expires),
	}
	if !time.Now().Before(sess.ExpiresAt) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token) //nolint:erraudit // best-effort hygiene; the read verdict above already returned dead
		return Session{}, false
	}
	return sess, true
}

// Delete drops a session (logout).
func (s *SQLiteStore) Delete(token string) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token) //nolint:erraudit // logout stays best-effort, parity with MemStore
}
