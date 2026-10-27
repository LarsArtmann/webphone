package server

// The machine-API contract: /openapi.json is GENERATED from the wire
// DTOs via go-cqrs-lite/catalog, not hand-written — the request and
// response schemas are derived from the same named types the handlers
// decode and encode (loginRequest, sessionIdentity, api.ContactSaveRequest,
// …), so a DTO change moves the spec and the drift pin tests catch a
// route that lost its contract. Field-level documentation rides the
// DTOs' description struct tags; per-operation response descriptions
// carry the honest status inventory (superset of what the handlers
// plus their limiters and the CSRF gate emit).
//
// Modeling rules for this registry:
//   - POSTs ride catalog.Command[T] with T = the request DTO (the
//     exporter emits the request body from T).
//   - GETs ride catalog.Query[T] with T = struct{} and an explicit
//     typed Response[T] for the 200.
//   - Mutations that carry NO body (DELETEs) also ride the query kind
//     with an explicit MsgOperation method override: a command would
//     make the exporter invent a required request body the handler
//     ignores. The kind name is catalog taxonomy; the wire shape is
//     the contract.
//   - The 429 limiter responses document the Retry-After header in
//     their DESCRIPTION text: the exporter cannot express response
//     headers at catalog v4.9.0. The contacts/calls pins assert the
//     clause; revisit when headers land upstream.
//
// The /hooks/* lanes and the /events SSE stream are machine-to-machine
// contracts with their own home (docs/contracts/, AsyncAPI) — they stay
// out of this document by the wire-contract split rule.

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4"
	catalogopenapi "github.com/larsartmann/go-cqrs-lite/catalog/v4/openapi"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/server/api"
	"github.com/larsartmann/webphone/internal/session"
)

const (
	apiContractTitle   = "webphone machine API"
	apiContractVersion = "1.0.0"

	// retryAfterClause documents the flood limiters' Retry-After contract
	// in 429 description text (see the modeling rules above).
	retryAfterClause = "the answer carries a Retry-After header naming the seconds until the bucket refills"

	// csrfClause is the shared CSRF-gate wording for unsafe methods.
	csrfClause = "CSRF token missing or invalid"
)

