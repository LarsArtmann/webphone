// The transcripts backend (SUPERB Phase 4, A27): the dispatch-backed
// transcript store — the ONE implementation of
// readmodels.TranscriptGate (Append with both retention bounds, the
// reassembly reads, the end-user delete, the metrics count, the
// retention sweep). The readmodels engine store that used to implement
// the gate beside it retired with the 2026-11-11 storage-ladder train A.
package cqrs

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// Transcripts is the dispatch-backed transcript store.
type Transcripts struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.TranscriptSegmentView]
	vers   streamVersions
	mu     sync.Mutex
}

var _ readmodels.TranscriptGate = (*Transcripts)(nil)

// NewTranscripts builds the backend over a System.
func NewTranscripts(sys *system.System) *Transcripts {
	return &Transcripts{
		sys:    sys,
		reader: metaengine.NewReader[readmodels.TranscriptSegmentView](sys.MetaEngine(), transcriptsCollection),
		vers:   streamVersions{},
	}
}

// dispatchCounting warms the stream's tail cache BEFORE dispatching (so
// the post-dispatch tail is exactly this command's batch), dispatches,
// and counts the committed events of matchType — the decision-authoritative
// count (expunges the decider actually emitted).
func (s *Transcripts) dispatchCounting(ctx context.Context, name, streamKey string,
	build func(*command.BasicCommand) command.Command,
	matchType func(event.Event) bool,
) (int, error) {
	if _, err := committedTail(ctx, s.sys, TranscriptsStreamType, streamKey, s.vers); err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_tail", "warm %s tail", streamKey)
	}
	if err := dispatchOwnerCommand(ctx, s.sys, name, streamKey, build); err != nil {
		return 0, err
	}
	tail, err := committedTail(ctx, s.sys, TranscriptsStreamType, streamKey, s.vers)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_tail", "read %s tail", streamKey)
	}
	count := 0
	for _, evt := range tail {
		if matchType(evt) {
			count++
		}
	}
	return count, nil
}

// Append stores one segment and enforces both retention bounds: the
// per-CALL floor answers from the decision (expunge-oldest batch); the
// per-OWNER grouped cap rides EnforceCaps on the system's projection
// store, exactly like the engine store did on its own.
func (s *Transcripts) Append(ctx context.Context, seg domain.CallTranscriptSegment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := TranscriptStreamKey(seg.Owner.String(), seg.CallID)
	segID := domain.GenerateTranscriptSegmentKey()
	if err := dispatchOwnerCommand(ctx, s.sys, "transcript.append", key,
		func(base *command.BasicCommand) *AppendTranscriptSegment {
			return &AppendTranscriptSegment{
				BasicCommand: base,
				Owner:        seg.Owner.String(),
				CallID:       seg.CallID,
				SegID:        segID.ID,
				Direction:    seg.Direction,
				Remote:       seg.Remote,
				StartedAt:    seg.StartedAt,
				Seq:          time.Now().UnixNano(),
				Text:         seg.Text,
			}
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.transcript_append", "dispatch transcript.append")
	}
	// Enforce the per-owner cap only once the projection worker has
	// folded this append: the EnforceCaps write path and the worker's
	// fold path take the store's fold lock and the system engine's connection
	// pool in OPPOSITE orders, so running them concurrently can
	// deadlock (pinned by the Phase-4 battery as an order-dependent
	// hang). After the await the worker is idle past our event; the
	// enforcement itself is best-effort under a bounded context — the
	// cap is the durability story, never an invariant, and the next
	// append retries (the file header's convergence posture).
	awaitOwnWrite(ctx, s.sys, TranscriptsStreamType, key, transcriptsCollection, s.vers)
	enforceCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := s.sys.MetaEngine().EnforceCaps(enforceCtx); err != nil {
		slog.Warn("transcript per-owner cap enforcement skipped (converges at the next append)", "err", err)
	}
	return nil
}

// Recent reassembles the owner's newest transcribed calls.
func (s *Transcripts) Recent(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, owner.String(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(transcriptFetchWindow),
	)
	if err != nil {
		return nil, err
	}
	return readmodels.ReassembleTranscripts(rows, maxCalls), nil
}

// Search finds the owner's segments whose text contains the needle and
// reassembles the matches into call groups (matching lines only).
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
	return readmodels.ReassembleTranscripts(matches, maxCalls), nil
}

// RecentAll reassembles the newest transcribed calls across EVERY owner
// (the admin surface's cross-line inventory).
func (s *Transcripts) RecentAll(ctx context.Context, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, "all owners",
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(transcriptFetchWindow),
	)
	if err != nil {
		return nil, err
	}
	return readmodels.ReassembleTranscripts(rows, maxCalls), nil
}

