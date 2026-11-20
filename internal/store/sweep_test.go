package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// The retention sweep's behavior pins, seeded with raw SQL against the
// FROZEN legacy tables (the Messages type that used to seed them was
// deleted 2026-11-20 with the frozen-ladder cleanup — the tables stay
// until the §5 DROP migration, and the sweep must keep proving it
// empties them).
func TestSweepDeletesOnlyExpiredContent(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	now := time.Now()
	old := now.Add(-48 * time.Hour)

	seedLegacyThread(t, db, "t-old-full", "1001", "+441632960961", old)
	seedLegacyMessage(t, db, "m-old", "t-old-full", "1001", "old one", "old-ref", old)
	if _, err := db.ExecContext(ctx,
		`INSERT INTO attachments (id, message_id, name, mime_type, size_bytes, path)
		 VALUES ('a-old', 'm-old', 'old.png', 'image/png', 4, 'old-blob')`); err != nil {
		t.Fatal(err)
	}
	seedLegacyThread(t, db, "t-live", "1001", "+441632960977", now)
	seedLegacyMessage(t, db, "m-live", "t-live", "1001", "keep me", "", now)

	res, err := Sweep(ctx, db, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res.Messages != 1 || res.Threads != 1 {
		t.Fatalf("sweep counts wrong: %+v", res)
	}
	if len(res.BlobPaths) != 1 || res.BlobPaths[0] != "old-blob" {
		t.Fatalf("expected the old attachment path, got %v", res.BlobPaths)
	}

	// The live thread and its message survive; the swept provider ref
	// is gone (idempotency guard: a second sweep is a no-op).
	if n := countRows(t, db, `SELECT COUNT(*) FROM messages WHERE id = 'm-live'`); n != 1 {
		t.Fatalf("live message lost: %d", n)
	}
	if n := countRows(t, db, `SELECT COUNT(*) FROM messages WHERE provider_ref = 'old-ref'`); n != 0 {
		t.Fatal("old provider ref must not resolve after the sweep")
	}
	if res2, err := Sweep(ctx, db, now.Add(-24*time.Hour)); err != nil || res2.Messages != 0 || res2.Threads != 0 {
		t.Fatalf("second sweep must be a no-op: %+v %v", res2, err)
	}

	// Empty OLD threads go; empty RECENT threads stay (the thread row
	// is the conversation handle, not content).
	seedLegacyThread(t, db, "t-old", "1001", "+441632960999", old)
	seedLegacyThread(t, db, "t-new", "1001", "+441632960988", now)
	res3, err := Sweep(ctx, db, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res3.Threads != 1 {
		t.Fatalf("only the old empty thread must go: %+v", res3)
	}
	if n := countRows(t, db, `SELECT COUNT(*) FROM threads WHERE id = 't-new'`); n != 1 {
		t.Fatalf("recent empty thread lost: %d", n)
	}
	if n := countRows(t, db, `SELECT COUNT(*) FROM threads WHERE id = 't-old'`); n != 0 {
		t.Fatalf("old empty thread survived: %d", n)
	}
}

// seedLegacyThread inserts one legacy threads row (the v5 ladder's
// final shape; network defaults to 'sms' and the flag columns to 0).
func seedLegacyThread(t *testing.T, db *sql.DB, id, owner, remote string, lastActivity time.Time) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO threads (id, owner, remote, last_activity_at, unread) VALUES (?,?,?, ?,0)`,
		id, owner, remote, lastActivity.Unix()); err != nil {
		t.Fatalf("seed thread %s: %v", id, err)
	}
}

// seedLegacyMessage inserts one legacy messages row; direction/channel
// carry their production literals, which the sweep never reads (it
// keys on created_at, thread_id and the attachment path).
func seedLegacyMessage(t *testing.T, db *sql.DB, id, threadID, owner, body, providerRef string, createdAt time.Time) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO messages (id, thread_id, owner, remote, direction, channel, body, status, provider_ref, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, threadID, owner, remoteOfThread(t, db, threadID), "outbound", "sms", body, "sent", providerRef, createdAt.Unix()); err != nil {
		t.Fatalf("seed message %s: %v", id, err)
	}
}

func remoteOfThread(t *testing.T, db *sql.DB, threadID string) string {
	t.Helper()
	var remote string
	if err := db.QueryRowContext(context.Background(),
		`SELECT remote FROM threads WHERE id = ?`, threadID).Scan(&remote); err != nil {
		t.Fatalf("thread %s remote: %v", threadID, err)
	}
	return remote
}

func countRows(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
