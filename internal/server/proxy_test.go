package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestPhoneAPIProxyInjectsCredentialsAndNudgesVoicemail(t *testing.T) {
	var (
		mu      sync.Mutex
		sawAuth string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.URL.Path == "/phone-api/voicemail/1001/summary" && r.Method == http.MethodGet:
			sawAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"new":1,"old":0}`))
		case r.URL.Path == "/phone-api/voicemail/1001/messages" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"messages":[{"uuid":"abc-123","cid_number":"+491700000000","cid_name":"Fax Machine","seconds":12,"created":1750000000,"read":false,"audio_url":"/rec/abc.wav"}]}`))
		case r.URL.Path == "/phone-api/history" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"entries":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(upstream.Close)

	server := newTestServerWithPhoneAPI(t, upstream.URL)
	events := subscribeEvents(t, server)
	c := signIn(t, server)

	// Proxied voicemail read: same JSON, session credentials injected.
	resp, body := c.do(http.MethodGet, "/phone-api/voicemail/1001/summary", nil, "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"new":1`) {
		t.Fatalf("proxied summary: %d %s", resp.StatusCode, body)
	}
	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "json") {
		t.Errorf("content-type not passed through: %q", got)
	}
	mu.Lock()
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("1001:pw"))
	if sawAuth != wantAuth {
		mu.Unlock()
		t.Fatalf("upstream auth = %q, want %q", sawAuth, wantAuth)
	}
	mu.Unlock()

	// The island's voicemail poll nudges the voicemail tab.
	nudge := expectEvent(t, events, "voicemail")
	if nudge.Data != "" {
		t.Fatalf("voicemail nudge must be payload-less, got %.80q", nudge.Data)
	}

	// Non-voicemail proxy traffic does not nudge.
	resp, _ = c.do(http.MethodGet, "/phone-api/history?limit=5", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("proxied history: %d", resp.StatusCode)
	}
	assertNoEvent(t, events, "voicemail")
	assertNoEvent(t, events, "voicemail")

	// Unknown upstream paths surface as 404, not 500.
	resp, _ = c.do(http.MethodGet, "/phone-api/nope", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("proxied unknown path: %d (want 404)", resp.StatusCode)
	}
}

func TestPhoneAPIProxyDisabledReturns503(t *testing.T) {
	c := signIn(t, newTestServer(t))
	resp, body := c.do(http.MethodGet, "/phone-api/history", nil, "")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("proxy without phone api: %d %s (want 503)", resp.StatusCode, body)
	}
}
