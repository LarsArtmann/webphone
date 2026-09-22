package server

import (
	"archive/zip"
	"bytes"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/vcard"
)

// exportLimit bounds each store read behind the export: the per-extension
// data set is small by product scale, but a bound keeps a pathological
// thread or fax list from making the response unbounded.
const exportLimit = 10000

// The export wire shapes are deliberately explicit (not the domain
// structs): the zip is a PORTABILITY artifact, so its fields must stay
// stable even when domain types gain columns.

type exportAttachment struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type exportMessage struct {
	Direction     string             `json:"direction"`
	Channel       string             `json:"channel"`
	Body          string             `json:"body"`
	Status        string             `json:"status,omitempty"`
	FailureKind   string             `json:"failureKind,omitempty"`
	FailureDetail string             `json:"failureDetail,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
	Attachments   []exportAttachment `json:"attachments,omitempty"`
}

type exportThread struct {
	Remote   string          `json:"remote"`
	Unread   bool            `json:"unread"`
	Messages []exportMessage `json:"messages"`
}

type exportFax struct {
	Remote    string    `json:"remote"`
	Direction string    `json:"direction"`
	Status    string    `json:"status"`
	Pages     int       `json:"pages"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// exportData packages EVERYTHING the signed-in extension owns into one
// zip: message threads with their messages (messages.json), fax jobs
// (faxes.json) and personal contacts (contacts.vcf). A plain link
// download — GET, session-gated, owner-scoped by the store queries.
// Attachment and fax DOCUMENT blobs stay out of the archive (names and
// sizes are listed in the JSON): the zip is a portable manifest, and the
// blobs remain in the server's blob store where they are swept by
// retention.
func (h *handlers) exportData(w http.ResponseWriter, r *http.Request) {
	sess, ok := h.requireSession(w, r)
	if !ok {
		return
	}

	threads, err := h.deps.Messages.ListThreads(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list threads", http.StatusInternalServerError)
		return
	}
	faxes, err := h.deps.Faxes.List(r.Context(), sess.Extension, exportLimit)
	if err != nil {
		http.Error(w, "could not list faxes", http.StatusInternalServerError)
		return
	}
	contacts, err := h.deps.Contacts.List(r.Context(), sess.Extension)
	if err != nil {
		http.Error(w, "could not list contacts", http.StatusInternalServerError)
		return
	}

	outThreads := make([]exportThread, 0, len(threads))
	for _, summary := range threads {
		msgs, err := h.deps.Messages.ListMessages(r.Context(), sess.Extension, summary.Thread.ID, exportLimit)
		if err != nil {
			http.Error(w, "could not list messages", http.StatusInternalServerError)
			return
		}
		thread := exportThread{Remote: summary.Thread.Remote.String(), Unread: summary.Thread.Unread, Messages: make([]exportMessage, 0, len(msgs))}
		for _, msg := range msgs {
			message := exportMessage{
				Direction:     msg.Direction.String(),
				Channel:       msg.Channel.String(),
				Body:          msg.Body,
				Status:        msg.Status.String(),
				FailureKind:   msg.FailureKind,
				FailureDetail: msg.FailureDetail,
				CreatedAt:     msg.CreatedAt,
			}
			for _, att := range msg.Attachments {
				message.Attachments = append(message.Attachments, exportAttachment{Name: att.Name, MimeType: att.MimeType, Size: att.SizeBytes})
			}
			thread.Messages = append(thread.Messages, message)
		}
		outThreads = append(outThreads, thread)
	}

	outFaxes := make([]exportFax, 0, len(faxes))
	for _, fax := range faxes {
		outFaxes = append(outFaxes, exportFax{
			Remote:    fax.Remote.String(),
			Direction: fax.Direction.String(),
			Status:    fax.Status.String(),
			Pages:     fax.Pages,
			Error:     fax.Error,
			CreatedAt: fax.CreatedAt,
		})
	}

	messagesJSON, err := json.Marshal(outThreads)
	if err != nil {
		http.Error(w, "could not render threads", http.StatusInternalServerError)
		return
	}
	faxesJSON, err := json.Marshal(outFaxes)
	if err != nil {
		http.Error(w, "could not render faxes", http.StatusInternalServerError)
		return
	}
	cards := make([]vcard.Card, 0, len(contacts))
	for _, contact := range contacts {
		cards = append(cards, vcard.Card{Name: contact.Name, Number: contact.Phone.String()})
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=webphone-export-%s-%s.zip", sess.Extension.String(), time.Now().UTC().Format("20060102")))

	zw := zip.NewWriter(w)
	for _, entry := range []struct {
		name    string
		content []byte
	}{
		{"messages.json", messagesJSON},
		{"faxes.json", faxesJSON},
		{"contacts.vcf", vcard.Encode(cards)},
	} {
		file, err := zw.Create(entry.name)
		if err != nil {
			slog.WarnContext(r.Context(), "export zip entry failed", "entry", entry.name, "error", err)
			return
		}
		if _, err := file.Write(entry.content); err != nil {
			slog.WarnContext(r.Context(), "export zip write failed", "entry", entry.name, "error", err)
			return
		}
	}
	if err := zw.Close(); err != nil {
		slog.WarnContext(r.Context(), "export zip close failed", "error", err)
	}
}
