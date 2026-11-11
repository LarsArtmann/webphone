package cqrs

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// faxBattery boots the fax aggregate over a real system (the
// code-complete definition: this battery is green or the aggregate
// does not exist).
func faxBattery(t *testing.T) (*system.System, *Faxes) {
	t.Helper()
	sys := bootDomainSystem(t, t.TempDir(), FaxesDomainDecl())
	return sys, NewFaxes(sys)
}

func faxJob(t *testing.T, owner domain.Extension, remote string, id domain.FaxID, at time.Time) domain.FaxJob {
	t.Helper()
	return domain.NewOutboundFaxJob(id, owner, domain.MustParsePhone(remote), "faxes/a.pdf", at)
}

// TestFaxRecordAdvanceAndRefresh pins the lifecycle lane: birth queued,
// the accepted/settled drive, an out-of-order verdict, and the
// same-status refresh carrying a richer page count (faxMoves' doctrine
// — a later payload refreshes the row).
func TestFaxRecordAdvanceAndRefresh(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	at := time.Now().Add(-time.Hour)

	job := faxJob(t, owner, "+491701", domain.GenerateFaxID(), at)
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create: %v", err)
	}
	advances := []struct {
		delivery domain.FaxDelivery
	}{
		{delivery: domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "ref-1"}},
		{delivery: domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 2}},
		{delivery: domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 5}}, // same-status refresh
	}
	for i, step := range advances {
		if err := faxes.UpdateStatus(ctx, job.ID, step.delivery); err != nil {
			t.Fatalf("UpdateStatus %d: %v", i, err)
		}
	}
	got, err := faxes.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != domain.FaxTransmitted || got.ProviderRef != "ref-1" || got.Pages != 5 {
		t.Fatalf("advanced job = %+v, want transmitted/ref-1/5 pages", got)
	}

	// The failed→transmitted out-of-order verdict stays legal (a late
	// provider correction); the row keeps the LAST payload.
	other := faxJob(t, owner, "+491702", domain.GenerateFaxID(), at)
	if err := faxes.Create(ctx, other); err != nil {
		t.Fatalf("Create other: %v", err)
	}
	if err := facesDrive(t, faxes, ctx, other.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "no answer"}); err != nil {
		t.Fatalf("drive failed: %v", err)
	}
	if err := faxes.UpdateStatus(ctx, other.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 3}); err != nil {
		t.Fatalf("out-of-order verdict: %v", err)
	}
	got, err = faxes.Get(ctx, owner, other.ID)
	if err != nil || got.Status != domain.FaxTransmitted || got.Error != "" || got.Pages != 3 {
		t.Fatalf("corrected job = (%+v, %v), want transmitted/3 pages/no error", got, err)
	}
}

func facesDrive(t *testing.T, faxes *Faxes, ctx context.Context, id domain.FaxID, delivery domain.FaxDelivery) error {
	t.Helper()
	return faxes.UpdateStatus(ctx, id, delivery)
}

// TestFaxInboundFactHasNoLifecycle pins the received rule: an inbound
// row ignores every move (inert success, no write, no nudge), and the
// provider-ref lookup resolves it for the webhook lane.
func TestFaxInboundFactHasNoLifecycle(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	at := time.Now()
	job := domain.NewInboundFaxJob(domain.GenerateFaxID(), owner, domain.MustParsePhone("+491703"),
		"faxes/in.pdf", "in-ref-9", 3, at)
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create inbound: %v", err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted}); err != nil {
		t.Fatalf("inbound move must be inert, not an error: %v", err)
	}
	got, err := faxes.Get(ctx, owner, job.ID)
	if err != nil || got.Status != domain.FaxReceived {
		t.Fatalf("inbound row after move = (%+v, %v), want still received", got, err)
	}
	byRef, err := faxes.ByProviderRef(ctx, "in-ref-9")
	if err != nil || byRef.ID != job.ID {
		t.Fatalf("ByProviderRef = (%+v, %v), want the inbound job", byRef, err)
	}
}

// TestFaxDuplicateAndMissing pins the decision rejections: a stream
// holds exactly one job, and an advance against a missing job is a
// Rejection (the readmodels fax_missing contract).
func TestFaxDuplicateAndMissing(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	job := faxJob(t, owner, "+491704", domain.GenerateFaxID(), time.Now())
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := faxes.Create(ctx, job); err == nil {
		t.Fatal("duplicate record must be rejected")
	}
	missing := domain.GenerateFaxID()
	if err := faxes.UpdateStatus(ctx, missing, domain.FaxDelivery{Status: domain.FaxSending}); err == nil {
		t.Fatal("advance against a missing job must be rejected")
	}
}

