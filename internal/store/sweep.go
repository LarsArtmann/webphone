package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/go-error-family"
)

// SweepResult reports one retention pass: what was deleted and which
// blob files the CALLER must unlink (the store layer only owns rows).
// branching-flow COMPOSITION_mixin (baselined): retention tallies vs
// store Counts — coincidental int64 field shape, different domains.
type SweepResult struct {
	Messages  int64
	Threads   int64
	BlobPaths []string
}

// Sweep deletes stored content older than the cutoff: messages (their
// attachment rows cascade) and the threads left empty by those
// deletions. Fax jobs and call transcripts ride the ENGINE sweep
// (readmodels Faxes/Transcripts SweepBefore) since the T11/T13 swaps;
// the caller composes all three.
// Attachment blob paths are collected BEFORE the rows go
// and returned for unlinking — a crash between the SQL and the unlink
// leaves an orphaned blob (harmless disk, cleaned by the next sweep of
// the same rows... which no longer exist, so orphans stay until an
// operator prunes the blob tree; bounded and documented, never a
// correctness issue).
func Sweep(ctx context.Context, db *sql.DB, cutoff time.Time) (SweepResult, error) {
	var res SweepResult

	// Blob paths first: once the rows are gone the paths are unknown.
	paths, err := listRows(ctx, db, "collect expirable blob paths", `
		SELECT path FROM attachments WHERE message_id IN (
			SELECT id FROM messages WHERE created_at < ?
		)
	`, []any{cutoff.Unix()}, func(row rowScanner) (string, error) {
		var path string
		if err := row.Scan(&path); err != nil {
			return "", errorfamily.WrapInfrastructuref(err, "store.sweep_scan", "scan blob path")
		}
		return path, nil
	})
	if err != nil {
		return res, err
	}
	res.BlobPaths = paths

	msgs, err := execRows(ctx, db, "delete old messages", `
		DELETE FROM messages WHERE created_at < ?
	`, cutoff.Unix())
	if err != nil {
		return res, err
	}
	res.Messages = msgs

	threads, err := execRows(ctx, db, "delete empty threads", `
		DELETE FROM threads WHERE last_activity_at < ? AND NOT EXISTS (
			SELECT 1 FROM messages WHERE messages.thread_id = threads.id
		)
	`, cutoff.Unix())
	if err != nil {
		return res, err
	}
	res.Threads = threads

	return res, nil
}

func execRows(ctx context.Context, db *sql.DB, op, query string, args ...any) (int64, error) {
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "store.exec", "%s", op)
	}
	n, _ := res.RowsAffected() //nolint:erraudit // count is informational
	return n, nil
}

// Contacts are deliberately out of scope: they are the user's
// address book, not transient content — retention never touches them.

// Counts are the AGGREGATE table sizes the metrics surface reports
// (plan T26a): totals across ALL extensions, never per-owner values.
type Counts struct {
	Threads  int64
	Messages int64
	Sessions int64
}

// ReadCounts reads the aggregate sizes in one call. Contacts, faxes,
// and transcripts are counted by the read-model engine (readmodels
// Count) — their SQL tables are retired storage.
func ReadCounts(ctx context.Context, db *sql.DB) (Counts, error) {
	var c Counts
	err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM threads),
			(SELECT COUNT(*) FROM messages),
			(SELECT COUNT(*) FROM sessions)
	`).Scan(&c.Threads, &c.Messages, &c.Sessions)
	if err != nil {
		return c, errorfamily.WrapInfrastructuref(err, "store.counts", "count aggregates")
	}
	return c, nil
}
