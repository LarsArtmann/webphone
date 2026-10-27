package readmodels

import (
	"context"
	"errors"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// NotesMaxPerExtension bounds one extension's noted calls (the
// contacts-cap posture: hand-written content must never be silently
// expunged, so the cap REJECTS a new note past the bound instead of
// trimming — the plan's trim wording is superseded by the
// data-loss ruling; errata noted in the plan doc).
const NotesMaxPerExtension = 2000

// ErrNoteListFull is returned when a NEW call's note would exceed the
// per-owner cap; re-saving an existing call's note stays open.
var ErrNoteListFull = errors.New("call note list full") //nolint:erraudit // sentinel: identity, not an error family (plan guardrail #4); classified via init registration

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrNoteListFull: errorfamily.Rejection,
	})
}

// CallNoteView is the engine-side projection of one call's review
// slot: the freeform note plus the operator's disposition verdict.
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema; CallID is the opaque island token (strong-id wire ruling)
type CallNoteView struct {
	Owner       string `json:"owner"`
	CallID      string `json:"call_id"`
	Body        string `json:"body"`
	Disposition string `json:"disposition"`
	UpdatedAt   int64  `json:"updated_at"`
}

// notesCollection names the engine collection (query name).
const notesCollection = "call_notes"

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type callNoteInput struct{ Key domain.CallNoteKey }

// notesQuery declares the call-notes read model: ONE slot per
// (owner, call), upserted by CallNoteSaved facts that snapshot the
// whole slot (body + disposition — the seam pre-reads and fills both),
// filtered by owner, sorted by update time.
var notesQuery = metaengine.Query[callNoteInput, CallNoteView](notesCollection,
	metaengine.OnRecord(domain.CallNoteSaved{},
		func(_ record.Record, e domain.CallNoteSaved) (domain.CallNoteKey, CallNoteView) {
			return e.Key, CallNoteView{
				Owner:       e.Owner,
				CallID:      e.CallID,
				Body:        e.Body,
				Disposition: e.Disposition.String(),
				UpdatedAt:   e.UpdatedAt.Unix(),
			}
		}),
	metaengine.FilterOnField[CallNoteView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[CallNoteView]("updated_at", false),
	metaengine.Volume(humanScaleVolume),
)

// Notes is the engine-backed post-call-note store: one freeform body
// per (owner, call id), owner-scoped Get/Set, per-owner cap for NEW
// calls (re-saving an existing call's note never hits the cap).
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Snippets/Contacts) — deliberate
// symmetry, not a mixin candidate.
type Notes struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[CallNoteView]
	mu     sync.Mutex
}

func newNotes(models *metaengine.Store) *Notes {
	return &Notes{
		models: models,
		reader: metaengine.NewReader[CallNoteView](models, notesCollection),
	}
}

// Get returns the owner's note for one call, or nil when none exists
// (the Transcripts.Call unknown-call precedent: the caller owns the
// absent-vs-empty vocabulary).
func (s *Notes) Get(ctx context.Context, owner domain.Extension, callID string) (*CallNoteView, error) {
	view, found, err := s.reader.Get(ctx, domain.CallNoteKey{Owner: owner.String(), CallID: callID})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.note_get", "load note slot for %s", callID)
	}
	if !found || view.Owner != owner.String() {
		return nil, nil
	}
	return &view, nil
}

// Set upserts the owner's note for one call: the slot is the (owner,
// call) pair, so a re-save replaces the body and refreshes the stamp —
// no duplicate rows, no creation-time ceremony. The fact snapshots the
// WHOLE slot: the existing disposition rides along untouched. A NEW
// call's note past the per-owner cap is refused with ErrNoteListFull.
func (s *Notes) Set(ctx context.Context, owner domain.Extension, callID, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, _, err := s.claimSlot(ctx, owner, callID)
	if err != nil {
		return err
	}
	disposition := domain.Disposition("")
	if existing != nil {
		disposition = domain.Disposition(existing.Disposition)
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.CallNoteSaved{}), domain.CallNoteSaved{
		Key:         domain.CallNoteKey{Owner: owner.String(), CallID: callID},
		Owner:       owner.String(),
		CallID:      callID,
		Body:        body,
		Disposition: disposition,
		UpdatedAt:   time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.note_save", "apply note save")
	}
	return nil
}

// SetDisposition records the operator's post-call verdict on the same
// (owner, call) slot as the note. The fact snapshots the WHOLE slot:
// the existing note body rides along untouched. It shares the note's
// cap: a NEW call's verdict past NotesMaxPerExtension is refused — the
// cap guards the owner's review slots, not just their prose.
func (s *Notes) SetDisposition(ctx context.Context, owner domain.Extension, callID string, disposition domain.Disposition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, _, err := s.claimSlot(ctx, owner, callID)
	if err != nil {
		return err
	}
	body := ""
	if existing != nil {
		body = existing.Body
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.CallNoteSaved{}), domain.CallNoteSaved{
		Key:         domain.CallNoteKey{Owner: owner.String(), CallID: callID},
		Owner:       owner.String(),
		CallID:      callID,
		Body:        body,
		Disposition: disposition,
		UpdatedAt:   time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.disposition_save", "apply disposition save")
	}
	return nil
}

// claimSlot loads the slot under the store mutex and enforces the
// per-owner cap for a NEW call (existing slots always stay open). The
// caller emits its fact while still holding the mutex.
func (s *Notes) claimSlot(ctx context.Context, owner domain.Extension, callID string) (*CallNoteView, bool, error) {
	key := domain.CallNoteKey{Owner: owner.String(), CallID: callID}
	view, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return nil, false, errorfamily.WrapInfrastructuref(err, "readmodels.note_save", "load note slot")
	}
	if !found {
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
			metaengine.WithLimit(NotesMaxPerExtension+1),
		)
		if err != nil {
			return nil, false, errorfamily.WrapInfrastructuref(err, "readmodels.note_save", "count notes for %s", owner.String())
		}
		if len(views) >= NotesMaxPerExtension {
			return nil, false, errorfamily.WrapRejectionf(ErrNoteListFull, "readmodels.note_list_full",
				"%s at %d noted calls", owner.String(), NotesMaxPerExtension)
		}
	}
	if !found {
		return nil, false, nil
	}
	return &view, true, nil
}
