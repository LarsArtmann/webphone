// Package hooks owns the machine-facing webhook ingest: the provider
// callbacks (inbound messages and faxes, outbound verdicts), the
// shared-secret bearer gate that fronts them, and the idempotency
// windows the callback lane and the island's post-call report ride
// on. The HTTP wiring (mount, flood budget) stays in internal/server.
package hooks

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/store"
)

// Deps are the seams the webhook ingest rides on, wired by
// internal/server at construction.
type Deps struct {
	// Secret authenticates every /hooks/* callback (fail closed when
	// empty — see BearerGate).
	Secret string
	// Messaging persists inbound messages and applies outbound verdicts.
	Messaging *messaging.Service
	// Fax persists inbound faxes and applies outbound verdicts.
	Fax *fax.Service
	// Unread drops the owner's memoized nav-badge total when an inbound
	// message lands (the badge must not wait out the cache TTL).
	Unread UnreadCache
	// Dedup is the OPTIONAL engine-backed dedup window; nil degrades to
	// the in-process map (byte-identical legacy behavior).
	Dedup DedupWindow
}

// UnreadCache is the narrow nav-badge seam the message hook
// invalidates: one method, implemented by the server's unread cache.
type UnreadCache interface {
	Drop(owner domain.Extension)
}

// Handlers is the webhook ingest surface: one construction per server,
// methods are the http.HandlerFuncs the wiring mounts.
type Handlers struct {
	secret    string
	messaging *messaging.Service
	fax       *fax.Service
	unread    UnreadCache
	idem      Gate
}

// New builds the webhook ingest handlers, including their idempotency
// gate (the engine window when wired, the in-process map otherwise).
func New(deps Deps) *Handlers {
	return &Handlers{
		secret:    deps.Secret,
		messaging: deps.Messaging,
		fax:       deps.Fax,
		unread:    deps.Unread,
		idem:      NewGate(deps.Dedup, hookIdempotencyTTL),
	}
}

// webhookFail answers a provider-facing 5xx without leaking internals:
// the full error goes to the server log (runbook-greppable), the body gets
// the library's redacted SafeDetail text. 500, not 503, on purpose — the
// provider already retries any 5xx, and changing the observed status
// semantics is not this helper's job.
func webhookFail(w http.ResponseWriter, kind string, err error) {
	slog.Warn("webhook apply failed", "kind", kind, "err", err)
	http.Error(w, cqrshtmx.SafeDetail(err, http.StatusInternalServerError, false), http.StatusInternalServerError)
}

// RetryAfterHint is the Retry-After value the 503 writers send: one
// second, matching cqrs-htmx's DefaultRetryAfter (deliberately
// conservative — standards-compliant clients back off without a retry
// storm). Sent where the CONSUMER honors it: provider retry machinery
// on the fail-closed hooks and any HTTP client on the proxy. The
// startupz latch deliberately does NOT send it (its consumers are
// liveness probes that ignore response headers).
const RetryAfterHint = "1"

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

