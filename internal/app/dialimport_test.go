package app

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/samber/do/v2"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// TestCqrsCarriesTheDialDomain is risk R5's gate: since Phase 2 the
// composition root's domain is NOT hollow — a dial dispatch through
// the service must answer from a real decider.
func TestCqrsCarriesTheDialDomain(t *testing.T) {
	svc, err := NewCqrs(testConfig(t, false), discardLogger())
	if err != nil {
		t.Fatalf("NewCqrs: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })
	if err := svc.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	owner := mustExtension(t, "1001")
	allowed, _, err := svc.Dials().Attempt(context.Background(), owner, "+15550009999", time.Hour)
	if err != nil || !allowed {
		t.Fatalf("a fresh slot must be dialable through the wired domain: allowed=%v err=%v", allowed, err)
	}
}

// TestDialImportPortsLegacyRows pins the A18 contract: legacy engine
// rows survive the journal swap with their stamps (remaining-window
// math unchanged), the gate DENIES inside the ported window, the
// marker latches, and a second boot imports nothing new.
func TestDialImportPortsLegacyRows(t *testing.T) {
	cfg := testConfig(t, false)

	// Seed the LEGACY engine store exactly as the old product wrote it.
	legacyEngine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	ctx := context.Background()
	owner := mustExtension(t, "1001")
	seeded, seededStamp, err := legacyEngine.DialWindows().Attempt(ctx, owner, "+15550008888", 0)
	if err != nil || !seeded {
		t.Fatalf("seed legacy window: allowed=%v err=%v", seeded, err)
	}
	if err := legacyEngine.DialWindows().ArmOverride(ctx, owner, "+15550007777"); err != nil {
		t.Fatalf("seed legacy override: %v", err)
	}
	if err := legacyEngine.Shutdown(); err != nil {
		t.Fatalf("close legacy engine: %v", err)
	}

	// First boot: the importer ports the rows into the journal.
	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)

	// Timestamp fidelity: the ported window carries the legacy stamp.
	deadline := time.Now().Add(10 * time.Second)
	for {
		last, found, err := svc.Dials().Last(ctx, owner, "+15550008888")
		if err != nil {
			t.Fatalf("Last: %v", err)
		}
		if found && last.Unix() == seededStamp.Unix() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported window never reached the journal projection: found=%v last=%v want=%v", found, last, seededStamp)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// The ported window still advises: an attempt inside it is denied.
	allowed, stamp, err := svc.Dials().Attempt(ctx, owner, "+15550008888", time.Hour)
	if err != nil || allowed {
		t.Fatalf("attempt inside the ported window must be denied: allowed=%v err=%v", allowed, err)
	}
	if stamp.Unix() != seededStamp.Unix() {
		t.Fatalf("deny stamp must carry the ported stamp: got %v want %v", stamp, seededStamp)
	}

	// The ported override still spends: the next attempt consumes it.
	allowed, _, err = svc.Dials().Attempt(ctx, owner, "+15550007777", time.Hour)
	if err != nil || !allowed {
		t.Fatalf("the ported override must spend on its next attempt: allowed=%v err=%v", allowed, err)
	}

	// The marker latched: a second boot imports nothing new.
	done, err := store.FlagPresent(ctx, mustSQL(t, application), flagDialJournalImported)
	if err != nil || !done {
		t.Fatalf("import marker must latch after a full import: done=%v err=%v", done, err)
	}
	streamID, err := id.ParseStreamID(owner.String())
	if err != nil {
		t.Fatal(err)
	}
	ref := id.NewStreamRef("dial", streamID)
	before, err := svc.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream: %v", err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	second, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	svc2 := mustResolveCqrs(t, second)
	after, err := svc2.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream after second boot: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("second boot must not append import facts: before=%d after=%d", len(before), len(after))
	}
	if err := second.Shutdown(); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

// TestAgentAuditImportPortsLegacyRows pins the A23 continuity: legacy
// engine audit rows survive the journal swap with their ids, actors
// and seq stamps (Settings' newest-20 window is unchanged), the
// marker latches, and a second boot imports nothing new.
func TestAgentAuditImportPortsLegacyRows(t *testing.T) {
	cfg := testConfig(t, false)

	legacyEngine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	ctx := context.Background()
	owner := mustExtension(t, "1001")
	if err := legacyEngine.AgentAudit().Append(ctx, owner, "ops@example.com", "approve", "draft-1", "sent"); err != nil {
		t.Fatalf("seed legacy audit row 1: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := legacyEngine.AgentAudit().Append(ctx, owner, "ops@example.com", "discard", "draft-2", "dropped"); err != nil {
		t.Fatalf("seed legacy audit row 2: %v", err)
	}
	if err := legacyEngine.Shutdown(); err != nil {
		t.Fatalf("close legacy engine: %v", err)
	}

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)

	deadline := time.Now().Add(10 * time.Second)
	for {
		rows, err := svc.AgentAuditSvc().Recent(ctx, owner, 20)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		if len(rows) == 2 {
			if rows[0].Subject != "draft-2" || rows[0].Actor != "ops@example.com" || rows[1].Subject != "draft-1" {
				t.Fatalf("ported audit rows drifted: %+v", rows)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("audit rows never reached the journal projection: got %d", len(rows))
		}
		time.Sleep(10 * time.Millisecond)
	}

	done, err := store.FlagPresent(ctx, mustSQL(t, application), flagAgentAuditJournalImported)
	if err != nil || !done {
		t.Fatalf("import marker must latch after a full import: done=%v err=%v", done, err)
	}
	streamID, err := id.ParseStreamID(owner.String())
	if err != nil {
		t.Fatal(err)
	}
	ref := id.NewStreamRef("agent_audit", streamID)
	before, err := svc.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream: %v", err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	second, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	svc2 := mustResolveCqrs(t, second)
	after, err := svc2.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("load stream after second boot: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("second boot must not append import facts: before=%d after=%d", len(before), len(after))
	}
	if err := second.Shutdown(); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

func mustExtension(t *testing.T, raw string) domain.Extension {
	t.Helper()
	ext, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatalf("parse extension %q: %v", raw, err)
	}
	return ext
}

func mustResolveCqrs(t *testing.T, application *App) *CqrsService {
	t.Helper()
	svc, err := do.InvokeNamed[*CqrsService](application.injector, cqrsProviderName)
	if err != nil {
		t.Fatalf("resolve cqrs: %v", err)
	}
	return svc
}

func mustSQL(t *testing.T, application *App) *sql.DB {
	t.Helper()
	db, err := do.InvokeNamed[*store.Database](application.injector, "sqlite")
	if err != nil {
		t.Fatalf("resolve sqlite: %v", err)
	}
	return db.SQL()
}
