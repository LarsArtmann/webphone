package readmodels

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// TranscriptSegmentsMaxPerExtension bounds one extension's transcript
// history (≈5.5 h of continuous speech at the 4 s segment cadence).
// The cap is the durability story like ContactsMaxPerExtension: Recent
// renders the newest window, so trimming the oldest rows costs the
// deep past, never the current view.
const TranscriptSegmentsMaxPerExtension = 5000

// TranscriptSegmentsMaxPerCall bounds ONE call's transcript (≈100 min of
// continuous speech at the 4 s cadence): a single marathon call must not
// eat the owner's whole budget before the other calls' history. Same trim
// shape as the owner cap — the interaction (3+ long calls trimming each
// other's oldest rows) is intended, the per-call floor keeps every call
// reassemblable.
const TranscriptSegmentsMaxPerCall = 1500

// transcriptsGlobalCeiling is the collection-global safety net ABOVE the
// per-owner grouped cap (see dialWindowsGlobalCeiling for the rationale):
// sized for hundreds of owners at the full per-extension budget without
// ever competing with the grouped bound.
const transcriptsGlobalCeiling = 1_000_000

// transcriptFetchWindow is the segment read window behind Recent and
// Search: wide enough that the maxCalls newest calls reassemble whole
// even when one call alone produced hundreds of segments.
const transcriptFetchWindow = 2000

// TranscriptSegmentView is the engine-side projection of one stored
// transcript segment. Seq is the append order (nanos at the seam): the
// engine's single sort field reproducing the SQL stream's
// `started_at DESC, rowid DESC` — live segments arrive in spoken
// order, so append order IS the display order.
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the retired SQL row verbatim (parity port).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema; CallID is the opaque island token (strong-id wire ruling)
type TranscriptSegmentView struct {
	Key       domain.TranscriptSegmentKey `json:"key"`
	ID        string                      `json:"id"`
	Owner     string                      `json:"owner"`
	CallID    string                      `json:"call_id"`
	Direction domain.Direction            `json:"direction"`
	Remote    string                      `json:"remote"`
	StartedAt int64                       `json:"started_at"`
	Seq       int64                       `json:"seq"`
	Text      string                      `json:"text"`
}

// transcriptsCollection names the engine collection (query name).
const transcriptsCollection = "call_transcripts"

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type transcriptInput struct{ ID string }

// transcriptsQuery declares the transcript read model: an append-only
// keyed projection (segments are never updated in place) removed by
// the trim/delete/retention expunge facts. Filtered by owner, call,
// and seq (the sweep), sorted by append order.
var transcriptsQuery = metaengine.Query[transcriptInput, TranscriptSegmentView](transcriptsCollection,
	metaengine.OnRecord(domain.TranscriptAppended{},
		func(_ record.Record, e domain.TranscriptAppended) (domain.TranscriptSegmentKey, TranscriptSegmentView) {
			return e.Key, TranscriptSegmentView{
				Key:       e.Key,
				ID:        e.Key.ID,
				Owner:     e.Owner,
				CallID:    e.CallID,
				Direction: e.Direction,
				Remote:    e.Remote,
				StartedAt: e.StartedAt.Unix(),
				Seq:       e.Seq,
				Text:      e.Text,
			}
		}),
	metaengine.OnRecord(domain.TranscriptExpunged{}, metaengine.Remove[TranscriptSegmentView]()),
	metaengine.FilterOnField[TranscriptSegmentView]("owner", metaengine.FilterEq),
	metaengine.FilterOnField[TranscriptSegmentView]("call_id", metaengine.FilterEq),
	metaengine.FilterOnField[TranscriptSegmentView]("seq", metaengine.FilterLt),
	metaengine.SortOnField[TranscriptSegmentView]("seq", true),
	metaengine.WithCap(transcriptsGlobalCeiling, "seq"),
	metaengine.WithCapPerGroup("owner", TranscriptSegmentsMaxPerExtension),
	metaengine.Volume(humanScaleVolume),
)

