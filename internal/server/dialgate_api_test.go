package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/config"
)

// dialServer builds a server with the dial gate enabled for cooldown.
func dialServer(t *testing.T, cooldown time.Duration, mutate ...func(*Deps)) *testServer {
	t.Helper()
	return newTestServerWithConfig(t, "", func(c *config.Config) {
		c.Dial.Cooldown = cooldown
	}, mutate...)
}

// postDial is the advisory pre-invite gate ask (the body is raw JSON so
// the malformed-body case can send deliberate garbage).
func postDial(t *testing.T, c *client, body string) (*http.Response, []byte) {
	t.Helper()
	if body == "" {
		return c.do(http.MethodPost, "/api/dials", nil, "application/json")
	}
	return c.do(http.MethodPost, "/api/dials", []byte(body), "application/json")
}

// decodeDial reads the verdict.
func decodeDial(t *testing.T, body []byte) (allowed bool, lastAt, window int64) {
	t.Helper()
	var got struct {
		Allowed bool  `json:"allowed"`
		LastAt  int64 `json:"last_at"`
		Window  int64 `json:"window"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode dial verdict: %v (%s)", err, body)
	}
	return got.Allowed, got.LastAt, got.Window
}

// TestDialGateOffByDefault: the cooldown is OFF with zero config — the
// mounted route answers the disabled-seam 404 even with the store wired
// (the CONFIG is the gate, not the store) and /config.js carries no
// dialCooldownSeconds key (the opt-in-off byte-identical pin).
func TestDialGateOffByDefault(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")
	if resp, _ := postDial(t, c, `{"remote":"+4989123456"}`); resp.StatusCode != http.StatusNotFound {
		t.Errorf("gate with the cooldown off: %d, want 404", resp.StatusCode)
	}
	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js: %d", resp.StatusCode)
	}
	if strings.Contains(string(body), "dialCooldownSeconds") {
		t.Errorf("config.js must not carry the cooldown key when the gate is off: %s", body)
	}
}

// TestDialGateAllowsRecordsThenDenies: the first ask records the
// attempt and allows; the immediate replay inside the window denies
// with the slot's last-at stamp and the window length, carrying the
// config.js echo so the island can phrase the remaining wait.
func TestDialGateAllowsRecordsThenDenies(t *testing.T) {
	srv := dialServer(t, time.Hour)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	resp, body := postDial(t, c, `{"remote":"+4989123456"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first dial: %d, want 200", resp.StatusCode)
	}
	allowed, lastAt, window := decodeDial(t, body)
	if !allowed || lastAt != 0 || window != 0 {
		t.Errorf("first verdict: allowed=%v last_at=%d window=%d, want true/0/0", allowed, lastAt, window)
	}

	resp, body = postDial(t, c, `{"remote":"+4989123456"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("replay dial: %d, want 200", resp.StatusCode)
	}
	allowed, lastAt, window = decodeDial(t, body)
	if allowed {
		t.Errorf("replay inside the window must deny, got allowed=true")
	}
	if lastAt <= 0 {
		t.Errorf("denied verdict must carry the slot's last-at stamp, got %d", lastAt)
	}
	if window != 3600 {
		t.Errorf("denied verdict must carry the window length in seconds, got %d", window)
	}
}

// TestDialGateDenyNeverExtendsTheWindow: a denied ask records NOTHING —
// after the window runs out from the LAST ALLOWED dial, the next ask is
// allowed again (a retry loop must not grow the window it fights).
func TestDialGateDenyNeverExtendsTheWindow(t *testing.T) {
	srv := dialServer(t, 3*time.Second)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	if resp, _ := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("first dial: %d, want 200", resp.StatusCode)
	}
	if resp, body := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("immediate replay: %d", resp.StatusCode)
	} else if allowed, _, _ := decodeDial(t, body); allowed {
		t.Fatalf("replay inside the 3s window must deny")
	}
	time.Sleep(3400 * time.Millisecond) // unix-second stamps: wait past the window from the allowed stamp
	resp, body := postDial(t, c, `{"remote":"+4930123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("post-window dial: %d, want 200", resp.StatusCode)
	}
	if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Errorf("dial after the window must be allowed (denies must not extend the window)")
	}
}

// TestDialGateOverrideIsOneShot: the click-through arms the bypass and
// dials mid-window; the override is SPENT by that allowed attempt — the
// next plain ask inside the (re-stamped) window denies again.
func TestDialGateOverrideIsOneShot(t *testing.T) {
	srv := dialServer(t, time.Hour)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	if resp, body := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("seed dial: %d", resp.StatusCode)
	} else if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Fatalf("seed dial must be allowed")
	}
	resp, body := postDial(t, c, `{"remote":"+4930123"}`)
	if allowed, _, _ := decodeDial(t, body); resp.StatusCode != http.StatusOK || allowed {
		t.Fatalf("second dial must deny (seeds the window), got %d allowed=%v", resp.StatusCode, allowed)
	}

	resp, body = postDial(t, c, `{"remote":"+4930123","override":true}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("override dial: %d, want 200", resp.StatusCode)
	}
	if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Errorf("armed override must allow mid-window")
	}

	resp, body = postDial(t, c, `{"remote":"+4930123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("post-override dial: %d, want 200", resp.StatusCode)
	}
	if allowed, _, _ := decodeDial(t, body); allowed {
		t.Errorf("the override is one-shot: the next plain ask must deny again")
	}
}

