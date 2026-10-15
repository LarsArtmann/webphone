package readmodels

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/webphone/internal/domain"
)

func newTestTranscripts(t *testing.T) *Transcripts {
	t.Helper()
	return openTestEngine(t).Transcripts()
}

func seg(owner, callID, direction, remote, text string, startedAt time.Time) domain.CallTranscriptSegment {
	return domain.CallTranscriptSegment{
		Owner:     domain.MustParseExtension(owner),
		CallID:    callID,
		Direction: domain.Direction(direction),
		Remote:    remote,
		StartedAt: startedAt,
		Text:      text,
	}
}

// bulkSeed inserts n segments straight through the engine (no seam
// trims) — the engine twin of the SQL store's raw-INSERT seed: the
// trim math under test needs a flooded store without paying Append's
// per-segment trim scans.
func bulkSeed(t *testing.T, s *Transcripts, owner, callID string, n int, text func(i int) string) {
	t.Helper()
	ctx := context.Background()
	for i := range n {
		if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.TranscriptAppended{}), domain.TranscriptAppended{
			Key:       domain.GenerateTranscriptSegmentKey(),
			Owner:     owner,
			CallID:    callID,
			Direction: domain.DirectionOutbound,
			Remote:    "+4930",
			StartedAt: time.Unix(1760000000, 0),
			Seq:       int64(1_000_000 + i),
			Text:      text(i),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTranscriptsRecentGroupsByCallInSpokenOrder(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	base := time.Unix(1760000000, 0)

	// Call A: three segments, appended in spoken order.
	for _, text := range []string{"first", "second", "third"} {
		if err := s.Append(ctx, seg("1001", "call-a", "out", "+4930", text, base)); err != nil {
			t.Fatal(err)
		}
	}
	// Call B: one newer segment.
	if err := s.Append(ctx, seg("1001", "call-b", "in", "+4989", "neu", base.Add(time.Minute))); err != nil {
		t.Fatal(err)
	}

	got, err := s.Recent(ctx, owner, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("groups: %d (want 2)", len(got))
	}
	if got[0].CallID != "call-b" {
		t.Errorf("newest call first: got %s", got[0].CallID)
	}
	if got[1].CallID != "call-a" || got[1].Remote != "+4930" || got[1].Direction != domain.DirectionOutbound {
		t.Errorf("call-a identity: %+v", got[1])
	}
	if len(got[1].Lines) != 3 || got[1].Lines[0] != "first" || got[1].Lines[2] != "third" {
		t.Errorf("spoken order: %v", got[1].Lines)
	}
}

func TestTranscriptsAreOwnerScoped(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	if err := s.Append(ctx, seg("1001", "call-a", "out", "+4930", "mine", base)); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(ctx, seg("1002", "call-b", "in", "+4989", "theirs", base)); err != nil {
		t.Fatal(err)
	}

	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Lines[0] != "mine" {
		t.Errorf("owner scoping violated: %+v", got)
	}
}

func TestTranscriptsRecentCapsGroupCountButKeepsWholeCalls(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	// Three calls, each with two segments; the newest call wins the
	// single group slot and must arrive WHOLE.
	for i, call := range []string{"call-1", "call-2", "call-3"} {
		started := base.Add(time.Duration(i) * time.Minute)
		for _, text := range []string{"a", "b"} {
			if err := s.Append(ctx, seg("1001", call, "out", "+4930", text, started)); err != nil {
				t.Fatal(err)
			}
		}
	}
	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("groups: %d (want the cap)", len(got))
	}
	if got[0].CallID != "call-3" {
		t.Errorf("newest call: got %s", got[0].CallID)
	}
	if len(got[0].Lines) != 2 || got[0].Lines[0] != "a" {
		t.Errorf("the capped window must reassemble the newest call whole: %v", got[0].Lines)
	}
}

// TestTranscriptsSearchMatchesLiteralCaseInsensitive pins the History
// search arm: ASCII case-insensitive contains, metacharacters literal.
func TestTranscriptsSearchMatchesLiteralCaseInsensitive(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	if err := s.Append(ctx, seg("1001", "call-a", "out", "+4930", "Project ALPHA is live", base)); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(ctx, seg("1001", "call-b", "out", "+4930", "nothing here", base)); err != nil {
		t.Fatal(err)
	}
	got, err := s.Search(ctx, domain.MustParseExtension("1001"), "alpha %", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("metacharacter must match literally: %+v", got)
	}
	got, err = s.Search(ctx, domain.MustParseExtension("1001"), "ALPHA", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Lines) != 1 || got[0].Lines[0] != "Project ALPHA is live" {
		t.Fatalf("search: %+v", got)
	}
}

func TestTranscriptsAppendTrimsPastThePerCallCap(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()

	bulkSeed(t, s, "1001", "marathon", TranscriptSegmentsMaxPerCall+2, func(i int) string {
		return fmt.Sprintf("marathon-%05d", i+1)
	})
	bulkSeed(t, s, "1001", "other", 3, func(i int) string {
		return fmt.Sprintf("other-%05d", i+1)
	})

	// One more segment for the marathon call: 1503 rows for the call,
	// the trim keeps the newest 1500 (seeded 4..1502 + the fresh append).
	fresh := seg("1001", "marathon", "out", "+4930", "fresh", time.Unix(1760000000, 0).Add(time.Hour))
	if err := s.Append(ctx, fresh); err != nil {
		t.Fatal(err)
	}

	count := func(callID string) int {
		t.Helper()
		got, err := s.Export(ctx, domain.MustParseExtension("1001"), TranscriptSegmentsMaxPerExtension)
		if err != nil {
			t.Fatal(err)
		}
		for _, group := range got {
			if group.CallID == callID {
				return len(group.Lines)
			}
		}
		return 0
	}
	if got := count("marathon"); got != TranscriptSegmentsMaxPerCall {
		t.Errorf("marathon segments: %d (want the per-call cap %d)", got, TranscriptSegmentsMaxPerCall)
	}
	if got := count("other"); got != 3 {
		t.Errorf("the OTHER call must stay untouched by a marathon trim: %d rows", got)
	}
	// The survivors are the newest: the three oldest seeded rows are gone,
	// seed row #4 (the new oldest survivor) is present.
	got, err := s.Export(ctx, domain.MustParseExtension("1001"), TranscriptSegmentsMaxPerExtension)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, group := range got {
		for _, line := range group.Lines {
			seen[line] = true
		}
	}
	for _, gone := range []string{"marathon-00001", "marathon-00002", "marathon-00003"} {
		if seen[gone] {
			t.Errorf("oldest row %q survived the per-call trim", gone)
		}
	}
	if !seen["marathon-00004"] {
		t.Error("seed row #4 must survive as the oldest row")
	}
}

func TestTranscriptsAppendTrimsPastTheCap(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()

	// The flood spreads over four calls so the OWNER cap binds (a
	// single-call flood would stop at the per-call cap first — that
	// interaction is the per-call test above). The flood is bulk-seeded
	// straight through the engine: the trim math is under test, not
	// Append's per-segment cost.
	calls := []string{"call-a", "call-b", "call-c", "call-d"}
	flood := TranscriptSegmentsMaxPerExtension + 10
	for ci, call := range calls {
		share := flood / len(calls)
		if ci == 0 {
			share += flood % len(calls)
		}
		seeded := call
		bulkSeed(t, s, "1001", seeded, share, func(i int) string {
			return fmt.Sprintf("flood-%c-%05d", call[len(call)-1], i+1)
		})
	}

	// One more Append past the flood: the owner trim keeps the newest
	// 5000 rows — the oldest seeded segments go.
	if err := s.Append(ctx, seg("1001", "call-a", "out", "+4930", "fresh", time.Unix(1760000000, 0).Add(time.Hour))); err != nil {
		t.Fatal(err)
	}
	total, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total != TranscriptSegmentsMaxPerExtension {
		t.Errorf("stored segments: %d (want the cap %d)", total, TranscriptSegmentsMaxPerExtension)
	}
	// The SURVIVORS are the newest: Recent still reassembles whole calls.
	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Lines) == 0 {
		t.Fatalf("trimmed store still reassembles: %+v", got)
	}
}