// Transcripts is the engine-backed transcript store: the same contract
// the SQL store held (Append with both scalar-bounded trims, Recent /
// Search / Export reassembly, DeleteCall, plus the engine-count and
// retention arms the metrics and sweeper ride).
type Transcripts struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[TranscriptSegmentView]
	agg    metaengine.AggregateReader
	mu     sync.Mutex
}

func newTranscripts(models *metaengine.Store, agg metaengine.AggregateReader) *Transcripts {
	return &Transcripts{
		models: models,
		reader: metaengine.NewReader[TranscriptSegmentView](models, transcriptsCollection),
		agg:    agg,
	}
}

// Append stores one segment and enforces both retention bounds: the
// per-EXTENSION cap is a declarative WithCapPerGroup enforced via
// EnforceCaps (metaengine #114, ADR-0164 — it replaced this file's
// hand-rolled owner loop, the 2026-11-04 T07 ruling's reopen trigger);
// the per-CALL floor stays a hand-rolled loop because metaengine's
// grouped cap declares ONE group per collection and the per-call bound
// needs a second, nested grouping (owner+call).
func (s *Transcripts) Append(ctx context.Context, seg domain.CallTranscriptSegment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.TranscriptAppended{}), domain.TranscriptAppended{
		Key:       domain.GenerateTranscriptSegmentKey(),
		Owner:     seg.Owner.String(),
		CallID:    seg.CallID,
		Direction: seg.Direction,
		Remote:    seg.Remote,
		StartedAt: seg.StartedAt,
		Seq:       time.Now().UnixNano(),
		Text:      seg.Text,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.transcript_append", "append transcript")
	}
	if _, err := s.models.EnforceCaps(ctx); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.transcript_cap", "enforce per-owner transcript caps")
	}
	return s.trimCall(ctx, seg.Owner.String(), seg.CallID)
}

// trimCall expunges the call's oldest rows past the per-call floor: scan
// the newest cap+1 by seq and expunge everything the window cannot hold,
// repeating until the call fits — one bounded scan sees only cap+1 rows,
// so a marathon call converges over passes (each removes at least one
// row). The per-EXTENSION bound is declarative (WithCapPerGroup above);
// this loop exists because the per-call floor needs a second, nested
// grouping metaengine does not declare yet.

// brandid-lint:ignore(BD003) opaque island Call-ID passthrough into the engine filter (strong-id wire ruling)
func (s *Transcripts) trimCall(ctx context.Context, owner, callID string) error {
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("owner", metaengine.FilterEq, owner),
			metaengine.WithFilter("call_id", metaengine.FilterEq, callID),
			metaengine.WithSort("seq", true),
			metaengine.WithLimit(TranscriptSegmentsMaxPerCall+1),
		)
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.transcript_trim", "scan call trim window")
		}
		if len(views) <= TranscriptSegmentsMaxPerCall {
			return nil
		}
		for _, v := range views[TranscriptSegmentsMaxPerCall:] {
			if err := s.expunge(ctx, v.ID); err != nil {
				return err
			}
		}
	}
}

// brandid-lint:ignore(BD003) engine key passthrough into the expunge fact (strong-id wire ruling)
func (s *Transcripts) expunge(ctx context.Context, id string) error {
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.TranscriptExpunged{}), domain.TranscriptExpunged{
		Key: domain.TranscriptSegmentKey{ID: id},
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.transcript_trim", "expunge transcript segment")
	}
	return nil
}

// Recent reassembles the owner's newest transcribed calls: segments are
// fetched newest-first, grouped by call (newest call first, its lines
// in spoken order), and the group list is capped at maxCalls.
func (s *Transcripts) Recent(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, owner.String(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(transcriptFetchWindow),
	)
	if err != nil {
		return nil, err
	}
	return reassembleTranscripts(rows, maxCalls), nil
}

