package cqrs

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func bootTranscripts(t *testing.T, dir string) *Transcripts {
	t.Helper()
	return NewTranscripts(bootDomainSystem(t, dir, TranscriptsDomainDecl()))
}

var transcriptSeq atomic.Int64

func transcriptSegmentFor(t *testing.T, owner, callID, text string) domain.CallTranscriptSegment {
	t.Helper()
	transcriptSeq.Add(1)
	return domain.CallTranscriptSegment{
		Owner:     mustExt(t, owner),
		CallID:    callID,
		Direction: domain.DirectionOutbound,
		Remote:    "+15550001",
		StartedAt: time.Now(),
		Text:      text,
	}
}

// awaitTranscriptCount polls the projection until the owner holds want
// segments.
func awaitTranscriptCount(t *testing.T, s *Transcripts, owner domain.Extension, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		calls, err := s.Recent(context.Background(), owner, 50)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		got := 0
		for _, c := range calls {
			got += len(c.Lines)
		}
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("transcript projection never reached %d segments: got %d (%+v)", want, got, calls)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestTranscriptAppendRecentReassembles pins the append + reassembly
// read: segments group per call, the newest call leads, each call's
// lines are in spoken order, and Call reads ONE call.
func TestTranscriptAppendRecentReassembles(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "first line")); err != nil {
		t.Fatalf("append a1: %v", err)
	}
	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "second line")); err != nil {
		t.Fatalf("append a2: %v", err)
	}
	awaitTranscriptCount(t, s, owner, 2)
	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-b", "other call")); err != nil {
		t.Fatalf("append b1: %v", err)
	}
	awaitTranscriptCount(t, s, owner, 3)

	calls, err := s.Recent(ctx, owner, 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("want 2 call groups: %+v", calls)
	}
	if calls[0].CallID != "call-b" {
		t.Fatalf("newest call must lead: %+v", calls)
	}
	if calls[1].CallID != "call-a" || len(calls[1].Lines) != 2 ||
		calls[1].Lines[0] != "first line" || calls[1].Lines[1] != "second line" {
		t.Fatalf("call-a must reassemble whole in spoken order: %+v", calls[1])
	}

	one, err := s.Call(ctx, owner, "call-a")
	if err != nil || one == nil {
		t.Fatalf("Call(call-a): %v %+v", err, one)
	}
	if one.CallID != "call-a" || len(one.Lines) != 2 {
		t.Fatalf("Call must return the whole call: %+v", one)
	}

	// Owner scoping: a second owner sees none of it.
	other, err := s.Recent(ctx, mustExt(t, "1002"), 10)
	if err != nil {
		t.Fatalf("Recent other: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("owner 1002 must not see 1001's transcripts: %+v", other)
	}
}

// TestTranscriptSearchFindsLines pins the search read: matching lines
// reassemble into call groups (matching lines only).
func TestTranscriptSearchFindsLines(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "the widget exploded")); err != nil {
		t.Fatalf("append: %v", err)
	}
	awaitTranscriptCount(t, s, owner, 1)
	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "all quiet")); err != nil {
		t.Fatalf("append: %v", err)
	}
	awaitTranscriptCount(t, s, owner, 2)

	hits, err := s.Search(ctx, owner, "WIDGET", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 || len(hits[0].Lines) != 1 || hits[0].Lines[0] != "the widget exploded" {
		t.Fatalf("search must find the one matching line, case-insensitively: %+v", hits)
	}
}

