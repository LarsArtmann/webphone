package server

import (
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// Webhook contracts (inbound). The gateway secret authenticates both.
//
// POST /hooks/message — inbound SMS/MMS:
//
//	{
//	  "secret":   "...",
//	  "owner":    "1001",              // extension receiving the message
//	  "from":     "+441632960961",
//	  "body":     "hello",
//	  "attachments": [                 // optional (MMS)
//	    {"name": "pic.jpg", "mime_type": "image/jpeg", "data_base64": "..."}
//	  ]
//	}
//
// POST /hooks/fax — inbound fax document:
//
//	{
//	  "secret":   "...",
//	  "owner":    "1001",
//	  "from":     "+441632960961",
//	  "pages":    2,
//	  "provider_ref": "gw-123",        // optional
//	  "pdf_base64": "..."
//	}
//
// POST /hooks/fax/status — outbound fax verdict from the provider:
//
//	{
//	  "secret":   "...",
//	  "provider_ref": "gw-123",
//	  "status":   "transmitted" | "failed",
//	  "pages":    2,
//	  "error":    "remote hung up"
//	}
//
// POST /hooks/message/status — outbound message verdict from the provider:
//
//	{
//	  "secret":   "...",
//	  "provider_ref": "gw-123",
//	  "status":   "delivered" | "failed",
//	  "error":    "invalid number"
//	}

// secretGate guards the /hooks/* surface with the configured shared
// secret. Without a configured secret the hooks stay closed (fail closed).
func (h *handlers) secretGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := h.deps.Config.Gateway.WebhookSecret
		if secret == "" {
			http.Error(w, "webhooks not configured", http.StatusServiceUnavailable)
			return
		}
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if provided != secret {
			http.Error(w, "bad secret", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handlers) webhooks(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/hooks/message":
		h.hookMessage(w, r)
	case "/hooks/message/status":
		h.hookMessageStatus(w, r)
	case "/hooks/fax":
		h.hookFax(w, r)
	case "/hooks/fax/status":
		h.hookFaxStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *handlers) hookMessage(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Owner       string `json:"owner"`
		From        string `json:"from"`
		Body        string `json:"body"`
		Attachments []struct {
			Name     string `json:"name"`
			MimeType string `json:"mime_type"`
			DataB64  string `json:"data_base64"`
		} `json:"attachments"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	owner, from, ok := parseOwnerFrom(w, payload.Owner, payload.From)
	if !ok {
		return
	}

	inbound := domain.InboundMessage{Owner: owner, From: from, Body: payload.Body, ReceivedAt: time.Now()}
	for _, att := range payload.Attachments {
		data, err := base64.StdEncoding.DecodeString(att.DataB64)
		if err != nil {
			http.Error(w, "attachment is not valid base64", http.StatusBadRequest)
			return
		}
		inbound.Attachments = append(inbound.Attachments, domain.AttachmentContent{
			Name: att.Name, MimeType: att.MimeType, Bytes: data,
		})
	}

	if _, err := h.deps.Messaging.Receive(r.Context(), inbound); err != nil {
		http.Error(w, "could not store message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.unread.drop(owner)
	w.WriteHeader(http.StatusAccepted)
}

// flexPages accepts a page count as a JSON number or a numeric string
// ("2"), so providers that format the count differently still update
// the job instead of failing the whole webhook with a type error.
type flexPages int

func (p *flexPages) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	value, err := dec.ReadValue()
	if err != nil {
		return err
	}
	if raw := strings.TrimSpace(string(value)); raw == "null" {
		*p = 0
		return nil
	}
	trimmed := strings.Trim(strings.TrimSpace(string(value)), `"`)
	pages, err := strconv.Atoi(trimmed)
	if err != nil {
		return fmt.Errorf("page count must be a number, got %s", value)
	}
	if pages < 0 {
		return fmt.Errorf("page count must not be negative, got %d", pages)
	}
	*p = flexPages(pages)
	return nil
}

// faxStatusPages collects the field names providers actually use for
// page counts; the first non-zero wins.
type faxStatusPages struct {
	Pages     flexPages `json:"pages"`
	PageCount flexPages `json:"page_count"`
	NumPages  flexPages `json:"num_pages"`
}

func (p faxStatusPages) count() int {
	for _, candidate := range []flexPages{p.Pages, p.PageCount, p.NumPages} {
		if candidate != 0 {
			return int(candidate)
		}
	}
	return 0
}

func (h *handlers) hookFax(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Owner string `json:"owner"`
		From  string `json:"from"`
		faxStatusPages
		ProviderRef string `json:"provider_ref"`
		PDFB64      string `json:"pdf_base64"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	owner, from, ok := parseOwnerFrom(w, payload.Owner, payload.From)
	if !ok {
		return
	}
	pdf, err := base64.StdEncoding.DecodeString(payload.PDFB64)
	if err != nil {
		http.Error(w, "pdf is not valid base64", http.StatusBadRequest)
		return
	}

	if _, err := h.deps.Fax.Receive(r.Context(), domain.InboundFax{
		Owner: owner, From: from, Pages: payload.count(), PDFBytes: pdf,
		Received: time.Now(), ProviderRef: payload.ProviderRef,
	}); err != nil {
		http.Error(w, "could not store fax: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) hookFaxStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProviderRef string `json:"provider_ref"`
		Status      string `json:"status"`
		faxStatusPages
		Error string `json:"error"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	status := domain.FaxStatus(payload.Status)
	switch status {
	case domain.FaxTransmitted, domain.FaxFailed:
	default:
		http.Error(w, "status must be transmitted or failed", http.StatusBadRequest)
		return
	}

	if _, err := h.deps.Fax.UpdateProviderStatus(r.Context(), payload.ProviderRef, status, payload.count(), payload.Error); err != nil {
		http.Error(w, "could not update fax: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) hookMessageStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProviderRef string `json:"provider_ref"`
		Status      string `json:"status"`
		Error       string `json:"error"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	if payload.ProviderRef == "" {
		http.Error(w, "provider_ref is required", http.StatusBadRequest)
		return
	}
	status := domain.OutboundStatus(payload.Status)
	switch status {
	case domain.StatusDelivered, domain.StatusFailed:
	default:
		http.Error(w, "status must be delivered or failed", http.StatusBadRequest)
		return
	}

	if _, err := h.deps.Messaging.DeliveryReceipt(r.Context(), payload.ProviderRef, status, payload.Error); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "could not update message: "+err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "could not update message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) error {
	decoder := jsontext.NewDecoder(http.MaxBytesReader(w, r.Body, 40<<20))
	if err := json.UnmarshalDecode(decoder, out); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}

// parseOwnerFrom validates the webhook's owner extension and from number;
// on failure it has already written the error response.
func parseOwnerFrom(w http.ResponseWriter, ownerRaw, fromRaw string) (domain.Extension, domain.Phone, bool) {
	owner, err := domain.ParseExtension(ownerRaw)
	if err != nil {
		http.Error(w, "invalid owner extension", http.StatusBadRequest)
		return domain.Extension{}, domain.Phone{}, false
	}
	from, err := domain.ParsePhone(fromRaw)
	if err != nil {
		http.Error(w, "invalid from number", http.StatusBadRequest)
		return domain.Extension{}, domain.Phone{}, false
	}
	return owner, from, true
}
