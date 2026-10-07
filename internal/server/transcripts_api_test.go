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
