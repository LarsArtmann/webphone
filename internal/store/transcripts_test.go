package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func newTranscripts(t *testing.T) *Transcripts {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewTranscripts(db)
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

func TestTranscriptsRecentGroupsByCallInSpokenOrder(t *testing.T) {
	s := newTranscripts(t)
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
	s := newTranscripts(t)
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
	s := newTranscripts(t)
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

func TestTranscriptsAppendTrimsPastTheCap(t *testing.T) {
	s := newTranscripts(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	base := time.Unix(1760000000, 0)
	for i := range TranscriptSegmentsMaxPerExtension + 10 {
		text := string(rune('a'+i%26)) + "-seg"
		if err := s.Append(ctx, seg("1001", "call-many", "out", "+4930", text, base.Add(time.Duration(i)*time.Second))); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Recent(ctx, owner, 10)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, group := range got {
		total += len(group.Lines)
	}
	if total != TranscriptSegmentsMaxPerExtension {
		t.Errorf("stored segments: %d (want the cap %d)", total, TranscriptSegmentsMaxPerExtension)
	}
}