// apiCatalog declares the session-gated JSON surface as a catalog:
// one service, cookie-auth security, one message per (path, method).
func apiCatalog() *catalog.Catalog {
	builder := catalog.NewBuilder(apiContractTitle, apiContractVersion)
	builder.Registry().AddSecurityScheme(catalog.SecurityScheme{
		ID:          "cookieAuth",
		Type:        "apiKey",
		In:          "cookie",
		Name:        session.CookieName,
		Description: "The tab session cookie minted by POST /api/session; every session-gated surface requires it.",
	})

	builder.AddService(
		"webphone-api", apiContractTitle, apiContractVersion,
		"Session-gated JSON API behind the webphone UI: sessions, CSRF, "+
			"personal contacts, and the CRM call journal. Credentials are "+
			"verified against the PBX directory before a session is minted.",
		// --- session ---
		catalog.Command[loginRequest]("session.create",
			catalog.MsgOperation(http.MethodPost, "/api/session"),
			catalog.WithName("Create Session"),
			catalog.WithSummary("Verify the credentials against the PBX directory and mint the tab session (sets the HttpOnly cookie, rotates CSRF)."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[sessionIdentity]("201", "Session created. Credential brokering on answers the session-derived REGISTER credential instead of the password; no-store keeps either credential out of caches."),
			catalog.WithResponse("400", "Invalid extension or request body."),
			catalog.WithResponse("401", "Credentials rejected by the PBX directory."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("429", "Per-client login flood budget exhausted; "+retryAfterClause+"."),
			catalog.WithResponse("500", "Session store failure."),
			catalog.WithResponse("502", "PBX credential verification unavailable."),
		),
		catalog.Query[struct{}]("session.resume",
			catalog.MsgOperation(http.MethodGet, "/api/session"),
			catalog.WithName("Resume Session"),
			catalog.WithSummary("Island boot probe: a live cookie gets the SIP credentials back so the browser re-registers without the login form. no-store."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[sessionIdentity]("200", "The cookie's live session. Exactly one credential field rides: sip_credential (session-derived, credential brokering on) or password (legacy wire)."),
			catalog.WithResponse("401", "No live session — the island shows the login form."),
		),
		catalog.Query[struct{}]("session.destroy",
			catalog.MsgOperation(http.MethodDelete, "/api/session"),
			catalog.WithName("Destroy Session"),
			catalog.WithSummary("Sign the tab session out: clears the session cookie and rotates CSRF. Idempotent — 204 whether or not a live session existed."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[struct{}]("204", "No response body. Session destroyed (any live session deleted, cookies cleared)."),
			catalog.WithResponse("403", csrfClause+"."),
		),
		// --- CSRF ---
		catalog.Query[struct{}]("csrf.refresh",
			catalog.MsgOperation(http.MethodGet, "/api/csrf"),
			catalog.WithName("Refresh CSRF"),
			catalog.WithSummary("Fetch the fresh masked CSRF token after the login rotation invalidated the old one. Needs no session; the answer pairs with the csrf_token cookie and is no-store."),
			catalog.Response[api.CSRFTokenResponse]("200", "Current masked CSRF token."),
			catalog.WithResponse("429", "Per-client flood budget exhausted; "+retryAfterClause+"."),
			catalog.WithResponse("500", "No CSRF token available."),
		),
		// --- contacts ---
		catalog.Query[struct{}]("contacts.list",
			catalog.MsgOperation(http.MethodGet, "/api/contacts"),
			catalog.WithName("List Contacts"),
			catalog.WithSummary("The session extension's personal contacts plus the configured shared ones. The LIST is the only drift-free id source (renames keep the old id)."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[api.ContactListResponse]("200", "Personal (store-backed, per extension) and shared (config) contacts."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("500", "Store failure."),
		),
		catalog.Command[api.ContactSaveRequest]("contacts.save",
			catalog.MsgOperation(http.MethodPost, "/api/contacts"),
			catalog.WithName("Save Contact"),
			catalog.WithSummary("Upsert one personal contact: a repeated number renames. Mutations answer 204 and the island re-fetches the list."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[struct{}]("204", "No response body. Saved (upsert by number within the session's extension)."),
			catalog.WithResponse("400", "Invalid body (or body over "+humanize.IBytes(api.ContactsBodyLimit)+".)"),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("422", "Number failed dialable validation, or the per-extension list is full (500; renames never count)."),
			catalog.WithResponse("429", "Per-client write flood budget exhausted (generous: the legacy import bursts one POST per row); "+retryAfterClause+"."),
			catalog.WithResponse("500", "Store failure."),
		),
		catalog.Query[struct{}]("contacts.delete",
			catalog.MsgOperation(http.MethodDelete, "/api/contacts"),
			catalog.WithName("Delete Contact"),
			catalog.WithSummary("Delete one personal contact by id, scoped to the session's extension."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("id", "query", "Contact id from the LIST response.", true),
			catalog.Response[struct{}]("204", "No response body. Deleted."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "Unknown id (or another extension's contact)."),
		),
		// --- call journal ---
		catalog.Command[api.CallLogRequest]("calls.log",
			catalog.MsgOperation(http.MethodPost, "/api/calls"),
			catalog.WithName("Log Call"),
			catalog.WithSummary("Journal one finished call on the CRM contact that owns the number (no-op when the CRM is disabled or the number matches no contact — the integration never creates contacts)."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[struct{}]("204", "No response body. Logged, or deliberately dropped (CRM disabled, or the number matches no CRM contact)."),
			catalog.WithResponse("400", "Malformed body, or direction not in/out."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("422", "Number failed dialable validation."),
			catalog.WithResponse("429", "Per-client write flood budget exhausted (shares the contacts budget); "+retryAfterClause+"."),
			catalog.WithResponse("502", "The CRM could not record the call (unreachable or 5xx) — the island toasts."),
		),
		// --- post-call summarize (textagent-gated: a disabled seam answers 404) ---
		catalog.Command[api.SummaryDraftRequest]("calls.summaryDraft",
			catalog.MsgOperation(http.MethodPost, "/api/calls/{id}/summary-draft"),
			catalog.WithName("Summarize Call"),
			catalog.WithSummary("Turn one call's stored transcript into an AI follow-up draft in the review queue (the Messages drafts panel reviews it). Idempotent within the window: a replay answers duplicate:true."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("id", "path", "The call's island Call-ID (the transcript's key).", true),
			catalog.Response[api.SummaryDraftResponse]("200", "The drafted follow-up (or duplicate:true when this call was already summarized in the idempotency window)."),
			catalog.WithResponse("400", "Malformed call id."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "The textagent seam is not configured, or the call has no transcript for this extension."),
			catalog.WithResponse("429", "Per-client write flood budget exhausted (shares the contacts budget); "+retryAfterClause+"."),
			catalog.WithResponse("500", "Transcript store failure."),
			catalog.WithResponse("502", "The AI agent is unreachable — the island toasts."),
			catalog.WithResponse("503", "Transcript storage unavailable (or the agent answered 5xx)."),
		),
		// --- post-call notes (M16: the History tab's note affordance) ---
		catalog.Query[struct{}]("calls.noteGet",
			catalog.MsgOperation(http.MethodGet, "/api/calls/{id}/note"),
			catalog.WithName("Get Call Note"),
			catalog.WithSummary("Read the signed-in extension's freeform note for one call. Always 200: an absent note is the empty body — the History form renders the same box either way."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("id", "path", "The call's island Call-ID (the transcript's key).", true),
			catalog.Response[api.CallNoteResponse]("200", "The stored note (empty when none)."),
			catalog.WithResponse("400", "Malformed call id."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("429", "Per-client read flood budget exhausted (shares the contacts budget); "+retryAfterClause+"."),
			catalog.WithResponse("500", "Note store failure."),
			catalog.WithResponse("503", "Note storage unavailable (test compositions)."),
		),
		catalog.Command[api.CallNoteRequest]("calls.noteSet",
			catalog.MsgOperation(http.MethodPut, "/api/calls/{id}/note"),
			catalog.WithName("Set Call Note"),
			catalog.WithSummary("Upsert the signed-in extension's freeform note for one call (the slot is the owner+call pair — no duplicates). An empty body clears the note. A full note list is a 409: hand-written content is never silently expunged."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("id", "path", "The call's island Call-ID (the transcript's key).", true),
			catalog.Response[struct{}]("204", "No response body"),
			catalog.WithResponse("400", "Malformed call id or body."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("409", "The extension's note list is at its cap — delete another note first."),
			catalog.WithResponse("413", "Note body past the 4000-char budget."),
			catalog.WithResponse("429", "Per-client write flood budget exhausted (shares the contacts budget); "+retryAfterClause+"."),
			catalog.WithResponse("500", "Note store failure."),
			catalog.WithResponse("503", "Note storage unavailable (test compositions)."),
		),
		// --- ASR seam (config-gated: a disabled seam answers 404) ---
		catalog.Query[struct{}]("asr.transcribe",
			catalog.MsgOperation(http.MethodPost, "/api/transcribe"),
			catalog.WithName("Transcribe Audio"),
			catalog.WithSummary("Transcribe one audio segment via the configured ASR provider (live calls, voicemail, MMS audio all ride this endpoint). "+
				"The request body is the raw audio bytes as a multipart/form-data part (honest Content-Type, up to "+humanize.IBytes(api.TranscribeBodyLimit)+"); "+
				"filename and lang ride the query string. Session-gated and CSRF-protected like every JSON post."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("filename", "query", "Original filename hint for the provider (optional).", false),
			catalog.WithParam("lang", "query", "Language hint for the provider (optional).", false),
			catalog.Response[api.TranscriptionResponse]("200", "The transcribed text."),
			catalog.WithResponse("400", "The audio could not be read, or the segment was empty."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "The ASR seam is not configured on this server — the surface does not exist."),
			catalog.WithResponse("429", "Per-client flood budget exhausted; "+retryAfterClause+"."),
			catalog.WithResponse("502", "The transcription provider failed (including rejected upstream credentials — never a client problem)."),
			catalog.WithResponse("503", "The transcription concurrency ceiling is busy — retry shortly."),
		),
		catalog.Command[api.TranscriptSaveRequest]("transcripts.save",
			catalog.MsgOperation(http.MethodPost, "/api/transcripts"),
			catalog.WithName("Save Transcript"),
			catalog.WithSummary("Persist one live-call transcript segment to the owner-scoped store behind the History tab's transcript section. Fire-and-forget: 204 whether stored or deliberately shaped."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[struct{}]("204", "No response body. Stored (or deliberately shaped — the island's capture loop never retries)."),
			catalog.WithResponse("400", "Malformed segment (bad direction, missing callId/startedAt/remote/text, oversized body)."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "The ASR seam is not configured — the surface does not exist."),
			catalog.WithResponse("413", "The segment text exceeds the 4000-character bound."),
			catalog.WithResponse("429", "Per-client flood budget exhausted; "+retryAfterClause+"."),
			catalog.WithResponse("500", "Store failure."),
			catalog.WithResponse("503", "Transcript storage is unavailable."),
		),
		catalog.Query[struct{}]("transcripts.delete",
			catalog.MsgOperation(http.MethodDelete, "/api/transcripts"),
			catalog.WithName("Delete Transcript"),
			catalog.WithSummary("Erase one call's transcript (the end-user leg of the retention story). Idempotent: an unknown or already-deleted call is the same 204."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("call", "query", "The call id whose transcript to erase.", true),
			catalog.Response[struct{}]("204", "No response body. Erased (or already absent)."),
			catalog.WithResponse("400", "The call query parameter is missing or oversized."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "The ASR seam is not configured — the surface does not exist."),
			catalog.WithResponse("429", "Per-client flood budget exhausted; "+retryAfterClause+"."),
			catalog.WithResponse("500", "Store failure."),
			catalog.WithResponse("503", "Transcript storage is unavailable."),
		),
		// --- push-device registry (config-gated: a disabled seam answers 404) ---
		catalog.Query[struct{}]("push.devices.list",
			catalog.MsgOperation(http.MethodGet, "/api/push-devices"),
			catalog.WithName("List Push Devices"),
			catalog.WithSummary("The session extension's registered push devices, newest first. The wake address is write-only and never echoed."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[api.PushDeviceListResponse]("200", "The registered devices."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("404", "No push transport is configured — a shell reads that as \"this server cannot ring closed apps\"."),
			catalog.WithResponse("500", "Store failure."),
		),
		catalog.Command[api.PushDeviceRegisterRequest]("push.devices.register",
			catalog.MsgOperation(http.MethodPost, "/api/push-devices"),
			catalog.WithName("Register Push Device"),
			catalog.WithSummary("Register (or re-register) the calling install's wake address; re-registration replaces the stored row so FCM token churn never strands a device."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.Response[struct{}]("204", "No response body. Registered."),
			catalog.WithResponse("400", "Invalid body (or body over "+humanize.IBytes(api.PushDeviceBodyLimit)+".)"),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "No push transport is configured — a shell must not believe registration succeeded."),
			catalog.WithResponse("422", "Bad device id or token shape, an unconfigured transport, or the per-extension device list is full (8; deregister one first)."),
			catalog.WithResponse("429", "Per-client write flood budget exhausted (shares the contacts budget); "+retryAfterClause+"."),
			catalog.WithResponse("500", "Store failure."),
		),
		catalog.Query[struct{}]("push.devices.delete",
			catalog.MsgOperation(http.MethodDelete, "/api/push-devices"),
			catalog.WithName("Deregister Push Device"),
			catalog.WithSummary("Deregister the calling install (logout). An unknown or foreign id is the same 204 — the deregistration outcome is identical."),
			catalog.MsgSecurity("cookieAuth"),
			catalog.WithParam("id", "query", "The shell-minted device id to deregister.", true),
			catalog.Response[struct{}]("204", "No response body. Deregistered (unknown and foreign ids included)."),
			catalog.WithResponse("401", "No live session."),
			catalog.WithResponse("403", csrfClause+"."),
			catalog.WithResponse("404", "No push transport is configured — the surface does not exist."),
			catalog.WithResponse("500", "Store failure."),
		),
	)

	return builder.Build()
}

// renderAPIContract exports the catalog to the served OpenAPI document.
var openAPIContract = sync.OnceValues(func() ([]byte, error) {
	doc := catalogopenapi.NewExporter(apiContractTitle, apiContractVersion,
		catalogopenapi.WithDescription(
			"Generated from the handler wire types (go-cqrs-lite/catalog). "+
				"The webhook lanes under /hooks/ and the /events SSE stream are "+
				"machine-to-machine contracts documented separately (docs/contracts/).",
		)).Export(apiCatalog())
	body, err := doc.MarshalJSON()
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(err, "webphone.server.openapi_render",
			"marshal the generated OpenAPI document")
	}
	return body, nil
})

func openapiHandler(w http.ResponseWriter, r *http.Request) {
	body, err := openAPIContract()
	if err != nil {
		// A render failure is a build-time bug (the contract tests
		// validate the catalog); serve the honest 500 instead of stale
		// hand-written bytes.
		slog.ErrorContext(r.Context(), "openapi render failed", "error", err)
		http.Error(w, "could not render the API contract", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/schema+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := w.Write(body); err != nil {
		slog.WarnContext(r.Context(), "openapi spec write failed mid-response", "error", err)
	}
}
