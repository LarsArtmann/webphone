// Package platform holds the machine-facing and infrastructure
// surfaces carved out of the server wiring: the embedded asset tree,
// /metrics and its recorder types, the CSP report intake, the
// flight-recorder snapshot seam, the /phone-api proxy, the FreeSWITCH
// directory bridge, /config.js, and the per-extension data export. The
// HTTP wiring (mux, middleware chain, flood budgets) stays in
// internal/server; these handlers reach the session gate and the HTML
// error plumbing through the narrow Shell seam, and the voicemail SSE
// nudge keeps its one home in the server package.
package platform

import (
	"database/sql"
	"net/http"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/textagent"
)

// Deps is the platform surfaces' narrowed dependency view: exactly what
// the asset tree, the metrics exposition, the proxy and the export read.
type Deps struct {
	// Version is the running build's display string (the composition
	// root's DisplayVersion answer) for webphone_build_info.
	Version string
	Config  config.Config
	DB      *sql.DB
	// Sessions answers ActiveTokens for the directory bridge.
	Sessions session.Store
	Messages readmodels.ThreadGate
	Faxes    readmodels.FaxGate
	Contacts readmodels.ContactGate
	// Transcripts is nil-safe: the export degrades to no transcripts.json
	// leg when a hand-composed test composition omits it.
	Transcripts readmodels.TranscriptGate
	CRM         *crm.Resolver
	// Shared is the config contacts feed /config.js publishes.
	Shared    []domain.SharedContact
	PhoneAPI  *pbx.Client
	TextAgent *textagent.Client
	// ASR drives /config.js's asr flag; nil-safe (the island renders no
	// transcribe affordances).
	ASR asr.Provider
	// ASRLocality is the config-honest trust label riding /config.js.
	ASRLocality string
	// Trace is the OPTIONAL flight-recorder seam; nil means the snapshot
	// surface does not exist.
	Trace TraceRecorder
	// Latency and Island are the /metrics recorders (built by the server
	// wiring: the middleware writes the latency one, the beacon intake
	// writes the island one).
	Latency *LatencyRecorder
	Island  *IslandLatency
}

// Shell is the narrow wiring seam the carved handlers ride on,
// implemented by the server's handlers hub: the session gate and the
// render/error plumbing the machine surfaces borrow. Exported methods so
// internal/server can implement it.
type Shell interface {
	// RequireSession gates a handler on the attached session; on
	// failure it has already written the 401.
	RequireSession(w http.ResponseWriter, r *http.Request) (session.Session, bool)
	// NotFoundPage renders the styled 404 (browser navigations only —
	// the asset tree's miss arm).
	NotFoundPage(w http.ResponseWriter, r *http.Request)
	// NotFound renders the surface-appropriate 404 (the machineSurface
	// prologue's missing arm).
	NotFound(w http.ResponseWriter, r *http.Request)
	// InternalError logs the op and answers the terse 500.
	InternalError(w http.ResponseWriter, r *http.Request, op string, err error)
	// NotifyVoicemail nudges the owner's surfaces with the payload-less
	// "voicemail" SSE event (the proxy's post-read nudge; the event
	// vocabulary's ONE home stays in the server package).
	NotifyVoicemail(extension domain.Extension)
}

// Hub is the platform surface hub: every file in this package hangs
// methods off it. The server's wiring mounts the routes; the Shell seam
// is the server's handlers hub.
type Hub struct {
	deps  Deps
	shell Shell
}

// New wires the platform hub. The server's New builds the narrowed Deps
// and mounts the routes through the returned hub.
func New(deps Deps, shell Shell) *Hub {
	return &Hub{deps: deps, shell: shell}
}

// asrOn is the config-honest ASR gate (nil-safe over the optional seam).
func (p *Hub) asrOn() bool {
	return p.deps.ASR != nil && p.deps.ASR.Enabled()
}
