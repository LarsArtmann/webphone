// Package store persists webphone data in SQLite (modernc.org/sqlite,
// pure Go — no cgo, cross-builds cleanly under Nix).
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// Sentinel families (family-adoption train, 2026-09-30): absence and
// input caps are user-facing outcomes, so both sentinels classify
// Rejection — classifyForUser and any future handler can branch on the
// family without another ladder. Registration follows the library's
// documented pattern for stable error values; the sentinels keep their
// errors.New identity.
func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrNotFound: errorfamily.Rejection,
		ErrListFull: errorfamily.Rejection,
	})
}

// Open opens (creating if needed) the database at path and applies the
// schema. Path may be ":memory:" for tests.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "store.open", "open sqlite %s", path)
	}
	// modernc sqlite is happiest with one writer connection; the app is
	// single-process and low-traffic, so serialize everything.
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrate(ctx, db); err != nil {
		_ = db.Close()                                      //nolint:erraudit // close-after-use: nothing left to do on failure
		return nil, fmt.Errorf("migrate %s: %w", path, err) //nolint:erraudit // family-neutral propagation: the inner error owns the family
	}

	slog.Info("store ready", "path", path)

	return db, nil
}

// Database adapts the SQLite handle to the container's health and
// lifecycle contracts: HealthCheck pings the live connection,
// Shutdown closes it exactly once on the container's shutdown cascade.
// The service constructors keep taking *sql.DB (tests compose the raw
// handle); the composition root wraps via OpenDatabase and exposes the
// handle back through SQL. Structural conformance to the container's
// HealthcheckerWithContext / ShutdownerWithError interfaces is asserted
// in internal/app — service packages stay framework-free.
type Database struct {
	db     *sql.DB
	closed bool
}

// OpenDatabase opens (creating if needed) the database at path and
// wraps it in the container-adapted Database. Path may be ":memory:"
// for tests.
func OpenDatabase(path string) (*Database, error) {
	db, err := Open(path)
	if err != nil {
		return nil, err //nolint:erraudit // family-neutral propagation: Open owns the classification
	}
	return &Database{db: db}, nil
}

// SQL returns the underlying handle for the service constructors and
// the /healthz readiness wiring.
func (d *Database) SQL() *sql.DB { return d.db }

// HealthCheck pings the database: a closed or corrupt handle, or a
// lost database file, fails here.
func (d *Database) HealthCheck(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.ping", "ping sqlite")
	}
	return nil
}

// Shutdown closes the database. Closing an already-closed Database is
// inert — the shutdown cascade must stay idempotent.
func (d *Database) Shutdown() error {
	if d.closed {
		return nil
	}
	d.closed = true
	if err := d.db.Close(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.close", "close sqlite")
	}
	return nil
}

// schemaVersion is the version migrate() brings a database to. Bump it
// when appending a migration step; never renumber or edit shipped steps.
const schemaVersion = 3

// migration is one versioned schema step. Statements run inside a single
// transaction: a failure aborts at this version and a retry resumes there
// (SQLite DDL is transactional). Only the v1 baseline tolerates "duplicate
// column name" — it must converge legacy databases that already carry
// parts of the schema from the pre-versioning era (the ad-hoc
// failure-column ALTERs of 2026-09-24). Later steps are strict: the
// version row guarantees they run exactly once per database.
type migration struct {
	version  int
	tolerant bool
	stmts    []string
}

// migrations is the ordered chain from an empty database to
// schemaVersion. Fresh and legacy databases walk the SAME chain — v1
// recreates the historical shape, later steps evolve it — so there is
// exactly one shape-history and no drift between fresh and migrated
// databases.
var migrations = []migration{
	{
		version:  1,
		tolerant: true,
		stmts: []string{
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
			// Additive column migrations (2026-09-24): CREATE IF NOT EXISTS
			// never extends an EXISTING table, so each late-added column
			// needs an ALTER — duplicate-tolerant here because legacy
			// databases already applied them.
			`ALTER TABLE messages ADD COLUMN failure_kind TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE messages ADD COLUMN failure_detail TEXT NOT NULL DEFAULT ''`,
		},
	},
	{
		version: 2,
		stmts: []string{
			// Thread organization flags (M22): three independent axes,
			// 0/1 like unread. Archived threads leave the default list,
			// pinned sort first, muted suppress attention surfaces only.
			`ALTER TABLE threads ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE threads ADD COLUMN archived INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE threads ADD COLUMN muted INTEGER NOT NULL DEFAULT 0`,
			// Reply snippets (M21): per-extension canned texts, quick ones
			// double as composer chips.
			`CREATE TABLE IF NOT EXISTS snippets (
			id         TEXT PRIMARY KEY,
			owner      TEXT NOT NULL,
			body       TEXT NOT NULL,
			quick      INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL
		)`,
			`CREATE INDEX IF NOT EXISTS idx_snippets_owner ON snippets(owner, created_at)`,
		},
	},
	{
		version: 3,
		stmts: []string{
			// Live-call transcripts (2026-10-07): one row per transcribed
			// segment. Append-only; the per-owner cap in the transcripts
			// store bounds growth (the blob sweep never touches these —
			// they are plain rows, not content).
			`CREATE TABLE IF NOT EXISTS call_transcripts (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			owner      TEXT NOT NULL,
			call_id    TEXT NOT NULL,
			direction  TEXT NOT NULL,
			remote     TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			text       TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
			`CREATE INDEX IF NOT EXISTS idx_call_transcripts_owner ON call_transcripts(owner, started_at)`,
			`CREATE INDEX IF NOT EXISTS idx_call_transcripts_call ON call_transcripts(owner, call_id)`,
		},
	},
}

// migrate brings the schema to schemaVersion through the ordered
// migration chain. A database without a version row (legacy or fresh)
// starts at 0 and walks every step; the single-row schema_version table
// is itself created idempotently first so the runner can always read.
func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.migrate", "create schema_version")
	}
	var current int
	if err := db.QueryRowContext(ctx, `SELECT version FROM schema_version LIMIT 1`).Scan(&current); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errorfamily.WrapInfrastructuref(err, "store.migrate", "read schema_version")
		}
		// No row: a database from before versioning (or brand new) —
		// the v1 baseline converges both.
		current = 0
	}
	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if err := applyMigration(ctx, db, m); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration runs one migration step atomically and stamps its
// version. The stamp is delete-then-insert inside the same transaction:
// the table is single-row by construction and has no key to upsert on.
func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.migrate", "begin v%d", m.version)
	}
	defer func() { _ = tx.Rollback() }() //nolint:erraudit // best-effort rollback; the commit path owns the outcome
	for _, stmt := range m.stmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			if m.tolerant && strings.Contains(err.Error(), "duplicate column name") {
				continue // already applied pre-versioning
			}
			return errorfamily.WrapInfrastructuref(err, "store.migrate", "v%d: apply %q", m.version, firstLine(stmt))
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_version`); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.migrate", "v%d: clear version", m.version)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_version (version) VALUES (?)`, m.version); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.migrate", "v%d: stamp version", m.version)
	}
	if err := tx.Commit(); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.migrate", "v%d: commit", m.version)
	}
	slog.Info("store migrated", "version", m.version)
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
		return nil, errorfamily.WrapInfrastructuref(err, "store.query", "%s", op)
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
		return nil, errorfamily.WrapInfrastructuref(err, "store.query", "%s rows", op)
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
