package api

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
	//brandid-lint:ignore(BD003) wire DTO: raw id at the JSON edge; the delete route re-parses client ids strictly (strong-id ruling)
	ID     string `json:"id"` //nolint:branching-flow // wire DTO projecting domain.ContactID at the JSON edge; the branded home is the domain type and the delete route re-parses client ids strictly
	Name   string `json:"name"`
	Number string `json:"number"`
}

// branching-flow DUPLICATE_TYPE (baselined): wire-DTO layer of the
// boundary trio (vcard.Card, domain.SharedContact); mirrors the json
// tags on purpose (registry 2026-10-05).
type apiSharedContact struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

// ContactListResponse is the GET /api/contacts wire shape: the session
// extension's personal contacts plus the configured shared ones. Named
// (not inline) so the machine-API contract derives its schema from this
// type (internal/server/apidocs.go).
type ContactListResponse struct {
	Personal []apiContact       `description:"The extension's store-backed personal contacts." json:"personal"`
	Shared   []apiSharedContact `description:"The operator's configured shared contacts." json:"shared"`
}

// ContactSaveRequest is the POST /api/contacts wire shape: a repeated
// number renames the existing entry.
type ContactSaveRequest struct {
	Name   string `description:"Display name; may be empty." json:"name"`
	Number string `description:"Dialable phone number, e.g. +4917012345678." json:"number"`
}

// ListContacts answers the island's panel read: the extension's
// personal contacts plus the operator's shared contacts from config
// (the same merge the Contacts tab renders).
func (h *Handlers) ListContacts(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	contacts, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list contacts", http.StatusInternalServerError)
		return
	}
	body := ContactListResponse{
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
	writeJSONBody(w, body)
}

// ContactSaveFailed answers both contact-write surfaces (the tab form
// and the JSON API) with the same 500 and the same words: one fact, one
// home. Exported because the tab's form handlers (internal/server)
// share it.
func ContactSaveFailed(w http.ResponseWriter) {
	http.Error(w, "could not save the contact", http.StatusInternalServerError)
}

// contactSaved closes a successful JSON contact mutation: nudge the
// extension's other surfaces, then the bare 204 the island's re-fetch
// contract expects. The tab handlers share the nudge but answer with a
// toast + partial instead — this epilogue belongs to the JSON API only.
func (h *Handlers) contactSaved(w http.ResponseWriter, extension domain.Extension) {
	h.deps.Server.NotifyContactsChanged(extension)
	w.WriteHeader(http.StatusNoContent)
}

// SaveContact upserts one personal contact (same store semantics as
// the tab: a repeated number renames the entry). Mutations answer 204
// and the island re-fetches the list — the store mints IDs on insert
// and keeps the old ID on rename, so a GET is the only ID source that
// cannot drift.
func (h *Handlers) SaveContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
	if !ok {
		return
	}
	var body ContactSaveRequest
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
		if storeAtCapacity(w, err, "contact list is full — delete one first") {
			return
		}
		ContactSaveFailed(w)
		return
	}
	h.contactSaved(w, sess.Extension)
}

// DeleteContact removes one personal contact scoped to the session's
// extension (an id from another owner is a plain 404).
func (h *Handlers) DeleteContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.deps.Server.RequireSession(w, r)
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
	h.contactSaved(w, sess.Extension)
}