// BearerGate guards a secret-authenticated surface with the configured
// shared secret. Without a configured secret the surface stays closed
// (fail closed). The comparison is constant-time over SHA-256 digests:
// no early-exit on the first mismatched byte and no secret-length leak,
// the same bar the CRM seam holds its bearer token to. `unavailable` is
// the 503 body when the deployment has no secret configured.
func BearerGate(secret, unavailable string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			w.Header().Set("Retry-After", RetryAfterHint)
			http.Error(w, unavailable, http.StatusServiceUnavailable)
			return
		}
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		wantHash := sha256.Sum256([]byte(secret))
		gotHash := sha256.Sum256([]byte(provided))
		if subtle.ConstantTimeCompare(wantHash[:], gotHash[:]) != 1 {
			http.Error(w, "bad secret", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecretGate guards the /hooks/* surface with the configured shared
// secret (see BearerGate).
func (h *Handlers) SecretGate(next http.Handler) http.Handler {
	return BearerGate(h.secret, "webhooks not configured", next)
}

func (h *Handlers) Webhooks(w http.ResponseWriter, r *http.Request) {
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

func (h *Handlers) hookMessage(w http.ResponseWriter, r *http.Request) {
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

	if _, err := h.messaging.Receive(r.Context(), inbound); err != nil {
		webhookFail(w, "message-store", err)
		return
	}
	h.unread.Drop(owner)
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
		return errorfamily.Newf(errorfamily.Rejection, "webhook.fax_pages", "page count must be a number, got %s", value)
	}
	if pages < 0 {
		return errorfamily.Newf(errorfamily.Rejection, "webhook.fax_pages", "page count must not be negative, got %d", pages)
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

func (h *Handlers) hookFax(w http.ResponseWriter, r *http.Request) {
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

	// The fax lane is telephony-only: the sms-parsed address round-trips
	// through the same dialable rule, so the phone parse cannot fail.
	if _, err := h.fax.Receive(r.Context(), domain.InboundFax{
		Owner: owner, From: domain.MustParsePhone(from.String()), Pages: payload.count(), PDFBytes: pdf,
		Received: time.Now(), ProviderRef: payload.ProviderRef,
	}); err != nil {
		webhookFail(w, "fax-store", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// applyStatusWebhook is the shared tail of both provider verdict hooks:
// an empty ref is a 400, a replayed ref answers 202 inertly, the apply
// runs once (404 names an unknown ref so the provider stops retrying it,
// 500 means retryable), and the idempotency key is recorded only on
// success so failures stay retryable. kind namespaces the key
// ("<kind>/<ref>") and names the failure surfaces.
func (h *Handlers) applyStatusWebhook(w http.ResponseWriter, r *http.Request, kind, ref string, apply func(context.Context, string) error) {
	if ref == "" {
		http.Error(w, "provider_ref is required", http.StatusBadRequest)
		return
	}
	key := kind + "/" + ref
	if h.idem.Seen(key) {
		w.WriteHeader(http.StatusAccepted) // replay: the original verdict already applied
		return
	}
	if err := apply(r.Context(), ref); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "could not update "+kind+": "+err.Error(), http.StatusNotFound)
			return
		}
		webhookFail(w, kind+"-status", err)
		return
	}
	h.idem.Record(key) // only successes are deduped; failures stay retryable
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) hookFaxStatus(w http.ResponseWriter, r *http.Request) {
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
	h.applyStatusWebhook(w, r, "fax", payload.ProviderRef,
		func(ctx context.Context, ref string) error {
			_, err := h.fax.UpdateProviderStatus(ctx, ref, status, payload.count(), payload.Error)
			return err
		})
}

func (h *Handlers) hookMessageStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProviderRef string `json:"provider_ref"`
		Status      string `json:"status"`
		Error       string `json:"error"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		return
	}
	status := domain.OutboundStatus(payload.Status)
	switch status {
	case domain.StatusDelivered, domain.StatusFailed:
	default:
		http.Error(w, "status must be delivered or failed", http.StatusBadRequest)
		return
	}
	h.applyStatusWebhook(w, r, "message", payload.ProviderRef,
		func(ctx context.Context, ref string) error {
			_, err := h.messaging.DeliveryReceipt(ctx, ref, status, payload.Error)
			return err
		})
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
func parseOwnerFrom(w http.ResponseWriter, ownerRaw, fromRaw string) (domain.Extension, domain.Address, bool) {
	owner, err := domain.ParseExtension(ownerRaw)
	if err != nil {
		http.Error(w, "invalid owner extension", http.StatusBadRequest)
		return domain.Extension{}, domain.Address{}, false
	}
	// The webhook lane is telephony-only by contract (the Telnyx
	// bridge): the from value is a phone number on the sms network.
	from, err := domain.ParseAddress(domain.NetworkSMS, fromRaw)
	if err != nil {
		http.Error(w, "invalid from number", http.StatusBadRequest)
		return domain.Extension{}, domain.Address{}, false
	}
	return owner, from, true
}
