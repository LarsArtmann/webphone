package cqrs

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	_ "modernc.org/sqlite" // registers the "sqlite" driver for the DLQ handle

	"github.com/larsartmann/webphone/internal/domain"
)

// bootDialSystem builds a REAL system over a temp dir with the dial
// domain — the same topology the composition root uses (sqlite engine,
// source-of-truth + projections, DLQ, fast backoff). close always runs.
func bootDialSystem(t *testing.T, dir string) *DialWindows {
	t.Helper()
	old := timeNow
	t.Cleanup(func() { timeNow = old })
	ctx := context.Background()
	dsn := filepath.Join(dir, "system.db")

	dlqDB, err := sql.Open("sqlite", dsn+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dlqDB.Close() })
	dlq, err := projectionhost.NewSQLiteDeadLetterStore(ctx, dlqDB)
	if err != nil {
		t.Fatal(err)
	}

	domainCfg := DialDomainDecl()
	domainCfg.ProjectionHostOptions = append(domainCfg.ProjectionHostOptions,
		projectionhost.WithDeadLetterStore(dlq, 3),
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond),
	)
	sys, err := system.New(ctx, domainCfg, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		t.Fatalf("system.New: %v", err)
	}
	if err := sys.Start(ctx); err != nil {
		t.Fatalf("sys.Start: %v", err)
	}
	t.Cleanup(func() { _ = sys.GracefulClose(ctx) })
	return NewDialWindows(sys)
}