// Search finds the owner's transcript segments whose text contains the
// needle (ASCII case-insensitive, literal — the SQL LIKE's
// metacharacter escaping needs no engine equivalent for a Contains
// scan) and reassembles the matches into call groups — newest match
// first, MATCHING lines only: the silent half of a hit call is not the
// searcher's answer, and the History hint says what is shown.
func (s *Transcripts) Search(ctx context.Context, owner domain.Extension, needle string, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, owner.String(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(transcriptFetchWindow),
	)
	if err != nil {
		return nil, err
	}
	haystack := strings.ToLower(needle)
	matches := make([]domain.CallTranscriptSegment, 0, len(rows))
	for _, seg := range rows {
		if strings.Contains(strings.ToLower(seg.Text), haystack) {
			matches = append(matches, seg)
		}
	}
	return reassembleTranscripts(matches, maxCalls), nil
}

// RecentAll reassembles the newest transcribed calls across EVERY
// owner (the admin surface's cross-line call inventory — the deliberate
// seam without the owner filter; owner-scoping stays the rule for every
// other read). Grouped and capped exactly like Recent.
func (s *Transcripts) RecentAll(ctx context.Context, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, "all owners",
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(transcriptFetchWindow),
	)
	if err != nil {
		return nil, err
	}
	return reassembleTranscripts(rows, maxCalls), nil
}

// Export reassembles the owner's WHOLE transcript history for the
// settings export zip: the read window is the owner cap (not the render
// window), so every stored call arrives — newest first, lines in spoken
// order, the group list capped at maxCalls.
func (s *Transcripts) Export(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, owner.String(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(TranscriptSegmentsMaxPerExtension),
	)
	if err != nil {
		return nil, err
	}
	return reassembleTranscripts(rows, maxCalls), nil
}

// fetchSegments reads segments newest-first behind the given scan
// options — the ONE home for the transcript read so Recent, RecentAll,
// Search and Export cannot drift apart. scope names the read's reach
// in the failure message (one owner, or all owners for the admin
// seam); the owner FILTER is an explicit scan option each caller
// passes (or omits, deliberately, on the admin path).
func (s *Transcripts) fetchSegments(ctx context.Context, scope string, opts ...metaengine.ScanOption) ([]domain.CallTranscriptSegment, error) {
	views, err := s.reader.Scan(ctx, opts...)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.transcript_scan", "scan transcripts for %s", scope)
	}
	rows := make([]domain.CallTranscriptSegment, 0, len(views))
	for _, v := range views {
		rows = append(rows, domain.CallTranscriptSegment{
			Owner:     domain.MustParseExtension(v.Owner),
			CallID:    v.CallID,
			Direction: v.Direction,
			Remote:    v.Remote,
			StartedAt: time.Unix(v.StartedAt, 0),
			Text:      v.Text,
		})
	}
	return rows, nil
}

// Call reassembles ONE call's transcript for one owner — the
// summary-draft seam's read. Nil (no error) when the owner holds no
// such call: the caller answers 404 and never another owner's words.
// brandid-lint:ignore(BD003) opaque island Call-ID, verbatim passthrough (strong-id wire ruling)
func (s *Transcripts) Call(ctx context.Context, owner domain.Extension, callID string) (*domain.CallTranscript, error) { //nolint:branching-flow // the opaque island Call-ID, verbatim passthrough (see DeleteCall)
	s.mu.Lock()
	defer s.mu.Unlock()

	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithFilter("call_id", metaengine.FilterEq, callID),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(TranscriptSegmentsMaxPerCall),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.transcript_call", "scan call transcript for %s", owner.String())
	}
	rows := make([]domain.CallTranscriptSegment, 0, len(views))
	for _, v := range views {
		rows = append(rows, domain.CallTranscriptSegment{
			Owner:     owner,
			CallID:    v.CallID,
			Direction: v.Direction,
			Remote:    v.Remote,
			StartedAt: time.Unix(v.StartedAt, 0),
			Text:      v.Text,
		})
	}
	groups := reassembleTranscripts(rows, 1)
	if len(groups) == 0 {
		return nil, nil
	}
	return &groups[0], nil
}

