package server

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
)

// getCallNote / putCallNote are the History note affordance's wire.
func getCallNote(t *testing.T, c *client, callID string) (*http.Response, []byte) {
	t.Helper()
	return c.do(http.MethodGet, "/api/calls/"+callID+"/note", nil, "")
}

func putCallNote(t *testing.T, c *client, callID, body string) (*http.Response, []byte) {
	t.Helper()
	payload, err := json.Marshal(struct {
		Body string `json:"body"`
	}{Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return c.do(http.MethodPut, "/api/calls/"+callID+"/note", payload, "application/json")
}

// TestCallNoteRoundTripOwnerScoped: a save is a 204, the read answers
// the body, another extension's call id is the same empty box, and an
// empty body clears the note while the slot stays.
func TestCallNoteRoundTripOwnerScoped(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	if resp, _ := getCallNote(t, c, "call-n1"); resp.StatusCode != http.StatusOK {
		t.Fatalf("unknown note read: %d, want 200", resp.StatusCode)
	}
	if resp, _ := putCallNote(t, c, "call-n1", "rueckruf am freitag"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("note save: %d", resp.StatusCode)
	}
	_, body := getCallNote(t, c, "call-n1")
	var got struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v (%s)", err, body)
	}
	if got.Body != "rueckruf am freitag" {
		t.Fatalf("note mismatch: %q", got.Body)
	}

	// Another extension's same call id: the empty box.
	other := newClient(t)
	other.login("2002", "pw")
	_, obody := getCallNote(t, other, "call-n1")
	var ogot struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(obody, &ogot); err != nil || ogot.Body != "" {
		t.Fatalf("another owner's note must be empty: %s (%v)", obody, err)
	}

	// Empty body clears; the read stays 200.
	if resp, _ := putCallNote(t, c, "call-n1", ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("clear: %d", resp.StatusCode)
	}
	_, body = getCallNote(t, c, "call-n1")
	if strings.Contains(string(body), "rueckruf") {
		t.Fatalf("cleared note must read empty: %s", body)
	}
}

// TestCallNoteGuards: malformed id, oversized body, and anonymous
// writes all fail honestly.
func TestCallNoteGuards(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	if resp, _ := getCallNote(t, c, ""); resp.StatusCode != http.StatusNotFound {
		t.Errorf("empty id never matches the route: %d, want 404 (styled catch-all)", resp.StatusCode)
	}
	if resp, _ := putCallNote(t, c, "bad%2Fid", "x"); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("path-escaping id: %d, want 400", resp.StatusCode)
	}
	long := strings.Repeat("x", 4001)
	if resp, _ := putCallNote(t, c, "call-n2", long); resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized note: %d, want 413", resp.StatusCode)
	}

	anon := newClient(t)
	anon.token = ""
	if resp, _ := putCallNote(t, anon, "call-n3", "hi"); resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Errorf("anonymous save: %d (want 401/403)", resp.StatusCode)
	}
}

// TestCallNoteNilSeamIs503: a composition without the notes store
// answers 503 (the transcript surface's honest-unavailable contract).
func TestCallNoteNilSeamIs503(t *testing.T) {
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Notes = nil })
	c := clientFor(t, server)
	c.login("1001", "pw")
	if resp, _ := getCallNote(t, c, "call-n4"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil-seam read: %d, want 503", resp.StatusCode)
	}
	if resp, _ := putCallNote(t, c, "call-n4", "x"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil-seam save: %d, want 503", resp.StatusCode)
	}
}