// TestDialGateOwnerScoping: windows key on (owner, remote) — one
// extension's cooldown never touches another's (two clients against the
// SAME server).
func TestDialGateOwnerScoping(t *testing.T) {
	srv := dialServer(t, time.Hour)
	c := clientFor(t, srv)
	c.login("1001", "pw")
	other := clientFor(t, srv)
	other.login("2002", "pw")

	if resp, body := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("first owner dial: %d", resp.StatusCode)
	} else if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Fatalf("first owner dial must be allowed")
	}
	resp, body := postDial(t, other, `{"remote":"+4930123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("second owner dial: %d, want 200", resp.StatusCode)
	}
	if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Errorf("another extension's window must be independent, got denied")
	}
	resp, body = postDial(t, other, `{"remote":"+4930123"}`)
	if allowed, _, _ := decodeDial(t, body); resp.StatusCode != http.StatusOK || allowed {
		t.Errorf("second owner's replay must deny within its own window, got %d allowed=%v", resp.StatusCode, allowed)
	}
}

// TestDialGateBadInput: malformed bodies, unknown fields, and remotes
// that fail the bound/charset checks are 400 before any store touch;
// the gate sits behind the session (401) and CSRF gates like every
// session surface.
func TestDialGateBadInput(t *testing.T) {
	srv := dialServer(t, time.Hour)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	for name, body := range map[string]string{
		"no body":        "",
		"garbage":        `not json`,
		"unknown field":  `{"remote":"+4930123","bogus":1}`,
		"empty remote":   `{"remote":"  "}`,
		"oversized":      `{"remote":"` + strings.Repeat("1", 65) + `"}`,
		"slash remote":   `{"remote":"49/30"}`,
		"percent remote": `{"remote":"49%2030"}`,
	} {
		if resp, _ := postDial(t, c, body); resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", name, resp.StatusCode)
		}
	}

	anon := clientFor(t, srv)
	if resp, _ := postDial(t, anon, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unauthenticated ask: %d, want 401", resp.StatusCode)
	}
}

// TestDialGateStoreFailureIs502: a broken window store answers the
// honest 502 — the island dials anyway (the gate is advisory), but the
// failure is visible, not swallowed into a fake allow.
func TestDialGateStoreFailureIs502(t *testing.T) {
	dead := newTestEngine(t)
	closed := dead.DialWindows()
	if err := dead.Shutdown(); err != nil {
		t.Fatalf("close engine: %v", err)
	}
	srv := dialServer(t, time.Hour, func(d *Deps) { d.Dials = closed })
	c := clientFor(t, srv)
	c.login("1001", "pw")
	if resp, _ := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusBadGateway {
		t.Errorf("dead store: %d, want 502", resp.StatusCode)
	}
}
