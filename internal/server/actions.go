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
	"github.com/larsartmann/go-error-family"
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
		h.sendFailure(w, r, sess, views.TabMessages, err, "message", sendFailureKeys{
			rejected:  "err.messageRejected",
			transport: "err.messageTransport",
		})
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
		h.sendFailure(w, r, sess, views.TabFax, err, "fax", sendFailureKeys{
			rejected:  "err.faxRejected",
			transport: "err.faxTransport",
		})
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
	defer func() { _ = document.Close() }() //nolint:erraudit // read-side close on defer; nothing left to act on
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", "fax-"+job.ID.String()+".pdf"))
	if _, err := io.Copy(w, document); err != nil {
		slog.WarnContext(r.Context(), "fax document stream broke mid-response", "error", err)
	}
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
	defer func() { _ = file.Close() }() //nolint:erraudit // read-side close on defer; nothing left to act on
	w.Header().Set("Content-Type", attachment.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.Name))
	if _, err := io.Copy(w, file); err != nil {
		slog.WarnContext(r.Context(), "attachment stream broke mid-response", "error", err)
	}
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
	h.notifyContactsChanged(sess.Extension)
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
	h.notifyContactsChanged(sess.Extension)
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

// sendFailureKeys bundles the per-lane i18n keys of the send failure
// surfaces; message and fax share one classification ladder.
type sendFailureKeys struct {
	rejected  string // formats with the provider's refusal detail
	transport string // the system-side copy carrying the retry advice
}

// classifyForUser maps a send-path failure to its user-facing HTTP status.
// Provider refusals keep their pinned 502-with-detail surface (the status
// must never drift from the copy their fast path renders); for everything
// else the family decides: Rejection is a 422 the caller can act on, every
// system-side family is a 502 whose transport copy carries the retry
// advice. ErrInvalidSend/ErrInvalidFax never reach the system-side arm —
// their fast path renders the service's own reason at 422.
func classifyForUser(err error) int {
	if _, ok := errors.AsType[*gateway.ErrProviderRejected](err); ok { //nolint:erraudit // presence check only: the typed value is intentionally unused, ok is checked
		return http.StatusBadGateway
	}
	if errorfamily.Classify(err) == errorfamily.Rejection {
		return http.StatusUnprocessableEntity
	}

	return http.StatusBadGateway
}

// sendFailure is the ONE failure→feedback ladder for outbound sends
// (SUPERB error-excellence T04): typed fast paths keep the authored copy
// (service validation reasons, provider refusal details — both pinned by
// tests), the family switch decides status and copy for everything else.
// Rendered strings are byte-identical to the per-handler ladders it
// replaced; the family is logged so an unclassified error is visible.
func (h *handlers) sendFailure(
	w http.ResponseWriter, r *http.Request, sess session.Session, tab views.Tab, err error, lane string, k sendFailureKeys,
) {
	if invalid, ok := errors.AsType[*messaging.ErrInvalidSend](err); ok {
		h.renderPanelError(w, r, sess, tab, http.StatusUnprocessableEntity, invalid.Reason)
		return
	}
	if invalid, ok := errors.AsType[*fax.ErrInvalidFax](err); ok {
		h.renderPanelError(w, r, sess, tab, http.StatusUnprocessableEntity, invalid.Reason)
		return
	}
	if rejected, ok := errors.AsType[*gateway.ErrProviderRejected](err); ok {
		// The provider ANSWERED with an actionable refusal (invalid
		// destination, provider policy) — the gateway itself is fine.
		// Surface the reason; the generic transport banner would
		// misdiagnose a working system (2026-09-21 self-send burn).
		slog.WarnContext(r.Context(), lane+" send rejected by provider", "status", rejected.Status,
			"family", errorfamily.Classify(err).String())
		h.renderPanelError(w, r, sess, tab, http.StatusBadGateway, fmt.Sprintf(h.T(r, k.rejected), rejected.Detail))
		return
	}
	// System-side failure: the send is safe in the thread as failed; the
	// detail stays in the log (gateway errors can carry internal URLs)
	// while the user gets reassurance + a retry path. The family rides the
	// log line so an unclassified error (defaults to transient) is
	// visible without changing what the user sees.
	slog.ErrorContext(r.Context(), lane+" send gateway failure", "error", err,
		"family", errorfamily.Classify(err).String())
	h.renderPanelError(w, r, sess, tab, classifyForUser(err), h.T(r, k.transport))
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
	if _, err := fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, templ.EscapeString(message)); err != nil {
		slog.WarnContext(r.Context(), "error banner write failed after committed response", "error", err)
	}
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
	skipped, firstSkip := 0, ""
	for _, card := range vcard.Decode(data) {
		phone, err := domain.ParsePhone(card.Number)
		if err != nil { //nolint:erraudit // batch import: counted + first reason feeds the one-line import log (T05)
			skipped++
			if firstSkip == "" {
				firstSkip = "invalid number: " + card.Number
			}
			continue
		}
		// The dialable alphabet also carries letters (SIP user parts);
		// a vCard number without a single digit can never be dialed.
		if !strings.ContainsAny(phone.String(), "0123456789") {
			skipped++
			if firstSkip == "" {
				firstSkip = "no digits: " + card.Number
			}
			continue
		}
		contact := domain.Contact{
			ID:        domain.GenerateContactID(),
			Owner:     sess.Extension,
			Name:      card.Name,
			Phone:     phone,
			CreatedAt: time.Now(),
		}
		if err := h.deps.Contacts.Save(r.Context(), contact); err != nil { //nolint:erraudit // batch import: counted + first reason (incl. the store error) feeds the one-line import log (T05)
			skipped++
			if firstSkip == "" {
				firstSkip = "store rejected " + phone.String() + ": " + err.Error()
			}
			continue
		}
		imported++
	}
	if skipped > 0 {
		// One line, not one per card: the count sizes the problem, the
		// first reason names it (SUPERB error-excellence T05 — skips are
		// invisible to the operator by design, but never to the log).
		slog.WarnContext(r.Context(), "contacts import skipped cards",
			"skipped", skipped, "imported", imported, "first_reason", firstSkip)
	}
	if imported == 0 {
		h.renderPanelError(w, r, sess, views.TabContacts, http.StatusUnprocessableEntity,
			h.T(r, "contacts.importNone.pre")+templ.EscapeString(header.Filename)+h.T(r, "contacts.importNone.post"))
		return
	}
	notifyToast(w, "ok", h.T(r, "toast.imported.pre")+strconv.Itoa(imported)+h.T(r, "toast.imported.post"))
	h.notifyContactsChanged(sess.Extension)
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
	if _, err := w.Write(vcard.Encode(cards)); err != nil {
		slog.WarnContext(r.Context(), "contacts export stream broke mid-response", "error", err)
	}
}
