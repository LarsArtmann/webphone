package server

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/asr"
)

// providerSeen records what the fake provider received on the wire so a
// test can assert the pass-through end to end.
type providerSeen struct {
	Audio       string
	Filename    string
	ContentType string
	Language    string
}

// fakeProvider is an OpenAI-compatible transcription stub. It records the
// last request it saw so a test can assert the bytes and the multipart
// shape forwarded end to end.
func fakeProvider(t *testing.T, status int, text string) (*httptest.Server, *providerSeen) {
	t.Helper()
	seen := &providerSeen{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("provider path: %q", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("provider multipart: %v", err)
		}
		if fhs := r.MultipartForm.File["file"]; len(fhs) == 1 {
			fh := fhs[0]
			seen.Filename = fh.Filename
			seen.ContentType = fh.Header.Get("Content-Type")
			if f, err := fh.Open(); err == nil {
				defer func() { _ = f.Close() }()
				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				seen.Audio = string(buf[:n])
			}
		}
		seen.Language = r.FormValue("language")
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_ = json.MarshalWrite(w, struct { //nolint:erraudit // test stub write
			Text string `json:"text"`
		}{Text: text})
	}))
	t.Cleanup(server.Close)
	return server, seen
}

func asrClientFor(t *testing.T, providerURL string) *asr.Client {
	t.Helper()
	client, err := asr.NewClient(providerURL, "", "")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestTranscribeRequiresSession(t *testing.T) {
	server := newTestServer(t)
	anon := clientFor(t, server)
	anon.token = ""
	resp, _ := anon.do(http.MethodPost, "/api/transcribe", []byte("audio"), "audio/webm")
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Errorf("anonymous transcribe: %d (want 401/403)", resp.StatusCode)
	}
}

func TestTranscribeAbsentSeamIs404(t *testing.T) {
	server := newTestServer(t) // Deps.ASR nil
	c := signIn(t, server)
	resp, _ := c.do(http.MethodPost, "/api/transcribe", []byte("audio"), "audio/webm")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("disabled seam: %d (want 404)", resp.StatusCode)
	}
}

func TestTranscribeRoundTrip(t *testing.T) {
	provider, seen := fakeProvider(t, http.StatusOK, "hello from the provider")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)

	resp, body := c.do(http.MethodPost, "/api/transcribe?filename=clip.webm", []byte("RAWAUDIO"), "audio/webm")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("transcribe: %d %s", resp.StatusCode, body)
	}
	var got struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode %q: %v", body, err)
	}
	if got.Text != "hello from the provider" {
		t.Errorf("text: got %q", got.Text)
	}
	if seen.Audio != "RAWAUDIO" {
		t.Errorf("provider saw %q, want the forwarded bytes", seen.Audio)
	}
}

func TestTranscribeEmptyBodyIs400(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)
	resp, _ := c.do(http.MethodPost, "/api/transcribe", nil, "audio/webm")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty body: %d (want 400)", resp.StatusCode)
	}
}

func TestTranscribeProviderFailureIs502(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusUnauthorized, "")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)
	resp, _ := c.do(http.MethodPost, "/api/transcribe", []byte("audio"), "audio/webm")
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("provider rejected credentials: %d (want 502)", resp.StatusCode)
	}
}

// TestTranscribeOversizedBodyIs400 pins the 25 MiB body cap at the
// endpoint: a runaway client gets 400 and the provider never sees the
// segment (the 16-54 §f #32 debt).
func TestTranscribeOversizedBodyIs400(t *testing.T) {
	provider, seen := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)
	oversized := bytes.Repeat([]byte("a"), (25<<20)+1)
	resp, _ := c.do(http.MethodPost, "/api/transcribe", oversized, "audio/webm")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("oversized body: %d (want 400)", resp.StatusCode)
	}
	if seen.Audio != "" {
		t.Errorf("provider must not see an oversized segment: %d bytes", len(seen.Audio))
	}
}

// TestTranscribeWirePassThrough pins the pass-through contract: the
// filename query, the HONEST Content-Type header and the lang query reach
// the provider multipart verbatim (the 16-54 §f #33 debt).
func TestTranscribeWirePassThrough(t *testing.T) {
	provider, seen := fakeProvider(t, http.StatusOK, "hello")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := signIn(t, server)
	resp, _ := c.do(http.MethodPost, "/api/transcribe?filename=clip.ogg&lang=de", []byte("RAW"), "audio/ogg")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("transcribe: %d", resp.StatusCode)
	}
	if seen.Filename != "clip.ogg" || seen.ContentType != "audio/ogg" || seen.Language != "de" || seen.Audio != "RAW" {
		t.Errorf("wire pass-through drifted: %+v", seen)
	}
}

// TestConfigJSCarriesASRFlag pins the island's gate: the served
// window.PBX_CONFIG must mirror the seam's Enabled() state.
func TestConfigJSCarriesASRFlag(t *testing.T) {
	provider, _ := fakeProvider(t, http.StatusOK, "x")
	server := newTestServerWithPhoneAPI(t, "", func(d *Deps) { d.ASR = asrClientFor(t, provider.URL) })
	c := clientFor(t, server)
	resp, body := c.do(http.MethodGet, "/config.js", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config.js: %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"asr":true`) {
		t.Errorf("config.js must carry asr:true when enabled: %.300s", body)
	}
}
