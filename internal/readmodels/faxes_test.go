package readmodels

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

func mustFax(t *testing.T, owner, remote string, direction domain.FaxDirection, status domain.FaxStatus, created time.Time) domain.FaxJob {
	t.Helper()
	return domain.FaxJob{
		ID:        domain.GenerateFaxID(),
		Owner:     domain.MustParseExtension(owner),
		Remote:    domain.MustParsePhone(remote),
		Direction: direction,
		Status:    status,
		CreatedAt: created,
		UpdatedAt: created,
	}
}

func TestFaxesCreateListGet(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)

	outbound := mustFax(t, "1001", "+15550000001", domain.FaxOutbound, domain.FaxQueued, base)
	inbound := mustFax(t, "1001", "+15550000002", domain.FaxInbound, domain.FaxReceived, base.Add(time.Minute))
	other := mustFax(t, "1002", "+15550000003", domain.FaxOutbound, domain.FaxQueued, base)
	for _, job := range []domain.FaxJob{outbound, inbound, other} {
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("create %s: %v", job.ID, err)
		}
	}

	list, err := faxes.List(ctx, domain.MustParseExtension("1001"), 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list length %d, want 2", len(list))
	}
	if list[0].ID != inbound.ID || list[1].ID != outbound.ID {
		t.Fatalf("list order: want newest first, got [%s, %s]", list[0].ID, list[1].ID)
	}
	if list[1].Direction != domain.FaxOutbound || list[0].Direction != domain.FaxInbound {
		t.Fatalf("direction lost in projection: %+v", list)
	}

	// The limit is the page shape (the SQL store's LIMIT).
	page, err := faxes.List(ctx, domain.MustParseExtension("1001"), 1)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if len(page) != 1 || page[0].ID != inbound.ID {
		t.Fatalf("limit broke the page: %+v", page)
	}

	// Get is owner-scoped; misses and cross-owner reads are ErrNotFound.
	job, err := faxes.Get(ctx, domain.MustParseExtension("1001"), outbound.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if job.ID != outbound.ID || job.Status != domain.FaxQueued {
		t.Fatalf("get round-trip: %+v", job)
	}
	if _, err := faxes.Get(ctx, domain.MustParseExtension("1002"), outbound.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("cross-owner get: %v, want ErrNotFound", err)
	}
	if _, err := faxes.Get(ctx, domain.MustParseExtension("1001"), domain.GenerateFaxID()); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing get: %v, want ErrNotFound", err)
	}
}

// TestFaxesUpdateStatusParity pins the SQL UPDATE's exact arms: the
// correlation id lands on the row, a later verdict with an EMPTY ref
// keeps the existing one, error text and page count are stored
// verbatim, and a missing row answers the not-found sentinel.
func TestFaxesUpdateStatusParity(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	job := mustFax(t, "1001", "+15550000001", domain.FaxOutbound, domain.FaxQueued, time.Unix(1700000000, 0))
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Acceptance: sending + correlation id.
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "prov-ref-1"}); err != nil {
		t.Fatalf("accept: %v", err)
	}
	got, err := faxes.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.FaxSending || got.ProviderRef != "prov-ref-1" {
		t.Fatalf("accept did not land: %+v", got)
	}

	// Deferred verdict: transmitted + page count; the ref rides empty
	// and MUST be kept (the SQL CASE arm).
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 3}); err != nil {
		t.Fatalf("verdict: %v", err)
	}
	got, err = faxes.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.FaxTransmitted || got.ProviderRef != "prov-ref-1" || got.Pages != 3 {
		t.Fatalf("verdict broke the row: %+v", got)
	}

	// A failure verdict stores the verbatim text (SQL: error = ?).
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "busy signal"}); err != nil {
		t.Fatalf("failure verdict: %v", err)
	}
	got, _ = faxes.Get(ctx, owner, job.ID)
	if got.Status != domain.FaxFailed || got.Error != "busy signal" {
		t.Fatalf("failure text lost: %+v", got)
	}

	// A missing row is the not-found sentinel (updatedOrNotFound arm).
	if err := faxes.UpdateStatus(ctx, domain.GenerateFaxID(), domain.FaxDelivery{Status: domain.FaxFailed}); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("update missing: %v, want ErrNotFound", err)
	}
}

