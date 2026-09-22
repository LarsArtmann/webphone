package server

import (
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

// decodePBXConfig strips the window.PBX_CONFIG assignment and decodes the
// JSON payload, so tests assert on parsed values instead of substring
// greps (a greppable "username":"..." would pass with stale quotes).
func decodePBXConfig(t *testing.T, body []byte) (iceServers []config.ICEServer) {
	t.Helper()
	payload := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(string(body), "window.PBX_CONFIG =")), ";"))
	var parsed struct {
		ICEServers []config.ICEServer `json:"iceServers"`
	}
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		t.Fatalf("config.js payload is not JSON: %v\n%s", err, payload)
	}
	return parsed.ICEServers
}

// TestConfigJSTurnRESTCredentials pins the coturn REST derivation: with
// turn_rest.secret set, every TURN entry ships a fresh unix-expiry
// username plus the HMAC-SHA1 credential (verifiable with the shared
// secret), static passwords never leak, and STUN-only entries stay
// credential-free.
func TestConfigJSTurnRESTCredentials(t *testing.T) {
	const secret = "coturn-shared-secret"
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.ICEServers = []config.ICEServer{
			{URLs: []string{"turn:turn.example.org:3478?transport=udp"}, Username: "static-user", Credential: "static-password"},
			{URLs: []string{"stun:stun.example.org:3478"}},
		}
		cfg.TURN = config.TURNREST{Secret: secret, TTL: 2 * time.Hour}
	})
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js status %d", resp.StatusCode)
	}
	iceServers := decodePBXConfig(t, body)
	if len(iceServers) != 2 {
		t.Fatalf("iceServers: got %d entries, want 2", len(iceServers))
	}

	turn := iceServers[0]
	if turn.Username == "static-user" || turn.Credential == "static-password" {
		t.Errorf("static TURN password leaked into /config.js: %+v", turn)
	}
	expiry, err := strconv.ParseInt(turn.Username, 10, 64)
	if err != nil {
		t.Fatalf("username %q is not the coturn unix-expiry form", turn.Username)
	}
	remaining := time.Until(time.Unix(expiry, 0))
	if remaining <= 0 || remaining > 2*time.Hour+time.Minute {
		t.Errorf("username expiry %s from now, want inside the configured 2h TTL", remaining)
	}
	if !verifyTURNPair(secret, turn.Username, turn.Credential) {
		t.Errorf("credential %q does not verify as base64(HMAC-SHA1(secret, username))", turn.Credential)
	}

	if iceServers[1].Username != "" || iceServers[1].Credential != "" {
		t.Errorf("STUN-only entry must stay credential-free: %+v", iceServers[1])
	}
}

// TestConfigJSStaticICEWithoutTurnSecret pins the passthrough: with no
// turn_rest.secret, configured static credentials ship verbatim — the
// pre-REST behavior stays available (e.g. static-auth coturn setups).
func TestConfigJSStaticICEWithoutTurnSecret(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		cfg.ICEServers = []config.ICEServer{
			{URLs: []string{"turn:turn.example.org:3478"}, Username: "static-user", Credential: "static-password"},
		}
	})
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js status %d", resp.StatusCode)
	}
	iceServers := decodePBXConfig(t, body)
	if len(iceServers) != 1 {
		t.Fatalf("iceServers: got %d entries, want 1", len(iceServers))
	}
	if iceServers[0].Username != "static-user" || iceServers[0].Credential != "static-password" {
		t.Errorf("static ICE credentials must pass through verbatim without turn_rest.secret: %+v", iceServers[0])
	}
}
