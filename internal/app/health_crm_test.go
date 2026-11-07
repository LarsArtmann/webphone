package app

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCRMLegNeverGatesReadiness pins the criticality DECISION for the CRM
// health leg (TODO §2026-11-06 r6, closed 2026-11-07).
//
// The question the row asked: a CRM outage reds `/healthz` readiness even
// though the phone keeps working. The answer, verified here rather than
// assumed: it does NOT. Only the critical pair (`sqlite`, `blob-dir`) routes
// on readiness — `health.WithCriticalServices("sqlite", "blob-dir")` — so a
// CRM leg can only ever be an alarm channel. That is the correct posture: a
// dead CRM costs toasts and raw caller names, never call flow, and nothing
// load-balances on `/healthz` today (the unit is `Type=simple` and Caddy
// fences the probes with `remote_ip` matchers).
//
// The leg is still DISCOVERED and REPORTED, which is what makes the outage
// visible to an operator instead of silent. This test holds both halves: the
// endpoint stays ready, and the failing leg is named in the body.
func TestCRMLegNeverGatesReadiness(t *testing.T) {
	cfg := testConfig(t, false)
	// A port that nothing listens on: the health probe must fail against it.
	cfg.CRM.URL = "http://127.0.0.1:1"
	cfg.CRM.Token = "test-token"

	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(cancel)
	t.Cleanup(func() {
		if err := application.Shutdown(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	})

	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz = %d with an unreachable CRM — the crm leg must NEVER gate readiness; body:\n%s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"crm"`) {
		t.Errorf("the crm leg must still be REPORTED (an outage has to be visible, not silent); body:\n%s", body)
	}
}

// TestNoCRMLegWhenSeamIsOff pins the other direction: zero config means the
// leg does not exist at all, so a CRM-less deployment's probe surface is
// byte-identical to the pre-CRM product.
func TestNoCRMLegWhenSeamIsOff(t *testing.T) {
	application := startApp(t, false)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), `"crm"`) {
		t.Errorf("a CRM-less deployment must not carry a crm leg; body:\n%s", body)
	}
}
