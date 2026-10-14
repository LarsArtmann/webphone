package readmodels

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testCollection = "webphone.idem"

func TestOpenCreatesDatabaseAndHealthCheckPasses(t *testing.T) {
	dir := t.TempDir()
	engine, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })

	if _, err := os.Stat(filepath.Join(dir, fileName)); err != nil {
		t.Fatalf("engine database missing: %v", err)
	}
	if err := engine.HealthCheck(context.Background()); err != nil {
		t.Fatalf("health check: %v", err)
	}
}

// TestDedupSurvivesRestart pins the defect fix the engine swap exists
// for: dedup memory must outlive the process, so a restart inside the
// TTL window no longer re-arms a seen webhook key.
func TestDedupSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	now := time.Now()

	first, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	dedup := first.DedupStore()
	seen, err := dedup.DedupCheckAndRecord(ctx, testCollection, "telnyx-ref-1", time.Hour, now)
	if err != nil {
		t.Fatalf("check-and-record: %v", err)
	}
	if seen {
		t.Fatal("first record reported already-seen")
	}
	if err := first.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	second, err := Open(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = second.Shutdown() })

	dedup = second.DedupStore()
	seen, err = dedup.DedupSeen(ctx, testCollection, "telnyx-ref-1", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("seen after restart: %v", err)
	}
	if !seen {
		t.Fatal("key forgotten across restart — restart durability broken")
	}
	seen, err = dedup.DedupCheckAndRecord(ctx, testCollection, "telnyx-ref-1", time.Hour, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("replay check-and-record: %v", err)
	}
	if !seen {
		t.Fatal("replay after restart not detected as duplicate")
	}
}

func TestDedupExpiryAndSweep(t *testing.T) {
	dir := t.TempDir()
	engine, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })

	ctx := context.Background()
	dedup := engine.DedupStore()
	now := time.Now()

	if _, err := dedup.DedupCheckAndRecord(ctx, testCollection, "expired-key", time.Minute, now); err != nil {
		t.Fatalf("record: %v", err)
	}

	// Past the window the key reads unseen and may be re-claimed…
	seen, err := dedup.DedupSeen(ctx, testCollection, "expired-key", now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("expired seen: %v", err)
	}
	if seen {
		t.Fatal("expired key read as seen")
	}

	// …and the sweep reclaims the expired row.
	removed, err := dedup.DedupSweep(ctx, testCollection, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if removed != 1 {
		t.Fatalf("sweep removed %d rows, want 1", removed)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	engine, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := engine.Shutdown(); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}
	if err := engine.Shutdown(); err != nil {
		t.Fatalf("second shutdown: %v", err)
	}
}