// TestFaxesMovesLegalityMirror drives every row of the faxMoves table
// through the seam (the projection stores what the machine — the ONE
// legality home — sends it) and pins the resulting statuses.
func TestFaxesMovesLegalityMirror(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	// Every {from, event, to} row of faxMoves, as store-level moves:
	// create at `from`, UpdateStatus to `to`.
	moves := []struct {
		from domain.FaxStatus
		to   domain.FaxStatus
	}{
		{domain.FaxQueued, domain.FaxSending},          // accepted
		{domain.FaxQueued, domain.FaxFailed},           // submitFailed
		{domain.FaxSending, domain.FaxTransmitted},     // settled
		{domain.FaxQueued, domain.FaxTransmitted},      // verdict
		{domain.FaxSending, domain.FaxTransmitted},     // verdict
		{domain.FaxFailed, domain.FaxTransmitted},      // verdict (late reversal)
		{domain.FaxTransmitted, domain.FaxTransmitted}, // same-status refresh
		{domain.FaxQueued, domain.FaxFailed},           // verdict
		{domain.FaxSending, domain.FaxFailed},          // verdict
		{domain.FaxTransmitted, domain.FaxFailed},      // verdict (late reversal)
		{domain.FaxFailed, domain.FaxFailed},           // same-status refresh
	}
	for i, m := range moves {
		job := mustFax(t, "1001", "+15550000009", domain.FaxOutbound, m.from, time.Unix(1700000000, 0).Add(time.Duration(i)*time.Second))
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("move %d create: %v", i, err)
		}
		if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: m.to}); err != nil {
			t.Fatalf("move %d (%s -> %s): %v", i, m.from, m.to, err)
		}
		got, err := faxes.Get(ctx, owner, job.ID)
		if err != nil {
			t.Fatalf("move %d get: %v", i, err)
		}
		if got.Status != m.to {
			t.Fatalf("move %d: status %s, want %s", i, got.Status, m.to)
		}
	}
}

// TestFaxesDeferredResolutionAndReplay pins the out-of-order story:
// the deferred arm (accept, then the webhook's verdict lands later),
// a same-status replay that refreshes the payload without duplicating
// the row, and a contradictory late verdict overwriting a settled
// state (the faxMoves doctrine the service enforces before the seam).
func TestFaxesDeferredResolutionAndReplay(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	job := mustFax(t, "1001", "+15550000007", domain.FaxOutbound, domain.FaxQueued, time.Unix(1700000000, 0))
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "ref-7"}); err != nil {
		t.Fatalf("accept: %v", err)
	}
	// The deferred verdict: transmitted, 2 pages.
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 2}); err != nil {
		t.Fatalf("verdict: %v", err)
	}
	// A replayed verdict (richer page count) refreshes IN PLACE.
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 5}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	list, err := faxes.List(ctx, owner, 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Status != domain.FaxTransmitted || list[0].Pages != 5 || list[0].ProviderRef != "ref-7" {
		t.Fatalf("replay broke the row: %+v", list)
	}
	// A contradictory late verdict overwrites the settled state
	// (faxMoves keeps {transmitted, verdict, failed} legal).
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "reversal"}); err != nil {
		t.Fatalf("late reversal: %v", err)
	}
	got, _ := faxes.Get(ctx, owner, job.ID)
	if got.Status != domain.FaxFailed || got.Error != "reversal" {
		t.Fatalf("reversal lost: %+v", got)
	}
}

