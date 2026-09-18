package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
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
	owner, err := domain.ParseExtension(payload.Owner)
	if err != nil {
		http.Error(w, "invalid owner extension", http.StatusBadRequest)
		return
	}
	from, err := domain.ParsePhone(payload.From)
	if err != nil {
		http.Error(w, "invalid from number", http.StatusBadRequest)
		return
	}

	inbound := domain.InboundMessage{Owner: owner, From: from, Body: payload.Body, ReceivedAt: time.Now()}
	for _, att := range payload.Attachments {
		data, err := base64.StdEncoding.DecodeString(att.DataB64)
		if err != nil {
			http.Error(w, "attachment is not valid base64", http.StatusBadRequest)
			return
		}
		inbound.Attachments = append(inbound.Attachments, domain.InboundAttachment{
			Name: att.Name, MimeType: att.MimeType, Bytes: data,
		})
	}

	if _, err := h.deps.Messaging.Receive(r.Context(), inbound); err != nil {
		http.Error(w, "could not store message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) hookFax(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Owner       string `json:"owner"`
		From        string `json:"from"`
		Pages       int    `json:"pages"`
		ProviderRef string `json:"provider_ref"`
		PDFB64      string `json:"pdf_base64"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	owner, err := domain.ParseExtension(payload.Owner)
	if err != nil {
		http.Error(w, "invalid owner extension", http.StatusBadRequest)
		return
	}
	from, err := domain.ParsePhone(payload.From)
	if err != nil {
		http.Error(w, "invalid from number", http.StatusBadRequest)
		return
	}
	pdf, err := base64.StdEncoding.DecodeString(payload.PDFB64)
	if err != nil {
		http.Error(w, "pdf is not valid base64", http.StatusBadRequest)
		return
	}

	if _, err := h.deps.Fax.Receive(r.Context(), domain.InboundFax{
		Owner: owner, From: from, Pages: payload.Pages, PDFBytes: pdf,
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
		Pages       int    `json:"pages"`
		Error       string `json:"error"`
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

	if _, err := h.deps.Fax.UpdateProviderStatus(r.Context(), payload.ProviderRef, status, payload.Pages, payload.Error); err != nil {
		http.Error(w, "could not update fax: "+err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 40<<20))
	if err := decoder.Decode(out); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}
