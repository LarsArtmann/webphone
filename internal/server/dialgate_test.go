// Dial-gate end-to-end pins (issue #7): the advisory pre-invite gate
// behind the real route — off answers 404, an allowed ask records, the
// same remote inside the window is denied with the state the island
// phrases (last_at + window), and the one-shot override spends exactly
// once. The store's own semantics are pinned in readmodels.
package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

func newDialGateClient(t *testing.T) *client {
	t.Helper()
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {
		// The tweak runs before the deps wire: the window flows into
		// api.Deps the same way the composition root's would.
		cfg.Dial.Cooldown = time.Hour
	}, func(deps *Deps) {
		deps.Dials = newTestEngine(t).DialWindows()
	})
	c := clientFor(t, server)
	c.login("1001", "pw")
	return c
}

func TestDialGateAdvisesAndRecords(t *testing.T) {
	c := newDialGateClient(t)

	post := func(body string) (*http.Response, []byte) {
		return c.do(http.MethodPost, "/api/dials", []byte(body), "application/json")
	}

	// The first ask is allowed and recorded.
	resp, body := post(`{"remote":"+4917012345678"}`)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"allowed":true`) {
		t.Fatalf("first ask: %d %s", resp.StatusCode, body)
	}
	// The same remote inside the window is denied, carrying the state
	// the island phrases: the last-at stamp and the window length.
	resp, body = post(`{"remote":"+4917012345678"}`)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"allowed":false`) {
		t.Fatalf("second ask: %d %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"window":3600`) || !strings.Contains(string(body), `"last_at"`) {
		t.Fatalf("denied verdict must carry last_at + window: %s", body)
	}
	// A different remote is a different slot.
	if resp, _ = post(`{"remote":"+4917099887766"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("other remote: %d", resp.StatusCode)
	}
	// Malformed remotes never reach the store.
	resp, _ = post(`{"remote":"sip:user@host/pct%"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed remote: %d", resp.StatusCode)
	}
}

func TestDialGateOverrideSpendsOnce(t *testing.T) {
	c := newDialGateClient(t)
	post := func(body string) *http.Response {
		resp, _ := c.do(http.MethodPost, "/api/dials", []byte(body), "application/json")
		return resp
	}

	post(`{"remote":"+17155550132"}`) // allowed, recorded
	if post(`{"remote":"+17155550132"}`).StatusCode != http.StatusOK {
		t.Fatal("window ask")
	}
	// The click-through override is allowed and spent…
	resp, body := c.do(http.MethodPost, "/api/dials", []byte(`{"remote":"+17155550132","override":true}`), "application/json")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"allowed":true`) {
		t.Fatalf("override ask: %d %s", resp.StatusCode, body)
	}
	// …so the NEXT plain ask is denied again (a spent override never
	// becomes a standing bypass).
	resp, body = c.do(http.MethodPost, "/api/dials", []byte(`{"remote":"+17155550132"}`), "application/json")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"allowed":false`) {
		t.Fatalf("post-override ask must be denied: %d %s", resp.StatusCode, body)
	}
}

func TestDialGateOffAnswers404AndAnonIs401(t *testing.T) {
	// Off: the route itself answers 404 (disabled-seam convention).
	c := newClient(t)
	c.login("1001", "pw")
	resp, body := c.do(http.MethodPost, "/api/dials", []byte(`{"remote":"+17155550132"}`), "application/json")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("gate off: %d %s", resp.StatusCode, body)
	}

	// Session-gated: an anonymous ask never reaches the store.
	anon := newClient(t)
	if resp, _ = anon.do(http.MethodPost, "/api/dials", []byte(`{"remote":"+17155550132"}`), "application/json"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous ask: %d", resp.StatusCode)
	}
}