// TestFaxesAdvanceIdempotent pins the in-process replay lane: the same
// idemKey applies once (a duplicate is a no-op), a different key applies.
func TestFaxesAdvanceIdempotent(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	job := mustFax(t, "1001", "+15550000005", domain.FaxOutbound, domain.FaxSending, time.Unix(1700000000, 0))
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := faxes.AdvanceIdempotent(ctx, "fax:ref-5:transmitted", job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 3}); err != nil {
		t.Fatalf("advance: %v", err)
	}
	// The replayed webhook: same key, richer payload — inert.
	if err := faxes.AdvanceIdempotent(ctx, "fax:ref-5:transmitted", job.ID, domain.FaxDelivery{Status: domain.FaxTransmitted, Pages: 9}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	got, err := faxes.Get(ctx, owner, job.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.FaxTransmitted || got.Pages != 3 {
		t.Fatalf("idempotent replay was not inert: %+v", got)
	}
	// A different key applies.
	if err := faxes.AdvanceIdempotent(ctx, "fax:ref-5:failed", job.ID, domain.FaxDelivery{Status: domain.FaxFailed, Error: "late"}); err != nil {
		t.Fatalf("second key: %v", err)
	}
	got, _ = faxes.Get(ctx, owner, job.ID)
	if got.Status != domain.FaxFailed {
		t.Fatalf("second key did not apply: %+v", got)
	}
}

func TestFaxesByProviderRef(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()

	job := mustFax(t, "1001", "+15550000006", domain.FaxOutbound, domain.FaxSending, time.Unix(1700000000, 0))
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: domain.FaxSending, ProviderRef: "ref-6"}); err != nil {
		t.Fatalf("accept: %v", err)
	}

	got, err := faxes.ByProviderRef(ctx, "ref-6")
	if err != nil {
		t.Fatalf("by ref: %v", err)
	}
	if got.ID != job.ID {
		t.Fatalf("by ref resolved the wrong row: %s", got.ID)
	}
	if _, err := faxes.ByProviderRef(ctx, "unknown"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown ref: %v, want ErrNotFound", err)
	}
	if _, err := faxes.ByProviderRef(ctx, ""); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("empty ref: %v, want ErrNotFound", err)
	}
}

// TestFaxesSurviveRestart pins engine durability: the projected views
// live in readmodels.db, so a restart keeps every fax job.
func TestFaxesSurviveRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	first, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	job := mustFax(t, "1001", "+15550000008", domain.FaxInbound, domain.FaxReceived, time.Unix(1700000000, 0))
	job.ProviderRef = "durable-ref"
	if err := first.Faxes().Create(ctx, job); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := first.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	second, err := Open(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = second.Shutdown() })
	list, err := second.Faxes().List(ctx, domain.MustParseExtension("1001"), 100)
	if err != nil {
		t.Fatalf("list after restart: %v", err)
	}
	if len(list) != 1 || list[0].ID != job.ID || list[0].ProviderRef != "durable-ref" {
		t.Fatalf("faxes lost across restart: %+v", list)
	}
}

// TestFaxesSweepBefore pins the retention engine arm: rows older than
// the cutoff are expunged and their spooled document paths returned
// for the caller to unlink; newer rows and empty paths are untouched.
func TestFaxesSweepBefore(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	now := time.Now()
	old := now.Add(-48 * time.Hour)

	oldJob := mustFax(t, "1001", "+15550000011", domain.FaxInbound, domain.FaxReceived, old)
	oldJob.DocumentPath = "old-fax-doc"
	liveJob := mustFax(t, "1001", "+15550000012", domain.FaxOutbound, domain.FaxTransmitted, now)
	for _, job := range []domain.FaxJob{oldJob, liveJob} {
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	paths, removed, err := faxes.SweepBefore(ctx, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if removed != 1 || len(paths) != 1 || paths[0] != "old-fax-doc" {
		t.Fatalf("sweep removed %d with paths %v, want the old fax doc only", removed, paths)
	}
	if _, err := faxes.Get(ctx, domain.MustParseExtension("1001"), oldJob.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("old fax survived the sweep: %v", err)
	}
	if _, err := faxes.Get(ctx, domain.MustParseExtension("1001"), liveJob.ID); err != nil {
		t.Fatalf("live fax lost: %v", err)
	}
}
