// Package asr is the server-side speech-to-text seam: it forwards a chunk
// of audio to an OpenAI-compatible transcription provider and returns the
// text. A client without a configured base URL is "disabled": every call
// returns ErrDisabled and the server never registers the /api/transcribe
// surface. The seam is deliberately provider-agnostic — a self-hosted
// whisper.cpp server, a faster-whisper HTTP wrapper, or any OpenAI-
// compatible API all speak the same /v1/audio/transcriptions shape.
//
// Audio never accumulates here: the caller hands one whole (segmented)
// audio file per request, the provider answers with the text, and nothing
// is stored. Live call transcription is client-driven — the island sends
// short MediaRecorder segments and appends the returned text.
package asr

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
)

// Sentinels mirror the pbx/crm seam: disabled is Infrastructure (nil
// deps), a rejected bearer token is a Rejection (the operator must fix
// the config). The sentinels keep their errors.New identity so callers
// compare by equality; classifications are registered here.
func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrDisabled:     errorfamily.Infrastructure,
		ErrUnauthorized: errorfamily.Rejection,
	})
}

// Provider is the transcription seam every consumer depends on: one
// call, one whole audio segment, one text back. The OpenAI-compatible
// Client (client.go) and the GoogleClient (google.go) both implement
// it; the composition root picks the kind per config, and nothing
// outside this package needs to know which wire is on the other end.
type Provider interface {
	// Enabled reports whether a provider is wired up.
	Enabled() bool
	// Describe names the provider kind and model for the Settings tab's
	// display (raw service vocabulary, like the gateway-mode row).
	Describe() string
	// Transcribe sends one audio file to the provider and returns the text.
	Transcribe(ctx context.Context, req Request) (string, error)
}

// ErrDisabled is returned when no ASR provider is configured.
var ErrDisabled = errors.New("asr not configured")

// ErrUnauthorized is returned when the provider rejected the bearer token
// (HTTP 401/403): the integration is half-configured and must surface.
var ErrUnauthorized = errors.New("asr rejected the bearer token")

// Locality classifies where a provider URL processes caller audio — the
// config-honest trust label the Settings tab and the caption strip
// render (M17). "local" names infrastructure the operator runs
// (loopback, RFC1918/private, .lan/.home.arpa, or a dotless hostname);
// everything else is "remote". An unparseable URL answers "remote"
// (the honest default: assume the outside world).
func LocalityOf(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return "remote"
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || host == "::1" || strings.HasSuffix(host, ".lan") || strings.HasSuffix(host, ".home.arpa") {
		return "local"
	}
	if strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.168.") {
		return "local"
	}
	if strings.HasPrefix(host, "172.") {
		if second, ok := strings.CutPrefix(host, "172."); ok {
			if octet, _, found := strings.Cut(second, "."); found {
				if n, err := strconv.Atoi(octet); err == nil && n >= 16 && n <= 31 {
					return "local"
				}
			}
		}
	}
	if !strings.Contains(host, ".") {
		return "local"
	}
	return "remote"
}

const (
	// requestTimeout bounds one transcription round-trip. Long voicemail
	// clips are minutes, not hours; a provider that cannot answer in this
	// window is down, and the island's UX is a retry, not an endless spin.
	requestTimeout = 90 * time.Second
	// defaultModel is the OpenAI-compatible default when the config names
	// none; whisper.cpp variants accept it and ignore unknown names.
	defaultModel = "whisper-1"
	// defaultFilename is the multipart filename fallback. Providers sniff
	// the bytes (and this package sends the honest Content-Type), so the
	// name is cosmetic.
	defaultFilename = "audio.webm"
	// transcriptionsPath is the OpenAI-compatible transcription endpoint.
	transcriptionsPath = "/v1/audio/transcriptions"
	// inferencePath is the native whisper.cpp server endpoint: the same
	// multipart + {"text"} wire as the OpenAI-compatible lane, posted to
	// /inference (asr.provider = "whispercpp").
	inferencePath = "/inference"
)

// Request is one transcription job: a whole (segmented) audio file plus
// the metadata the provider needs to decode and route it.
type Request struct {
	// Audio is the raw encoded bytes (webm/opus from MediaRecorder, or
	// wav/mp3 for stored voicemail). Never stored.
	Audio []byte
	// Filename names the upload (cosmetic; providers sniff). Empty keeps
	// a default.
	Filename string
	// ContentType is the HONEST media type of Audio. Empty falls back to
	// application/octet-stream.
	ContentType string
	// Language is an optional ISO-639-1 hint; empty lets the provider
	// auto-detect.
	Language string
}
