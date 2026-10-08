package server

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
)

// postTranscript is the island's fire-and-forget segment report.
func postTranscript(t *testing.T, c *client, callID, direction, remote, text string) (*http.Response, []byte) {
	t.Helper()
	payload, err := json.Marshal(struct {
		CallID    string `json:"callId"`
		Direction string `json:"direction"`
		Remote    string `json:"remote"`
		StartedAt int64  `json:"startedAt"`
		Text      string `json:"text"`
	}{CallID: callID, Direction: direction, Remote: remote, StartedAt: 1760000000000, Text: text})
	if err != nil {
		t.Fatal(err)
	}
	return c.do(http.MethodPost, "/api/transcripts", payload, "application/json")
}

func TestSaveTranscriptRequiresSession(t *testing.T) {
	server := newTestServer(t)
	anon := clientFor(t, server)
	anon.token = ""
	payload := []byte(`{"callId":"c","direction":"out","remote":"+49","startedAt":1,"text":"hi"}`)
	resp, _ := anon.do(http.MethodPost, "/api/transcripts", payload, "application/json")
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Errorf("anonymous transcript save: %d (want 401/403)", resp.StatusCode)
	}
}

func TestSaveTranscriptAbsentSeamIs404(t *testing.T) {
	server := newTestServer(t) // Deps.ASR nil
	c := signIn(t, server)
	resp, _ := postTranscript(t, c, "c1", "out", "+4930", "hi")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("disabled seam: %d (want 404)", resp.StatusCode)
	}
}

func TestSaveTranscriptRoundTripAndHistoryRendering(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	for _, text := range []string{"guten tag", "wie gehts"} {
		resp, body := postTranscript(t, c, "call-1", "in", "+4989123456", text)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("save: %d %s", resp.StatusCode, body)
		}
	}
	resp, _ := postTranscript(t, c, "call-2", "out", "+4930111", "second call")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save 2: %d", resp.StatusCode)
	}

	// The History tab re-renders the extension's own records: newest
	// call first, spoken order inside.
	resp, body := c.do(http.MethodGet, "/partials/history", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history: %d", resp.StatusCode)
	}
	page := string(body)
	for _, want := range []string{"Call transcripts", "+4930111", "+4989123456", "guten tag wie gehts"} {
		if !strings.Contains(page, want) {
			t.Errorf("history transcript section missing %q; page:\n%s", want, page)
		}
	}
	if strings.Index(page, "+4930111") > strings.Index(page, "+4989123456") {
		t.Errorf("newest call must lead the section")
	}

	// Owner scoping: another extension's history never shows the rows.
	other := clientFor(t, server)
	other.login("1002", "pw")
	resp, body = other.do(http.MethodGet, "/partials/history", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("other history: %d", resp.StatusCode)
	}
	if strings.Contains(string(body), "guten tag") {
		t.Errorf("transcript leaked across extensions")
	}
}

func TestSaveTranscriptValidation(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	cases := []struct {
		name       string
		payload    string
		wantStatus int
	}{
		{"bad direction", `{"callId":"c","direction":"sideways","remote":"+49","startedAt":1,"text":"hi"}`, http.StatusBadRequest},
		{"empty text", `{"callId":"c","direction":"out","remote":"+49","startedAt":1,"text":"  "}`, http.StatusBadRequest},
		{"no callId", `{"direction":"out","remote":"+49","startedAt":1,"text":"hi"}`, http.StatusBadRequest},
		{"no startedAt", `{"callId":"c","direction":"out","remote":"+49","text":"hi"}`, http.StatusBadRequest},
		{"oversized text", `{"callId":"c","direction":"out","remote":"+49","startedAt":1,"text":"` + strings.Repeat("x", 4001) + `"}`, http.StatusRequestEntityTooLarge},
		{"not json", `not-json`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		resp, _ := c.do(http.MethodPost, "/api/transcripts", []byte(tc.payload), "application/json")
		if resp.StatusCode != tc.wantStatus {
			t.Errorf("%s: %d (want %d)", tc.name, resp.StatusCode, tc.wantStatus)
		}
	}
}

func TestSaveTranscriptNilStoreIs503(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) {
		d.ASR = asrClientFor(t, provider.URL)
		d.Transcripts = nil
	})
	c := signIn(t, server)
	resp, body := postTranscript(t, c, "c1", "out", "+4930", "hi")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("nil store: %d %s (want 503)", resp.StatusCode, body)
	}
}

