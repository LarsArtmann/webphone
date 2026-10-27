package api

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// noteBodyMax bounds one call note: the transcript segment's own text
// budget — a post-call note is a paragraph, not a document.
const noteBodyMax = 4000

// noteCallIDMax bounds the opaque island Call-ID before it reaches the
// store key (the summary lane's guard, same vocabulary).
const noteCallIDMax = 128

// CallNoteRequest is the PUT /api/calls/{id}/note wire shape. Named
// (not inline) so the machine-API contract derives its schema from
// this type (internal/server/apidocs.go).
type CallNoteRequest struct {
	Body string `description:"The note's plain text (max 4000 chars); empty clears the note." json:"body"`
}

// CallNoteResponse is the GET /api/calls/{id}/note answer: the stored
// body, empty when no note exists (absent and empty are the same
// surface vocabulary — the History form renders the same box).
type CallNoteResponse struct {
	Body string `description:"The note's plain text (empty when none)." json:"body"`
}

// validNoteCallID is the summary lane's path guard: the id is opaque,
// but the bounds keep malformed paths client-bug-shaped.
func validNoteCallID(callID string) bool {
	return callID != "" && len(callID) <= noteCallIDMax && !strings.ContainsAny(callID, "/%")
}

// GetCallNote answers the owner's note for one call — always 200 with
// the (possibly empty) body, so the History form needs no absent-state
// branch. Another extension's call id is the same empty box (owner
// scoping at the store key).
func (h *Handlers) GetCallNote(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.Notes == nil {
		http.Error(w, "note storage is unavailable", http.StatusServiceUnavailable)
		return
	}
	callID := r.PathValue("id")
	if !validNoteCallID(callID) {
		http.Error(w, "invalid call id", http.StatusBadRequest)
		return
	}
	view, err := h.deps.Notes.Get(r.Context(), sess.Extension, callID)
	if err != nil {
		slog.Warn("note read failed", "error", err, "extension", sess.Extension.String())
		http.Error(w, "could not read the note", http.StatusInternalServerError)
		return
	}
	body := ""
	if view != nil {
		body = view.Body
	}
	writeJSONBody(w, CallNoteResponse{Body: body})
}

// SetCallNote upserts the owner's note for one call — 204, idempotent
// by nature (the slot is the owner+call pair). An empty body clears
// the note. A full note list (per-owner cap) is a 409: hand-written
// content is never silently expunged, the user must delete another
// note first.
func (h *Handlers) SetCallNote(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	if h.deps.Notes == nil {
		http.Error(w, "note storage is unavailable", http.StatusServiceUnavailable)
		return
	}
	callID := r.PathValue("id")
	if !validNoteCallID(callID) {
		http.Error(w, "invalid call id", http.StatusBadRequest)
		return
	}
	var body CallNoteRequest
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, transcriptBodyLimit), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if len(body.Body) > noteBodyMax {
		http.Error(w, "note too long", http.StatusRequestEntityTooLarge)
		return
	}
	if err := h.deps.Notes.Set(r.Context(), sess.Extension, callID, body.Body); err != nil {
		if errors.Is(err, readmodels.ErrNoteListFull) {
			http.Error(w, "the note list is full — delete another note first", http.StatusConflict)
			return
		}
		wrapped := errorfamily.WrapInfrastructuref(err, "api.note_save", "save call note")
		slog.Warn("note save failed", "error", wrapped, "extension", sess.Extension.String())
		http.Error(w, "could not save the note", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
