package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
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
	if !turnPairVerifies(secret, turn.Username, turn.Credential) {
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

// TestConfigJSCarriesNoSharedDirectory pins the contacts-data gate
// (2026-11-13 mesh review, T10): /config.js is mounted PRE-AUTH, so the
// operator's shared contact directory must never ride it — an anonymous
// fetcher must not be able to enumerate who shares this phone. The
// directory reaches the island exclusively through the session-gated
// GET /api/contacts answer ({shared}, lowercase name/number wire keys —
// the shape the island typeahead reads; the capitalized Go-style shape
// is asserted ABSENT there so the tags cannot regress unnoticed).
//
// Cross-repo seam: the nix-international-telephony stack's
// tests/configjs_check.py fixture and tests/browser-e2e.py contacts
// round-trip consume this contract — they were updated in the same
// train. Change all three together.
func TestConfigJSCarriesNoSharedDirectory(t *testing.T) {
	server := newTestServerWithConfig(t, "", nil, func(d *Deps) {
		d.Shared = []domain.SharedContact{{Name: `O"Brien`, Number: "1000"}}
	})

	// The anonymous boot probe is the leak: the directory must be absent
	// from the payload it fetches without a session.
	anon := clientFor(t, server)

	anonResp, anonBody := anon.do(http.MethodGet, "/config.js", nil, "")
	if anonResp.StatusCode != http.StatusOK {
		t.Fatalf("anonymous config.js status %d", anonResp.StatusCode)
	}
	if strings.Contains(string(anonBody), "contacts") {
		t.Errorf("anonymous /config.js leaks the shared directory:\n%s", anonBody)
	}

	// The signed-in payload carries the same absence (one contract, both
	// states; the island adopts the directory post-login instead).
	c := signIn(t, server)
	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js status %d", resp.StatusCode)
	}
	if strings.Contains(string(body), "contacts") {
		t.Errorf("signed-in /config.js still carries the shared directory:\n%s", body)
	}

	// The authenticated read is where the directory lives now: the wire
	// keys stay lowercase name/number (the typeahead's contract), and
	// the quoted name doubles as the JSON-escaping regression.
	resp, body = c.do(http.MethodGet, "/api/contacts", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/contacts status %d", resp.StatusCode)
	}
	var parsed struct {
		Shared []map[string]string `json:"shared"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("/api/contacts payload is not JSON: %v\n%s", err, body)
	}
	if len(parsed.Shared) != 1 {
		t.Fatalf("shared: got %d entries, want 1\n%s", len(parsed.Shared), body)
	}
	contact := parsed.Shared[0]
	if contact["name"] != `O"Brien` || contact["number"] != "1000" {
		t.Errorf("shared wire keys must be lowercase name/number (the island typeahead reads them): got %v", contact)
	}
	if _, ok := contact["Name"]; ok {
		t.Errorf("capitalized Name leaked onto the wire: %v", contact)
	}
	if _, ok := contact["Number"]; ok {
		t.Errorf("capitalized Number leaked onto the wire: %v", contact)
	}
}

// turnPairVerifies recomputes the coturn HMAC for username and reports
// whether credential matches — the test-side oracle for the derivation
// (the derivation itself lives in the platform package's configjs).
func turnPairVerifies(secret, username, credential string) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	return hmac.Equal([]byte(credential), []byte(base64.StdEncoding.EncodeToString(mac.Sum(nil))))
}
