package server

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/asr"
)

// fakeProvider is an OpenAI-compatible transcription stub. It records the
// last audio it saw so a test can assert the bytes forwarded end to end.
func fakeProvider(t *testing.T, status int, text string) (*httptest.Server, *string) {
	t.Helper()
	var lastAudio string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("provider path: %q", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("provider multipart: %v", err)
		}
		if f, _, err := r.FormFile("file"); err == nil {
			defer f.Close()
			buf := make([]byte, 1024)
			n, _ := f.Read(buf)
			lastAudio = string(buf[:n])
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_ = json.MarshalWrite(w, struct { //nolint:erraudit // test stub write
			Text string `json:"text"`
		}{Text: text})
	}))
	t.Cleanup(server.Close)
	return server, &lastAudio
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
	provider, lastAudio := fakeProvider(t, http.StatusOK, "hello from the provider")
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
	if *lastAudio != "RAWAUDIO" {
		t.Errorf("provider saw %q, want the forwarded bytes", *lastAudio)
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
