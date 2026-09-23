package server

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// notifyContactsChanged nudges the extension's other surfaces (open
// Contacts tab, the island's contacts dropdown) with the payload-less
// "contacts" SSE event — the voicemail-nudge pattern: both consumers
// re-fetch with their own session credentials instead of trusting a
// push payload. A missing hub wiring (fuzz/minimal harnesses) skips
// the push: one lost nudge is cosmetic, the next mutation catches up.
func (h *handlers) notifyContactsChanged(extension domain.Extension) {
	if h.deps.Hubs == nil {
		return
	}
	h.deps.Hubs.Publish(extension, sseEventContacts, "")
}

// JSON contacts API for the SIP island's contact panel. The island used
// to keep its personal contacts in localStorage while the Contacts tab
// read the server store — two homes for the same fact. These endpoints
// expose the same per-extension store as JSON so the island reads and
// writes the one true home; the HTML-partial routes the tab uses
// (/contacts/save|delete|import|export) stay untouched.

type apiContact struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
}

type apiSharedContact struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// apiListContacts answers the island's panel read: the extension's
// personal contacts plus the operator's shared contacts from config
// (the same merge the Contacts tab renders).
func (h *handlers) apiListContacts(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	contacts, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list contacts", http.StatusInternalServerError)
		return
	}
	body := struct {
		Personal []apiContact       `json:"personal"`
		Shared   []apiSharedContact `json:"shared"`
	}{
		Personal: make([]apiContact, 0, len(contacts)),
		Shared:   make([]apiSharedContact, 0, len(h.deps.Shared)),
	}
	for _, contact := range contacts {
		body.Personal = append(body.Personal, apiContact{
			ID:     contact.ID.String(),
			Name:   contact.Name,
			Number: contact.Phone.String(),
		})
	}
	for _, shared := range h.deps.Shared {
		body.Shared = append(body.Shared, apiSharedContact{Name: shared.Name, Number: shared.Number})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.MarshalWrite(w, body) //nolint:erraudit // best-effort write; the response is already committed
}

// contactSaveFailed answers both contact-write surfaces (the tab form
// and the JSON API) with the same 500 and the same words: one fact, one
// home.
func contactSaveFailed(w http.ResponseWriter) {
	http.Error(w, "could not save the contact", http.StatusInternalServerError)
}

// apiContactSaved closes a successful JSON contact mutation: nudge the
// extension's other surfaces, then the bare 204 the island's re-fetch
// contract expects. The tab handlers share the nudge but answer with a
// toast + partial instead — this epilogue belongs to the JSON API only.
func (h *handlers) apiContactSaved(w http.ResponseWriter, extension domain.Extension) {
	h.notifyContactsChanged(extension)
	w.WriteHeader(http.StatusNoContent)
}

// apiSaveContact upserts one personal contact (same store semantics as
// the tab: a repeated number renames the entry). Mutations answer 204
// and the island re-fetches the list — the store mints IDs on insert
// and keeps the old ID on rename, so a GET is the only ID source that
// cannot drift.
func (h *handlers) apiSaveContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	var body struct {
		Name   string `json:"name"`
		Number string `json:"number"`
	}
	if err := json.UnmarshalRead(http.MaxBytesReader(w, r.Body, 4096), &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	phone, err := domain.ParsePhone(body.Number)
	if err != nil {
		http.Error(w, "enter a valid number", http.StatusUnprocessableEntity)
		return
	}
	contact := domain.Contact{
		ID:        domain.GenerateContactID(),
		Owner:     sess.Extension,
		Name:      body.Name,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	if err := h.deps.Contacts.Save(r.Context(), contact); err != nil {
		if errors.Is(err, store.ErrListFull) {
			http.Error(w, "contact list is full — delete one first", http.StatusUnprocessableEntity)
			return
		}
		contactSaveFailed(w)
		return
	}
	h.apiContactSaved(w, sess.Extension)
}

// apiDeleteContact removes one personal contact scoped to the session's
// extension (an id from another owner is a plain 404).
func (h *handlers) apiDeleteContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	contactID, err := domain.ParseContactID(r.URL.Query().Get("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.deps.Contacts.Delete(r.Context(), sess.Extension, contactID); err != nil {
		http.NotFound(w, r)
		return
	}
	h.apiContactSaved(w, sess.Extension)
}
