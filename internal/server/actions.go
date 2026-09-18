package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/session"
)

// uploadLimit bounds multipart bodies (5 attachments + a PDF fit easily).
const uploadLimit = 60 << 20

// sendMessage handles the composer forms (new conversation and reply).
func (h *handlers) sendMessage(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(uploadLimit); err != nil {
		http.Error(w, "could not read the form: "+err.Error(), http.StatusBadRequest)
		return
	}

	to, err := domain.ParsePhone(r.FormValue("to"))
	if err != nil {
		h.renderMessagesError(w, r, sess, "Enter a valid number to send to.")
		return
	}

	uploads := make([]messaging.Upload, 0)
	for _, fileHeader := range r.MultipartForm.File["attachment"] {
		file, err := fileHeader.Open()
		if err != nil {
			h.renderMessagesError(w, r, sess, "Could not read an attachment.")
			return
		}
		content, readErr := io.ReadAll(io.LimitReader(file, messaging.MaxAttachmentSize()+1))
		_ = file.Close()
		if readErr != nil || int64(len(content)) > messaging.MaxAttachmentSize() {
			h.renderMessagesError(w, r, sess, "An attachment is too large (10 MiB each).")
			return
		}
		uploads = append(uploads, messaging.Upload{
			Name:     fileHeader.Filename,
			MimeType: fileHeader.Header.Get("Content-Type"),
			Bytes:    content,
		})
	}

	if _, err := h.deps.Messaging.Send(r.Context(), sess.Extension, to, r.FormValue("body"), uploads); err != nil {
		h.renderMessagesError(w, r, sess, sendErrorMessage(err))
		return
	}

	// Reply keeps the thread open; a new conversation returns to the list.
	threadParam := r.URL.Query().Get("thread")
	if threadParam != "" {
		component, err := h.threadPanel(r, sess, domain.MustThreadID(threadParam))
		if err != nil {
			http.Error(w, "load conversation: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
		}
		return
	}
	h.partial(w, r, tabFromPath("/messages"))
}

func sendErrorMessage(err error) string {
	var invalid *messaging.ErrInvalidSend
	if errors.As(err, invalid) {
		return invalid.Reason
	}
	return "The gateway rejected the message: " + err.Error()
}

func (h *handlers) renderMessagesError(w http.ResponseWriter, r *http.Request, sess session.Session, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	component, err := h.messagesPanel(r, sess)
	if err != nil {
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}
	_ = component.Render(r.Context(), w)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, message)
}

// sendFax handles the fax upload form.
func (h *handlers) sendFax(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(uploadLimit); err != nil {
		http.Error(w, "could not read the form: "+err.Error(), http.StatusBadRequest)
		return
	}
	to, err := domain.ParsePhone(r.FormValue("to"))
	if err != nil {
		h.renderFaxError(w, r, sess, "Enter a valid fax number.")
		return
	}
	file, header, err := r.FormFile("document")
	if err != nil {
		h.renderFaxError(w, r, sess, "Attach a PDF to send.")
		return
	}
	pdf, readErr := io.ReadAll(io.LimitReader(file, fax.MaxPDFSize()+1))
	_ = file.Close()
	if readErr != nil || int64(len(pdf)) > fax.MaxPDFSize() {
		h.renderFaxError(w, r, sess, "The PDF is too large (20 MiB maximum).")
		return
	}

	if _, err := h.deps.Fax.Send(r.Context(), sess.Extension, to, header.Filename, pdf); err != nil {
		h.renderFaxError(w, r, sess, faxErrorMessage(err))
		return
	}
	h.partial(w, r, tabFromPath("/fax"))
}

func faxErrorMessage(err error) string {
	var invalid *fax.ErrInvalidFax
	if errors.As(err, invalid) {
		return invalid.Reason
	}
	return "The gateway rejected the fax: " + err.Error()
}

func (h *handlers) renderFaxError(w http.ResponseWriter, r *http.Request, sess session.Session, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	component, err := h.faxPanel(r, sess)
	if err != nil {
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}
	_ = component.Render(r.Context(), w)
	_, _ = fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, message)
}

// faxDocument streams a job's PDF (session-gated).
func (h *handlers) faxDocument(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	job, err := h.deps.Fax.Get(r.Context(), sess.Extension, domain.MustFaxID(r.PathValue("id")))
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
	_, _ = io.Copy(w, document)
}

// attachment streams one MMS attachment (session-gated, owner-scoped).
func (h *handlers) attachment(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	attachment, err := h.deps.Messaging.AttachmentByID(r.Context(), sess.Extension, domain.MustAttachmentID(r.PathValue("id")))
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
	_, _ = io.Copy(w, file)
}

// deleteVoicemail removes a message through the phone API.
func (h *handlers) deleteVoicemail(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "missing message id", http.StatusBadRequest)
		return
	}
	if err := h.deps.PhoneAPI.DeleteVoicemail(r.Context(), pbx.Credentials{
		Extension: sess.Extension.String(), Password: sess.Password,
	}, uuid); err != nil {
		h.renderVoicemailError(w, r, sess, "Could not delete the message — try again.")
		return
	}
	h.partial(w, r, tabFromPath("/voicemail"))
}

func (h *handlers) renderVoicemailError(w http.ResponseWriter, r *http.Request, sess session.Session, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	component, err := h.voicemailPanel(r, sess)
	if err != nil {
		http.Error(w, message, http.StatusBadGateway)
		return
	}
	_ = component.Render(r.Context(), w)
	_, _ = fmt.Fprintf(w, `<p class="wp-error" role="alert">%s</p>`, message)
}

// saveContact upserts a personal contact.
func (h *handlers) saveContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "could not read the form", http.StatusBadRequest)
		return
	}
	phone, err := domain.ParsePhone(r.FormValue("number"))
	if err != nil {
		http.Error(w, "enter a valid number", http.StatusUnprocessableEntity)
		return
	}
	contact := domain.Contact{
		ID:        domain.GenerateContactID(),
		Owner:     sess.Extension,
		Name:      r.FormValue("name"),
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	if err := h.deps.Contacts.Save(r.Context(), contact); err != nil {
		http.Error(w, "could not save the contact", http.StatusInternalServerError)
		return
	}
	h.partial(w, r, tabFromPath("/contacts"))
}

// deleteContact removes a personal contact.
func (h *handlers) deleteContact(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.From(r.Context())
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}
	if err := h.deps.Contacts.Delete(r.Context(), sess.Extension, domain.MustContactID(r.URL.Query().Get("id"))); err != nil {
		http.NotFound(w, r)
		return
	}
	h.partial(w, r, tabFromPath("/contacts"))
}

// countUnread sums unread messages across threads for the nav badge.
func (h *handlers) countUnread(r *http.Request, sess session.Session) int {
	threads, err := h.deps.Messaging.Threads(r.Context(), sess.Extension)
	if err != nil {
		return 0
	}
	total := 0
	for _, summary := range threads {
		total += summary.Thread.Unread
	}
	return total
}

// countVoicemail reads the new-message count through the phone API.
func (h *handlers) countVoicemail(r *http.Request, sess session.Session) int {
	if !h.deps.PhoneAPI.Enabled() {
		return 0
	}
	summary, err := h.deps.PhoneAPI.VoicemailSummary(r.Context(), pbx.Credentials{
		Extension: sess.Extension.String(), Password: sess.Password,
	})
	if err != nil {
		return 0
	}
	return summary.New
}