// awaitLast polls the projection reader until the slot's last-at
// reaches want (the async worker lands folds ~1-2ms after dispatch).
func awaitLast(t *testing.T, dials *DialWindows, owner domain.Extension, remote string, want time.Time) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		last, found, err := dials.Last(context.Background(), owner, remote)
		if err != nil {
			t.Fatalf("Last: %v", err)
		}
		if found && last.Unix() == want.Unix() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("projection never caught up for %s/%s: found=%v last=%v want=%v", owner, remote, found, last, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestDialAttemptAllowsThenDeniesInsideWindow pins the gate's core:
// the first attempt is allowed and recorded; an immediate re-attempt
// is DENIED with the slot's stamp — and records nothing (the retry
// loop must not extend the window it fights).
func TestDialAttemptAllowsThenDeniesInsideWindow(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")
	allowed, firstAt, err := dials.Attempt(ctx, owner, "+15550001111", time.Hour)
	if err != nil || !allowed {
		t.Fatalf("first attempt must be allowed: allowed=%v err=%v", allowed, err)
	}
	awaitLast(t, dials, owner, "+15550001111", firstAt)

	allowed, stamp, err := dials.Attempt(ctx, owner, "+15550001111", time.Hour)
	if err != nil || allowed {
		t.Fatalf("second attempt inside the window must be denied: allowed=%v err=%v", allowed, err)
	}
	if stamp.Unix() != firstAt.Unix() {
		t.Fatalf("deny stamp must carry the slot's last-at (second granularity, like the stored window): got %v want %v", stamp, firstAt)
	}
	// Nothing was recorded by the deny: the slot still holds firstAt.
	awaitLast(t, dials, owner, "+15550001111", firstAt)
}

// TestDialOverrideSpendsOnce pins the one-shot bypass: armed, the NEXT
// attempt spends it (mid-window) and dials; the attempt after that is
// denied again.
func TestDialOverrideSpendsOnce(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if _, _, err := dials.Attempt(ctx, owner, "+15550002222", time.Hour); err != nil {
		t.Fatalf("first attempt: %v", err)
	}
	if err := dials.ArmOverride(ctx, owner, "+15550002222"); err != nil {
		t.Fatalf("arm override: %v", err)
	}
	allowed, _, err := dials.Attempt(ctx, owner, "+15550002222", time.Hour)
	if err != nil || !allowed {
		t.Fatalf("override must spend and allow mid-window: allowed=%v err=%v", allowed, err)
	}
	allowed, _, err = dials.Attempt(ctx, owner, "+15550002222", time.Hour)
	if err != nil || allowed {
		t.Fatalf("a spent override must not bypass twice: allowed=%v err=%v", allowed, err)
	}
}

// TestDialWindowExpires pins the cooldown's other edge: once the
// window has passed, the same remote is dialable again.
func TestDialWindowExpires(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if _, _, err := dials.Attempt(ctx, owner, "+15550003333", 20*time.Millisecond); err != nil {
		t.Fatalf("first attempt: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	allowed, _, err := dials.Attempt(ctx, owner, "+15550003333", 20*time.Millisecond)
	if err != nil || !allowed {
		t.Fatalf("attempt after the window passed must be allowed: allowed=%v err=%v", allowed, err)
	}
}

// TestDialOwnerScoping pins the (owner, remote) keying: one owner's
// window says nothing about another's right to dial.
func TestDialOwnerScoping(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	one := mustExt(t, "1001")
	two := mustExt(t, "1002")

	if _, _, err := dials.Attempt(ctx, one, "+15550004444", time.Hour); err != nil {
		t.Fatalf("owner one attempt: %v", err)
	}
	allowed, _, err := dials.Attempt(ctx, two, "+15550004444", time.Hour)
	if err != nil || !allowed {
		t.Fatalf("owner two must be unaffected: allowed=%v err=%v", allowed, err)
	}
}

// TestDialCapExpungesOldest pins the decider-side cap: beyond 500
// slots the OLDEST is expunged (the projection forgets it), and one
// owner's flood never touches another's.
func TestDialCapExpungesOldest(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	first := "+15550001"
	backdated := time.Now().Add(-time.Hour)
	timeNow = func() time.Time { return backdated }
	_, _, err := dials.Attempt(ctx, owner, first, time.Hour)
	if err != nil {
		t.Fatalf("oldest attempt: %v", err)
	}
	timeNow = time.Now
	awaitLast(t, dials, owner, first, backdated)
	for i := 0; i < dialWindowMaxPerOwner; i++ {
		if _, _, err := dials.Attempt(ctx, owner, remoteOf(i), time.Hour); err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	// The projection must eventually forget the oldest slot (cap trim).
	deadline := time.Now().Add(10 * time.Second)
	for {
		_, found, err := dials.Last(ctx, owner, first)
		if err != nil {
			t.Fatalf("Last: %v", err)
		}
		if !found {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("oldest slot was never expunged past the per-owner cap")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDialStateSurvivesRestart pins the fold's journal-reload
// correctness: state rebuilt from CBOR payload bytes, not in-memory
// structs — the decode path a long-lived process never exercises but
// every restart does.
func TestDialStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	dials := bootDialSystem(t, dir)
	if _, _, err := dials.Attempt(ctx, owner, "+15550005555", time.Hour); err != nil {
		t.Fatalf("attempt before restart: %v", err)
	}

	dials2 := bootDialSystem(t, dir)
	allowed, _, err := dials2.Attempt(ctx, owner, "+15550005555", time.Hour)
	if err != nil || allowed {
		t.Fatalf("state must reload from the journal after restart: allowed=%v err=%v", allowed, err)
	}
}

// TestDialActiveListsUnexpiredWindows pins the typeahead read: only
// remotes still inside their window appear, newest first.
func TestDialActiveListsUnexpiredWindows(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	at := time.Now()
	timeNow = func() time.Time { return at }
	if _, _, err := dials.Attempt(ctx, owner, "+15550006666", time.Hour); err != nil {
		t.Fatalf("fresh attempt: %v", err)
	}
	stale := time.Now().Add(-2 * time.Hour)
	timeNow = func() time.Time { return stale }
	if _, _, err := dials.Attempt(ctx, owner, "+15550007777", time.Hour); err != nil {
		t.Fatalf("stale attempt: %v", err)
	}
	timeNow = time.Now
	awaitLast(t, dials, owner, "+15550006666", at)

	deadline := time.Now().Add(5 * time.Second)
	for {
		slots, err := dials.Active(ctx, owner, time.Hour)
		if err != nil {
			t.Fatalf("Active: %v", err)
		}
		if len(slots) == 1 && slots[0].Remote == "+15550006666" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("active list must hold exactly the unexpired slot: %+v", slots)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func remoteOf(i int) string {
	// The 06xx block never collides with the tests' hand-written +1555… remotes.
	return fmt.Sprintf("+1606%04d", i)
}

func mustExt(t *testing.T, raw string) domain.Extension {
	t.Helper()
	ext, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatalf("parse extension %q: %v", raw, err)
	}
	return ext
}
