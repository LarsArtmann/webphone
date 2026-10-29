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
// The base dial-gate pins (advise+record, one-shot override, off→404,
// anon→401) live in dialgate_test.go; this file holds the pins that
// file does not carry.
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

// TestDialGateConfigSilentWhenOff: the opt-in-off contract — /config.js
// carries NO dialCooldownSeconds key with zero config (byte-identical
// off) and the honest window length once the gate is on.
func TestDialGateConfigSilentWhenOff(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")
	_, body := c.do(http.MethodGet, "/config.js", nil, "")
	if strings.Contains(string(body), "dialCooldownSeconds") {
		t.Errorf("config.js must not carry the cooldown key when the gate is off: %s", body)
	}

	on := clientFor(t, dialServer(t, time.Hour))
	on.login("1001", "pw")
	resp, body := on.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js (gate on): %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"dialCooldownSeconds":3600`) {
		t.Errorf("config.js must carry the window when the gate is on: %s", body)
	}
}

// TestDialGateAllowedVerdictCarriesNoState: an ALLOWED answer is exactly
// the verdict — no last_at stamp, no window echo (those are the denied
// ask's cooldown-phrasing state, per the DTO contract).
func TestDialGateAllowedVerdictCarriesNoState(t *testing.T) {
	c := clientFor(t, dialServer(t, time.Hour))
	c.login("1001", "pw")
	resp, body := postDial(t, c, `{"remote":"+4989123456"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first dial: %d, want 200", resp.StatusCode)
	}
	if allowed, lastAt, window := decodeDial(t, body); !allowed || lastAt != 0 || window != 0 {
		t.Errorf("allowed verdict: allowed=%v last_at=%d window=%d, want true/0/0", allowed, lastAt, window)
	}
}

// TestDialGateDenyNeverExtendsTheWindow: a denied ask records NOTHING —
// after the window runs out from the LAST ALLOWED dial, the next ask is
// allowed again (a retry loop must not grow the window it fights).
func TestDialGateDenyNeverExtendsTheWindow(t *testing.T) {
	c := clientFor(t, dialServer(t, 3*time.Second))
	c.login("1001", "pw")

	if resp, body := postDial(t, c, `{"remote":"+4930123"}`); resp.StatusCode != http.StatusOK {
		t.Fatalf("first dial: %d", resp.StatusCode)
	} else if allowed, _, _ := decodeDial(t, body); !allowed {
		t.Fatalf("first dial must be allowed")
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
// that fail the bound/charset checks are 400 before any store touch.
func TestDialGateBadInput(t *testing.T) {
	c := clientFor(t, dialServer(t, time.Hour))
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
