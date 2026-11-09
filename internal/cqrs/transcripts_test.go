package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"

	"github.com/larsartmann/webphone/internal/domain"
)

func bootTranscripts(t *testing.T, dir string) *Transcripts {
	t.Helper()
	return NewTranscripts(bootDomainSystem(t, dir, TranscriptsDomainDecl()))
}

var transcriptSeq int

func transcriptSegmentFor(t *testing.T, owner, callID, text string) domain.CallTranscriptSegment {
	t.Helper()
	transcriptSeq++
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
