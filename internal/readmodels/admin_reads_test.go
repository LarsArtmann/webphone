package readmodels

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// The cross-owner admin seams: ListAll/GetAny (faxes) and RecentAll
// (transcripts) are the ONLY reads without the owner filter — these
// pins hold them to that contract.

func TestFaxesListAllCrossesOwnersNewestFirst(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)

	older := mustFax(t, "1001", "+15550000001", domain.FaxOutbound, domain.FaxQueued, base)
	newer := mustFax(t, "1002", "+15550000002", domain.FaxInbound, domain.FaxReceived, base.Add(time.Minute))
	for _, job := range []domain.FaxJob{older, newer} {
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatal(err)
		}
	}
	jobs, err := faxes.ListAll(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].ID != newer.ID || jobs[1].ID != older.ID {
		t.Errorf("ListAll = [%s %s], want newest-first across owners", jobs[0].ID, jobs[1].ID)
	}
	if jobs, err = faxes.ListAll(ctx, 1); err != nil || len(jobs) != 1 {
		t.Errorf("limit must bound the inventory: %d, %v", len(jobs), err)
	}
}

func TestFaxesGetAnyCrossesOwners(t *testing.T) {
	engine := openTestEngine(t)
	faxes := engine.Faxes()
	ctx := context.Background()
	foreign := mustFax(t, "1002", "+15550000009", domain.FaxInbound, domain.FaxReceived, time.Unix(1700000000, 0))
	if err := faxes.Create(ctx, foreign); err != nil {
		t.Fatal(err)
	}

	// The owner-scoped sibling refuses the foreign read; GetAny grants it.
	if _, err := faxes.Get(ctx, domain.MustParseExtension("1001"), foreign.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("owner-scoped Get must refuse: %v", err)
	}
	job, err := faxes.GetAny(ctx, foreign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Owner != foreign.Owner {
		t.Errorf("owner = %s, want %s (the row carries its line)", job.Owner, foreign.Owner)
	}
	if _, err = faxes.GetAny(ctx, domain.GenerateFaxID()); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("unknown id must answer not-found: %v", err)
	}
}

func TestTranscriptsRecentAllCrossesOwners(t *testing.T) {
	engine := openTestEngine(t)
	transcripts := engine.Transcripts()
	ctx := context.Background()
	first := domain.MustParseExtension("1001")
	second := domain.MustParseExtension("1002")
	base := time.Unix(1700000000, 0)

	for i, text := range []string{"first line one", "first line two"} {
		if err := transcripts.Append(ctx, domain.CallTranscriptSegment{
			Owner: first, CallID: "call-first", Direction: domain.DirectionInbound,
			Remote: "+15550000001", StartedAt: base, Text: text,
			Seq: int64(i),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := transcripts.Append(ctx, domain.CallTranscriptSegment{
		Owner: second, CallID: "call-second", Direction: domain.DirectionOutbound,
		Remote: "+15550000002", StartedAt: base.Add(time.Hour), Text: "second line words",
		Seq: 10,
	}); err != nil {
		t.Fatal(err)
	}

	groups, err := transcripts.RecentAll(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2 (one per line)", len(groups))
	}
	// Newest call first; each group keeps its own owner.
	if groups[0].Owner != second || groups[1].Owner != first {
		t.Errorf("owner order = [%s %s], want newest first", groups[0].Owner, groups[1].Owner)
	}
	if len(groups[1].Lines) != 2 {
		t.Errorf("first line's group must reassemble whole: %v", groups[1].Lines)
	}
	// The owner-scoped sibling still answers one line only.
	onlyFirst, err := transcripts.Recent(ctx, first, 10)
	if err != nil || len(onlyFirst) != 1 {
		t.Errorf("owner-scoped Recent must stay scoped: %d, %v", len(onlyFirst), err)
	}
}
