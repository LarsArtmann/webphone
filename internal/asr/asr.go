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
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
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
// Client (this file) and the GoogleClient (google.go) both implement
// it; the composition root picks the kind per config, and nothing
// outside this package needs to know which wire is on the other end.
type Provider interface {
	// Enabled reports whether a provider is wired up.
	Enabled() bool
	// Transcribe sends one audio file to the provider and returns the text.
	Transcribe(ctx context.Context, req Request) (string, error)
}

// ErrDisabled is returned when no ASR provider is configured.
var ErrDisabled = errors.New("asr not configured") //nolint:erraudit // sentinel: identity, not an error family (plan guardrail #4); classified via init registration

// ErrUnauthorized is returned when the provider rejected the bearer token
// (HTTP 401/403): the integration is half-configured and must surface.
var ErrUnauthorized = errors.New("asr rejected the bearer token") //nolint:erraudit // sentinel: identity, not an error family (plan guardrail #4); classified via init registration

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
)

// Client talks to one OpenAI-compatible transcription provider.
type Client struct {
	base   *url.URL
	token  string
	model  string
	client *http.Client
}

// NewClient builds the client; baseURL like "http://127.0.0.1:8081". An
// empty baseURL means "disabled": every call returns ErrDisabled and the
// server treats the whole seam as off. Config validation rejects a token
// without a URL before this runs.
func NewClient(baseURL, token, model string) (*Client, error) {
	if baseURL == "" {
		return &Client{}, nil
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "asr.url", "parse asr url %q", baseURL)
	}
	if model == "" {
		model = defaultModel
	}

	return &Client{
		base:   parsed,
		token:  token,
		model:  model,
		client: &http.Client{Timeout: requestTimeout},
	}, nil
}

// Enabled reports whether a provider is wired up.
func (c *Client) Enabled() bool { return c != nil && c.base != nil }

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

// Transcribe sends one audio file to the provider and returns the text.
func (c *Client) Transcribe(ctx context.Context, req Request) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	if len(req.Audio) == 0 {
		return "", errorfamily.NewRejection("asr.empty", "no audio to transcribe")
	}

	body, contentType, err := buildMultipart(c.model, req)
	if err != nil {
		return "", err
	}

	target := c.base.JoinPath(transcriptionsPath).String()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "asr.request", "asr: build request")
	}
	httpReq.Header.Set("Content-Type", contentType)
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.transport", "asr: request")
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", ErrUnauthorized
	case resp.StatusCode != http.StatusOK:
		family := errorfamily.Transient
		if resp.StatusCode < 500 {
			family = errorfamily.Rejection
		}
		return "", errorfamily.Newf(family, "asr.http", "asr: provider answered %d", resp.StatusCode)
	}

	return decodeTranscript(resp.Body)
}

// transcriptResponse is the OpenAI-compatible /v1/audio/transcriptions
// wire shape; providers that return a bare object or extra fields are
// tolerated (unknown fields ignored by the decoder).
type transcriptResponse struct {
	Text string `json:"text"`
}

func decodeTranscript(r io.Reader) (string, error) {
	raw, err := io.ReadAll(io.LimitReader(r, 1<<20))
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.read", "asr: read response")
	}
	var parsed transcriptResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.decode", "asr: decode response")
	}
	return strings.TrimSpace(parsed.Text), nil
}

// buildMultipart assembles the multipart body: the file part carries the
// HONEST Content-Type (a provider that sniffs trusts the header less than
// the bytes, but a truthful header never hurts), plus the model and the
// optional language hint as fields.
func buildMultipart(model string, req Request) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := textproto.MIMEHeader{}
	filename := req.Filename
	if filename == "" {
		filename = defaultFilename
	}
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: create file part")
	}
	if _, err := part.Write(req.Audio); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write audio part")
	}
	if err := writer.WriteField("model", model); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write model field")
	}
	if req.Language != "" {
		if err := writer.WriteField("language", req.Language); err != nil {
			return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: write language field")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: close multipart body")
	}

	return body, writer.FormDataContentType(), nil
}
