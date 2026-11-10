// The transcripts aggregate (SUPERB Phase 4, A27): one stream per
// (owner, call). The per-CALL floor (1500 segments) answers from the
// DECISION (append + expunge-oldest in one batch, the agent-audit
// pattern — the state IS the call's segments); the per-OWNER bound
// (5000) stays a PROJECTION-side grouped cap enforced via EnforceCaps
// after every append (the owner axis spans many streams, no single
// decision sees it — the nested-grouping gap stays untouched, per the
// plan). Journal/projection divergence from cap trims is the legacy
// posture: a replay re-grows rows until the next append's enforcement
// converges (the cap is the durability story, not an invariant).
package cqrs

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// TranscriptsStreamType keys per-(owner, call) transcript streams.
const TranscriptsStreamType = "transcripts"

// transcriptFloorPerCall mirrors readmodels.TranscriptSegmentsMaxPerCall
// at the aggregate (the per-call bound the decision enforces).
const transcriptFloorPerCall = readmodels.TranscriptSegmentsMaxPerCall

// transcriptsGlobalCeiling mirrors the readmodels declaration's
// collection-global safety net above the per-owner grouped cap.
const transcriptsGlobalCeiling = 1_000_000

// transcriptFetchWindow mirrors the readmodels read window behind
// Recent/Search (wide enough that the newest calls reassemble whole).
const transcriptFetchWindow = 2000

// transcriptSweepPageSize bounds one retention scan/expunge batch.
const transcriptSweepPageSize = 1000

// TranscriptStreamKey builds the per-(owner, call) stream key. Any
// non-empty UTF-8 string is a legal stream id; the key is never parsed
// back — both halves ride the facts explicitly.
func TranscriptStreamKey(owner, callID string) string {
	return owner + "|" + callID
}

// AppendTranscriptSegment appends one live segment; the decision emits
// the append plus the floor expunges when the call outgrew its bound.
// SegID and Seq are minted at the seam (the projection row identity and
// append order, the retired SQL rowid's role).
type AppendTranscriptSegment struct {
	*command.BasicCommand
	Owner     string
	CallID    string
	SegID     string
	Direction domain.Direction
	Remote    string
	StartedAt time.Time
	Seq       int64
	Text      string
}

// ExpungeTranscriptSegments removes the named segments (the retention
// sweep's arm); ids already gone from state are inert no-ops.
type ExpungeTranscriptSegments struct {
	*command.BasicCommand
	Owner  string
	CallID string
	IDs    []string
}

// DeleteCallTranscript erases the whole call transcript (the end-user
// delete); the decision expunges every segment the stream holds.
type DeleteCallTranscript struct {
	*command.BasicCommand
	Owner  string
	CallID string
}

// transcriptSeg is one segment's aggregate state.
type transcriptSeg struct {
	ID        string
	Direction domain.Direction
	Remote    string
	StartedAt int64 // unix seconds (the view's fidelity)
	Seq       int64
	Text      string
}

// transcriptState is the call's segments in seq order.
type transcriptState struct {
	Segs []transcriptSeg
}

func (transcriptState) initial() transcriptState { return transcriptState{} }

func applyTranscriptState(s transcriptState, evt event.Event) (transcriptState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.TranscriptAppended{})):
		fact, err := event.DecodePayloadAuto[domain.TranscriptAppended](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.transcript_fold", "decode TranscriptAppended")
		}
		s.Segs = append(s.Segs, transcriptSeg{
			ID: fact.Key.ID, Direction: fact.Direction, Remote: fact.Remote,
			StartedAt: fact.StartedAt.Unix(), Seq: fact.Seq, Text: fact.Text,
		})
		slices.SortStableFunc(s.Segs, func(a, b transcriptSeg) int {
			switch {
			case a.Seq < b.Seq:
				return -1
			case a.Seq > b.Seq:
				return 1
			}
			return 0
		})
	case event.Type(metaengine.EventTypeName(domain.TranscriptExpunged{})):
		fact, err := event.DecodePayloadAuto[domain.TranscriptExpunged](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.transcript_fold", "decode TranscriptExpunged")
		}
		s.Segs = slices.DeleteFunc(s.Segs, func(seg transcriptSeg) bool { return seg.ID == fact.Key.ID })
	}
	return s, nil
}