func TestHistoryTranscriptSearchMatchesOwnerScopedAndHinted(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	for _, text := range []string{"guten tag", "wie gehts"} {
		if resp, _ := postTranscript(t, c, "call-1", "in", "+4989123456", text); resp.StatusCode != http.StatusNoContent {
			t.Fatalf("save call-1: %d", resp.StatusCode)
		}
	}
	if resp, _ := postTranscript(t, c, "call-2", "out", "+4930111", "second call"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save call-2: %d", resp.StatusCode)
	}
	// Another extension owns the same words: search must never leak them.
	other := clientFor(t, server)
	other.login("1002", "pw")
	if resp, _ := postTranscript(t, other, "call-x", "in", "+49130", "guten tag from the neighbor"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save neighbor: %d", resp.StatusCode)
	}

	// Unfiltered: both of the owner's calls render, and no hint.
	_, body := c.do(http.MethodGet, "/partials/history", nil, "")
	page := string(body)
	for _, want := range []string{"guten tag", "second call"} {
		if !strings.Contains(page, want) {
			t.Errorf("unfiltered history missing %q:\n%s", want, page)
		}
	}
	if strings.Contains(page, "Transcript matches for:") {
		t.Errorf("no search hint without an active query")
	}

	// q=geht: the matching line renders with the hint; the non-matching
	// call's transcript hides under the active filter.
	_, body = c.do(http.MethodGet, "/partials/history?q=geht", nil, "")
	page = string(body)
	if !strings.Contains(page, "wie gehts") {
		t.Errorf("search must render the matching line:\n%s", page)
	}
	if !strings.Contains(page, "Transcript matches for:") || !strings.Contains(page, "<code>geht</code>") {
		t.Errorf("search hint missing:\n%s", page)
	}
	if strings.Contains(page, "second call") {
		t.Errorf("a non-matching transcript must hide under an active query:\n%s", page)
	}
	if strings.Contains(page, "from the neighbor") {
		t.Errorf("another extension's transcript leaked into the search")
	}

	// The neighbor finds their own words — owner scoping cuts both ways.
	_, body = other.do(http.MethodGet, "/partials/history?q=neighbor", nil, "")
	if !strings.Contains(string(body), "from the neighbor") {
		t.Errorf("the neighbor's own search must find their transcript")
	}

	// No hit at all: the transcript section disappears entirely.
	_, body = c.do(http.MethodGet, "/partials/history?q=zzz-nothing", nil, "")
	page = string(body)
	if strings.Contains(page, "wp-transcripts") || strings.Contains(page, "guten tag") {
		t.Errorf("a no-hit query must hide the transcript section:\n%s", page)
	}

	// LIKE metacharacters are literals: "100%" matches "100%", not
	// "100" + anything (q pre-encoded as %25).
	if resp, _ := postTranscript(t, c, "call-pct", "out", "+4930", "fifty 100% sure"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save pct: %d", resp.StatusCode)
	}
	_, body = c.do(http.MethodGet, "/partials/history?q=100%25", nil, "")
	page = string(body)
	if !strings.Contains(page, "fifty 100% sure") {
		t.Errorf("an escaped-percent query must match its literal:\n%s", page)
	}
	if strings.Contains(page, "guten tag") || strings.Contains(page, "second call") {
		t.Errorf("non-matching transcripts leaked past the literal-percent query:\n%s", page)
	}
}

func TestSettingsNamesTheASRProviderWhenSeamIsOn(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)
	_, body := c.do(http.MethodGet, "/partials/settings", nil, "")
	if page := string(body); !strings.Contains(page, "openai-compatible · whisper-1") {
		t.Errorf("settings must name the provider kind and model:\n%s", page)
	}

	// Seam off: the row stays honest ("not configured"), no detail.
	plain := newTestServer(t)
	other := signIn(t, plain)
	_, body = other.do(http.MethodGet, "/partials/settings", nil, "")
	if page := string(body); strings.Contains(page, "openai-compatible") {
		t.Errorf("no provider detail without the seam:\n%s", page)
	}
}

func TestDeleteTranscriptAbsentSeamIs404(t *testing.T) {
	server := newTestServer(t) // Deps.ASR nil
	c := signIn(t, server)
	resp, _ := c.do(http.MethodDelete, "/api/transcripts?call=c1", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("disabled seam: %d (want 404)", resp.StatusCode)
	}
}

func TestDeleteTranscriptRoundTripOwnerScopingAndIdempotency(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	for _, text := range []string{"guten tag", "wie gehts"} {
		if resp, _ := postTranscript(t, c, "call-1", "in", "+4989123456", text); resp.StatusCode != http.StatusNoContent {
			t.Fatalf("save call-1: %d", resp.StatusCode)
		}
	}
	if resp, _ := postTranscript(t, c, "call-2", "out", "+4930111", "second call"); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("save call-2: %d", resp.StatusCode)
	}

	// The History partial renders the delete affordance per call.
	resp, body := c.do(http.MethodGet, "/partials/history", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history: %d", resp.StatusCode)
	}
	page := string(body)
	if !strings.Contains(page, `data-delete-transcript="call-1"`) || !strings.Contains(page, `data-delete-transcript="call-2"`) {
		t.Errorf("history transcript rows carry no delete affordance:\n%s", page)
	}
	if !strings.Contains(page, `data-copy-transcript`) {
		t.Errorf("history transcript rows carry no copy affordance:\n%s", page)
	}

	// Owner scoping: another extension's DELETE of the same call id is a
	// 204 NO-OP — it must not erase the owner's rows.
	other := clientFor(t, server)
	other.login("1002", "pw")
	if resp, _ := other.do(http.MethodDelete, "/api/transcripts?call=call-1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("other's delete: %d (want 204)", resp.StatusCode)
	}
	resp, body = c.do(http.MethodGet, "/partials/history", nil, "")
	if !strings.Contains(string(body), "guten tag") {
		t.Errorf("another extension's delete erased the owner's transcript")
	}

	// The owner's own delete erases exactly one call; idempotent on repeat.
	if resp, _ := c.do(http.MethodDelete, "/api/transcripts?call=call-1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d (want 204)", resp.StatusCode)
	}
	if resp, _ := c.do(http.MethodDelete, "/api/transcripts?call=call-1", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("repeat delete: %d (want 204)", resp.StatusCode)
	}
	resp, body = c.do(http.MethodGet, "/partials/history", nil, "")
	page = string(body)
	if strings.Contains(page, "guten tag") || strings.Contains(page, "+4989123456") {
		t.Errorf("call-1 transcript survived its delete:\n%s", page)
	}
	if !strings.Contains(page, "second call") {
		t.Errorf("call-2 transcript was collateral damage:\n%s", page)
	}

	// Validation: the call parameter is required.
	if resp, _ := c.do(http.MethodDelete, "/api/transcripts", nil, ""); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing call param: %d (want 400)", resp.StatusCode)
	}
}