// Export reassembles the owner's WHOLE transcript history for the
// settings export zip (the read window is the owner cap).
func (s *Transcripts) Export(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error) {
	rows, err := s.fetchSegments(ctx, owner.String(),
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(readmodels.TranscriptSegmentsMaxPerExtension),
	)
	if err != nil {
		return nil, err
	}
	return readmodels.ReassembleTranscripts(rows, maxCalls), nil
}

// fetchSegments reads segments newest-first behind the given scan
// options — the ONE home for the transcript read so Recent, RecentAll,
// Search and Export cannot drift apart. scope names the read's reach
// in the failure message (one owner, or all owners for the admin
//
//	seam); the owner FILTER is an explicit scan option each caller
//	passes (or omits, deliberately, on the admin path).
func (s *Transcripts) fetchSegments(ctx context.Context, scope string, opts ...metaengine.ScanOption) ([]domain.CallTranscriptSegment, error) {
	//cqrs-lint:ignore(F031) every caller passes an explicit metaengine.WithLimit; the variadic forwarding hides it (2026-11-11 triage; re-placed after the ladder re-home)
	views, err := s.reader.Scan(ctx, opts...)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_scan", "scan transcripts for %s", scope)
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

// Call reassembles ONE call's transcript for one owner (the
// summary-draft seam's read). Nil (no error) when the owner holds no
// such call.
func (s *Transcripts) Call(ctx context.Context, owner domain.Extension, callID string) (*domain.CallTranscript, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithFilter("call_id", metaengine.FilterEq, callID),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(readmodels.TranscriptSegmentsMaxPerCall),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_call", "scan call transcript for %s", owner.String())
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
	groups := readmodels.ReassembleTranscripts(rows, 1)
	if len(groups) == 0 {
		return nil, nil
	}
	return &groups[0], nil
}

// DeleteCall erases one call's transcript for one owner; the removed
// count is decision-authoritative (the expunge events the decider
// emitted). Idempotent: an unknown call expunges nothing, which is the
// same success.
func (s *Transcripts) DeleteCall(ctx context.Context, owner domain.Extension, callID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expungeType := event.Type(metaengine.EventTypeName(domain.TranscriptExpunged{}))
	count, err := s.dispatchCounting(ctx, "transcript.delete_call", TranscriptStreamKey(owner.String(), callID),
		func(base *command.BasicCommand) command.Command {
			return &DeleteCallTranscript{BasicCommand: base, Owner: owner.String(), CallID: callID}
		},
		func(evt event.Event) bool { return evt.Type() == expungeType },
	)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_delete", "dispatch transcript.delete_call")
	}
	return int64(count), nil
}

// Count reports the total stored transcript segments across all owners
// (webphone_transcripts_total) via the system engine's COUNT.
func (s *Transcripts) Count(ctx context.Context) (int64, error) {
	n, err := engineCount(ctx, s.sys, transcriptsCollection)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_count", "count transcripts")
	}
	return n, nil
}

// SweepBefore deletes segments appended before the cutoff (the
// retention pass's engine arm): the projection scan names the victims,
// grouped per call stream into expunge commands whose decisions count
// the real removals. The loop converges because every batch removes
// what it scanned (projection lag can only leave rows for the next
// pass, never lose them).
func (s *Transcripts) SweepBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expungeType := event.Type(metaengine.EventTypeName(domain.TranscriptExpunged{}))
	var removed int64
	for {
		pageRemoved := int64(0)
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("seq", metaengine.FilterLt, cutoff.UnixNano()),
			metaengine.WithSort("seq", false),
			metaengine.WithLimit(transcriptSweepPageSize),
		)
		if err != nil {
			return removed, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_sweep", "scan expirable transcripts")
		}
		if len(views) == 0 {
			return removed, nil
		}
		perStream := map[string][]string{}
		for _, v := range views {
			key := TranscriptStreamKey(v.Owner, v.CallID)
			perStream[key] = append(perStream[key], v.ID)
		}
		for key, ids := range perStream {
			count, err := s.dispatchCounting(ctx, "transcript.expunge", key,
				func(base *command.BasicCommand) command.Command {
					return &ExpungeTranscriptSegments{BasicCommand: base, IDs: ids}
				},
				func(evt event.Event) bool { return evt.Type() == expungeType },
			)
			if err != nil {
				return removed, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_sweep", "dispatch transcript.expunge")
			}
			pageRemoved += int64(count)
		}
		removed += pageRemoved
		// Termination guard (the enroll sweep's shape, PER PAGE): a page
		// whose expunges all applied as no-ops (the projection already
		// removed them) must end the walk — rescanning the same window
		// would spin forever.
		if pageRemoved == 0 {
			return removed, nil
		}
	}
}
