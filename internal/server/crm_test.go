package server

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
)

// crmStub is a Ledger machine-API double for the server tests.
type crmStub struct {
	mu         sync.Mutex
	logBodies  []map[string]any
	lookupJSON string
	logStatus  int
}

func (s *crmStub) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/contacts/by-phone" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.lookupJSON)) //nolint:erraudit // test stub write
		return
	}
	if r.URL.Path == "/api/contacts/01M/calls" && r.Method == http.MethodPost {
		s.mu.Lock()
		var got map[string]any
		_ = json.UnmarshalRead(r.Body, &got) //nolint:errcheck // test stub decode
		s.logBodies = append(s.logBodies, got)
		status := s.logStatus
		s.mu.Unlock()

		w.WriteHeader(status)
		return
	}
	http.NotFound(w, r)
}

// crmResolverFor boots a stub CRM that knows exactly one number.
func crmResolverFor(t *testing.T, stub *crmStub) *crm.Resolver {
	t.Helper()
	stub.lookupJSON = `{"results":[{"id":"01M","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}]}`
	stub.logStatus = http.StatusNoContent
	srv := httptest.NewServer(http.HandlerFunc(stub.handler))
	t.Cleanup(srv.Close)

	client, err := crm.NewClient(srv.URL, "crm-token")
	if err != nil {
		t.Fatalf("crm client: %v", err)
	}
	return crm.NewResolver(client, nil)
}

