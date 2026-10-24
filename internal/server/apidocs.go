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
		catalog.Query[struct{}]("ghost.probe",
			catalog.MsgOperation(http.MethodGet, "/api/ghost"),
			catalog.WithName("Ghost Probe"),
			catalog.Response[struct{}]("200", "never mounted"),
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
			catalog.WithResponse("400", "Invalid body (or body over 4 KiB)."),
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
