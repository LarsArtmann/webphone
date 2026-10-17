// Package api holds the island's session-gated JSON endpoints, carved
// out of the server wiring: the CSRF token refresh, the personal-
// contacts panel API, the CRM call journal, and the transcription
// surface. The HTTP wiring (mux, flood budgets, CSRF middleware,
// session attach) stays in internal/server; these handlers reach the
// session gate and the SSE nudges through the narrow Server seam, so
// the 401 wording keeps its one home there.
package api

import (
	"net/http"

	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/server/hooks"
	"github.com/larsartmann/webphone/internal/session"
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
}

// Handlers is the carved JSON surface: one construction per server,
// methods are the http.HandlerFuncs the wiring mounts.
type Handlers struct {
	deps Deps
}

// New builds the JSON endpoint handlers.
func New(deps Deps) *Handlers {
	return &Handlers{deps: deps}
}