// DeleteCall erases one call's transcript for one owner (the end-user
// leg of the retention story). Idempotent — deleting an unknown call
// removes zero rows, which is the same success as deleting an
// already-deleted one.
// brandid-lint:ignore(BD003) opaque island Call-ID, verbatim passthrough (strong-id wire ruling)
func (s *Transcripts) DeleteCall(ctx context.Context, owner domain.Extension, callID string) (int64, error) { //nolint:branching-flow // the opaque island Call-ID, verbatim passthrough (see domain.CallTranscriptSegment.CallID)
	s.mu.Lock()
	defer s.mu.Unlock()

	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithFilter("call_id", metaengine.FilterEq, callID),
		metaengine.WithLimit(TranscriptSegmentsMaxPerCall),
	)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "readmodels.transcript_delete", "scan call transcript")
	}
	for _, v := range views {
		if err := s.expunge(ctx, v.ID); err != nil {
			return 0, err
		}
	}
	return int64(len(views)), nil
}

// Count reports the total stored transcript segments across all owners
// (the metrics surface's webphone_transcripts_total). Served by the
// engine's native COUNT aggregate — no row materialization.
func (s *Transcripts) Count(ctx context.Context) (int64, error) {
	n, err := s.agg.Aggregate(ctx, transcriptsCollection, metaengine.AggregateCount, "", nil)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "readmodels.transcript_count", "count transcripts")
	}
	return int64(n), nil
}

// sweepPageSize bounds one retention scan/expunge batch; the loop
// converges because every batch removes what it scanned.
const transcriptSweepPageSize = 1000

// SweepBefore deletes transcript segments appended before the cutoff
// (the retention pass's engine arm). Transcript content carries no
// blobs — nothing to unlink, the count is the whole answer.
func (s *Transcripts) SweepBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var removed int64
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("seq", metaengine.FilterLt, cutoff.UnixNano()),
			metaengine.WithSort("seq", false),
			metaengine.WithLimit(transcriptSweepPageSize),
		)
		if err != nil {
			return removed, errorfamily.WrapInfrastructuref(err, "readmodels.transcript_sweep", "scan expirable transcripts")
		}
		if len(views) == 0 {
			return removed, nil
		}
		for _, v := range views {
			if err := s.expunge(ctx, v.ID); err != nil {
				return removed, err
			}
			removed++
		}
	}
}

// reassembleTranscripts groups newest-first segments into whole calls:
// newest call first, each call's lines in spoken order, the group list
// capped at maxCalls. Segments of already-included calls keep landing
// (the oldest included call must reassemble whole); only NEW groups are
// refused past the cap. Lines APPEND in fetch order (newest-first) and
// one finishing reverse per group restores spoken order — the naive
// prepend-per-segment was O(n²) on long calls (a 1500-segment marathon
// rebuilt its line slice for every segment). Ported verbatim from the
// retired SQL store with the metaengine swap.
func reassembleTranscripts(rows []domain.CallTranscriptSegment, maxCalls int) []domain.CallTranscript {
	groups := make([]domain.CallTranscript, 0, maxCalls)
	byCall := make(map[string]int, maxCalls)
	for _, seg := range rows {
		idx, seen := byCall[seg.CallID]
		if !seen {
			if len(groups) >= maxCalls {
				continue
			}
			idx = len(groups)
			byCall[seg.CallID] = idx
			groups = append(groups, domain.CallTranscript{
				Owner:     seg.Owner,
				CallID:    seg.CallID,
				Direction: seg.Direction,
				Remote:    seg.Remote,
				StartedAt: seg.StartedAt,
			})
		}
		group := &groups[idx] //nolint:branching-flow // idx comes only from byCall entries written as len(groups) at append time, or freshly the same way — an invariant the analyzer cannot see through the map
		group.Lines = append(group.Lines, seg.Text)
		if group.StartedAt.Before(seg.StartedAt) {
			group.StartedAt = seg.StartedAt
		}
	}
	for i := range groups {
		// The store yields oldest-first per call; the UI reads top-down, so
		// each call's lines are presented newest-first. slices.Reverse is the
		// stdlib home for this (the hand-rolled two-pointer helper it replaces
		// carried a nolint purely for its own arithmetic).
		slices.Reverse(groups[i].Lines)
	}
	return groups
}
