package server

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fakeMailbox stands in for the PBX's voicemail API: summary, message
// rows, and the DELETE the voicemail-delete lane proxies. Every request
// is recorded so the chain can prove the proxy actually reached the PBX.
type fakeMailbox struct {
	mu       sync.Mutex
	deleted  []string
	requests []string
}

func (f *fakeMailbox) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /phone-api/voicemail/{ext}/summary", func(w http.ResponseWriter, r *http.Request) {
		f.record("GET summary")
		writeJSON(w, map[string]int{"new": 1, "old": 1})
	})
	mux.HandleFunc("GET /phone-api/voicemail/{ext}/messages", func(w http.ResponseWriter, r *http.Request) {
		f.record("GET messages")
		writeJSON(w, map[string]any{"messages": []map[string]any{
			{
				"uuid": "vm-chain-1", "cid_number": "+441632960401", "cid_name": "",
				"seconds": 12, "created": 1795000000, "read": false,
				"audio_url": "/voicemail/audio/vm-chain-1.wav",
			},
			{
				"uuid": "vm-chain-2", "cid_number": "+441632960402", "cid_name": "",
				"seconds": 5, "created": 1795000100, "read": true,
				"audio_url": "/voicemail/audio/vm-chain-2.wav",
			},
		}})
	})
	mux.HandleFunc("DELETE /phone-api/voicemail/{ext}/messages/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.deleted = append(f.deleted, r.PathValue("uuid"))
		f.mu.Unlock()
		f.record("DELETE " + r.PathValue("uuid"))
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (f *fakeMailbox) record(what string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, what)
}

func (f *fakeMailbox) saw(needle string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.requests {
		if strings.Contains(r, needle) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.MarshalWrite(w, payload) //nolint:erraudit // test fake: nothing to act on
}

// The voicemail chain over the REAL proxy: the panel reads the PBX's
// rows, the delete lane proxies to the PBX and lands as the payload-less
// "voicemail" SSE nudge, and the refreshed panel re-reads the mailbox.
func TestVoicemailChainOverProxy(t *testing.T) {
	mailbox := &fakeMailbox{}
	pbx := httptest.NewServer(mailbox.handler())
	t.Cleanup(pbx.Close)

	server := newTestServerWithPhoneAPI(t, pbx.URL)
	c := signIn(t, server)
	events := subscribeEvents(t, server)

	_, body := c.do(http.MethodGet, "/partials/voicemail", nil, "")
	if !strings.Contains(string(body), "vm-chain-1") || !strings.Contains(string(body), "+441632960402") {
		t.Fatalf("voicemail panel missing the PBX rows: %.200s", body)
	}
	if !mailbox.saw("GET messages") {
		t.Fatal("panel did not reach the PBX")
	}

	resp, respBody := c.do(http.MethodPost, "/voicemail/delete?uuid=vm-chain-1", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("voicemail delete: %d %s", resp.StatusCode, respBody)
	}
	if !mailbox.saw("DELETE vm-chain-1") {
		t.Fatal("delete did not proxy to the PBX")
	}
	if ev := expectEvent(t, events, "voicemail"); ev.Data != "" {
		t.Fatalf("voicemail nudge must be payload-less, got %q", ev.Data)
	}
}
