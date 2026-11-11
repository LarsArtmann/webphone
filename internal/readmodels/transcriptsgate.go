// The transcript seam's GATE + VIEW surface: the interface the server
// consumes, the engine projection row the journal backend folds, and
// the reassembly both backends shared. The engine-backed store that
// used to implement the gate here retired with the readmodels engine
// (2026-11-11 storage-ladder train A) — internal/cqrs/transcriptsstore.go
// is the one implementation now.
package readmodels

import (
	"context"
	"slices"
	"time"

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

// TranscriptSegmentView is the projection row ONE journal fold produces
// for a stored transcript segment. Seq is the append order (nanos at
// the seam): the projection's single sort field reproducing the
// retired SQL stream's `started_at DESC, rowid DESC` — live segments
// arrive in spoken order, so append order IS the display order.
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the retired SQL row verbatim (parity port).
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

// TranscriptGate is the transcript seam as the server surfaces see it
// (append with both retention bounds, the reassembly reads, the
// end-user delete, the metrics count, the retention sweep). The
// journal-backed service (internal/cqrs) is the implementation.
type TranscriptGate interface {
	Append(ctx context.Context, seg domain.CallTranscriptSegment) error
	Recent(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error)
	Search(ctx context.Context, owner domain.Extension, needle string, maxCalls int) ([]domain.CallTranscript, error)
	RecentAll(ctx context.Context, maxCalls int) ([]domain.CallTranscript, error)
	Export(ctx context.Context, owner domain.Extension, maxCalls int) ([]domain.CallTranscript, error)
	Call(ctx context.Context, owner domain.Extension, callID string) (*domain.CallTranscript, error)
	DeleteCall(ctx context.Context, owner domain.Extension, callID string) (int64, error)
	Count(ctx context.Context) (int64, error)
	SweepBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// ReassembleTranscripts groups newest-first segments into whole calls:
// newest call first, each call's lines in spoken order, the group list
// capped at maxCalls. Segments of already-included calls keep landing
// (the oldest included call must reassemble whole); only NEW groups are
// refused past the cap. Lines APPEND in fetch order (newest-first) and
// one finishing reverse per group restores spoken order — the naive
// prepend-per-segment was O(n²) on long calls (a 1500-segment marathon
// rebuilt its line slice for every segment). Ported verbatim from the
// retired SQL store with the metaengine swap.
func ReassembleTranscripts(rows []domain.CallTranscriptSegment, maxCalls int) []domain.CallTranscript {
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