func TestAPICallLoggingContract(t *testing.T) {
	report := func(t *testing.T, c *client, payload map[string]any) (int, string) {
		t.Helper()
		resp, body := postJSONRaw(t, c, "/api/calls", payload)
		return resp.StatusCode, body
	}

	t.Run("anonymous callers are refused", func(t *testing.T) {
		server := newTestServer(t)
		anon := clientFor(t, server)
		anon.token = ""

		resp, _ := anon.do(http.MethodPost, "/api/calls", []byte(`{}`), "application/json")
		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
			t.Fatalf("anonymous report: %d (want 401/403)", resp.StatusCode)
		}
	})

	t.Run("matched number journals the call upstream", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)

		status, body := report(t, c, map[string]any{
			"number": "+493012345678", "direction": "in", "seconds": 123, "outcome": "answered",
		})
		if status != http.StatusNoContent {
			t.Fatalf("report: %d %s", status, body)
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()

		if len(stub.logBodies) != 1 {
			t.Fatalf("upstream saw %d call logs", len(stub.logBodies))
		}

		got := stub.logBodies[0]
		if got["direction"] != "in" || got["seconds"] != float64(123) || got["outcome"] != "answered" {
			t.Fatalf("forwarded facts: %+v", got)
		}
	})

	t.Run("unknown number is a silent 204", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		stub.lookupJSON = `{"results":[]}`
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)

		status, _ := report(t, c, map[string]any{"number": "+449900000000", "direction": "out"})
		if status != http.StatusNoContent {
			t.Fatalf("unknown-number report: %d (want 204)", status)
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()

		if len(stub.logBodies) != 0 {
			t.Fatalf("unknown numbers must never reach the CRM: %+v", stub.logBodies)
		}
	})

	t.Run("CRM outage is a 502", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		stub.logStatus = http.StatusInternalServerError
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)

		status, _ := report(t, c, map[string]any{"number": "+493012345678", "direction": "in"})
		if status != http.StatusBadGateway {
			t.Fatalf("CRM outage: %d (want 502)", status)
		}
	})

	t.Run("same key journals once, replay is an inert 204", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)
		payload := func(key string) map[string]any {
			return map[string]any{"number": "+493012345678", "direction": "in", "key": key}
		}

		if status, body := report(t, c, payload("11111111-1111-4111-8111-111111111111")); status != http.StatusNoContent {
			t.Fatalf("first report: %d %s", status, body)
		}
		if status, _ := report(t, c, payload("11111111-1111-4111-8111-111111111111")); status != http.StatusNoContent {
			t.Fatalf("replayed report: %d (want inert 204)", status)
		}
		if status, _ := report(t, c, payload("22222222-2222-4222-8222-222222222222")); status != http.StatusNoContent {
			t.Fatalf("different-key report: %d (want 204)", status)
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()
		if len(stub.logBodies) != 2 {
			t.Fatalf("dedupe is key-scoped: journaled %d times (want 2: once per key)", len(stub.logBodies))
		}
	})

	t.Run("502 keeps the key retryable", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		stub.logStatus = http.StatusInternalServerError
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)
		payload := map[string]any{"number": "+493012345678", "direction": "in", "key": "33333333-3333-4333-8333-333333333333"}

		if status, _ := report(t, c, payload); status != http.StatusBadGateway {
			t.Fatalf("failing report: %d (want 502)", status)
		}

		stub.mu.Lock()
		stub.logStatus = http.StatusNoContent
		stub.mu.Unlock()

		if status, body := report(t, c, payload); status != http.StatusNoContent {
			t.Fatalf("retry after 502: %d %s (want 204)", status, body)
		}
		// The stub journals before answering, so the 502 attempt already
		// left one entry (received-but-failed). The retry must ADD the
		// second; a further replay must then be deduped by the key.
		if status, _ := report(t, c, payload); status != http.StatusNoContent {
			t.Fatalf("replay after success: %d (want inert 204)", status)
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()
		if len(stub.logBodies) != 2 {
			t.Fatalf("retry+replay journaled %d times (want 2: failed attempt, retry; replay deduped)", len(stub.logBodies))
		}
	})

	t.Run("absent key never dedupes", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)
		payload := map[string]any{"number": "+493012345678", "direction": "in"}

		for i := 0; i < 2; i++ {
			if status, _ := report(t, c, payload); status != http.StatusNoContent {
				t.Fatalf("report %d: got %d (want 204)", i+1, status)
			}
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()
		if len(stub.logBodies) != 2 {
			t.Fatalf("legacy no-key shape must journal every call, saw %d", len(stub.logBodies))
		}
	})

	t.Run("unknown-number drop consumes the key", func(t *testing.T) {
		stub := &crmStub{}
		resolver := crmResolverFor(t, stub)
		stub.lookupJSON = `{"results":[]}`
		server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
		c := signIn(t, server)
		key := "44444444-4444-4444-8444-444444444444"

		if status, _ := report(t, c, map[string]any{"number": "+449900000000", "direction": "out", "key": key}); status != http.StatusNoContent {
			t.Fatalf("dropped report: %d (want 204)", status)
		}

		stub.mu.Lock()
		stub.lookupJSON = `{"results":[{"id":"01M","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}]}`
		stub.mu.Unlock()

		if status, _ := report(t, c, map[string]any{"number": "+493012345678", "direction": "in", "key": key}); status != http.StatusNoContent {
			t.Fatalf("same-key report for a known number: %d (want inert 204)", status)
		}

		stub.mu.Lock()
		defer stub.mu.Unlock()
		if len(stub.logBodies) != 0 {
			t.Fatalf("a consumed key must never journal later, saw %+v", stub.logBodies)
		}
	})

	t.Run("disabled CRM is a no-op 204", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		status, _ := report(t, c, map[string]any{"number": "+493012345678", "direction": "in"})
		if status != http.StatusNoContent {
			t.Fatalf("disabled CRM report: %d (want 204)", status)
		}
	})

	t.Run("malformed bodies are 400", func(t *testing.T) {
		server := newTestServer(t)
		c := signIn(t, server)

		resp, _ := c.do(http.MethodPost, "/api/calls", []byte(`{"direction":"sideways"}`), "application/json")
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("bad direction: %d (want 400)", resp.StatusCode)
		}

		resp, _ = c.do(http.MethodPost, "/api/calls", []byte(`not-json`), "application/json")
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("not-json: %d (want 400)", resp.StatusCode)
		}
	})
}

