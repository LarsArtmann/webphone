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
// the swap verbatim. Writes carry the D7 awaitOwnWrite: the History
// panel re-fetches the note right behind the SSE nudge, so the
// read-your-writes beat (the engine store's old synchronous fold) is
// the honest contract.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// DeleteCallNote removes one call's note slot outright (the explicit
// DELETE lane — the empty-body clear keeps the slot, the delete frees
// the per-owner cap). A missing slot answers the ErrNoteMissing
// Rejection (the snippet-delete precedent; the handler maps it 404).
type DeleteCallNote struct {
	*command.BasicCommand
	Owner  string
	CallID string
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
	event.Type(metaengine.EventTypeName(domain.CallNoteDeleted{})),
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
	case event.Type(metaengine.EventTypeName(domain.CallNoteDeleted{})):
		fact, err := event.DecodePayloadAuto[domain.CallNoteDeleted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.note_fold", "decode CallNoteDeleted")
		}
		delete(s.Slots, fact.Key.CallID)
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
	// An unchanged body is not a revision: the note editors autosave on
	// every typing pause, and the journal must record CHANGES, not
	// echoes — a same-body save answers as the settled state (the
	// handler still returns its 204; the idempotent no-op is the same
	// verdict the delete lane gives a vanished slot). Existing slot
	// only: a first save keeps today's slot-minting behavior.
	if _, exists := s.Slots[cmd.CallID]; exists && prev.Body == cmd.Body {
		return nil, nil
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
	// The same echo guard the note body rides: the disposition chips are
	// click-shaped, and the API contract already promises "the same
	// verdict again is an inert 204" — the journal must record CHANGES,
	// not re-clicks (each event would bump UpdatedAt and fire the notes
	// SSE nudge for a fact that did not change). Existing slot only: a
	// first disposition keeps minting the slot.
	if _, exists := s.Slots[cmd.CallID]; exists && prev.Disposition == cmd.Disposition {
		return nil, nil
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

func decideDeleteNote(s notesState, ver event.Version, cmd DeleteCallNote) ([]event.Event, error) {
	if _, ok := s.Slots[cmd.CallID]; !ok {
		// The sentinel wrap (the cap verdict's pattern): the handler's
		// errors.Is mapping keys on readmodels.ErrNoteMissing while the
		// code stays greppable.
		return nil, errorfamily.WrapRejectionf(readmodels.ErrNoteMissing, "cqrs.note_missing", "note slot %s", cmd.CallID)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.CallNoteDeleted{})), cmd.StreamID(), NotesStreamType, ver.Increment(), domain.CallNoteDeleted{
		Key: domain.CallNoteKey{Owner: cmd.Owner, CallID: cmd.CallID},
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.note_delete", "build CallNoteDeleted")
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
	if err := system.RegisterCommand[*DeleteCallNote, notesState](sys, "note.delete",
		func(_ context.Context, cmd *DeleteCallNote) system.Op[notesState] {
			return system.Execute[notesState](context.Background(), cmd.StreamID(), NotesStreamType,
				func(state notesState, ver event.Version) ([]event.Event, error) {
					return decideDeleteNote(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.note_register", "register note.delete")
	}
	return nil
}

// notesKeyInput is the projection's key-extraction input.
type notesKeyInput struct {
	Key domain.CallNoteKey
}

// notesCollection names the system engine's note projection.
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
		metaengine.OnRecord(domain.CallNoteDeleted{}, metaengine.Remove[readmodels.CallNoteView]()),
		metaengine.FilterOnField[readmodels.CallNoteView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.CallNoteView]("updated_at", false),
		metaengine.Volume(humanScaleVolume),
	)
}

// NotesDomainDecl is the call-notes aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and
// panics loudly).
func NotesDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterNotes(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register call-notes aggregate: %v", err)) //cqrs-lint:ignore(C009) boot-time fail-fast on a composition miswire (DO-1; the decl comment owns the why)
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(notesProjection())},
		Events:      notesStreamEventTypes,
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Notes is the dispatch-backed post-call-note store: the contract
// readmodels' gate declares (nil-when-absent owner-scoped Get,
// body upsert merging the verdict, disposition upsert merging the
// prose, per-owner cap for NEW calls as the 409 Conflict).
type Notes struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.CallNoteView]
	vers   streamVersions
	mu     sync.Mutex
}

var _ readmodels.NotesGate = (*Notes)(nil)

// NewNotes builds the backend over a System.
func NewNotes(sys *system.System) *Notes {
	return &Notes{sys: sys, reader: metaengine.NewReader[readmodels.CallNoteView](sys.MetaEngine(), notesCollection), vers: streamVersions{}}
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

// noteDispatchError keeps the decision's coded verdicts verbatim (the
// 409 Conflict cap contract AND the missing-slot Rejection the delete
// door's 404 mapping keys on) and codes every other dispatch failure
// as the seam's Infrastructure wrap — all three doors share it.
func noteDispatchError(err error, op string) error {
	if errors.Is(err, readmodels.ErrNoteListFull) || errors.Is(err, readmodels.ErrNoteMissing) {
		return err
	}
	return errorfamily.WrapInfrastructuref(err, NoteDispatchCode, "dispatch %s", op)
}

// upsertSlot is the TWO upsert lanes' shared discipline (one home):
// lock, pre-read, echo-skip, dispatch under the same lock, then the D7
// read-your-writes await. isEcho names what an unchanged slot looks
// like; dispatch runs the command; op names the wrap for a dispatch
// failure. The bool answers the gate contract: true only when a fresh
// event was dispatched (echoes and read failures answer false).
func (s *Notes) upsertSlot(ctx context.Context, owner domain.Extension, callID, op string, isEcho func(*readmodels.CallNoteView) bool, dispatch func() error) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	view, err := s.Get(ctx, owner, callID)
	if err != nil {
		return false, err
	}
	if view != nil && isEcho(view) {
		return false, nil
	}
	if err := dispatch(); err != nil {
		return false, noteDispatchError(err, op)
	}
	awaitOwnWrite(ctx, s.sys, NotesStreamType, owner.String(), notesCollection, s.vers)
	return true, nil
}

// Set upserts the owner's note for one call; the decision keeps an
// existing disposition. A NEW call past the cap refuses with the
// ErrNoteListFull Conflict contract. The D7 await keeps the read-your-writes
// beat (the History panel re-fetches right behind the SSE nudge).
//
// The returned bool names whether the slot changed: the editors
// autosave on every typing pause, and a save whose body already is the
// slot's body is an echo (what a re-mounted editor PUTs right after
// its GET). Echoes skip the dispatch entirely; the decider's same-body
// guard stays the authority under races (a pre-read that races a
// concurrent write at worst dispatches a save the decider then no-ops,
// which is the pre-echo behavior, not a new one). A first save always
// changes (the slot mints, whatever the body).
func (s *Notes) Set(ctx context.Context, owner domain.Extension, callID, body string) (bool, error) {
	return s.upsertSlot(ctx, owner, callID, "note.save",
		func(view *readmodels.CallNoteView) bool { return view.Body == body },
		func() error {
			return dispatchOwnerCommand(ctx, s.sys, "note.save", owner.String(),
				func(base *command.BasicCommand) *SaveCallNote {
					return &SaveCallNote{BasicCommand: base, Owner: owner.String(), CallID: callID, Body: body, At: time.Now()}
				})
		})
}

// SetDisposition records the operator's verdict on the same slot; the
// decision keeps an existing note body. It shares the note's cap.
//
// The returned bool is Set's symmetry: whether the slot CHANGED. The
// chips are click-shaped, so a same-verdict re-click is an echo (what
// a double-tap or a re-mounted editor POSTs right after its read).
// Echoes skip the dispatch entirely; the decider's same-disposition
// guard stays the authority under races. A first disposition always
// changes (the slot mints).
func (s *Notes) SetDisposition(ctx context.Context, owner domain.Extension, callID string, disposition domain.Disposition) (bool, error) {
	return s.upsertSlot(ctx, owner, callID, "note.disposition_save",
		func(view *readmodels.CallNoteView) bool { return view.Disposition == disposition.String() },
		func() error {
			return dispatchOwnerCommand(ctx, s.sys, "note.disposition_save", owner.String(),
				func(base *command.BasicCommand) *SaveCallDisposition {
					return &SaveCallDisposition{BasicCommand: base, Owner: owner.String(), CallID: callID, Disposition: disposition, At: time.Now()}
				})
		})
}

// Delete removes one call's note slot outright: the per-owner cap
// counts open slots, so freeing one is the 409 remedy made real. A
// missing slot answers the ErrNoteMissing Rejection (the handler's
// 404). The D7 await keeps the read-your-writes beat for the follow-up
// read.
func (s *Notes) Delete(ctx context.Context, owner domain.Extension, callID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "note.delete", owner.String(),
		func(base *command.BasicCommand) *DeleteCallNote {
			return &DeleteCallNote{BasicCommand: base, Owner: owner.String(), CallID: callID}
		}); err != nil {
		return noteDispatchError(err, "note.delete")
	}
	awaitOwnWrite(ctx, s.sys, NotesStreamType, owner.String(), notesCollection, s.vers)
	return nil
}

// Count answers how many note slots the owner holds (the Settings
// cap-visibility row reads it).
func (s *Notes) Count(ctx context.Context, owner domain.Extension) (int, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithLimit(notesMaxPerOwner+1),
	)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.note_count", "list note slots for %s", owner.String())
	}
	return len(views), nil
}

// SearchBodies answers the owner's note slots whose body contains the
// needle (ASCII case-insensitive contains — the transcript Search
// contract), newest first. The scan walks the owner's whole slot set,
// which the per-owner cap already bounds, mirroring how the transcript
// search reads the full owner corpus rather than a window.
func (s *Notes) SearchBodies(ctx context.Context, owner domain.Extension, needle string) ([]readmodels.CallNoteView, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("updated_at", true),
		metaengine.WithLimit(0),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.note_search", "search note bodies of %s", owner.String())
	}
	haystack := strings.ToLower(needle)
	out := make([]readmodels.CallNoteView, 0, len(views))
	for _, v := range views {
		if strings.Contains(strings.ToLower(v.Body), haystack) {
			out = append(out, v)
		}
	}
	return out, nil
}
