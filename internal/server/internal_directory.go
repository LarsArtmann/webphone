package server

import (
	"net"
	"net/http"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/server/hooks"
	"github.com/larsartmann/webphone/internal/sipcred"
)

// The internal directory endpoint: GET /internal/directory/{extension}
// answers the derived REGISTER credentials of every LIVE session of that
// extension. Consumer: the stack's FreeSWITCH directory bridge
// (mod_xml_curl) — loopback + Bearer secret, never the browser. The
// directory password itself never appears here; the credentials are
// session-scoped derivations (sipcred), so logout/revocation is
// inherent: a dead session's credential dies at the next REGISTER.

// sipDirectoryGate guards the internal directory endpoint with the
// derivation secret (the bridge authenticates with the same secret the
// server derives under — one secret, two sides of one seam).
func (h *handlers) sipDirectoryGate(next http.Handler) http.Handler {
	return hooks.BearerGate(h.deps.Config.SIP.CredentialSecret, "sip credentials not configured", next)
}

// loopbackPeer is the machine-surface posture: the FreeSWITCH directory
// bridge and the operator's flight-recorder probe both run on the same
// host (the module wires them as local processes talking to the loopback
// listener). A non-loopback peer is a misdeployment — fail closed before
// any secret even matters.
func loopbackPeer(r *http.Request) bool {
	ip := net.ParseIP(remoteHostKey(r))
	return ip != nil && ip.IsLoopback()
}

// machineSurface is the loopback-only machine-surface prologue: a disabled
// seam answers like a missing surface (missing picks that surface's own
// 404), then a non-loopback peer is refused before any secret matters.
func machineSurface( //nolint:branching-flow // machineSurface bool-last is a recorded open-roadmap item (2026-10-25 review); the enabled gate is the seam's whole contract
	w http.ResponseWriter, r *http.Request, enabled bool, missing func(http.ResponseWriter, *http.Request),
) bool {
	if !enabled {
		missing(w, r)
		return false
	}
	if !loopbackPeer(r) {
		http.Error(w, "internal", http.StatusForbidden)
		return false
	}
	return true
}

func (h *handlers) internalDirectory(w http.ResponseWriter, r *http.Request) {
	if !machineSurface(w, r, h.deps.Config.SIP.DerivedCredentials(), http.NotFound) {
		return
	}
	extension, err := domain.ParseExtension(r.PathValue("extension"))
	if err != nil {
		http.Error(w, "invalid extension", http.StatusBadRequest)
		return
	}
	credentials := []string{}
	for _, token := range h.deps.Sessions.ActiveTokens(extension) {
		credentials = append(credentials, sipcred.Derive(h.deps.Config.SIP.CredentialSecret, token.String()))
	}
	// Unknown or session-less extensions answer an EMPTY list (200): the
	// bridge's static-password fallback owns the decision for anyone the
	// webphone has no live session for.
	writeJSON(w, http.StatusOK, map[string]any{"credentials": credentials})
}
