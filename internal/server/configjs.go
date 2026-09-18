package server

import (
	"encoding/json"
	"net/http"
)

// configJS renders window.PBX_CONFIG for the island — the same contract
// the static site consumed from the serving PBX, now produced by this
// server from its config. TURN credentials can be short-lived because this
// response is dynamic; today the values come straight from config.
func (h *handlers) configJS(w http.ResponseWriter, _ *http.Request) {
	type iceServer struct {
		URLs       []string `json:"urls"`
		Username   string   `json:"username,omitempty"`
		Credential any      `json:"credential,omitempty"`
	}
	pbxConfig := struct {
		SIPDomain     string     `json:"sipDomain"`
		WebsocketPath string     `json:"websocketPath"`
		ICEServers    []iceServer `json:"iceServers,omitempty"`
		PhoneAPI      bool       `json:"phoneApi"`
		Contacts      []struct {
			Name   string `json:"name"`
			Number string `json:"number"`
		} `json:"contacts,omitempty"`
	}{
		SIPDomain:     h.deps.Config.SIPDomain,
		WebsocketPath: h.deps.Config.WebsocketPath,
		PhoneAPI:      h.deps.PhoneAPI.Enabled(),
	}
	for _, server := range h.deps.Config.ICEServers {
		entry := iceServer{URLs: server.URLs, Username: server.Username}
		if server.Credential != "" {
			entry.Credential = server.Credential
		}
		pbxConfig.ICEServers = append(pbxConfig.ICEServers, entry)
	}
	for _, contact := range h.deps.Shared {
		pbxConfig.Contacts = append(pbxConfig.Contacts, struct {
			Name   string `json:"name"`
			Number string `json:"number"`
		}{Name: contact.Name, Number: contact.Number})
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
