package app

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCRMLegGatesReadinessDeliberately pins the criticality DECISION for the
// CRM health leg (TODO §2026-11-06 r6, closed 2026-11-07).
//
// The row asked whether a CRM outage should red `/healthz` readiness although
// the phone keeps working. Measured first, then decided: it DOES red it, and
// that is now the INTENDED behaviour — the code and the doctrine disagreed
// before (a comment in the composition root called the leg "non-critical by
// the seam's own doctrine", which is true of the go-health probe and false of
// this endpoint, which renders a list of checks and fails if any one fails).
//
// Why keep it gating:
//
//   - It is the only surface that reports a REVOKED CRM token when the health
//     dashboard is off (the check is authenticated on purpose, issue #9). The
//     alternative — dropping the leg — silently loses a detection lane, which
//     this repo treats as a bug, not a simplification.
//   - Nothing routes traffic on `/healthz`: the unit is `Type=simple` and the
//     consuming stack fences the probes with `remote_ip` matchers, so a 503 is
//     an ALARM read by an operator, never a traffic decision.
//   - The two audiences stay coherent: `/healthz` is the strict "every
//     dependency works" alarm, while the go-health probe (dashboard, /livez,
//     /startupz) grades the same leg as a non-critical WARN — same fact, two
//     readings, both true.
//
// If a future deployment puts a load balancer in front of `/healthz`, this
// test is the place to reconsider: a CRM outage must never remove the phone
// from service, and today the only thing that can act on the 503 is a human.
func TestCRMLegGatesReadinessDeliberately(t *testing.T) {
	cfg := testConfig(t, false)
	// A port nothing listens on: the authenticated probe must fail against it.
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

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/healthz = %d with an unreachable CRM, want 503 (the leg is a deliberate gate); body:\n%s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"crm"`) {
		t.Errorf("the 503 must NAME the failing leg, otherwise the operator cannot read the fix off the body; body:\n%s", body)
	}
	// The critical pair must still read green: a CRM outage says nothing about
	// the product's own storage.
	for _, name := range []string{`"sqlite"`, `"blob-dir"`} {
		if !strings.Contains(string(body), name) {
			t.Errorf("%s missing from the readiness body; body:\n%s", name, body)
		}
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

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz = %d without a CRM, want 200; body:\n%s", resp.StatusCode, body)
	}
	if strings.Contains(string(body), `"crm"`) {
		t.Errorf("a CRM-less deployment must not carry a crm leg; body:\n%s", body)
	}
}
