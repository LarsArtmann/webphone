package server

// The machine-to-machine wire contracts, declared as a catalog so the
// `webphone contracts` command can export them as AsyncAPI
// (docs/contracts/asyncapi.yaml). These are the lanes the consuming
// stack and the native shells build against — previously documented
// only as Go comments (hooks/webhooks.go) and AGENTS prose:
//
//   - INBOUND (/hooks/*, Bearer gateway secret, fail-closed): the
//     stack's Telnyx bridge and FS dialplan call webphone here. The
//     JSON payloads ride the exported hooks.Hook*Request types — the
//     same structs the handlers decode, so the schema cannot drift
//     from the decoder.
//   - OUTBOUND: webphone's webhook delivery rides the gateway seam
//     (multipart POSTs to the configured bridge URL); the receipt
//     fact and the SSE stream are declared as events.
//
// The browser-facing JSON API lives in apidocs.go (OpenAPI); the
// split rule: /hooks/* and SSE belong HERE, never in the OpenAPI
// document (see docs/contracts/README.md).

import (
	"github.com/larsartmann/go-cqrs-lite/catalog/v4"
	"github.com/larsartmann/go-cqrs-lite/catalog/v4/asyncapi"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/server/hooks"
)

const (
	// WireContractsTitle and WireContractsVersion identify the
	// generated AsyncAPI document. Version stamps the CONTRACT content:
	// bump it when a declared lane's payload changes shape.
	WireContractsTitle   = "webphone wire contracts"
	WireContractsVersion = "1.0.0"
)

// BuildWireContracts declares webphone's machine-to-machine lanes from
// the real wire types. Pure declaration — no config, no network.
func BuildWireContracts() *catalog.Catalog {
	builder := catalog.NewBuilder(WireContractsTitle, WireContractsVersion)

	builder.AddService(
		"webphone-wire", WireContractsTitle, WireContractsVersion,
		"Machine-to-machine lanes of the webphone binary: the secret-gated "+
			"/hooks/* ingest the telephony stack calls, the webhook delivery "+
			"receipts webphone answers with, and the per-extension SSE stream "+
			"the browser tabs live on.",
		// --- inbound: the stack calls webphone (Bearer gateway secret) ---
		catalog.Command[hooks.RingRequest]("wire.hooks.ring",
			catalog.MsgOperation("POST", "/hooks/ring"),
			catalog.WithName("Ring Hook"),
			catalog.WithSummary("POST /hooks/ring — inbound-call wake from the stack's FS dialplan (ring-while-closed seam). "+
				"Fans a push out to the owner's registered devices; state \"end\" stands them down. "+
				"Answers 202 inertly whether or not push is configured — the dialplan must never wait on push health. "+
				"Bearer gateway secret; fail-closed 503 without one."),
		),
		catalog.Command[hooks.HookMessageRequest]("wire.hooks.message",
			catalog.MsgOperation("POST", "/hooks/message"),
			catalog.WithName("Inbound Message Hook"),
			catalog.WithSummary("POST /hooks/message — inbound SMS/MMS from the stack's Telnyx bridge. "+
				"Persists into the owner's thread, nudges the SSE surfaces, auto-unarchives. "+
				"Bearer gateway secret; status answers are idempotent (1 h window; replays answer 202)."),
		),
		catalog.Command[hooks.HookFaxRequest]("wire.hooks.fax",
			catalog.MsgOperation("POST", "/hooks/fax"),
			catalog.WithName("Inbound Fax Hook"),
			catalog.WithSummary("POST /hooks/fax — inbound fax document (base64 PDF) from the bridge. "+
				"Bearer gateway secret."),
		),
		catalog.Command[hooks.HookFaxStatusRequest]("wire.hooks.fax.status",
			catalog.MsgOperation("POST", "/hooks/fax/status"),
			catalog.WithName("Fax Status Hook"),
			catalog.WithSummary("POST /hooks/fax/status — outbound fax verdict from the provider (transmitted | failed). "+
				"Bearer gateway secret; idempotent (1 h window)."),
		),
		catalog.Command[hooks.HookMessageStatusRequest]("wire.hooks.message.status",
			catalog.MsgOperation("POST", "/hooks/message/status"),
			catalog.WithName("Message Status Hook"),
			catalog.WithSummary("POST /hooks/message/status — outbound message verdict from the provider (delivered | failed). "+
				"Bearer gateway secret; idempotent (1 h window)."),
		),
		catalog.Event[gateway.ReceiptEnvelope]("wire.gateway.receipt",
			catalog.Receives,
			catalog.WithName("Gateway Receipt"),
			catalog.WithSummary("The acceptance receipt the bridge answers webphone's webhook POST with: "+
				"the provider_ref to address later status hooks at. A small non-JSON answer is taken "+
				"verbatim as the ref (the bare-token alternative); the receipt only ever acknowledges "+
				"acceptance — settlement waits for the status hook."),
		),
		// --- the per-extension SSE stream (/events) ---
		catalog.Event[struct{}]("wire.sse.threads",
			catalog.Sends,
			catalog.WithName("SSE threads"),
			catalog.WithSummary("GET /events event \"threads\": the thread LIST morphs (list-only by design — "+
				"a zero thread id never pushes the transcript, so an open conversation is never wiped)."),
		),
		catalog.Event[struct{}]("wire.sse.thread",
			catalog.Sends,
			catalog.WithName("SSE thread"),
			catalog.WithSummary("GET /events event \"thread\": one open conversation's transcript re-renders."),
		),
		catalog.Event[struct{}]("wire.sse.fax",
			catalog.Sends,
			catalog.WithName("SSE fax"),
			catalog.WithSummary("GET /events event \"fax\": the fax list re-renders."),
		),
		catalog.Event[struct{}]("wire.sse.voicemail",
			catalog.Sends,
			catalog.WithName("SSE voicemail"),
			catalog.WithSummary("GET /events event \"voicemail\": payload-less nudge — the panel re-fetches."),
		),
		catalog.Event[struct{}]("wire.sse.contacts",
			catalog.Sends,
			catalog.WithName("SSE contacts"),
			catalog.WithSummary("GET /events event \"contacts\": payload-less nudge — the panel re-fetches."),
		),
	)

	return builder.Build()
}

// renderWireContracts exports the declared lanes as an AsyncAPI
// document (JSON render; the CLI writes the YAML form of the same
// document). The exporter's default server stanza is a kafka broker —
// pure fabrication for these HTTP lanes — so the empty host suppresses
// the Servers section entirely.
func renderWireContracts() ([]byte, error) {
	doc := asyncapi.NewExporter(WireContractsTitle, WireContractsVersion,
		asyncapi.WithServer("webphone", "", ""),
		asyncapi.WithDescription(
			"Generated from the handler wire types (go-cqrs-lite/catalog). "+
				"The lanes are plain HTTPS calls: the telephony stack POSTs the "+
				"/hooks/* channels with the Bearer gateway secret, and webphone's "+
				"browser clients live on the /events SSE stream.",
		)).Export(BuildWireContracts())
	body, err := doc.MarshalJSON()
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(err, "webphone.server.wire_render",
			"marshal the generated AsyncAPI wire-contract document")
	}
	return body, nil
}
