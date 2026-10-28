// Package api holds the island's session-gated JSON endpoints, carved
// out of the server wiring: the CSRF token refresh, the personal-
// contacts panel API, the CRM call journal, the transcription
// surface, and the push-device registry the native shells ride. The
// HTTP wiring (mux, flood budgets, CSRF middleware, session attach)
// stays in internal/server; these handlers reach the session gate and
// the SSE nudges through the narrow Server seam, so the 401 wording
// keeps its one home there.
package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/server/hooks"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/textagent"
)

// Server is the narrow wiring seam the carved handlers ride on,
// implemented by the server's handlers hub: the session gate (the ONE
// 401 writer for missing sessions) and the contacts SSE nudge (the
// hubs are wiring-owned).
//
// branching-flow IFACE_COMPLETE (baselined): wiring seam by design —
// the scan cannot see the test stub (contextSessionGate) as the second
// implementer, and the seam exists precisely so the 401 wording and the
// SSE push keep their one home in internal/server.
type Server interface {
	// RequireSession gates a handler on the attached session; on
	// failure it has already written the 401.
	RequireSession(w http.ResponseWriter, r *http.Request) (session.Session, bool)
	// NotifyContactsChanged nudges the extension's surfaces with the
	// payload-less "contacts" SSE event.
	NotifyContactsChanged(extension domain.Extension)
	// AuditAgent writes one greppable agent-action audit line (the
	// ONE home of the vocabulary is the server's agentAudit; the JSON
	// lane reaches it through this seam).
	AuditAgent(event, actor, draftID, outcome string)
}

// Deps are the seams the JSON endpoints ride on, wired by
// internal/server at construction. Nil-optionality follows the
// server's Deps docs for each seam.
type Deps struct {
	Server      Server
	Contacts    *readmodels.Contacts
	Shared      []domain.SharedContact
	CRM         *crm.Resolver
	CallsIdem   hooks.Gate
	ASR         asr.Provider
	Transcripts *readmodels.Transcripts
	// Notes is the post-call-note store (readmodels-backed; present
	// whenever the composition built the engine). Nil: test
	// compositions only — the handlers answer 503 like the transcript
	// surface does.
	Notes *readmodels.Notes
	// Reviews is the post-call-review store (readmodels-backed; the
	// automatic review's durable artifact slot, issue #8). Nil: test
	// compositions only — the artifact endpoints answer 503 and the
	// job skips the writes.
	Reviews *readmodels.Reviews
	// TextAgent is the OPTIONAL supervised text-agent seam (the stack's
	// review API). Nil: the summarize endpoint answers 404 — a client
	// reads that as "this server does not draft".
	TextAgent *textagent.Client
	// Push is the OPTIONAL push fanout behind the device registry API.
	// Nil or disabled: the routes answer 404 (a native shell reads
	// that as "this server does not ring closed apps").
	Push *push.Fanout
	// Devices is the registered-device store (readmodels-backed;
	// present whenever the composition built the engine).
	Devices *readmodels.PushDevices
	// Dials is the dial-cooldown window store (readmodels-backed;
	// present whenever the composition built the engine). Nil: test
	// compositions only — the gate answers 503.
	Dials *readmodels.DialWindows
	// DialCooldown is the advisory window length (issue #7). Zero:
	// the gate route 404s and the island never asks.
	DialCooldown time.Duration
}

// Handlers is the carved JSON surface: one construction per server,
// methods are the http.HandlerFuncs the wiring mounts.
type Handlers struct {
	deps Deps
	// transcribeSlots bounds CONCURRENT transcription requests
	// (#47): each holds up to TranscribeBodyLimit buffered bytes plus
	// a provider round trip, so the ceiling is the product's honest
	// memory bound for the whole surface.
	transcribeSlots chan struct{}
	// reviewInFlight holds the automatic review jobs' idem keys for
	// their whole run (issue #8): the idem window records successes
	// only, so the guard is what keeps a racing manual click from
	// drafting twice. Zero value ready.
	reviewInFlight sync.Map
}

// New builds the JSON endpoint handlers.
func New(deps Deps) *Handlers {
	return &Handlers{deps: deps, transcribeSlots: make(chan struct{}, transcribeConcurrency)}
}
