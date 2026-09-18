package server

import (
	"encoding/json"
	"net/http"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
)

// configJS renders window.PBX_CONFIG for the island — the same contract
// the static site consumed from the serving PBX, now produced by this
// server from its config. TURN credentials can be short-lived because this
// response is dynamic; today the values come straight from config.
func (h *handlers) configJS(w http.ResponseWriter, _ *http.Request) {
	// config.ICEServer and domain.SharedContact carry the JSON tags of the
	// window.PBX_CONFIG wire contract (see README) — no local mirror types.
	pbxConfig := struct {
		SIPDomain     string                 `json:"sipDomain"`
		WebsocketPath string                 `json:"websocketPath"`
		ICEServers    []config.ICEServer     `json:"iceServers,omitempty"`
		PhoneAPI      bool                   `json:"phoneApi"`
		Contacts      []domain.SharedContact `json:"contacts,omitempty"`
	}{
		SIPDomain:     h.deps.Config.SIPDomain,
		WebsocketPath: h.deps.Config.WebsocketPath,
		ICEServers:    h.deps.Config.ICEServers,
		PhoneAPI:      h.deps.PhoneAPI.Enabled(),
		Contacts:      h.deps.Shared,
	}

	payload, err := json.Marshal(pbxConfig)
	if err != nil {
		http.Error(w, "render config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte("window.PBX_CONFIG = "))
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte(";\n"))
}