// TestTranscriptPerCallFloorEvictsOldest pins the per-CALL floor at the
// DECISION (1500 appends through the backend would be too slow to be
// the boundary pin): one append past the floor emits the append PLUS
// the oldest segment's expunge in the SAME batch; below the floor an
// append emits exactly one event.
func TestTranscriptPerCallFloorEvictsOldest(t *testing.T) {
	full := transcriptState{Segs: make([]transcriptSeg, 0, transcriptFloorPerCall)}
	for i := range transcriptFloorPerCall {
		full.Segs = append(full.Segs, transcriptSeg{ID: fmt.Sprintf("seg-%04d", i), Seq: int64(i + 1)})
	}
	cmd := AppendTranscriptSegment{BasicCommand: basicCommand(t, "transcript.append", "1001|call-a"), Owner: "1001", CallID: "call-a", SegID: "seg-new", Seq: int64(transcriptFloorPerCall + 1)}
	events, err := decideAppendTranscript(full, event.Version(len(full.Segs)), cmd)
	if err != nil {
		t.Fatalf("decide at floor: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("append at the floor must emit append + one expunge: %d events", len(events))
	}
	if events[1].Type() != event.Type(metaengine.EventTypeName(domain.TranscriptExpunged{})) {
		t.Fatalf("the second event must be the expunge: %v", events[1].Type())
	}

	below := transcriptState{Segs: full.Segs[:transcriptFloorPerCall-1]}
	events, err = decideAppendTranscript(below, event.Version(len(below.Segs)), cmd)
	if err != nil {
		t.Fatalf("decide below floor: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("append below the floor must emit exactly one event: %d", len(events))
	}
}

// TestTranscriptDeleteCallCountsAndIsIdempotent pins the end-user
// delete: the count is decision-authoritative, a second delete of the
// same call is the same success with count 0, and the projection
// eventually holds only the untouched call.
func TestTranscriptDeleteCallCountsAndIsIdempotent(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for _, seg := range []domain.CallTranscriptSegment{
		transcriptSegmentFor(t, "1001", "call-a", "a1"),
		transcriptSegmentFor(t, "1001", "call-a", "a2"),
		transcriptSegmentFor(t, "1001", "call-b", "b1"),
	} {
		if err := s.Append(ctx, seg); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	awaitTranscriptCount(t, s, owner, 3)

	removed, err := s.DeleteCall(ctx, owner, "call-a")
	if err != nil {
		t.Fatalf("DeleteCall: %v", err)
	}
	if removed != 2 {
		t.Fatalf("DeleteCall must count its expunges: got %d want 2", removed)
	}
	removed, err = s.DeleteCall(ctx, owner, "call-a")
	if err != nil {
		t.Fatalf("second DeleteCall: %v", err)
	}
	if removed != 0 {
		t.Fatalf("an empty delete is inert: got %d", removed)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		calls, err := s.Recent(ctx, owner, 10)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		if len(calls) == 1 && calls[0].CallID == "call-b" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("call-a must leave the projection: %+v", calls)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestTranscriptSweepBeforeConverges pins the retention arm: a cutoff
// after every append sweeps them all, the count is exact, and a second
// sweep is a 0-converged no-op.
func TestTranscriptSweepBeforeConverges(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for range 4 {
		if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "line")); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	awaitTranscriptCount(t, s, owner, 4)

	removed, err := s.SweepBefore(ctx, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("SweepBefore: %v", err)
	}
	if removed != 4 {
		t.Fatalf("sweep must remove exactly what it scanned: got %d want 4", removed)
	}
	removed, err = s.SweepBefore(ctx, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("second SweepBefore: %v", err)
	}
	if removed != 0 {
		t.Fatalf("a converged sweep removes nothing: got %d", removed)
	}
	awaitTranscriptCount(t, s, owner, 0)
}

// TestTranscriptStateSurvivesRestart pins the journal-reload fold.
func TestTranscriptStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s := bootTranscripts(t, dir)
	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "across boots")); err != nil {
		t.Fatalf("append: %v", err)
	}

	s2 := bootTranscripts(t, dir)
	deadline := time.Now().Add(5 * time.Second)
	for {
		calls, err := s2.Recent(ctx, owner, 10)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		if len(calls) == 1 && calls[0].CallID == "call-a" && len(calls[0].Lines) == 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("transcript did not survive the restart: %+v", calls)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// bulkSeedTranscripts folds n segments STRAIGHT into the projection
// store (no seam, no journal) — the engine store's raw-seed twin: the
// window/cap trim math under test needs a flooded projection without
// paying Append's per-segment dispatch + await. seqBase must keep the
// seeded range DISJOINT from other calls' (seq is the projection's
// only sort field — colliding ranges interleave arbitrarily).
func bulkSeedTranscripts(t *testing.T, s *Transcripts, owner, callID string, n, seqBase int, text func(i int) string) {
	t.Helper()
	ctx := context.Background()
	for i := range n {
		if err := s.sys.MetaEngine().Apply(ctx, metaengine.EventTypeName(domain.TranscriptAppended{}), domain.TranscriptAppended{
			Key:       domain.GenerateTranscriptSegmentKey(),
			Owner:     owner,
			CallID:    callID,
			Direction: domain.DirectionOutbound,
			Remote:    "+4930",
			StartedAt: time.Unix(1760000000, 0),
			Seq:       int64(seqBase + i),
			Text:      text(i),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

// TestTranscriptRecentWindowKeepsWholeNewestCalls pins the caps ×
// window contract over a flooded projection (2230 stored segments,
// 2000 read window): the render cap keeps WHOLE newest calls, the
// window's only cut is the OLDEST call's head (the group nobody
// renders at that depth), and the reassembled total never exceeds the
// window. Nano-scale seq bases match the seam's real width (UnixNano).
func TestTranscriptRecentWindowKeepsWholeNewestCalls(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	nano := base.UnixNano()

	bulkSeedTranscripts(t, s, "1001", "old-1", 1100, int(nano-3_000_000_000), func(i int) string { return fmt.Sprintf("old1-%04d", i) })
	bulkSeedTranscripts(t, s, "1001", "old-2", 1100, int(nano-2_000_000_000), func(i int) string { return fmt.Sprintf("old2-%04d", i) })
	for _, call := range []string{"new-1", "new-2", "new-3"} {
		for i := range 10 {
			seg := transcriptSegmentFor(t, "1001", call, fmt.Sprintf("%s-%02d", call, i))
			seg.StartedAt = base.Add(time.Duration(i) * time.Second)
			if err := s.Append(ctx, seg); err != nil {
				t.Fatal(err)
			}
		}
	}

	cases := []struct {
		name      string
		maxCalls  int
		wantCalls int
	}{
		{"render cap keeps the newest whole calls", 2, 2},
		{"a deep render cap surfaces every stored call", 10, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.Recent(ctx, mustExt(t, "1001"), tc.maxCalls)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.wantCalls {
				t.Fatalf("groups: %d (want %d)", len(got), tc.wantCalls)
			}
			if got[0].CallID != "new-3" || len(got[0].Lines) != 10 || got[0].Lines[0] != "new-3-00" {
				t.Errorf("newest call must lead, whole: %+v", got[0])
			}
			total := 0
			for _, call := range got {
				total += len(call.Lines)
			}
			if total > transcriptFetchWindow {
				t.Errorf("reassembled %d lines, want the %d-segment window as the ceiling", total, transcriptFetchWindow)
			}
		})
	}

	// The window's honest shape: 2230 stored, 2000 read — exactly the
	// 230 oldest segments (old-1's head) are absent; old-2 stays whole.
	got, err := s.Recent(ctx, mustExt(t, "1001"), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 || got[4].CallID != "old-1" || len(got[4].Lines) != 1100-230 {
		t.Errorf("window cut shape: %d groups, oldest = %s/%d — want old-1 head-truncated to %d",
			len(got), got[len(got)-1].CallID, len(got[len(got)-1].Lines), 1100-230)
	}
	if got[3].CallID != "old-2" || len(got[3].Lines) != 1100 {
		t.Errorf("old-2 must ride the window whole: %+v", got[3])
	}
}

// TestTranscriptOwnerCapTrimsOldestOnAppend pins the per-OWNER bound
// (5000, the declarative WithCapPerGroup): flooding past the cap and
// appending once converges the projection to exactly the cap, the
// oldest seeded segments go, and Recent still reassembles whole calls.
// The cap enforcement rides the store's EnforceCaps after the append's
// read-your-writes await (ADR-0164; the nested-grouping gap that keeps
// the per-call floor at the decision is the recorded reopen trigger).
func TestTranscriptOwnerCapTrimsOldestOnAppend(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()

	calls := []string{"call-a", "call-b", "call-c", "call-d"}
	flood := readmodels.TranscriptSegmentsMaxPerExtension + 10
	for ci, call := range calls {
		share := flood / len(calls)
		if ci == 0 {
			share += flood % len(calls)
		}
		nano := time.Unix(1760000000, 0).UnixNano()
		bulkSeedTranscripts(t, s, "1001", call, share, int(nano)+ci*1_000_000_000, func(i int) string {
			return fmt.Sprintf("flood-%c-%05d", call[len(call)-1], i+1)
		})
	}

	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "fresh")); err != nil {
		t.Fatal(err)
	}
	total, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total != readmodels.TranscriptSegmentsMaxPerExtension {
		t.Errorf("stored segments: %d (want the cap %d)", total, readmodels.TranscriptSegmentsMaxPerExtension)
	}
	got, err := s.Recent(ctx, mustExt(t, "1001"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Lines) == 0 {
		t.Fatalf("trimmed store still reassembles: %+v", got)
	}
}

// TestTranscriptSearchMatchesMetacharactersLiterally pins the History
// search arm's LITERAL half (the case-insensitive half lives in
// TestTranscriptSearchFindsLines): a Contains scan has no LIKE
// metacharacters — "%" matches only a literal percent sign.
func TestTranscriptSearchMatchesMetacharactersLiterally(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-a", "Project ALPHA is live")); err != nil {
		t.Fatal(err)
	}
	awaitTranscriptCount(t, s, owner, 1)
	if err := s.Append(ctx, transcriptSegmentFor(t, "1001", "call-b", "nothing here")); err != nil {
		t.Fatal(err)
	}
	awaitTranscriptCount(t, s, owner, 2)

	hits, err := s.Search(ctx, owner, "alpha %", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("metacharacter must match literally: %+v", hits)
	}
	hits, err = s.Search(ctx, owner, "ALPHA", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || len(hits[0].Lines) != 1 || hits[0].Lines[0] != "Project ALPHA is live" {
		t.Fatalf("search: %+v", hits)
	}
}

// TestTranscriptConcurrentAppendHoldsNoLocksNoLoss: live segments land
// from many goroutines at once (one flush loop per live call, several
// sessions per server). The seam mutex serializes the dispatches; the
// contract is NO errors, NO lost rows, and whole calls after the dust
// settles.
func TestTranscriptConcurrentAppendHoldsNoLocksNoLoss(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")
	const writers, perWriter = 8, 25
	var wg sync.WaitGroup
	errs := make(chan error, writers*perWriter)
	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range perWriter {
				call := "call-" + strconv.Itoa(w%2)
				seg := transcriptSegmentFor(t, "1001", call, fmt.Sprintf("w%d-s%02d", w, i))
				if err := s.Append(ctx, seg); err != nil {
					errs <- err
					return
				}
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent append failed: %v", err)
	}
	awaitTranscriptCount(t, s, owner, writers*perWriter)
	total, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total != writers*perWriter {
		t.Errorf("rows: %d (want %d — nothing lost)", total, writers*perWriter)
	}
	got, err := s.Recent(ctx, owner, 2)
	if err != nil {
		t.Fatal(err)
	}
	lines := 0
	for _, group := range got {
		lines += len(group.Lines)
	}
	if lines != writers*perWriter {
		t.Errorf("reassembled lines: %d (want %d)", lines, writers*perWriter)
	}
}

// TestTranscriptRecentAllCrossesOwners pins the cross-owner admin
// seam: RecentAll is the ONLY transcript read without the owner filter
// — this pin holds it to that contract (the deliberate
// readmodels.{}RecentAll-class seam; owner-scoping stays the rule
// everywhere else).
func TestTranscriptRecentAllCrossesOwners(t *testing.T) {
	s := bootTranscripts(t, t.TempDir())
	ctx := context.Background()
	first := mustExt(t, "1001")
	second := mustExt(t, "1002")

	for _, text := range []string{"first line one", "first line two"} {
		seg := transcriptSegmentFor(t, "1001", "call-first", text)
		seg.Owner = first
		seg.Direction = domain.DirectionInbound
		seg.Remote = "+15550000001"
		if err := s.Append(ctx, seg); err != nil {
			t.Fatal(err)
		}
	}
	segSecond := transcriptSegmentFor(t, "1002", "call-second", "second line words")
	segSecond.Owner = second
	segSecond.Remote = "+15550000002"
	if err := s.Append(ctx, segSecond); err != nil {
		t.Fatal(err)
	}

	groups, err := s.RecentAll(ctx, 10)
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
	onlyFirst, err := s.Recent(ctx, first, 10)
	if err != nil || len(onlyFirst) != 1 {
		t.Errorf("owner-scoped Recent must stay scoped: %d, %v", len(onlyFirst), err)
	}
}
