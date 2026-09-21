package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/vcard"
	"github.com/larsartmann/webphone/internal/web/views"
)

// uploadLimit bounds multipart bodies (5 attachments + a PDF fit easily).
const uploadLimit = 60 << 20

// sendMessage handles the composer forms (new conversation and reply).
func (h *handlers) sendMessage(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSessionMultipart(w, r)
	if !ok {
		return
	}

	to, err := domain.ParsePhone(r.FormValue("to"))
	if err != nil {
		h.renderPanelError(w, r, sess, views.TabMessages, http.StatusUnprocessableEntity, h.T(r, "err.invalidTo"))
		return
	}

	uploads := make([]domain.AttachmentContent, 0)
	for _, fileHeader := range r.MultipartForm.File["attachment"] {
		file, err := fileHeader.Open()
		if err != nil {
			h.renderPanelError(w, r, sess, views.TabMessages, http.StatusUnprocessableEntity, h.T(r, "err.attachmentRead"))
			return
		}
		content, readErr := io.ReadAll(io.LimitReader(file, messaging.MaxAttachmentSize+1))
		_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
		if readErr != nil || int64(len(content)) > messaging.MaxAttachmentSize {
			h.renderPanelError(w, r, sess, views.TabMessages, http.StatusUnprocessableEntity, h.T(r, "err.attachmentLarge"))
			return
		}
		uploads = append(uploads, domain.AttachmentContent{
			Name:     fileHeader.Filename,
			MimeType: fileHeader.Header.Get("Content-Type"),
			Bytes:    content,
		})
	}

	if _, err := h.deps.Messaging.Send(r.Context(), sess.Extension, to, r.FormValue("body"), uploads); err != nil {
		if invalid, ok := errors.AsType[*messaging.ErrInvalidSend](err); ok {
			h.renderPanelError(w, r, sess, views.TabMessages, http.StatusUnprocessableEntity, invalid.Reason)
			return
		}
		if rejected, ok := errors.AsType[*gateway.ErrProviderRejected](err); ok {
			// The provider ANSWERED with an actionable refusal (invalid
			// destination, provider policy) — the gateway itself is fine.
			// Surface the reason; the generic "unreachable" banner would
			// misdiagnose a working system (2026-09-21 self-send burn).
			slog.WarnContext(r.Context(), "message send rejected by provider", "status", rejected.Status)
			h.renderPanelError(w, r, sess, views.TabMessages, http.StatusBadGateway, rejected.Detail)
			return
		}
		// Transport failure: say 502, keep the detail in the log (the
		// gateway error can carry internal URLs), and tell the user the
		// message was saved as failed.
		slog.ErrorContext(r.Context(), "message send gateway failure", "error", err)
		h.renderPanelError(w, r, sess, views.TabMessages, http.StatusBadGateway, h.T(r, "err.gatewayUnavailable"))
		return
	}

	// Reply keeps the thread open; a new conversation returns to the list.
	threadParam := r.URL.Query().Get("thread")
	if threadParam != "" {
		threadID, err := domain.ParseThreadID(threadParam)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		component, err := h.threadPanel(r, sess, threadID, 0)
		if err != nil {
			http.Error(w, "load conversation: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		notifyToast(w, "ok", h.T(r, "toast.messageSent"))
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
		}
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.messageSent"))
	h.partial(w, r, tabFromPath("/messages"))
}

// sendFax handles the fax upload form.
func (h *handlers) sendFax(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSessionMultipart(w, r)
	if !ok {
		return
	}
	to, err := domain.ParsePhone(r.FormValue("to"))
	if err != nil {
		h.renderPanelError(w, r, sess, views.TabFax, http.StatusUnprocessableEntity, h.T(r, "err.invalidFaxTo"))
		return
	}
	file, header, err := r.FormFile("document")
	if err != nil {
		h.renderPanelError(w, r, sess, views.TabFax, http.StatusUnprocessableEntity, h.T(r, "err.attachPDF"))
		return
	}
	pdf, readErr := io.ReadAll(io.LimitReader(file, fax.MaxPDFSize+1))
	_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
	if readErr != nil || int64(len(pdf)) > fax.MaxPDFSize {
		h.renderPanelError(w, r, sess, views.TabFax, http.StatusUnprocessableEntity, h.T(r, "err.pdfLarge"))
		return
	}

	if _, err := h.deps.Fax.Send(r.Context(), sess.Extension, to, header.Filename, pdf); err != nil {
		if invalid, ok := errors.AsType[*fax.ErrInvalidFax](err); ok {
			h.renderPanelError(w, r, sess, views.TabFax, http.StatusUnprocessableEntity, invalid.Reason)
			return
		}
		if rejected, ok := errors.AsType[*gateway.ErrProviderRejected](err); ok {
			// Provider refusal with its own reason (e.g. "fax not wired"):
			// show it instead of implying the gateway is down.
			slog.WarnContext(r.Context(), "fax send rejected by provider", "status", rejected.Status)
			h.renderPanelError(w, r, sess, views.TabFax, http.StatusBadGateway, rejected.Detail)
			return
		}
		// Upstream failure, not a user mistake: 502 + log detail, same
		// policy as the message send path.
		slog.ErrorContext(r.Context(), "fax send gateway failure", "error", err)
		h.renderPanelError(w, r, sess, views.TabFax, http.StatusBadGateway, h.T(r, "err.faxGatewayUnavailable"))
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.faxSent"))
	h.partial(w, r, tabFromPath("/fax"))
}

// faxDocument streams a job's PDF (session-gated).
func (h *handlers) faxDocument(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	faxID, err := domain.ParseFaxID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	job, err := h.deps.Fax.Get(r.Context(), sess.Extension, faxID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	document, err := h.deps.Fax.Document(job)
	if err != nil {
		http.Error(w, "document missing", http.StatusGone)
		return
	}
	defer func() { _ = document.Close() }()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", "fax-"+job.ID.String()+".pdf"))
	_, _ = io.Copy(w, document) //nolint:erraudit // best-effort write; the response is already committed
}

// attachment streams one MMS attachment (session-gated, owner-scoped).
func (h *handlers) attachment(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	attachmentID, err := domain.ParseAttachmentID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	attachment, err := h.deps.Messaging.AttachmentByID(r.Context(), sess.Extension, attachmentID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	file, err := h.deps.Messaging.OpenAttachment(attachment.Path)
	if err != nil {
		http.Error(w, "attachment missing", http.StatusGone)
		return
	}
	defer func() { _ = file.Close() }()
	w.Header().Set("Content-Type", attachment.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.Name))
	_, _ = io.Copy(w, file) //nolint:erraudit // best-effort write; the response is already committed
}

// deleteVoicemail removes a message through the phone API.
func (h *handlers) deleteVoicemail(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "missing message id", http.StatusBadRequest)
		return
	}
	if err := h.deps.PhoneAPI.DeleteVoicemail(r.Context(), sess.PBXCredentials(), uuid); err != nil {
		h.renderPanelError(w, r, sess, views.TabVoicemail, http.StatusBadGateway, h.T(r, "vm.deleteFailed"))
		return
	}
	// Nudge the extension's other tabs: the "voicemail" SSE event carries
	// no payload — the voicemail panel re-fetches its partial on receipt.
	h.deps.Hubs.Publish(sess.Extension, sseEventVoicemail, "")
	notifyToast(w, "ok", h.T(r, "toast.voicemailDeleted"))
	h.partial(w, r, tabFromPath("/voicemail"))
}

// saveContact upserts a personal contact.
func (h *handlers) saveContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	// FormValue transparently handles urlencoded AND multipart bodies.
	name := r.FormValue("name")
	phone, err := domain.ParsePhone(r.FormValue("number"))
	if err != nil {
		http.Error(w, "enter a valid number", http.StatusUnprocessableEntity)
		return
	}
	contact := domain.Contact{
		ID:        domain.GenerateContactID(),
		Owner:     sess.Extension,
		Name:      name,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	if err := h.deps.Contacts.Save(r.Context(), contact); err != nil {
		http.Error(w, "could not save the contact", http.StatusInternalServerError)
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.contactSaved"))
	h.partial(w, r, tabFromPath("/contacts"))
}

// deleteContact removes a personal contact.
func (h *handlers) deleteContact(w http.ResponseWriter, r *http.Request) {
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
	notifyToast(w, "ok", h.T(r, "toast.contactDeleted"))
	h.partial(w, r, tabFromPath("/contacts"))
}

// countUnread sums unread messages across threads for the nav badge,
// served from the short-TTL cache; every unread mutation drops its
// extension's entry (send, inbound webhook, mark-read), so the badge is
// recomputed only for real changes or after the TTL.
func (h *handlers) countUnread(r *http.Request, sess session.Session) int {
	if total, ok := h.unread.get(sess.Extension); ok {
		return total
	}
	threads, err := h.deps.Messaging.Threads(r.Context(), sess.Extension)
	if err != nil {
		return 0
	}
	total := 0
	for _, summary := range threads {
		total += summary.Thread.Unread
	}
	h.unread.put(sess.Extension, total)
	return total
}

// countVoicemail reads the new-message count through the phone API.
func (h *handlers) countVoicemail(r *http.Request, sess session.Session) int {
	if !h.deps.PhoneAPI.Enabled() {
		return 0
	}
	summary, err := h.deps.PhoneAPI.VoicemailSummary(r.Context(), sess.PBXCredentials())
	if err != nil {
		return 0
	}
	return summary.New
}

// renderPanelError re-renders a tab with an error banner appended — the
// failure lands inside the region the user is looking at, not on a blank
// page.
func (h *handlers) renderPanelError(
	w http.ResponseWriter, r *http.Request, sess session.Session, tab views.Tab, status int, message string,
) {
	notifyToast(w, "error", message) // header rides along with the re-rendered panel
	component, err := h.tabComponent(r, tab, sess)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		http.Error(w, message, status)
		return
	}
	w.WriteHeader(status)
	if err := component.Render(r.Context(), w); err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, templ.EscapeString(message)) //nolint:erraudit // best-effort write; the response is already committed
}

// markThreadRead records the read marker for one thread — the live-swap
// path: an SSE `thread` push replaces an open conversation's transcript
// without a GET, so the client marks read explicitly (the badge must not
// stay lit for the conversation on screen). Idempotent by store contract.
func (h *handlers) markThreadRead(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	threadID, err := domain.ParseThreadID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.deps.Messaging.MarkRead(r.Context(), sess.Extension, threadID); err != nil {
		http.Error(w, "could not mark read", http.StatusInternalServerError)
		return
	}
	h.unread.drop(sess.Extension)
	w.WriteHeader(http.StatusNoContent)
}

// requireSession gates a handler behind the signed-in extension; on
// failure it has already written the error response. Handlers keep their
// own gate (rather than relying on route middleware) so they are safe by
// construction no matter how they are wired.
func (h *handlers) requireSession(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, session.SignInFirst, http.StatusUnauthorized)
		return session.Session{}, false
	}
	return sess, true
}

