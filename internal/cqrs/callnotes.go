// The call-notes aggregate (SUPERB Phase 4, A28): the post-call note
// slot as a per-owner stream (one slot per (owner, call)). The stream
// is per OWNER — not per (owner, call) — because the 2000 cap must
// REJECT from the DECISION (hand-written content is never silently
// expunged; a projection-side trim cap would violate the data-loss
// ruling), and a per-call stream cannot see the owner's slot count.
// The body/disposition MERGE semantics (a note save keeps the verdict,
// a verdict save keeps the prose) moved out of the store's pre-read
// into the decision: the decider simply knows its slot. The
// ErrNoteListFull CONFLICT classification (the only true 409) survives
// the swap verbatim. No D7 await: the HTTP surfaces answer 204 and the
// SSE nudge is payload-less, so nothing re-reads this projection in
// the same request.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// NotesStreamType keys per-owner call-note streams (one stream holds
// the owner's every noted call; the slot is (owner, call)).
const NotesStreamType = "call_notes"

// notesMaxPerOwner mirrors readmodels.NotesMaxPerExtension at the
// aggregate.
const notesMaxPerOwner = readmodels.NotesMaxPerExtension

// SaveCallNote upserts the owner's freeform note for one call: an
// existing disposition rides along untouched (the merge semantics live
// in the decision). A NEW call's note past the cap answers the 409
// Conflict contract.
type SaveCallNote struct {
	*command.BasicCommand
	Owner  string
	CallID string
	Body   string
	At     time.Time // stamped at the service seam; the decision stays deterministic
}

// SaveCallDisposition records the operator's post-call verdict on the
// same slot: an existing note body rides along untouched.
type SaveCallDisposition struct {
	*command.BasicCommand
	Owner       string
	CallID      string
	Disposition domain.Disposition
	At          time.Time // stamped at the service seam; the decision stays deterministic
}

// noteSlot is one call's note-slot state.
type noteSlot struct {
	Body        string
	Disposition domain.Disposition
	UpdatedAt   int64 // unix seconds — the view's numeric order
}

// notesState is the owner's note slots keyed by call id.
type notesState struct {
	Slots map[string]noteSlot
}

func (notesState) initial() notesState { return notesState{Slots: map[string]noteSlot{}} }

// notesStreamEventTypes is the aggregate's stream vocabulary: the
// domain decl's Events entry AND the StrictApply guard list (an unknown
// type on the stream fails replay loudly instead of silently skipping).
var notesStreamEventTypes = []event.Type{
	event.Type(metaengine.EventTypeName(domain.CallNoteSaved{})),
}

func applyNotesState(s notesState, evt event.Event) (notesState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.CallNoteSaved{})):
		fact, err := event.DecodePayloadAuto[domain.CallNoteSaved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.note_fold", "decode CallNoteSaved")
		}
		s.Slots[fact.CallID] = noteSlot{
			Body:        fact.Body,
			Disposition: fact.Disposition,
			UpdatedAt:   fact.UpdatedAt.Unix(),
		}
	}
	return s, nil
}

// claimNoteSlot enforces the per-owner cap for a NEW call and returns
// the merged slot the save will snapshot (the old store's pre-read,
// answered from state).
func claimNoteSlot(s notesState, owner, callID string) (noteSlot, error) {
	if prev, ok := s.Slots[callID]; ok {
		return prev, nil
	}
	if len(s.Slots) >= notesMaxPerOwner {
		return noteSlot{}, errorfamily.WrapConflictf(readmodels.ErrNoteListFull, "cqrs.note_list_full",
			"%s at %d noted calls", owner, notesMaxPerOwner)
	}
	return noteSlot{}, nil
}

func decideSaveNote(s notesState, ver event.Version, cmd SaveCallNote) ([]event.Event, error) {
	prev, err := claimNoteSlot(s, cmd.Owner, cmd.CallID)
	if err != nil {
		return nil, err
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.CallNoteSaved{})), cmd.StreamID(), NotesStreamType, ver.Increment(), domain.CallNoteSaved{
		Key:         domain.CallNoteKey{Owner: cmd.Owner, CallID: cmd.CallID},
		Owner:       cmd.Owner,
		CallID:      cmd.CallID,
		Body:        cmd.Body,
		Disposition: prev.Disposition,
		UpdatedAt:   cmd.At,
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.note_save", "build CallNoteSaved")
	}
	return []event.Event{evt}, nil
}

func decideSaveDisposition(s notesState, ver event.Version, cmd SaveCallDisposition) ([]event.Event, error) {
	prev, err := claimNoteSlot(s, cmd.Owner, cmd.CallID)
	if err != nil {
		return nil, err
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.CallNoteSaved{})), cmd.StreamID(), NotesStreamType, ver.Increment(), domain.CallNoteSaved{
		Key:         domain.CallNoteKey{Owner: cmd.Owner, CallID: cmd.CallID},
		Owner:       cmd.Owner,
		CallID:      cmd.CallID,
		Body:        prev.Body,
		Disposition: cmd.Disposition,
		UpdatedAt:   cmd.At,
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.disposition_save", "build CallNoteSaved")
	}
	return []event.Event{evt}, nil
}

