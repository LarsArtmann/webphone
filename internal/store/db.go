// Package store persists webphone data in SQLite (modernc.org/sqlite,
// pure Go — no cgo, cross-builds cleanly under Nix).
package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// Open opens (creating if needed) the database at path and applies the
// schema. Path may be ":memory:" for tests.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	// modernc sqlite is happiest with one writer connection; the app is
	// single-process and low-traffic, so serialize everything.
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrate(ctx, db); err != nil {
		_ = db.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
		return nil, fmt.Errorf("migrate %s: %w", path, err)
	}

	slog.Info("store ready", "path", path)

	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS threads (
			id         TEXT PRIMARY KEY,
			owner      TEXT NOT NULL,
			remote     TEXT NOT NULL,
			last_activity_at INTEGER NOT NULL,
			unread     INTEGER NOT NULL DEFAULT 0,
			UNIQUE(owner, remote)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id           TEXT PRIMARY KEY,
			thread_id    TEXT NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
			owner        TEXT NOT NULL,
			remote       TEXT NOT NULL,
			direction    TEXT NOT NULL,
			channel      TEXT NOT NULL,
			body         TEXT NOT NULL DEFAULT '',
			status       TEXT NOT NULL DEFAULT '',
			provider_ref TEXT NOT NULL DEFAULT '',
			failure_kind   TEXT NOT NULL DEFAULT '',
			failure_detail TEXT NOT NULL DEFAULT '',
			created_at   INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_thread ON messages(thread_id, created_at)`,
		// Parity with idx_fax_provider_ref: a provider ref identifies
		// exactly one outbound message, so replayed status webhooks can
		// never double-apply. Empty refs (inbound, queued) are exempt.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_provider_ref ON messages(provider_ref) WHERE provider_ref != ''`,
		`CREATE TABLE IF NOT EXISTS attachments (
			id         TEXT PRIMARY KEY,
			message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
			name       TEXT NOT NULL,
			mime_type  TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			path       TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS fax_jobs (
			id            TEXT PRIMARY KEY,
			owner         TEXT NOT NULL,
			remote        TEXT NOT NULL,
			direction     TEXT NOT NULL,
			status        TEXT NOT NULL,
			pages         INTEGER NOT NULL DEFAULT 0,
			document_path TEXT NOT NULL DEFAULT '',
			provider_ref  TEXT NOT NULL DEFAULT '',
			error         TEXT NOT NULL DEFAULT '',
			created_at    INTEGER NOT NULL,
			updated_at    INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fax_owner ON fax_jobs(owner, created_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_fax_provider_ref ON fax_jobs(provider_ref) WHERE provider_ref != ''`,
		`CREATE TABLE IF NOT EXISTS contacts (
			id         TEXT PRIMARY KEY,
			owner      TEXT NOT NULL,
			name       TEXT NOT NULL,
			phone      TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			UNIQUE(owner, phone)
		)`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply %q: %w", firstLine(stmt), err)
		}
	}
	// Additive column migrations: CREATE IF NOT EXISTS never extends an
	// EXISTING table, so each late-added column needs an ALTER — which
	// SQLite has no IF NOT EXISTS form for. "duplicate column name" is
	// the already-applied signal; anything else is a real failure.
	alters := []string{
		`ALTER TABLE messages ADD COLUMN failure_kind TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE messages ADD COLUMN failure_detail TEXT NOT NULL DEFAULT ''`,
	}
	for _, stmt := range alters {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("apply %q: %w", firstLine(stmt), err)
			}
		}
	}
	return nil
}

// rowScanner is the shared Scan surface of *sql.Row and *sql.Rows the
// scan helpers read through.
type rowScanner interface{ Scan(dest ...any) error }

// listRows runs a list query and scans every row. It is the one home
// for the result-set lifecycle (close on all paths) and the shared
// error shape: query and iteration failures both wrap op.
func listRows[T any](
	ctx context.Context, db *sql.DB, op, query string, args []any,
	scan func(rowScanner) (T, error),
) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]T, 0)
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s rows: %w", op, err)
	}
	return out, nil
}

// updatedOrNotFound is the write-side twin of listRows' error shape: a
// status-advance UPDATE that matched no row (already gone, or an
// owner/ref mismatch) is the caller-visible ErrNotFound miss, not a
// storage failure. RowsAffected's own error carries nothing actionable
// for a completed modernc sqlite Exec.
func updatedOrNotFound(res sql.Result) error {
	if rows, _ := res.RowsAffected(); rows == 0 { //nolint:erraudit // only the count matters on a completed Exec
		return ErrNotFound
	}
	return nil
}

func firstLine(s string) string {
	for i := range s {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}