// requireSessionMultipart parses a multipart action form for the signed-in
// extension; on failure it has already written the error response.
func (h *handlers) requireSessionMultipart(w http.ResponseWriter, r *http.Request) (session.Session, bool) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return session.Session{}, false
	}
	if err := r.ParseMultipartForm(uploadLimit); err != nil {
		http.Error(w, "could not read the form: "+err.Error(), http.StatusBadRequest)
		return session.Session{}, false
	}
	return sess, true
}

// importContacts ingests an uploaded vCard file: every card with a
// valid number is upserted (same number = rename), invalid numbers are
// skipped so one bad row cannot block the import.
func (h *handlers) importContacts(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSessionMultipart(w, r)
	if !ok {
		return
	}
	file, header, err := r.FormFile("vcard")
	if err != nil {
		h.renderPanelError(w, r, sess, views.TabContacts, http.StatusUnprocessableEntity, h.T(r, "contacts.attachVCF"))
		return
	}
	data, readErr := io.ReadAll(io.LimitReader(file, 5<<20))
	_ = file.Close() //nolint:erraudit // close-after-use: nothing left to do on failure
	if readErr != nil {
		h.renderPanelError(w, r, sess, views.TabContacts, http.StatusUnprocessableEntity, h.T(r, "contacts.readFailed"))
		return
	}

	imported := 0
	for _, card := range vcard.Decode(data) {
		phone, err := domain.ParsePhone(card.Number)
		if err != nil { //nolint:erraudit // batch import: invalid cards are skipped, not surfaced
			continue
		}
		// The dialable alphabet also carries letters (SIP user parts);
		// a vCard number without a single digit can never be dialed.
		if !strings.ContainsAny(phone.String(), "0123456789") {
			continue
		}
		contact := domain.Contact{
			ID:        domain.GenerateContactID(),
			Owner:     sess.Extension,
			Name:      card.Name,
			Phone:     phone,
			CreatedAt: time.Now(),
		}
		if err := h.deps.Contacts.Save(r.Context(), contact); err != nil { //nolint:erraudit // batch import: per-card store failures skip the card, not the batch
			continue
		}
		imported++
	}
	if imported == 0 {
		h.renderPanelError(w, r, sess, views.TabContacts, http.StatusUnprocessableEntity,
			h.T(r, "contacts.importNone.pre")+templ.EscapeString(header.Filename)+h.T(r, "contacts.importNone.post"))
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.imported.pre")+strconv.Itoa(imported)+h.T(r, "toast.imported.post"))
	h.partial(w, r, tabFromPath("/contacts"))
}

// exportContacts streams the extension's personal contacts as vCard.
func (h *handlers) exportContacts(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}
	contacts, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list contacts", http.StatusInternalServerError)
		return
	}
	cards := make([]vcard.Card, 0, len(contacts))
	for _, contact := range contacts {
		cards = append(cards, vcard.Card{Name: contact.Name, Number: contact.Phone.String()})
	}
	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="webphone-contacts.vcf"`)
	_, _ = w.Write(vcard.Encode(cards)) //nolint:erraudit // best-effort write; the response is already committed
}
