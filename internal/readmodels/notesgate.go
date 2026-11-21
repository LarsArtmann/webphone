// The call-notes seam's GATE + VIEW surface: the interface the server
// consumes and the projection row the journal backend folds. The
// engine-backed store that used to implement the gate here retired
// with the readmodels engine (2026-11-11 storage-ladder train B) —
// internal/cqrs/callnotes.go is the one implementation now.
package readmodels

import (
	"context"
	"errors"

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
//
// Classified CONFLICT (not Rejection): the request is well-formed and
// acceptable — the OWNER'S STATE (2000 noted calls) conflicts with it,
// and the owner can resolve and resubmit (delete another note), which
// is the 409/Conflict contract the HTTP surface has always answered.
// The sibling caps (contacts, snippets, push) stay Rejections: their
// HTTP surfaces answer 422, where "this submission cannot be accepted"
// is the honest family.
var ErrNoteListFull = errors.New("call note list full")

// ErrNoteMissing is returned when a DELETE names a slot the owner does
// not hold. Classified Rejection (the snippet-delete precedent): the
// ask cannot be accepted because its subject does not exist — the HTTP
// surface maps it to the 404.
var ErrNoteMissing = errors.New("call note not found")

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrNoteListFull: errorfamily.Conflict,
		ErrNoteMissing:  errorfamily.Rejection,
	})
}

// CallNoteView is the projection row ONE journal fold produces for a
// call's note slot: the freeform note plus the operator's disposition
// verdict.
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema; CallID is the opaque island token (strong-id wire ruling)
type CallNoteView struct {
	Owner       string `json:"owner"`
	CallID      string `json:"call_id"`
	Body        string `json:"body"`
	Disposition string `json:"disposition"`
	UpdatedAt   int64  `json:"updated_at"`
}

// NotesGate is the post-call-note seam as the server surfaces see it
// (slot read, note upsert, disposition upsert, slot delete, slot
// count). The journal-backed service (internal/cqrs) is the
// implementation.
type NotesGate interface {
	Get(ctx context.Context, owner domain.Extension, callID string) (*CallNoteView, error)
	// Set upserts the owner's note for one call; the returned bool
	// names whether the slot CHANGED. The editors autosave on every
	// typing pause, so a same-body save over the existing slot is an
	// echo (what a re-mounted editor PUTs right after its GET):
	// echoes are journal-silent, and the SSE nudge the handler owes
	// other tabs rides changes only. A first save always changes (the
	// slot mints, whatever the body).
	Set(ctx context.Context, owner domain.Extension, callID, body string) (bool, error)
	// SetDisposition upserts the operator's verdict for one call; the
	// returned bool names whether the slot CHANGED — Set's symmetry. The
	// chips are click-shaped, so a same-verdict re-click is an echo:
	// echoes are journal-silent (the decider's same-disposition guard is
	// the authority under races), and a first disposition always
	// changes (the slot mints).
	SetDisposition(ctx context.Context, owner domain.Extension, callID string, disposition domain.Disposition) (bool, error)
	// Delete removes one call's note slot outright (the explicit
	// DELETE lane — the empty-body clear keeps the slot, the delete
	// frees the per-owner cap). A missing slot is the ErrNoteMissing
	// Rejection (the handler's 404 mapping; the snippet-delete
	// precedent); owner scoping rides the store key.
	Delete(ctx context.Context, owner domain.Extension, callID string) error
	// Count answers how many note slots the owner holds (the Settings
	// cap-visibility row). Infrastructure failure wraps and returns.
	Count(ctx context.Context, owner domain.Extension) (int, error)
	// SearchBodies answers the owner's note slots whose body contains
	// the needle (ASCII case-insensitive contains — the transcript
	// Search contract), newest first. The History query joins note
	// prose into the transcript result set: a call whose ONLY match is
	// its note must still surface, exactly as segment text does.
	SearchBodies(ctx context.Context, owner domain.Extension, needle string) ([]CallNoteView, error)
}
