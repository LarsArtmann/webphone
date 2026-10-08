package store

import (
	"context"
	"fmt"
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

func TestTranscriptsAppendTrimsPastThePerCallCap(t *testing.T) {
	s := newTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)

	// Bulk-seed past the per-call cap in ONE transaction: the cap math,
	// not Append's per-segment cost, is under test here (the owner-cap
	// test above already walks the Append path).
	seed := func(callID string, n int) {
		t.Helper()
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		for i := range n {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO call_transcripts (owner, call_id, direction, remote, started_at, text, created_at)
				VALUES (?, ?, 'out', '+4930', ?, ?, 0)
			`, "1001", callID, base.Unix(), fmt.Sprintf("%s-%05d", callID, i+1)); err != nil {
				_ = tx.Rollback()
				t.Fatal(err)
			}
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	seed("marathon", TranscriptSegmentsMaxPerCall+2)
	seed("other", 3)

	// One more segment for the marathon call: 1503 rows for the call,
	// the trim keeps the newest 1500 (seeded 4..1502 + the fresh append).
	fresh := seg("1001", "marathon", "out", "+4930", "fresh", base.Add(time.Hour))
	if err := s.Append(ctx, fresh); err != nil {
		t.Fatal(err)
	}

	count := func(callID string) int {
		t.Helper()
		var n int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM call_transcripts WHERE owner = '1001' AND call_id = ?`, callID,
		).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if got := count("marathon"); got != TranscriptSegmentsMaxPerCall {
		t.Errorf("marathon segments: %d (want the per-call cap %d)", got, TranscriptSegmentsMaxPerCall)
	}
	if got := count("other"); got != 3 {
		t.Errorf("the OTHER call must stay untouched by a marathon trim: %d rows", got)
	}
	// The survivors are the newest: the three oldest seeded rows are gone,
	// seed row #4 (the new oldest survivor) is present.
	for _, gone := range []string{"marathon-00001", "marathon-00002", "marathon-00003"} {
		var n int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM call_transcripts WHERE text = ?`, gone,
		).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("oldest row %q survived the per-call trim", gone)
		}
	}
	var n int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM call_transcripts WHERE text = 'marathon-00004'`,
	).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("seed row #4 must survive as the oldest row: %d", n)
	}
}

func TestTranscriptsAppendTrimsPastTheCap(t *testing.T) {
	s := newTranscripts(t)
	ctx := context.Background()
	base := time.Unix(1760000000, 0)
	// The flood spreads over four calls so the OWNER cap binds (a
	// single-call flood would stop at the per-call cap first — that
	// interaction is the per-call test above).
	calls := []string{"call-a", "call-b", "call-c", "call-d"}
	for i := range TranscriptSegmentsMaxPerExtension + 10 {
		text := string(rune('a'+i%26)) + "-seg"
		call := calls[i%len(calls)]
		if err := s.Append(ctx, seg("1001", call, "out", "+4930", text, base.Add(time.Duration(i)*time.Second))); err != nil {
			t.Fatal(err)
		}
	}
	// Count rows directly: Recent deliberately reads a narrower window
	// than the cap, so it is the wrong lens for the trim.
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM call_transcripts`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != TranscriptSegmentsMaxPerExtension {
		t.Errorf("stored segments: %d (want the cap %d)", total, TranscriptSegmentsMaxPerExtension)
	}
	// The SURVIVORS are the newest: the oldest ten segments are gone.
	got, err := s.Recent(ctx, domain.MustParseExtension("1001"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Lines) == 0 {
		t.Fatalf("trimmed store still reassembles: %+v", got)
	}
}
