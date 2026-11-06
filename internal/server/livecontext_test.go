package server

import (
	"net/http"
	"strings"
	"testing"
)

// liveContextServer boots a stub CRM that knows one number and publishes
// context for it, then wires the resolver into a phone-API-less server
// (the live panel needs no CDR). The container itself only exists when
// the seam is on, which the shell test below pins.
func liveContextServer(t *testing.T) *testServer {
	t.Helper()

	stub := &crmStub{}
	stub.lookupJSON = `{"results":[{"id":"01M","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}]}`
	stub.contextJSON = `{"company":"Analytical Engines","open_deals":[{"title":"Difference engine","stage":"proposal","value_cents":250000,"currency":"EUR"}],"open_deal_count":1,"open_tasks":3,"next_task":{"title":"Send the spec","due_at":"2026-11-09T17:00:00Z","overdue":false},"last_activity_at":"2026-11-03T09:00:00Z"}`
	stub.contextStatus = http.StatusOK

	return newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = crmResolverFor(t, stub) })
}

// TestLiveContextPartialRendersTheCallersContext pins issue #10's first
// half: the shell's slot answers with who is on the line and what is
// outstanding with them, from the SAME context product the conversation
// strip renders.
func TestLiveContextPartialRendersTheCallersContext(t *testing.T) {
	server := liveContextServer(t)
	c := signIn(t, server)

	resp, body := c.do(http.MethodGet, "/partials/live-context?remote=%2B441632960961", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("live context = %d, want 200", resp.StatusCode)
	}

	page := string(body)
	for _, want := range []string{
		"wp-live-context",
		"Ada Lovelace",
		"Analytical Engines",
		"Difference engine",
		"Send the spec",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("live context missing %q: %.400s", want, page)
		}
	}

	if strings.Contains(page, `id="wp-crm-context"`) {
		t.Error("the live panel must not reuse the conversation strip's id (a duplicate id is a broken page)")
	}
}

// TestLiveContextPartialStaysSilentOnEveryDegradation pins the quiet-half
// contract: no seam, no match, a non-phone remote and an empty remote all
// answer an EMPTY 200 body — during a live call there is no room for an
// error banner, and an empty answer renders nothing in the slot.
func TestLiveContextPartialStaysSilentOnEveryDegradation(t *testing.T) {
	t.Run("no seam configured", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		resp, body := c.do(http.MethodGet, "/partials/live-context?remote=%2B441632960961", nil, "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("disabled seam = %d, want a quiet 200", resp.StatusCode)
		}

		if strings.TrimSpace(string(body)) != "" {
			t.Errorf("disabled seam answered %q, want an empty body", string(body))
		}
	})

	t.Run("no contact match", func(t *testing.T) {
		// A CRM that knows nobody: the resolver is built with the canned
		// Ada answer (crmResolverFor owns that setup), then the stub is
		// switched to an empty result set — which is what a real miss
		// looks like on the wire.
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		stub.lookupJSON = `{"results":[]}`
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)

		resp, body := c.do(http.MethodGet, "/partials/live-context?remote=%2B491700000000", nil, "")
		if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "" {
			t.Errorf("unmatched number = %d %q, want a quiet empty 200", resp.StatusCode, string(body))
		}
	})

	t.Run("unparseable remote", func(t *testing.T) {
		server := liveContextServer(t)
		c := signIn(t, server)

		// Longer than any dialable number (the domain rule caps at 32
		// characters), so it can never reach a CRM lookup.
		remote := strings.Repeat("9", 40)

		resp, body := c.do(http.MethodGet, "/partials/live-context?remote="+remote, nil, "")
		if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "" {
			t.Errorf("oversized remote = %d %q, want a quiet empty 200", resp.StatusCode, string(body))
		}
	})

	t.Run("missing remote", func(t *testing.T) {
		server := liveContextServer(t)
		c := signIn(t, server)

		resp, body := c.do(http.MethodGet, "/partials/live-context", nil, "")
		if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "" {
			t.Errorf("no remote = %d %q, want a quiet empty 200", resp.StatusCode, string(body))
		}
	})
}

// TestLiveContextPartialIsSessionGated pins the door: the fragment renders
// CRM facts, so it answers only for a signed-in caller.
func TestLiveContextPartialIsSessionGated(t *testing.T) {
	server := liveContextServer(t)
	anon := clientFor(t, server)
	anon.token = ""

	resp, _ := anon.do(http.MethodGet, "/partials/live-context?remote=%2B441632960961", nil, "")
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("anonymous live context = %d, want 401/403", resp.StatusCode)
	}
}

// TestLiveContextSlotExistsOnlyWithTheSeam pins the served-markup half:
// the aside carries the slot when the CRM is configured (the shell needs
// somewhere to swap into) and omits it entirely otherwise, so a
// CRM-less deployment stays byte-identical.
func TestLiveContextSlotExistsOnlyWithTheSeam(t *testing.T) {
	t.Run("with the seam", func(t *testing.T) {
		server := liveContextServer(t)
		c := signIn(t, server)

		_, body := c.do(http.MethodGet, "/", nil, "")
		if !strings.Contains(string(body), `id="wp-live-context"`) {
			t.Error("a configured CRM must render the island's live-context slot")
		}
	})

	t.Run("without the seam", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		_, body := c.do(http.MethodGet, "/", nil, "")
		if strings.Contains(string(body), `id="wp-live-context"`) {
			t.Error("zero CRM config must not render the slot (byte-identical product)")
		}
	})
}
