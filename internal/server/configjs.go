package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
)

// configJS renders window.PBX_CONFIG for the island — the same contract
// the static site consumed from the serving PBX, now produced by this
// server from its config. TURN credentials are short-lived whenever the
// turn_rest secret is configured: each response derives a fresh pair
// (coturn REST API) instead of shipping a static password.
func (h *handlers) configJS(w http.ResponseWriter, _ *http.Request) {
	// config.ICEServer and domain.SharedContact carry the JSON tags of the
	// window.PBX_CONFIG wire contract (see README) — no local mirror types.
	iceServers := h.deps.Config.ICEServers
	if h.deps.Config.TURN.Secret != "" {
		expiry := time.Now().Add(h.deps.Config.TURN.TTL)
		iceServers = turnRESTCredentials(iceServers, h.deps.Config.TURN.Secret, expiry)
	}
	pbxConfig := struct {
		SIPDomain     string                 `json:"sipDomain"`
		WebsocketPath string                 `json:"websocketPath"`
		ICEServers    []config.ICEServer     `json:"iceServers,omitempty"`
		PhoneAPI      bool                   `json:"phoneApi"`
		CRM           bool                   `json:"crm"`
		Contacts      []domain.SharedContact `json:"contacts,omitempty"`
	}{
		SIPDomain:     h.deps.Config.SIPDomain,
		WebsocketPath: h.deps.Config.WebsocketPath,
		ICEServers:    iceServers,
		PhoneAPI:      h.deps.PhoneAPI.Enabled(),
		CRM:           h.deps.CRM.Enabled(),
		Contacts:      h.deps.Shared,
	}

	payload, err := json.Marshal(pbxConfig)
	if err != nil {
		http.Error(w, "render config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte("window.PBX_CONFIG = ")) //nolint:erraudit // best-effort write; the response is already committed
	_, _ = w.Write(payload)                        //nolint:erraudit // best-effort write; the response is already committed
	_, _ = w.Write([]byte(";\n"))                  //nolint:erraudit // best-effort write; the response is already committed
}

// turnRESTCredentials replaces the static username/credential of every
// entry carrying a turn:/turns: URL with a coturn REST pair derived from
// secret, valid until expiry: username is the unix expiry (coturn rejects
// pairs whose timestamp is past), credential is base64(HMAC-SHA1(secret,
// username)). STUN-only entries pass through verbatim (STUN needs no
// auth); one shared expiry keeps every TURN server of a response in
// lockstep, so the island never juggles mixed validity windows.
func turnRESTCredentials(servers []config.ICEServer, secret string, expiry time.Time) []config.ICEServer {
	out := make([]config.ICEServer, len(servers))
	copy(out, servers)
	username := strconv.FormatInt(expiry.Unix(), 10)
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	credential := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	for i := range out {
		if !hasTurnURL(out[i]) {
			continue
		}
		out[i].Username = username
		out[i].Credential = credential
	}
	return out
}

func hasTurnURL(server config.ICEServer) bool {
	for _, raw := range server.URLs {
		if strings.HasPrefix(raw, "turn:") || strings.HasPrefix(raw, "turns:") {
			return true
		}
	}
	return false
}

// verifyTURNPair recomputes the coturn HMAC for username and reports
// whether credential matches — the test-side oracle for the derivation.
func verifyTURNPair(secret, username, credential string) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	return hmac.Equal([]byte(credential), []byte(base64.StdEncoding.EncodeToString(mac.Sum(nil))))
}
