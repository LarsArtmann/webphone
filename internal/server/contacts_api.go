package server

import (
	"encoding/json/v2"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

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
		http.Error(w, "could not save the contact", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	w.WriteHeader(http.StatusNoContent)
}
