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

// The browser's media element asks for byte ranges (Safari refuses to play
// a 200-full-body stream; Chrome seeks through 206s). The upstream answers
// single-range 206s, so the proxy must carry the Range header upstream and
// the range response headers back down.
func TestPhoneAPIProxyForwardsRangeAndPassthroughHeaders(t *testing.T) {
	var (
		mu       sync.Mutex
		sawRange string
	)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/phone-api/voicemail/1001/summary" && r.Method == http.MethodGet {
			// signIn verifies the session credentials through this route.
			_, _ = w.Write([]byte(`{"new":0,"old":0}`))
			return
		}
		if r.URL.Path != "/phone-api/voicemail/1001/messages/abc-123/audio" {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		sawRange = r.Header.Get("Range")
		mu.Unlock()
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", "bytes 0-3/8")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("RIFF"))
	}))
	t.Cleanup(upstream.Close)

	c := signIn(t, newTestServerWithPhoneAPI(t, upstream.URL))
	req, err := http.NewRequest(http.MethodGet, c.base+"/phone-api/voicemail/1001/messages/abc-123/audio", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=0-3")
	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body := readAll(t, resp)
	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("proxied range request: %d %s (want 206)", resp.StatusCode, body)
	}
	mu.Lock()
	if sawRange != "bytes=0-3" {
		mu.Unlock()
		t.Fatalf("upstream Range = %q, want %q", sawRange, "bytes=0-3")
	}
	mu.Unlock()
	if got := resp.Header.Get("Content-Range"); got != "bytes 0-3/8" {
		t.Errorf("Content-Range not passed through: %q", got)
	}
	if got := resp.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges not passed through: %q", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "audio/wav" {
		t.Errorf("Content-Type not passed through: %q", got)
	}
}
