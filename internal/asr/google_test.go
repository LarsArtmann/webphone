package asr_test

import (
	"context"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/asr"
)

func TestGoogleDisabledWithoutProject(t *testing.T) {
	client, err := asr.NewGoogleClient(asr.GoogleConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if client.Enabled() {
		t.Fatal("no project must be disabled")
	}
	if _, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")}); !errors.Is(err, asr.ErrDisabled) {
		t.Fatalf("got %v, want ErrDisabled", err)
	}
}

// googleCapture records one recognize round-trip for shape assertions.
type googleCapture struct {
	method string
	path   string
	apiKey string
	auth   string
	body   struct {
		Config struct {
			Model         string   `json:"model"`
			LanguageCodes []string `json:"languageCodes"`
			AutoDecoding  struct{} `json:"autoDecodingConfig"`
		} `json:"config"`
		Content string `json:"content"`
	}
}

func googleServer(t *testing.T, capture *googleCapture, status int, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.method = r.Method
		capture.path = r.URL.Path
		capture.apiKey = r.Header.Get("x-goog-api-key")
		capture.auth = r.Header.Get("Authorization")
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if err := json.Unmarshal(raw, &capture.body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(response))
	}))
}

func TestGoogleRequestShape(t *testing.T) {
	var capture googleCapture
	server := googleServer(t, &capture, http.StatusOK, `{"results":[{"alternatives":[{"transcript":"hallo"}]},{"alternatives":[{"transcript":"  welt  "}]}]}`)
	defer server.Close()

	client, err := asr.NewGoogleClient(asr.GoogleConfig{
		Endpoint: server.URL,
		APIKey:   "gkey",
		Project:  "my-proj",
		Location: "europe-west3",
		Model:    "telephony",
		Language: "fr-FR",
	})
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("AUDIOBYTES")})
	if err != nil {
		t.Fatal(err)
	}
	if text != "hallo welt" {
		t.Errorf("text: got %q, want joined+trimmed results", text)
	}
	if capture.method != http.MethodPost {
		t.Errorf("method: got %q", capture.method)
	}
	wantPath := "/v2/projects/my-proj/locations/europe-west3/recognizers/_:recognize"
	if capture.path != wantPath {
		t.Errorf("path: got %q, want %q", capture.path, wantPath)
	}
	if capture.apiKey != "gkey" {
		t.Errorf("api key header: got %q", capture.apiKey)
	}
	if capture.auth != "" {
		t.Errorf("no bearer auth expected on the google wire, got %q", capture.auth)
	}
	if capture.body.Config.Model != "telephony" {
		t.Errorf("model: got %q", capture.body.Config.Model)
	}
	if len(capture.body.Config.LanguageCodes) != 1 || capture.body.Config.LanguageCodes[0] != "fr-FR" {
		t.Errorf("language fallback: got %v", capture.body.Config.LanguageCodes)
	}
	if want := base64.StdEncoding.EncodeToString([]byte("AUDIOBYTES")); capture.body.Content != want {
		t.Errorf("content: got %q, want base64 audio", capture.body.Content)
	}
}

func TestGoogleLanguageFallbackChain(t *testing.T) {
	cases := []struct {
		name     string
		cfgLang  string
		reqLang  string
		expected string
	}{
		{"request wins", "fr-FR", "en", "en"},
		{"config fallback", "fr-FR", "", "fr-FR"},
		{"default floor", "", "", "de-DE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capture googleCapture
			server := googleServer(t, &capture, http.StatusOK, `{"results":[{"alternatives":[{"transcript":"x"}]}]}`)
			defer server.Close()

			client, err := asr.NewGoogleClient(asr.GoogleConfig{Endpoint: server.URL, Project: "p", Language: tc.cfgLang})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x"), Language: tc.reqLang}); err != nil {
				t.Fatal(err)
			}
			if got := capture.body.Config.LanguageCodes; len(got) != 1 || got[0] != tc.expected {
				t.Errorf("languageCodes: got %v, want [%s]", got, tc.expected)
			}
		})
	}
}

func TestGoogleErrorFamilies(t *testing.T) {
	cases := []struct {
		name         string
		status       int
		family       errorfamily.Family
		unauthorized bool
	}{
		{"unauthorized", http.StatusUnauthorized, errorfamily.Rejection, true},
		{"forbidden", http.StatusForbidden, errorfamily.Rejection, true},
		{"rate limited", http.StatusTooManyRequests, errorfamily.Transient, false},
		{"bad request", http.StatusBadRequest, errorfamily.Rejection, false},
		{"server error", http.StatusInternalServerError, errorfamily.Transient, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capture googleCapture
			server := googleServer(t, &capture, tc.status, "")
			defer server.Close()

			client, err := asr.NewGoogleClient(asr.GoogleConfig{Endpoint: server.URL, Project: "p"})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")})
			if err == nil {
				t.Fatal("want error")
			}
			if tc.unauthorized {
				if !errors.Is(err, asr.ErrUnauthorized) {
					t.Fatalf("got %v, want ErrUnauthorized", err)
				}
				return
			}
			if errors.Is(err, asr.ErrUnauthorized) {
				t.Fatalf("must not be ErrUnauthorized: %v", err)
			}
			errorfamilytest.AssertFamily(t, err, tc.family)
			errorfamilytest.AssertCode(t, err, "asr.http")
		})
	}
}

func TestGoogleEmptyAudioIsRejection(t *testing.T) {
	var capture googleCapture
	server := googleServer(t, &capture, http.StatusOK, `{}`)
	defer server.Close()

	client, err := asr.NewGoogleClient(asr.GoogleConfig{Endpoint: server.URL, Project: "p"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Transcribe(context.Background(), asr.Request{})
	if err == nil {
		t.Fatal("want error")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "asr.empty")
	if !strings.Contains(err.Error(), "asr") {
		t.Errorf("error should name the seam: %v", err)
	}
}

// TestGoogleClientSatisfiesSeam pins the composition-root contract: both
// provider kinds are interchangeable through the seam interface.
func TestGoogleClientSatisfiesSeam(t *testing.T) {
	var _ asr.Provider = &asr.Client{}
	var _ asr.Provider = &asr.GoogleClient{}
}