// TestMetricsRendersCRMLookupCounters pins the metrics family: with the
// integration on, /metrics carries the resolver's upstream-outcome
// counters (one lookup each classified hit / miss / failure); aggregates
// only, never a number or a contact name.
func TestMetricsRendersCRMLookupCounters(t *testing.T) {
	stub := &crmStub{}
	stub.lookupJSON = `{"results":[{"id":"01M","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}]}`
	stub.logStatus = http.StatusNoContent
	upstream := httptest.NewServer(http.HandlerFunc(stub.handler))
	client, err := crm.NewClient(upstream.URL, "crm-token")
	if err != nil {
		t.Fatalf("crm client: %v", err)
	}
	resolver := crm.NewResolver(client, nil)

	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
	c := signIn(t, server)
	report := func(t *testing.T, number, key string) int {
		t.Helper()
		resp, _ := postJSONRaw(t, c, "/api/calls", map[string]any{"number": number, "direction": "in", "key": key})
		return resp.StatusCode
	}

	if status := report(t, "+493012345678", "61111111-1111-4111-8111-111111111111"); status != http.StatusNoContent {
		t.Fatalf("known-number report: %d", status)
	}

	stub.mu.Lock()
	stub.lookupJSON = `{"results":[]}`
	stub.mu.Unlock()
	if status := report(t, "+449900000001", "62222222-2222-4222-8222-222222222222"); status != http.StatusNoContent {
		t.Fatalf("unknown-number report: %d", status)
	}

	upstream.Close()
	if status := report(t, "+493012345679", "63333333-3333-4333-8333-333333333333"); status != http.StatusNoContent {
		t.Fatalf("unreachable-CRM report: %d", status)
	}

	resp, body := c.do(http.MethodGet, "/metrics", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics: %d", resp.StatusCode)
	}
	page := string(body)
	for _, want := range []string{
		"# TYPE webphone_crm_lookups_total counter",
		`webphone_crm_lookups_total{outcome="hit"} 1`,
		`webphone_crm_lookups_total{outcome="miss"} 1`,
		`webphone_crm_lookups_total{outcome="failure"} 1`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("metrics missing %q:\n%s", want, page)
		}
	}
	for _, leak := range []string{"+493012345678", "Ada"} {
		if strings.Contains(page, leak) {
			t.Errorf("metrics leaks %q — aggregates only:\n%s", leak, page)
		}
	}
}

// postJSONRaw posts a JSON payload without the >=500 guard of postJSON —
// the 502 path of the CRM contract is exactly what this test asserts.
func postJSONRaw(t *testing.T, c *client, path string, payload any) (*http.Response, string) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	resp, respBody := c.do(http.MethodPost, path, body, "application/json")
	return resp, string(respBody)
}

// TestHistoryRendersCRMNames pins the enrichment: with the integration
// enabled, a matched CDR party renders the CRM name (the number stays on
// the data-dial button), and the island learns about the integration via
// PBX_CONFIG.crm.
func TestHistoryRendersCRMNames(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.RequestURI(), "/phone-api/voicemail/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"new":0,"old":0}`))
			return
		}
		if !strings.HasPrefix(r.URL.RequestURI(), "/phone-api/history") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"entries":[
			{"context":"from-internal","caller_id_number":"1001","caller_id_name":"","destination_number":"+493012345678","start":"2026-09-18 10:00","billsec":12}
		]}`))
	}))
	t.Cleanup(upstream.Close)

	stub := &crmStub{}
	resolver := crmResolverFor(t, stub)
	server := newTestServerWithPhoneAPI(t, upstream.URL, func(d *Deps) { d.CRM = resolver })
	c := signIn(t, server)

	_, body := c.do(http.MethodGet, "/partials/history", nil, "")
	page := string(body)
	if !strings.Contains(page, "Ada Lovelace") {
		t.Errorf("CRM name missing from the history row: %.400s", page)
	}
	if !strings.Contains(page, `data-dial="+493012345678"`) {
		t.Errorf("the number must stay on the dial button: %.400s", page)
	}

	_, cfg := c.do(http.MethodGet, "/config.js", nil, "")
	if !strings.Contains(string(cfg), `"crm":true`) {
		t.Errorf("config.js must advertise the CRM integration: %.200s", cfg)
	}
}

// TestConfigJSCarriesCRMKey keeps the wire contract honest when the
// integration is off: the key exists (false), so the island's gate is a
// plain feature test, not an existence probe.
func TestConfigJSCarriesCRMKey(t *testing.T) {
	server := newTestServerWithConfig(t, "", func(cfg *config.Config) {})
	c := signIn(t, server)

	_, cfg := c.do(http.MethodGet, "/config.js", nil, "")
	if !strings.Contains(string(cfg), `"crm":false`) {
		t.Errorf("config.js must carry the crm key even when off: %.200s", cfg)
	}
}
