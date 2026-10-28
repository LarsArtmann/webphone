// SQLite-backed session store: sessions survive service restarts, so a
// deploy no longer signs every tab out (the "Tab session ended" toast
// becomes the rare path it was always meant to be). Design + threat
// model: docs/planning/2026-09-20_17-41_session-persistence-spike-verdict.md.
package session

import (
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// SQLiteStore persists sessions in the same webphone.db the rest of the
// app uses. TTL semantics match MemStore exactly: absolute ExpiresAt
// (no idle extension), expiry enforced on Get, sweep on Create.
type SQLiteStore struct {
	db  *sql.DB
	ttl time.Duration
	now func() time.Time
}

const sessionsSchema = `CREATE TABLE IF NOT EXISTS sessions (
	token      TEXT PRIMARY KEY,
	extension  TEXT NOT NULL,
	password   TEXT NOT NULL,
	login_email TEXT NOT NULL DEFAULT '',
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
)`

// sessionsLoginEmailColumn upgrades pre-identity tables (the multi-user
// substrate, 2026-10-28): existing rows keep the empty default, which
// reads as "classic extension login" — exactly the actor they had. The
// sessions table lives outside the store/db.go migration ladder (this
// package owns its schema with IF NOT EXISTS DDL at open), so the column
// upgrade follows the same self-contained, idempotent pattern: a guarded
// ALTER whose "duplicate column" failure is the already-migrated verdict.
const sessionsLoginEmailColumn = `ALTER TABLE sessions ADD COLUMN login_email TEXT NOT NULL DEFAULT ''`

const sessionsExpiryIndex = `CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expires_at)`

// Stamps are unix MILLIS: second-truncation would make sub-second TTLs
// untestable and blur expiry at the boundary.

// NewSQLiteStore applies the schema (idempotent) and returns the store.
// An empty table on first boot means existing users re-sign-in once —
// the honest migration, nothing is resurrected.
func NewSQLiteStore(db *sql.DB, ttl time.Duration) (*SQLiteStore, error) {
	if db == nil {
		return nil, errorfamily.NewInfrastructure("session.nil_db", "session store: nil db")
	}
	if ttl <= 0 {
		return nil, errorfamily.Newf(errorfamily.Infrastructure, "session.ttl", "session store: ttl must be positive, got %s", ttl)
	}
	for _, stmt := range []string{sessionsSchema, sessionsExpiryIndex} {
		if _, err := db.Exec(stmt); err != nil {
			return nil, errorfamily.WrapInfrastructuref(err, "session.migrate", "migrate sessions")
		}
	}
	if _, err := db.Exec(sessionsLoginEmailColumn); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return nil, errorfamily.WrapInfrastructuref(err, "session.migrate_login_email", "add sessions.login_email")
		}
	}
	return &SQLiteStore{db: db, ttl: ttl, now: time.Now}, nil
}

// Create mints a session row and returns its token. Expired rows are
// swept alongside, so restarts across long downtimes cannot accumulate
// corpses.
func (s *SQLiteStore) Create(extension domain.Extension, password, loginEmail string) (Token, error) {
	token, sess, err := makeSession(extension, password, loginEmail, s.ttl, s.now)
	if err != nil {
		return Token{}, err
	}

	now := sess.CreatedAt
	if _, err := s.db.Exec(
		`DELETE FROM sessions WHERE expires_at < ?`, now.UnixMilli(),
	); err != nil {
		return Token{}, errorfamily.WrapInfrastructuref(err, "session.sweep", "sweep sessions")
	}
	if _, err := s.db.Exec(
		`INSERT INTO sessions (token, extension, password, login_email, created_at, expires_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
		token.String(), extension.String(), password, sess.LoginEmail, now.UnixMilli(), sess.ExpiresAt.UnixMilli(),
	); err != nil {
		return Token{}, errorfamily.WrapInfrastructuref(err, "session.insert", "insert session")
	}
	return token, nil
}

// Get returns a live session by token; an expired row is deleted on
// sight (a session that died while the server was down must not
// resurrect on the first read after boot).
func (s *SQLiteStore) Get(token Token) (Session, bool) {
	var (
		extension, password, loginEmail string
		created, expires                int64
	)
	err := s.db.QueryRow(
		`SELECT extension, password, login_email, created_at, expires_at
			 FROM sessions WHERE token = ?`, token.String(),
	).Scan(&extension, &password, &loginEmail, &created, &expires)
	if err != nil {
		// Auth stays fail-closed on ANY error, but a driver failure
		// (locked/busy/degraded box) is NOT "unknown token": swallow it
		// and the 401 becomes an undiagnosable mystery (the 2026-10-14
		// TestUploadBodyIsBounded flake class). Log the distinction.
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Warn("session: lookup failed", "error", err)
		}
		return Session{}, false
	}
	sess := Session{
		Extension:  domain.MustParseExtension(extension),
		Password:   password,
		LoginEmail: loginEmail,
		CreatedAt:  time.UnixMilli(created),
		ExpiresAt:  time.UnixMilli(expires),
	}
	if !s.now().Before(sess.ExpiresAt) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token.String()) //nolint:erraudit // best-effort hygiene; the read verdict above already returned dead
		return Session{}, false
	}
	return sess, true
}

// Delete drops a session (logout).
func (s *SQLiteStore) Delete(token Token) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token.String()) //nolint:erraudit // logout stays best-effort, parity with MemStore
}

// Renew applies the sliding-renewal policy: Get enforces expiry (lazy
// delete of a row that died unseen), renewDue decides, the UPDATE
// persists. A lost update race at worst keeps the older expiry — the
// next past-half-life request renews again.
func (s *SQLiteStore) Renew(token Token, idle, maxAge time.Duration) (Session, bool) {
	sess, ok := s.Get(token)
	if !ok {
		return Session{}, false
	}
	extended, due := renewDue(sess, idle, maxAge, s.now())
	if !due {
		return Session{}, false
	}
	if _, err := s.db.Exec(
		`UPDATE sessions SET expires_at = ? WHERE token = ?`,
		extended.UnixMilli(), token.String(),
	); err != nil {
		return Session{}, false
	}
	sess.ExpiresAt = extended
	return sess, true
}

// ActiveTokens lists the live tokens of one extension; expired rows are
// dead on sight (the Get contract) and never listed.
func (s *SQLiteStore) ActiveTokens(extension domain.Extension) []Token {
	rows, err := s.db.Query(
		`SELECT token FROM sessions WHERE extension = ? AND expires_at > ?`,
		extension.String(), s.now().UnixMilli(),
	)
	if err != nil {
		return nil
	}
	defer rows.Close() //nolint:errcheck // read-side close on defer; rows.Err owns the verdict
	var out []Token
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil
		}
		out = append(out, ParseToken(token)) //nolint:branching-flow // the raw credential blob by design: at-rest accepted, swept on read (sessions ruling)
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}