// RegisterNotes wires the call-notes aggregate into a System.
func RegisterNotes(sys *system.System) error {
	if err := system.RegisterDecider[notesState](sys, NotesStreamType, decider.Decider[notesState]{
		Initial: notesState{}.initial(),
		Apply:   decider.StrictApply(applyNotesState, notesStreamEventTypes),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.note_register", "register notes decider")
	}
	if err := system.RegisterCommand[*SaveCallNote, notesState](sys, "note.save",
		func(_ context.Context, cmd *SaveCallNote) system.Op[notesState] {
			return system.Execute[notesState](context.Background(), cmd.StreamID(), NotesStreamType,
				func(state notesState, ver event.Version) ([]event.Event, error) {
					return decideSaveNote(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.note_register", "register note.save")
	}
	if err := system.RegisterCommand[*SaveCallDisposition, notesState](sys, "note.disposition_save",
		func(_ context.Context, cmd *SaveCallDisposition) system.Op[notesState] {
			return system.Execute[notesState](context.Background(), cmd.StreamID(), NotesStreamType,
				func(state notesState, ver event.Version) ([]event.Event, error) {
					return decideSaveDisposition(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.note_register", "register note.disposition_save")
	}
	return nil
}

// notesKeyInput is the projection's key-extraction input (the same
// shape the readmodels engine declares).
type notesKeyInput struct {
	Key domain.CallNoteKey
}

// notesCollection names the system engine's projection (the same name
// the readmodels engine declares — different databases, one shape).
const notesCollection = "call_notes"

// notesProjection folds note facts into the SAME view shape the server
// reads today (readmodels.CallNoteView).
func notesProjection() metaengine.QueryDecl[notesKeyInput, readmodels.CallNoteView] {
	return metaengine.Query[notesKeyInput, readmodels.CallNoteView](notesCollection,
		metaengine.OnRecord(domain.CallNoteSaved{},
			func(_ record.Record, e domain.CallNoteSaved) (domain.CallNoteKey, readmodels.CallNoteView) {
				return e.Key, readmodels.CallNoteView{
					Owner:       e.Owner,
					CallID:      e.CallID,
					Body:        e.Body,
					Disposition: e.Disposition.String(),
					UpdatedAt:   e.UpdatedAt.Unix(),
				}
			}),
		metaengine.FilterOnField[readmodels.CallNoteView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.CallNoteView]("updated_at", false),
	)
}

// NotesDomainDecl is the call-notes aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and
// panics loudly).
func NotesDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterNotes(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register call-notes aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(notesProjection())},
		Events: notesStreamEventTypes,
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Notes is the dispatch-backed post-call-note store: the same contract
// the engine-backed store exposed (nil-when-absent owner-scoped Get,
// body upsert merging the verdict, disposition upsert merging the
// prose, per-owner cap for NEW calls as the 409 Conflict).
type Notes struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.CallNoteView]
	mu     sync.Mutex
}

var _ readmodels.NotesGate = (*Notes)(nil)

// NewNotes builds the backend over a System.
func NewNotes(sys *system.System) *Notes {
	return &Notes{sys: sys, reader: metaengine.NewReader[readmodels.CallNoteView](sys.MetaEngine(), notesCollection)}
}

// Get returns the owner's note for one call, or nil when none exists.
func (s *Notes) Get(ctx context.Context, owner domain.Extension, callID string) (*readmodels.CallNoteView, error) {
	view, found, err := s.reader.Get(ctx, domain.CallNoteKey{Owner: owner.String(), CallID: callID})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.note_get", "load note slot for %s", callID)
	}
	if !found || view.Owner != owner.String() {
		return nil, nil
	}
	return &view, nil
}

// Exported because the arch registry scan resolves the code-position
// identifier from the file's own const block (block form is what the
// scanner's declaration regex can see).
const (
	// NoteDispatchCode codes the dispatch transport failure shared by
	// the two write doors (the failing operation is THE DISPATCH, not a
	// door's save; the door stays greppable through the message's op).
	NoteDispatchCode = "cqrs.note_dispatch"
)

// noteDispatchError keeps the decision's coded cap verdict (the 409
// Conflict contract passes through verbatim) and codes every other
// dispatch failure as the seam's Infrastructure wrap — both write
// doors share it.
func noteDispatchError(err error, op string) error {
	if errors.Is(err, readmodels.ErrNoteListFull) {
		return err
	}
	return errorfamily.WrapInfrastructuref(err, NoteDispatchCode, "dispatch %s", op)
}

// Set upserts the owner's note for one call; the decision keeps an
// existing disposition. A NEW call past the cap refuses with the
// ErrNoteListFull Conflict contract.
func (s *Notes) Set(ctx context.Context, owner domain.Extension, callID, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "note.save", owner.String(),
		func(base *command.BasicCommand) *SaveCallNote {
			return &SaveCallNote{BasicCommand: base, Owner: owner.String(), CallID: callID, Body: body, At: time.Now()}
		}); err != nil {
		return noteDispatchError(err, "note.save")
	}
	return nil
}

// SetDisposition records the operator's verdict on the same slot; the
// decision keeps an existing note body. It shares the note's cap.
func (s *Notes) SetDisposition(ctx context.Context, owner domain.Extension, callID string, disposition domain.Disposition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "note.disposition_save", owner.String(),
		func(base *command.BasicCommand) *SaveCallDisposition {
			return &SaveCallDisposition{BasicCommand: base, Owner: owner.String(), CallID: callID, Disposition: disposition, At: time.Now()}
		}); err != nil {
		return noteDispatchError(err, "note.disposition_save")
	}
	return nil
}
