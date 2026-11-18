package domain

import (
	"github.com/larsartmann/go-cqrs-lite/catalog/v4"
)

// CatalogTitle and CatalogVersion identify the generated event catalog
// (`webphone catalog` export, federation hub). CatalogVersion versions
// the CATALOG CONTENT, not the binary: bump it when a declared message
// changes shape — every exported message must carry it (the zlota44
// lesson: unstamped versions fail the hub's governance gate).
const (
	CatalogTitle   = "Webphone"
	CatalogVersion = "1.0.0"

	// CatalogServiceID is the service id in the catalog and the
	// federation-hub union — keep in sync with the repo's
	// .eventcatalog-hub.json name.
	CatalogServiceID = "webphone"
)

// The Catalog*Schema types mirror the service-seam events (events.go)
// for catalog schema derivation. The real events carry branded types
// (Extension, ThreadID, FaxID) whose values are unexported — invisible
// to the schema reflector — so the catalog documents each event's
// observable field set explicitly. TestCatalogSchemaParity fails the
// build when an event and its mirror drift.
type (
	// CatalogMessageReceivedSchema mirrors MessageReceived.
	CatalogMessageReceivedSchema struct {
		Owner  string `json:"owner"`  // Extension: the PBX extension owning the thread
		Thread string `json:"thread"` // ThreadID: the thread the message landed in
	}

	// CatalogMessageDeliveryAdvancedSchema mirrors MessageDeliveryAdvanced.
	CatalogMessageDeliveryAdvancedSchema struct {
		Owner  string `json:"owner"`  // Extension: the PBX extension owning the thread
		Thread string `json:"thread"` // ThreadID: the outbound message's thread
	}

	// CatalogThreadFlagChangedSchema mirrors ThreadFlagChanged.
	CatalogThreadFlagChangedSchema struct {
		Owner string `json:"owner"` // Extension: the PBX extension whose flag moved
	}

	// CatalogFaxAdvancedSchema mirrors FaxAdvanced.
	CatalogFaxAdvancedSchema struct {
		Owner string `json:"owner"` // Extension: the PBX extension owning the fax job
		Fax   string `json:"fax"`   // FaxID: the fax job that advanced
	}
)

// BuildCatalog constructs webphone's event catalog from the closed
// service-seam event set. There are no catalog commands by design:
// webphone's writes are HTTP-driven (session + htmx), not messages.
// The catalog is consumed by the `webphone catalog` command, which
// renders the EventCatalog project tree for the architecture-catalog
// hub. No database, no network — a pure render.
func BuildCatalog() *catalog.Catalog {
	builder := catalog.NewBuilder(CatalogTitle, CatalogVersion)

	builder.AddService(
		CatalogServiceID, CatalogTitle, CatalogVersion,
		"Single-binary unified-communications web app for pbx.artmann.tech: "+
			"SIP calls (SIP.js island), SMS/MMS threads, fax, voicemail, history "+
			"and contacts behind one login against the PBX directory.",
		// --- Events (the service seam emits these; the projection
		// turns each into SSE pushes for the live UI) ---
		catalog.Event[CatalogMessageReceivedSchema](
			catalog.MessageID("message.received"), catalog.Sends,
			catalog.Owners("lars"),
			catalog.WithName("Message Received"),
			catalog.WithSummary("An inbound message landed in an owner's thread."),
		),
		catalog.Event[CatalogMessageDeliveryAdvancedSchema](
			catalog.MessageID("message.delivery.advanced"), catalog.Sends,
			catalog.Owners("lars"),
			catalog.WithName("Message Delivery Advanced"),
			catalog.WithSummary("An outbound message's delivery state moved: send accepted or refused, provider verdict landed."),
		),
		catalog.Event[CatalogThreadFlagChangedSchema](
			catalog.MessageID("thread.flag.changed"), catalog.Sends,
			catalog.Owners("lars"),
			catalog.WithName("Thread Flag Changed"),
			catalog.WithSummary("The user toggled a thread organization flag; list-only by design, so an open transcript is never wiped."),
		),
		catalog.Event[CatalogFaxAdvancedSchema](
			catalog.MessageID("fax.advanced"), catalog.Sends,
			catalog.Owners("lars"),
			catalog.WithName("Fax Advanced"),
			catalog.WithSummary("A fax job moved through its lifecycle: submit, settle, verdict, or receive."),
		),
	)

	// Owner-of-record for the federation hub (catalog.home.lan): the
	// service page links to this user; the hub renders owner cards from
	// it. Same user card pbx-lens declares, so the hub union renders one.
	builder.AddUser(catalog.User{
		ID:    "lars",
		Name:  "Lars Artmann",
		Email: "git@lars.software",
	})
	builder.ConfigureService(CatalogServiceID, catalog.ServiceOwners("lars"),
		// Wire-contract cross-links (hub inventory page): the catalog
		// command renders BOTH spec files directly into
		// <out>/eventcatalog/services/webphone/ — the AsyncAPI doc from
		// server.RenderWireContractsYAML (the /hooks/* lanes, receipts,
		// SSE) and the OpenAPI doc from server.RenderOpenAPIContract
		// (the session-gated HTTP API).
		//
		// Paths MUST be bare filenames: @eventcatalog/core resolves a
		// specification through TWO resolvers — the lint/MDX resolvers
		// join the full path onto the resource page's directory
		// (core/src/utils/resource-files.ts), while the spec viewer
		// pages join only the path's BASENAME
		// (core/src/utils/collections/util.ts processSpecifications ->
		// filename: getBasename(path)). A "specifications/"
		// subdirectory satisfied the first and silently 404'd the
		// viewer in the federation hub ("No OpenAPI spec file found";
		// the AsyncAPI page parsed empty content and blamed a missing
		// `asyncapi` field). The earlier catalog-root layout
		// ("/specifications/...", files at the tree root) satisfied no
		// check locally and died at lint (`refs/file-exists`) and
		// render (astro ENOENT).
		catalog.ServiceSpecifications(
			catalog.Specification{
				Type: "asyncapi",
				Path: "asyncapi.yaml",
				Name: "Webphone wire contracts (hooks, receipts, SSE)",
			},
			catalog.Specification{
				Type: "openapi",
				Path: "openapi.json",
				Name: "Session-gated HTTP API",
			},
		),
	)

	// Same domain id pbx-lens declares: the hub unions both sources'
	// services into one telephony graph (first-source-wins on the body,
	// relationships union through).
	builder.AddDomain(
		"telephony", "Telephony", CatalogVersion,
		"User-facing telephony for pbx.artmann.tech: calls, message threads, "+
			"and fax per extension, live over SSE.",
		CatalogServiceID,
	)
	builder.ConfigureDomain("telephony", catalog.DomainOwners("lars"))

	return builder.Build()
}
