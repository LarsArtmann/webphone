package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// TestFaxImportDualRunParity is A40's gate: the engine fax history
// survives the journal port row-for-row (every field the list and the
// webhook lane read), the marker latches, a second boot imports
// nothing new, and a live lifecycle move continues an imported stream.
func TestFaxImportDualRunParity(t *testing.T) {
	cfg := testConfig(t, false)
	ctx := context.Background()

	// Seed the ENGINE store exactly as the pre-swap product wrote it.
	engine, err := readmodels.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("open legacy engine: %v", err)
	}
	owner := mustExtension(t, "1001")
	other := mustExtension(t, "1002")
	base := time.Now().Add(-48 * time.Hour)

	// An outbound job driven to transmitted with a provider ref, and an
	// inbound received fact on another owner (the scoping axis).
	outbound := domain.NewOutboundFaxJob(domain.GenerateFaxID(), owner,
		domain.MustParsePhone("+15550001111"), "faxes/legacy-out.pdf", base)
	if err := engine.Faxes().Create(ctx, outbound); err != nil {
		t.Fatalf("seed outbound: %v", err)
	}
	if err := engine.Faxes().UpdateStatus(ctx, outbound.ID, domain.FaxDelivery{
		Status: domain.FaxSending, ProviderRef: "legacy-fax-ref"}); err != nil {
		t.Fatalf("seed sending: %v", err)
	}
	if err := engine.Faxes().UpdateStatus(ctx, outbound.ID, domain.FaxDelivery{
		Status: domain.FaxTransmitted, Pages: 3}); err != nil {
		t.Fatalf("seed transmitted: %v", err)
	}
	inbound := domain.NewInboundFaxJob(domain.GenerateFaxID(), other,
		domain.MustParsePhone("+15550002222"), "faxes/legacy-in.pdf", "legacy-in-ref", 2, base.Add(time.Hour))
	if err := engine.Faxes().Create(ctx, inbound); err != nil {
		t.Fatalf("seed inbound: %v", err)
	}
	wantAll, err := engine.Faxes().ListAll(ctx, 0)
	if err != nil {
		t.Fatalf("legacy ListAll: %v", err)
	}
	if err := engine.Shutdown(); err != nil {
		t.Fatalf("close engine: %v", err)
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
	gate := svc.FaxesSvc()

	deadline := time.Now().Add(20 * time.Second)
	for {
		got, err := gate.Count(ctx)
		if err != nil {
			t.Fatalf("gate Count: %v", err)
		}
		if got == int64(len(wantAll)) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported faxes never reached the projection: got %d want %d", got, len(wantAll))
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Row-by-row diff (newest-first both lanes).
	gotAll, err := gate.ListAll(ctx, 0)
	if err != nil {
		t.Fatalf("gate ListAll: %v", err)
	}
	if len(gotAll) != len(wantAll) {
		t.Fatalf("jobs: got %d want %d", len(gotAll), len(wantAll))
	}
	for i := range wantAll {
		want, got := wantAll[i], gotAll[i]
		if got.ID != want.ID || got.Owner != want.Owner || got.Remote != want.Remote ||
			got.Direction != want.Direction || got.Status != want.Status ||
			got.Pages != want.Pages || got.DocumentPath != want.DocumentPath ||
			got.ProviderRef != want.ProviderRef || got.Error != want.Error {
			t.Errorf("job %d: got %+v want %+v", i, got, want)
		}
		// The journal's CBOR epoch quantizes time.Time payloads by
		// ~200 ns (the recorded constraint from the snippets train);
		// every render is second-granularity, so the honest parity bar
		// is microsecond equality.
		if diff := got.CreatedAt.Sub(want.CreatedAt); diff > time.Microsecond || diff < -time.Microsecond {
			t.Errorf("job %d CreatedAt drift %v: got %v want %v", i, diff, got.CreatedAt, want.CreatedAt)
		}
		if diff := got.UpdatedAt.Sub(want.UpdatedAt); diff > time.Microsecond || diff < -time.Microsecond {
			t.Errorf("job %d UpdatedAt drift %v: got %v want %v", i, diff, got.UpdatedAt, want.UpdatedAt)
		}
	}

	// The webhook lane's lookups survive.
	byRef, err := gate.ByProviderRef(ctx, "legacy-fax-ref")
	if err != nil || byRef.ID != outbound.ID || byRef.Status != domain.FaxTransmitted {
		t.Fatalf("provider-ref lookup: got %+v err %v", byRef, err)
	}
	scoped, err := gate.Get(ctx, owner, outbound.ID)
	if err != nil || scoped.ID != outbound.ID {
		t.Fatalf("owner-scoped get: %+v %v", scoped, err)
	}
	if _, err := gate.Get(ctx, other, outbound.ID); err != store.ErrNotFound {
		t.Fatalf("cross-owner get must be ErrNotFound: %v", err)
	}

	// The marker latched.
	flagDB, err := store.OpenDatabase(filepath.Join(cfg.DataDir, "webphone.db"))
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	imported, err := store.FlagPresent(ctx, flagDB.SQL(), flagFaxesJournalImported)
	if err != nil || !imported {
		t.Fatalf("import marker: present %v err %v", imported, err)
	}
	if err := flagDB.Shutdown(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	// Second boot: nothing re-imports, and a live move continues the
	// imported stream (same-status refresh doctrine).
	if err := application.Shutdown(); err != nil {
		t.Fatalf("shutdown app 1: %v", err)
	}
	application2, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New 2: %v", err)
	}
	if err := application2.Start(ctx); err != nil {
		t.Fatalf("Start 2: %v", err)
	}
	t.Cleanup(func() {
		if err := application2.Shutdown(); err != nil {
			t.Errorf("Shutdown 2: %v", err)
		}
	})
	svc2 := mustResolveCqrs(t, application2)
	if n, err := svc2.FaxesSvc().Count(ctx); err != nil || n != int64(len(wantAll)) {
		t.Fatalf("second boot count = (%d, %v), want %d (no re-import)", n, err, len(wantAll))
	}
	if err := svc2.FaxesSvc().UpdateStatus(ctx, outbound.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "late correction"}); err != nil {
		t.Fatalf("post-import move: %v", err)
	}
	after, err := svc2.FaxesSvc().Get(ctx, owner, outbound.ID)
	if err != nil || after.Status != domain.FaxFailed || after.Error != "late correction" {
		t.Fatalf("post-import row = (%+v, %v), want failed/late correction", after, err)
	}
}
