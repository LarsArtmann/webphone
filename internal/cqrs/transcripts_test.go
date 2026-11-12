package cqrs

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
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