// TestTranscriptsConcurrentAppendHoldsNoLocksNoLoss: live segments land
// from many goroutines at once (one flush loop per live call, several
// sessions per server). The seam mutex serializes the applies; the
// contract is NO errors, NO lost rows, and whole calls after the dust
// settles.
func TestTranscriptsConcurrentAppendHoldsNoLocksNoLoss(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	const writers, perWriter = 8, 25
	var wg sync.WaitGroup
	errs := make(chan error, writers*perWriter)
	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range perWriter {
				call := "call-" + strconv.Itoa(w%2)
				err := s.Append(ctx, seg("1001", call, "out", "+4930",
					fmt.Sprintf("w%d-s%02d", w, i), base.Add(time.Duration(w*perWriter+i)*time.Second)))
				if err != nil {
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
	total, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total != writers*perWriter {
		t.Errorf("rows: %d (want %d — nothing lost)", total, writers*perWriter)
	}
	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 2)
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

// TestTranscriptsDeleteCallAndSweep pin the removal arms: DeleteCall
// erases exactly one call (idempotently) and SweepBefore drops old
// appends by seq.
func TestTranscriptsDeleteCallAndSweep(t *testing.T) {
	s := newTestTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	if err := s.Append(ctx, seg("1001", "call-a", "out", "+4930", "keep", base)); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(ctx, seg("1001", "call-b", "out", "+4930", "drop", base)); err != nil {
		t.Fatal(err)
	}
	n, err := s.DeleteCall(ctx, domain.MustParseExtension("1001"), "call-b")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("delete removed %d rows, want 1", n)
	}
	n, err = s.DeleteCall(ctx, domain.MustParseExtension("1001"), "call-b")
	if err != nil || n != 0 {
		t.Fatalf("second delete must be inert: %d %v", n, err)
	}
	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 10)
	if err != nil || len(got) != 1 || got[0].CallID != "call-a" {
		t.Fatalf("post-delete store: %+v %v", got, err)
	}

	removed, err := s.SweepBefore(ctx, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("sweep removed %d, want 1", removed)
	}
	total, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Fatalf("post-sweep total: %d, want 0", total)
	}
}
