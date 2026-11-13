package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

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
