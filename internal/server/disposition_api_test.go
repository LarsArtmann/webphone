package server

import (
	"encoding/json/v2"
	"net/http"
	"testing"
)

// confirmDisposition is the operator-verdict affordance's wire.
func confirmDisposition(t *testing.T, c *client, callID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return c.do(http.MethodPost, "/api/calls/"+callID+"/disposition", payload, "application/json")
}

// TestCallDispositionCorrectsTheCRM pins the issue #6 confirm: the
// operator's verdict persists on the review slot, journals the
// correction upstream with the confirmed disposition, the SAME verdict
// again is an inert 204 without a second journal entry, and a
// DIFFERENT verdict is a correction that journals again.
func TestCallDispositionCorrectsTheCRM(t *testing.T) {
	stub := &crmStub{}
	resolver := crmResolverFor(t, stub)
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
	c := signIn(t, server)

	body := func(disposition string) map[string]any {
		return map[string]any{
			"disposition": disposition, "number": "+493012345678",
			"direction": "out", "seconds": 63,
		}
	}

	resp, respBody := confirmDisposition(t, c, "call-d1", body("answered"))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("confirm: %d %s", resp.StatusCode, respBody)
	}
	stub.mu.Lock()
	if len(stub.logBodies) != 1 {
		t.Fatalf("upstream saw %d corrections", len(stub.logBodies))
	}
	got := stub.logBodies[0]
	stub.mu.Unlock()
	if got["outcome"] != "answered" || got["seconds"] != float64(63) {
		t.Fatalf("correction must carry the confirmed verdict: %+v", got)
	}

	// Same verdict again: inert — the CRM is NOT journaled twice.
	if resp, _ := confirmDisposition(t, c, "call-d1", body("answered")); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("replay: %d (want inert 204)", resp.StatusCode)
	}
	stub.mu.Lock()
	journaled := len(stub.logBodies)
	stub.mu.Unlock()
	if journaled != 1 {
		t.Fatalf("same-verdict replay must not journal again: %d entries", journaled)
	}

	// A different verdict IS a correction: journals again.
	if resp, _ := confirmDisposition(t, c, "call-d1", body("voicemail")); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("correction: %d (want 204)", resp.StatusCode)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if len(stub.logBodies) != 2 {
		t.Fatalf("correction must journal again: %d entries", len(stub.logBodies))
	}
	if got := stub.logBodies[1]["outcome"]; got != "voicemail" {
		t.Fatalf("correction verdict: %q", got)
	}
}

// TestCallDispositionVerdictSurvivesCRMOutage pins the persist-first
// order: a CRM 502 leaves the verdict stored (a fixed retry then
// journals exactly the correction) — the retry re-POSTs the same body.
func TestCallDispositionVerdictSurvivesCRMOutage(t *testing.T) {
	stub := &crmStub{}
	resolver := crmResolverFor(t, stub)
	stub.logStatus = http.StatusInternalServerError
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
	c := signIn(t, server)

	body := map[string]any{
		"disposition": "answered", "number": "+493012345678",
		"direction": "in", "seconds": 12,
	}
	if resp, _ := confirmDisposition(t, c, "call-d2", body); resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("outage confirm: %d (want 502)", resp.StatusCode)
	}

	// The verdict is on the slot: the retry after the CRM heals
	// journals exactly once (the failed attempt stayed retryable).
	stub.mu.Lock()
	stub.logStatus = http.StatusNoContent
	stub.mu.Unlock()
	if resp, _ := confirmDisposition(t, c, "call-d2", body); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("retry after outage: %d (want 204)", resp.StatusCode)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if len(stub.logBodies) != 1 {
		t.Fatalf("retry must journal exactly the correction: %d entries", len(stub.logBodies))
	}
}

// TestCallDispositionWithoutCRM pins the disabled-seam shape: verdict
// only, 204, no journal attempt — and an unknown number is the same
// silent 204 (the never-mint-contacts doctrine).
func TestCallDispositionWithoutCRM(t *testing.T) {
	server := newTestServer(t) // no CRM wired
	c := signIn(t, server)

	if resp, _ := confirmDisposition(t, c, "call-d3", map[string]any{
		"disposition": "ivr", "number": "+493012345678", "direction": "in", "seconds": 5,
	}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("disabled-CRM confirm: %d (want 204)", resp.StatusCode)
	}

	stub := &crmStub{}
	resolver := crmResolverFor(t, stub)
	stub.lookupJSON = `{"results":[]}`
	noMatch := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.CRM = resolver })
	c2 := clientFor(t, noMatch)
	c2.login("1001", "pw")
	if resp, _ := confirmDisposition(t, c2, "call-d3", map[string]any{
		"disposition": "ivr", "number": "+449900000000", "direction": "in", "seconds": 5,
	}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unknown-number confirm: %d (want 204)", resp.StatusCode)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if len(stub.logBodies) != 0 {
		t.Fatalf("unknown numbers must never journal: %+v", stub.logBodies)
	}
}

// TestCallDispositionGuards: foreign vocabulary, bad direction,
// malformed id, nil seam, and anonymous writes all fail honestly.
func TestCallDispositionGuards(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	if resp, _ := confirmDisposition(t, c, "call-d4", map[string]any{
		"disposition": "picked-up", "number": "+493012345678", "direction": "in",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("foreign disposition: %d (want 400)", resp.StatusCode)
	}
	if resp, _ := confirmDisposition(t, c, "call-d4", map[string]any{
		"disposition": "answered", "number": "+493012345678", "direction": "sideways",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad direction: %d (want 400)", resp.StatusCode)
	}
	if resp, _ := confirmDisposition(t, c, "bad%2Fid", map[string]any{
		"disposition": "answered", "number": "+493012345678", "direction": "in",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("path-escaping id: %d (want 400)", resp.StatusCode)
	}
	if resp, _ := confirmDisposition(t, c, "", map[string]any{"disposition": "answered"}); resp.StatusCode != http.StatusNotFound {
		t.Errorf("empty id never matches the route: %d, want 404 (styled catch-all)", resp.StatusCode)
	}

	anon := newClient(t)
	anon.token = ""
	if resp, _ := confirmDisposition(t, anon, "call-d5", map[string]any{
		"disposition": "answered", "number": "+493012345678", "direction": "in",
	}); resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Errorf("anonymous confirm: %d (want 401/403)", resp.StatusCode)
	}

	withoutNotes := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Notes = nil })
	cNil := clientFor(t, withoutNotes)
	cNil.login("1001", "pw")
	if resp, _ := confirmDisposition(t, cNil, "call-d6", map[string]any{
		"disposition": "answered", "number": "+493012345678", "direction": "in",
	}); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil-seam confirm: %d (want 503)", resp.StatusCode)
	}
}