func transcriptEvent(fact any, streamID id.StreamID, ver event.Version) (event.Event, error) {
	evt, err := event.New(event.Type(metaengine.EventTypeName(fact)), streamID, TranscriptsStreamType, ver, fact, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.transcript_event", "build %T", fact)
	}
	return evt, nil
}

func decideAppendTranscript(s transcriptState, ver event.Version, cmd AppendTranscriptSegment) ([]event.Event, error) {
	fact := domain.TranscriptAppended{
		Key:       domain.TranscriptSegmentKey{ID: cmd.SegID},
		Owner:     cmd.Owner,
		CallID:    cmd.CallID,
		Direction: cmd.Direction,
		Remote:    cmd.Remote,
		StartedAt: cmd.StartedAt,
		Seq:       cmd.Seq,
		Text:      cmd.Text,
	}
	// The per-CALL floor: one append past the bound evicts the oldest
	// excess IN THE SAME BATCH — the stream never holds more than the
	// floor after a decision.
	kept := append([]transcriptSeg(nil), s.Segs...)
	kept = append(kept, transcriptSeg{ID: fact.Key.ID, Seq: fact.Seq})
	slices.SortStableFunc(kept, func(a, b transcriptSeg) int {
		switch {
		case a.Seq < b.Seq:
			return -1
		case a.Seq > b.Seq:
			return 1
		}
		return 0
	})
	events := make([]event.Event, 0, max(len(kept)-transcriptFloorPerCall, 0)+1)
	evt, err := transcriptEvent(fact, cmd.StreamID(), ver.Increment())
	if err != nil {
		return nil, err
	}
	events = append(events, evt)
	for _, seg := range kept[:max(len(kept)-transcriptFloorPerCall, 0)] {
		expunge, err := transcriptEvent(domain.TranscriptExpunged{Key: domain.TranscriptSegmentKey{ID: seg.ID}}, cmd.StreamID(), ver.Increment().Add(uint(len(events))))
		if err != nil {
			return nil, err
		}
		events = append(events, expunge)
	}
	return events, nil
}

func decideExpungeTranscript(s transcriptState, ver event.Version, cmd ExpungeTranscriptSegments) ([]event.Event, error) {
	present := make(map[string]struct{}, len(cmd.IDs))
	for _, seg := range s.Segs {
		present[seg.ID] = struct{}{}
	}
	events := make([]event.Event, 0, len(cmd.IDs))
	for _, segID := range cmd.IDs {
		if _, ok := present[segID]; !ok {
			continue
		}
		evt, err := transcriptEvent(domain.TranscriptExpunged{Key: domain.TranscriptSegmentKey{ID: segID}}, cmd.StreamID(), ver+event.Version(len(events))+1)
		if err != nil {
			return nil, err
		}
		events = append(events, evt)
	}
	return events, nil
}

func decideDeleteCallTranscript(s transcriptState, ver event.Version, cmd DeleteCallTranscript) ([]event.Event, error) {
	events := make([]event.Event, 0, len(s.Segs))
	for _, seg := range s.Segs {
		evt, err := transcriptEvent(domain.TranscriptExpunged{Key: domain.TranscriptSegmentKey{ID: seg.ID}}, cmd.StreamID(), ver+event.Version(len(events))+1)
		if err != nil {
			return nil, err
		}
		events = append(events, evt)
	}
	return events, nil
}