// TestFaxIdempotentAdvance pins the durable replay guard: the same
// idem key dispatches once; a replay inside the window is a no-op (the
// journal does not grow).
func TestFaxIdempotentAdvance(t *testing.T) {
	sys, faxes := faxBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	job := faxJob(t, owner, "+491705", domain.GenerateFaxID(), time.Now())
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create: %v", err)
	}
	idem := "faxadv:ref-x:transmitted"
	if err := faxes.AdvanceIdempotent(ctx, idem, job.ID, domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "ref-x"}); err != nil {
		t.Fatalf("first idempotent advance: %v", err)
	}
	ref := mustFaxStreamRef(t, job.ID)
	before, err := sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("journal load: %v", err)
	}
	if err := faxes.AdvanceIdempotent(ctx, idem, job.ID, domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "ref-x"}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	after, err := sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("journal load 2: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("replay grew the journal: %d -> %d events", len(before), len(after))
	}
	// A different key applies.
	if err := faxes.AdvanceIdempotent(ctx, "faxadv:ref-x:failed", job.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "late"}); err != nil {
		t.Fatalf("second key: %v", err)
	}
	got, err := faxes.Get(ctx, owner, job.ID)
	if err != nil || got.Status != domain.FaxFailed || got.Error != "late" {
		t.Fatalf("second key did not apply: (%+v, %v)", got, err)
	}
}

// TestFaxScopingAndAdminSeams pins the owner axis: Get withholds
// another owner's row, GetAny crosses by design, and the counts see
// every owner.
func TestFaxScopingAndAdminSeams(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	a := mustExt(t, "100")
	b := mustExt(t, "101")
	job := faxJob(t, b, "+491706", domain.GenerateFaxID(), time.Now())
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := faxes.Get(ctx, a, job.ID); err != store.ErrNotFound {
		t.Fatalf("cross-owner Get must be ErrNotFound, got %v", err)
	}
	any, err := faxes.GetAny(ctx, job.ID)
	if err != nil || any.Owner != b {
		t.Fatalf("GetAny = (%+v, %v), want the row with its owner", any, err)
	}
	if n, err := faxes.Count(ctx); err != nil || n != 1 {
		t.Fatalf("Count = (%d, %v), want 1", n, err)
	}
	listed, err := faxes.List(ctx, b, 10)
	if err != nil || len(listed) != 1 || listed[0].ID != job.ID {
		t.Fatalf("List = (%+v, %v)", listed, err)
	}
}

// TestFaxSweepDeletesJournalViewsAndPaths is the retention gate: an
// expired job leaves the views AND the journal (bytes gone) and its
// document path returns for unlinking; a fresh job survives.
func TestFaxSweepDeletesJournalViewsAndPaths(t *testing.T) {
	dir := t.TempDir()
	sys := bootDomainSystem(t, dir, FaxesDomainDecl())
	faxes := NewFaxes(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	old := faxJob(t, owner, "+491707", domain.GenerateFaxID(), time.Now().Add(-10*24*time.Hour))
	fresh := faxJob(t, owner, "+491708", domain.GenerateFaxID(), time.Now())
	for _, job := range []domain.FaxJob{old, fresh} {
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	paths, removed, err := faxes.SweepBefore(ctx, time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("SweepBefore: %v", err)
	}
	if removed != 1 || len(paths) != 1 || paths[0] != "faxes/a.pdf" {
		t.Fatalf("sweep = (%d removed, paths %v), want 1 and the old document", removed, paths)
	}
	if _, err := faxes.Get(ctx, owner, old.ID); err != store.ErrNotFound {
		t.Fatalf("expired job must be gone: %v", err)
	}
	events, err := sys.EventStore().Load(ctx, mustFaxStreamRef(t, old.ID))
	if err != nil || len(events) != 0 {
		t.Fatalf("trimmed stream holds %d events (err %v), want 0", len(events), err)
	}
	if _, err := faxes.Get(ctx, owner, fresh.ID); err != nil {
		t.Fatalf("fresh job must survive: %v", err)
	}
}

// TestFaxRestartReloadsState pins replay: a second system over the
// same directory rebuilds the rows from the journal.
func TestFaxRestartReloadsState(t *testing.T) {
	dir := t.TempDir()
	sys := bootDomainSystem(t, dir, FaxesDomainDecl())
	faxes := NewFaxes(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	job := faxJob(t, owner, "+491709", domain.GenerateFaxID(), time.Now())
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 4}); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := sys.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	sys2 := bootDomainSystem(t, dir, FaxesDomainDecl())
	defer func() { _ = sys2.Close() }()
	faxes2 := NewFaxes(sys2)
	got, err := faxes2.Get(ctx, owner, job.ID)
	if err != nil || got.Status != domain.FaxTransmitted || got.Pages != 4 {
		t.Fatalf("after restart = (%+v, %v), want transmitted/4", got, err)
	}
}

func mustFaxStreamRef(t *testing.T, faxID domain.FaxID) id.StreamRef {
	t.Helper()
	streamID, err := id.ParseStreamID(FaxStreamKey(faxID))
	if err != nil {
		t.Fatalf("parse fax stream key %q: %v", FaxStreamKey(faxID), err)
	}
	return id.NewStreamRef(FaxesStreamType, streamID)
}
