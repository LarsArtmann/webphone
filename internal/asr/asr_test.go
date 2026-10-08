package asr_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/asr"
)

func TestDisabledClient(t *testing.T) {
	client, err := asr.NewClient("", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if client.Enabled() {
		t.Fatal("empty base URL must be disabled")
	}
	if _, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")}); !errors.Is(err, asr.ErrDisabled) {
		t.Fatalf("got %v, want ErrDisabled", err)
	}
}

func TestEmptyAudioRejected(t *testing.T) {
	client, err := asr.NewClient("http://127.0.0.1:1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Transcribe(context.Background(), asr.Request{}); err == nil {
		t.Fatal("empty audio: want error")
	}
}

func TestTranscribeHappyPath(t *testing.T) {
	var gotAuth, gotModel, gotFilename, gotContentType, gotLanguage string
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
		}
		gotModel = r.FormValue("model")
		gotLanguage = r.FormValue("language")
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("form file: %v", err)
		} else {
			defer func() { _ = file.Close() }()
			gotFilename = header.Filename
			gotContentType = header.Header.Get("Content-Type")
			data, _ := io.ReadAll(file)
			gotBody = string(data)
		}
		_, _ = w.Write([]byte(`{"text": "  hello world  "}`))
	}))
	defer server.Close()

	client, err := asr.NewClient(server.URL, "secret", "large-v3")
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Transcribe(context.Background(), asr.Request{
		Audio:       []byte("AUDIOBYTES"),
		Filename:    "segment.webm",
		ContentType: "audio/webm",
		Language:    "de",
	})
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello world" {
		t.Errorf("text: got %q, want trimmed %q", text, "hello world")
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("auth: got %q", gotAuth)
	}
	if gotModel != "large-v3" {
		t.Errorf("model: got %q", gotModel)
	}
	if gotFilename != "segment.webm" {
		t.Errorf("filename: got %q", gotFilename)
	}
	if gotContentType != "audio/webm" {
		t.Errorf("content type: got %q", gotContentType)
	}
	if gotLanguage != "de" {
		t.Errorf("language: got %q", gotLanguage)
	}
	if gotBody != "AUDIOBYTES" {
		t.Errorf("audio body: got %q", gotBody)
	}
}

func TestDescribeNamesKindAndModel(t *testing.T) {
	client, _ := asr.NewClient("http://127.0.0.1:8081", "", "Systran/faster-whisper-large-v3-turbo")
	if got := client.Describe(); got != "openai-compatible · Systran/faster-whisper-large-v3-turbo" {
		t.Errorf("Describe: %q", got)
	}
	def, _ := asr.NewClient("http://127.0.0.1:8081", "", "")
	if got := def.Describe(); got != "openai-compatible · whisper-1" {
		t.Errorf("Describe with the default model: %q", got)
	}
}

func TestDefaultModelAndFilename(t *testing.T) {
	var gotModel, gotFilename string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(1 << 20)
		gotModel = r.FormValue("model")
		_, header, err := r.FormFile("file")
		if err == nil {
			gotFilename = header.Filename
		}
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	defer server.Close()

	client, _ := asr.NewClient(server.URL, "", "")
	if _, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if gotModel != "whisper-1" {
		t.Errorf("default model: got %q", gotModel)
	}
	if gotFilename != "audio.webm" {
		t.Errorf("default filename: got %q", gotFilename)
	}
}

func TestUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, _ := asr.NewClient(server.URL, "bad", "")
	if _, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")}); !errors.Is(err, asr.ErrUnauthorized) {
		t.Fatalf("got %v, want ErrUnauthorized", err)
	}
}

func TestProviderErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
	}{
		{"bad request", http.StatusBadRequest},
		{"server error", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			client, _ := asr.NewClient(server.URL, "", "")
			_, err := client.Transcribe(context.Background(), asr.Request{Audio: []byte("x")})
			if err == nil {
				t.Fatal("want error")
			}
			if errors.Is(err, asr.ErrUnauthorized) {
				t.Fatalf("transport error must not be ErrUnauthorized: %v", err)
			}
			if !strings.Contains(err.Error(), "asr") {
				t.Errorf("error should name the seam: %v", err)
			}
		})
	}
}
