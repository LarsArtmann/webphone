package server

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
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
	// The bound is characters, not bytes: 4000 CJK characters are 12000
	// bytes, and the editors' maxLength=4000 counts characters — a byte
	// comparison would reject a note the UI promised to accept.
	multibyte := strings.Repeat("字", 4000)
	if resp, _ := putCallNote(t, c, "call-n2", multibyte); resp.StatusCode != http.StatusNoContent {
		t.Errorf("4000-character multibyte note: %d, want 204", resp.StatusCode)
	}
	overChars := strings.Repeat("字", 4001)
	if resp, _ := putCallNote(t, c, "call-n2", overChars); resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("4001-character multibyte note: %d, want 413", resp.StatusCode)
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

// TestCallNoteSaveNudgesHistory pins P2-3: every CHANGING note save publishes
// exactly one payload-less "notes" SSE event to the OWNER's hub, so a second
// open tab showing the same call re-fetches instead of holding a stale box.
//
// The nudge is deliberately not published on FAILURE: a 409 (note list full) or
// a store error must leave other tabs exactly as they were. It is not
// published on ECHOES either: a same-body autosave save (what a re-mounted
// editor PUTs right after its GET) is journal-silent, and other tabs already
// hold exactly what they show.
func TestCallNoteSaveNudgesHistory(t *testing.T) {
	srv := newTestServer(t)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	owner := domain.MustParseExtension("1001")
	events := srv.hubs.get(owner).Hub().Subscribe()
	t.Cleanup(func() { srv.hubs.get(owner).Hub().Unsubscribe(events) })

	if resp, _ := putCallNote(t, c, "call-nudge-note", "called back"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("note save: %d", resp.StatusCode)
	}

	select {
	case ev := <-events:
		if ev.Event != "notes" {
			t.Fatalf("nudge event = %q, want \"notes\"", ev.Event)
		}
		if len(ev.Data) != 0 {
			t.Errorf("the notes nudge must be payload-less, got %.80s", ev.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no notes nudge within 2s — a second tab would keep a stale note box")
	}

	// Exactly one, never a burst.
	select {
	case ev := <-events:
		t.Fatalf("second push %q: a save owes exactly one nudge", ev.Event)
	case <-time.After(200 * time.Millisecond):
	}

	// Wait until the save is client-visible, then PUT the same body:
	// an autosave echo is still a 204 but owes NO nudge (the journal
	// already holds this note).
	deadline := time.Now().Add(2 * time.Second)
	for {
		_, body := getCallNote(t, c, "call-nudge-note")
		if strings.Contains(string(body), "called back") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("first save never became visible: %s", body)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if resp, _ := putCallNote(t, c, "call-nudge-note", "called back"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("echo save: %d, want the same 204", resp.StatusCode)
	}
	select {
	case ev := <-events:
		t.Fatalf("echo pushed %q: an unchanged body owes no nudge", ev.Event)
	case <-time.After(200 * time.Millisecond):
	}

	// A real revision publishes exactly one more.
	if resp, _ := putCallNote(t, c, "call-nudge-note", "called back twice"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("revision save: %d", resp.StatusCode)
	}
	select {
	case ev := <-events:
		if ev.Event != "notes" || len(ev.Data) != 0 {
			t.Fatalf("revision nudge = %q (%.40s), want payload-less notes", ev.Event, ev.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no nudge for a real revision within 2s")
	}
}

// deleteCallNote is the explicit DELETE lane's wire (the 409 remedy).
func deleteCallNote(t *testing.T, c *client, callID string) (*http.Response, []byte) {
	t.Helper()
	return c.do(http.MethodDelete, "/api/calls/"+callID+"/note", nil, "")
}

// TestCallNoteDeleteFreesTheSlot pins the delete contract end to end:
// 204 removes the slot outright (a later read is empty AND a re-save is
// a fresh save), a missing slot is the 404 (another extension's call id
// included), and the empty-body PUT's slot is NOT gone — the two arms
// of "clear" must stay distinguishable.
func TestCallNoteDeleteFreesTheSlot(t *testing.T) {
	c := newClient(t)
	c.login("1001", "pw")

	if resp, _ := putCallNote(t, c, "call-del", "to be deleted"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed note: %d", resp.StatusCode)
	}
	if resp, _ := deleteCallNote(t, c, "call-del"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d, want 204", resp.StatusCode)
	}
	_, body := getCallNote(t, c, "call-del")
	if strings.Contains(string(body), "to be deleted") {
		t.Fatalf("the deleted note must read empty: %s", body)
	}
	// The slot is freed: a fresh save works and is NOT a resurrection.
	if resp, _ := putCallNote(t, c, "call-del", "fresh"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("re-save after delete: %d", resp.StatusCode)
	}
	_, body = getCallNote(t, c, "call-del")
	if !strings.Contains(string(body), "fresh") {
		t.Fatalf("fresh save must land: %s", body)
	}

	// A missing slot (and another extension's call id) is the same 404.
	if resp, _ := deleteCallNote(t, c, "call-never-was"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("missing-slot delete: %d, want 404", resp.StatusCode)
	}
	other := newClient(t)
	other.login("2002", "pw")
	if resp, _ := deleteCallNote(t, other, "call-del"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("another owner's delete: %d, want 404", resp.StatusCode)
	}

	// The nil-seam composition answers the honest 503.
	dsrv := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.Notes = nil })
	dc := clientFor(t, dsrv)
	dc.login("1001", "pw")
	if resp, _ := deleteCallNote(t, dc, "call-del"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil-seam delete: %d, want 503", resp.StatusCode)
	}
}

// TestCallNoteDeleteNudgesHistory: the delete fires the same payload-less
// "notes" nudge the save does (other tabs must re-fetch a removed box).
func TestCallNoteDeleteNudgesHistory(t *testing.T) {
	srv := newTestServer(t)
	c := clientFor(t, srv)
	c.login("1001", "pw")

	if resp, _ := putCallNote(t, c, "call-del-nudge", "gone soon"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("seed note: %d", resp.StatusCode)
	}
	owner := domain.MustParseExtension("1001")
	events := srv.hubs.get(owner).Hub().Subscribe()
	t.Cleanup(func() { srv.hubs.get(owner).Hub().Unsubscribe(events) })

	if resp, _ := deleteCallNote(t, c, "call-del-nudge"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	select {
	case ev := <-events:
		if ev.Event != "notes" {
			t.Fatalf("nudge event = %q, want \"notes\"", ev.Event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no notes nudge after the delete — a second tab would keep a stale box")
	}
}
