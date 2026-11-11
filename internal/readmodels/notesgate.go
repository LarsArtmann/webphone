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

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrNoteListFull: errorfamily.Conflict,
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
// (slot read, note upsert, disposition upsert). The journal-backed
// service (internal/cqrs) is the implementation.
type NotesGate interface {
	Get(ctx context.Context, owner domain.Extension, callID string) (*CallNoteView, error)
	Set(ctx context.Context, owner domain.Extension, callID, body string) error
	SetDisposition(ctx context.Context, owner domain.Extension, callID string, disposition domain.Disposition) error
}
