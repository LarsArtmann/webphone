package platform

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
)

// configJS renders window.PBX_CONFIG for the island — the same contract
// the static site consumed from the serving PBX, now produced by this
// server from its config. TURN credentials are short-lived whenever the
// turn_rest secret is configured: each response derives a fresh pair
// (coturn REST API) instead of shipping a static password.
//
// The shared contact directory deliberately does NOT ride this payload:
// /config.js is mounted PRE-AUTH (the island boot probe reads it before
// any session exists), and the operator's directory is exactly the data
// an anonymous fetcher must not enumerate. It loads post-login from the
// session-gated GET /api/contacts answer ({shared}) instead.
func (p *Hub) ConfigJS(w http.ResponseWriter, _ *http.Request) {
	// config.ICEServer and domain.SharedContact carry the JSON tags of the
	// window.PBX_CONFIG wire contract (see README) — no local mirror types.
	iceServers := p.deps.Config.ICEServers
	if p.deps.Config.TURN.Secret != "" {
		expiry := time.Now().Add(p.deps.Config.TURN.TTL)
		iceServers = turnRESTCredentials(iceServers, p.deps.Config.TURN.Secret, expiry)
	}
	var cooldownSeconds *int
	if p.deps.Config.Dial.Cooldown > 0 {
		secs := int(p.deps.Config.Dial.Cooldown.Seconds())
		cooldownSeconds = &secs
	}
	var metricsBeacon *bool
	if p.deps.Config.Metrics.IslandBeacon {
		on := true
		metricsBeacon = &on
	}
	pbxConfig := struct {
		SIPDomain     string             `json:"sipDomain"`
		WebsocketPath string             `json:"websocketPath"`
		ICEServers    []config.ICEServer `json:"iceServers,omitempty"`
		PhoneAPI      bool               `json:"phoneApi"`
		CRM           bool               `json:"crm"`
		ASR           bool               `json:"asr"`
		// Agent tells the island the textagent seam is on: the ended
		// call card grows the summarize affordance (M12). Off by
		// default; the post-call row then carries only Close.
		Agent bool `json:"agent"`
		// DialCooldownSeconds is the advisory dial-cooldown window
		// (issue #7): absent = no gate — the island dials without
		// asking; positive = the island asks POST /api/dials first.
		// A POINTER because encoding/json/v2's omitempty keeps zero
		// NUMBERS: the opt-in-off contract needs the key ABSENT, not
		// zero.
		DialCooldownSeconds *int `json:"dialCooldownSeconds,omitempty"`
		// Metrics tells the island the latency-beacon intake is on: the
		// island may POST /api/metrics/latency (the measure-first
		// latency design). Absent = off — the island never beacons and
		// the product stays byte-identical. A POINTER following the
		// DialCooldownSeconds discipline: the opt-in-off contract needs
		// the key ABSENT, never a zero value of uncertain omitempty
		// semantics.
		Metrics *bool `json:"metrics,omitempty"`
		// ASRLocality is the config-honest trust label ("local"|"remote"):
		// the caption-strip footer chip names where caller audio goes.
		ASRLocality string `json:"asrLocality,omitempty"`
		// SIPCredential tells the island the login ORDER changes: the
		// session must be minted BEFORE the REGISTER (the derived
		// credential comes back in the login/resume answers).
		SIPCredential bool `json:"sipCredential"`
	}{
		SIPDomain:           p.deps.Config.SIPDomain,
		WebsocketPath:       p.deps.Config.WebsocketPath,
		ICEServers:          iceServers,
		PhoneAPI:            p.deps.PhoneAPI.Enabled(),
		CRM:                 p.deps.CRM.Enabled(),
		ASR:                 p.asrOn(),
		Agent:               p.deps.TextAgent != nil,
		DialCooldownSeconds: cooldownSeconds,
		Metrics:             metricsBeacon,
		ASRLocality:         p.deps.ASRLocality,
		SIPCredential:       p.deps.Config.SIP.DerivedCredentials(),
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