// RegisterTranscripts wires the transcripts aggregate into a System.
func RegisterTranscripts(sys *system.System) error {
	if err := system.RegisterDecider[transcriptState](sys, TranscriptsStreamType, decider.Decider[transcriptState]{
		Initial: transcriptState{}.initial(),
		Apply:   applyTranscriptState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.transcript_register", "register transcripts decider")
	}
	for _, reg := range []struct {
		name string
		wire func() error
	}{
		{"transcript.append", func() error {
			return system.RegisterCommand[*AppendTranscriptSegment, transcriptState](sys, "transcript.append",
				func(_ context.Context, cmd *AppendTranscriptSegment) system.Op[transcriptState] {
					return system.Execute[transcriptState](context.Background(), cmd.StreamID(), TranscriptsStreamType,
						func(state transcriptState, ver event.Version) ([]event.Event, error) {
							return decideAppendTranscript(state, ver, *cmd)
						})
				})
		}},
		{"transcript.expunge", func() error {
			return system.RegisterCommand[*ExpungeTranscriptSegments, transcriptState](sys, "transcript.expunge",
				func(_ context.Context, cmd *ExpungeTranscriptSegments) system.Op[transcriptState] {
					return system.Execute[transcriptState](context.Background(), cmd.StreamID(), TranscriptsStreamType,
						func(state transcriptState, ver event.Version) ([]event.Event, error) {
							return decideExpungeTranscript(state, ver, *cmd)
						})
				})
		}},
		{"transcript.delete_call", func() error {
			return system.RegisterCommand[*DeleteCallTranscript, transcriptState](sys, "transcript.delete_call",
				func(_ context.Context, cmd *DeleteCallTranscript) system.Op[transcriptState] {
					return system.Execute[transcriptState](context.Background(), cmd.StreamID(), TranscriptsStreamType,
						func(state transcriptState, ver event.Version) ([]event.Event, error) {
							return decideDeleteCallTranscript(state, ver, *cmd)
						})
				})
		}},
	} {
		if err := reg.wire(); err != nil {
			return errorfamily.WrapInfrastructuref(err, "cqrs.transcript_register", "register %s", reg.name)
		}
	}
	return nil
}

// transcriptsCollection names the system engine's projection (the same
// name the readmodels engine declares).
const transcriptsCollection = "call_transcripts"

// transcriptsKeyInput is the projection's key-extraction input (the
// same shape the readmodels engine declares).
type transcriptsKeyInput struct{ ID string }

// transcriptsProjection folds transcript facts into the SAME view
// shape the server reads today (readmodels.TranscriptSegmentView),
// caps declared exactly like the engine store's query (the per-owner
// grouped cap + the collection-global net).
func transcriptsProjection() metaengine.QueryDecl[transcriptsKeyInput, readmodels.TranscriptSegmentView] {
	return metaengine.Query[transcriptsKeyInput, readmodels.TranscriptSegmentView](transcriptsCollection,
		metaengine.OnRecord(domain.TranscriptAppended{},
			func(_ record.Record, e domain.TranscriptAppended) (domain.TranscriptSegmentKey, readmodels.TranscriptSegmentView) {
				return e.Key, readmodels.TranscriptSegmentView{
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
		metaengine.OnRecord(domain.TranscriptExpunged{}, metaengine.Remove[readmodels.TranscriptSegmentView]()),
		metaengine.FilterOnField[readmodels.TranscriptSegmentView]("owner", metaengine.FilterEq),
		metaengine.FilterOnField[readmodels.TranscriptSegmentView]("call_id", metaengine.FilterEq),
		metaengine.FilterOnField[readmodels.TranscriptSegmentView]("seq", metaengine.FilterLt),
		metaengine.SortOnField[readmodels.TranscriptSegmentView]("seq", true),
		metaengine.WithCap(transcriptsGlobalCeiling, "seq"),
		metaengine.WithCapPerGroup("owner", readmodels.TranscriptSegmentsMaxPerExtension),
	)
}

// TranscriptsDomainDecl is the transcripts aggregate's share of the
// system's DomainConfig.
func TranscriptsDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterTranscripts(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register transcripts aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(transcriptsProjection())},
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.TranscriptAppended{})),
			event.Type(metaengine.EventTypeName(domain.TranscriptExpunged{})),
		},
	}
}
