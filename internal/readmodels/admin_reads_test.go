package readmodels

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// The cross-owner admin seam: RecentAll (transcripts) is the ONLY
// read without the owner filter in this package — this pin holds it to
// that contract. (The fax ListAll/GetAny twins dissolved into the cqrs
// backend; their pins live in cqrs/faxes_test.go now.)

func TestTranscriptsRecentAllCrossesOwners(t *testing.T) {
	engine := openTestEngine(t)
	transcripts := engine.Transcripts()
	ctx := context.Background()
	first := domain.MustParseExtension("1001")
	second := domain.MustParseExtension("1002")
	base := time.Unix(1700000000, 0)

	for _, text := range []string{"first line one", "first line two"} {
		if err := transcripts.Append(ctx, domain.CallTranscriptSegment{
			Owner: first, CallID: "call-first", Direction: domain.DirectionInbound,
			Remote: "+15550000001", StartedAt: base, Text: text,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := transcripts.Append(ctx, domain.CallTranscriptSegment{
		Owner: second, CallID: "call-second", Direction: domain.DirectionOutbound,
		Remote: "+15550000002", StartedAt: base.Add(time.Hour), Text: "second line words",
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
