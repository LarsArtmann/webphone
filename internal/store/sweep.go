package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SweepResult reports one retention pass: what was deleted and which
// blob files the CALLER must unlink (the store layer only owns rows).
type SweepResult struct {
	Messages  int64
	Faxes     int64
	Threads   int64
	BlobPaths []string
}

// Sweep deletes stored content older than the cutoff: messages (their
// attachment rows cascade), fax jobs, and the threads left empty by
// those deletions. Attachment and document blob paths are collected
// BEFORE the rows go and returned for unlinking — a crash between the
// SQL and the unlink leaves an orphaned blob (harmless disk, cleaned
// by the next sweep of the same rows... which no longer exist, so
// orphans stay until an operator prunes the blob tree; bounded and
// documented, never a correctness issue).
func Sweep(ctx context.Context, db *sql.DB, cutoff time.Time) (SweepResult, error) {
	var res SweepResult

	// Blob paths first: once the rows are gone the paths are unknown.
	paths, err := listRows(ctx, db, "collect expirable blob paths", `
		SELECT path FROM attachments WHERE message_id IN (
			SELECT id FROM messages WHERE created_at < ?
		)
		UNION ALL
		SELECT document_path FROM fax_jobs WHERE created_at < ? AND document_path != ''
	`, []any{cutoff.Unix(), cutoff.Unix()}, func(row rowScanner) (string, error) {
		var path string
		if err := row.Scan(&path); err != nil {
			return "", fmt.Errorf("scan blob path: %w", err)
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

	faxes, err := execRows(ctx, db, "delete old fax jobs", `
		DELETE FROM fax_jobs WHERE created_at < ?
	`, cutoff.Unix())
	if err != nil {
		return res, err
	}
	res.Faxes = faxes

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
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	n, _ := res.RowsAffected() //nolint:erraudit // count is informational
	return n, nil
}

// Contacts are deliberately out of scope: they are the user's
// address book, not transient content — retention never touches them.

// Counts are the AGGREGATE table sizes the metrics surface reports
// (plan T26a): totals across ALL extensions, never per-owner values.
type Counts struct {
	Threads   int64
	Messages  int64
	Faxes     int64
	Contacts  int64
	Sessions  int64
}

// ReadCounts reads the aggregate sizes in one call.
func ReadCounts(ctx context.Context, db *sql.DB) (Counts, error) {
	var c Counts
	err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM threads),
			(SELECT COUNT(*) FROM messages),
			(SELECT COUNT(*) FROM fax_jobs),
			(SELECT COUNT(*) FROM contacts),
			(SELECT COUNT(*) FROM sessions)
	`).Scan(&c.Threads, &c.Messages, &c.Faxes, &c.Contacts, &c.Sessions)
	if err != nil {
		return c, fmt.Errorf("count aggregates: %w", err)
	}
	return c, nil
}
