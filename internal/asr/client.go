package asr

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/httpclient"
)

// Client talks to one OpenAI-compatible transcription provider.
// branching-flow COMPOSITION_mixin (baselined): the client-chokepoint
// idiom shared with crm/pbx by design; a mixin would weld independent
// seams (registry 2026-10-05).
type Client struct {
	base   *url.URL
	token  string
	model  string
	path   string
	client *http.Client
	stats  outcomes
}

// RecentOutcomes exposes the rolling round-trip state — the health
// leg's truth (see outcomes.go).
func (c *Client) RecentOutcomes() OutcomeSnapshot { return c.stats.snapshot() }

// HealthCheck is the optional NON-CRITICAL probe leg (#22): the
// outcomes verdict, never a provider round trip.
func (c *Client) HealthCheck(ctx context.Context) error { return c.stats.HealthCheck(ctx) }

// NewClient builds the client; baseURL like "http://127.0.0.1:8081". An
// empty baseURL means "disabled": every call returns ErrDisabled and the
// server treats the whole seam as off. Config validation rejects a token
// without a URL before this runs.
// ClientConfig is the OpenAI-compatible client's wiring struct: the
// three primitives travel together, so a constructor-order swap cannot
// compile (the 2026-10-25 wiring-struct roadmap item; Google's
// GoogleConfig is the precedent).
type ClientConfig struct {
	BaseURL string
	Token   string
	Model   string
}

func NewClient(cfg ClientConfig) (*Client, error) { //nolint:branching-flow // single composition-root caller; boot validation rejects token-without-URL before this runs (triage 2026-10-11)
	if cfg.BaseURL == "" {
		return &Client{}, nil
	}

	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "asr.url", "parse asr url %q", cfg.BaseURL)
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}

	return &Client{
		base:   parsed,
		token:  cfg.Token,
		model:  model,
		path:   transcriptionsPath,
		client: httpclient.Isolated(requestTimeout),
	}, nil
}

// NewWhisperCppClient is NewClient for the native whisper.cpp server
// wire (#21): the multipart body and the {"text"} response are the same
// shape, the endpoint path is /inference. Unknown form fields (model,
// language) are ignored by the server, so the body stays shared.
func NewWhisperCppClient(cfg ClientConfig) (*Client, error) { //nolint:branching-flow // config primitives validated at load; parse-at-the-door ruling (strong-id/BD003 class)
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	client.path = inferencePath
	return client, nil
}

// Enabled reports whether a provider is wired up.
func (c *Client) Enabled() bool { return c != nil && c.base != nil }

// Describe names the wire kind and the configured model.
func (c *Client) Describe() string {
	kind := "openai-compatible"
	if c.path == inferencePath {
		kind = "whisper.cpp"
	}
	return kind + " · " + c.model
}

// Transcribe sends one audio file to the provider and returns the text.
// Every round trip feeds the outcomes tracker (latency metric + the
// optional health leg) and one debug log line — the operator-greppable
// latency signal (#23).
func (c *Client) Transcribe(ctx context.Context, req Request) (text string, err error) {
	started := time.Now()
	defer func() {
		latency := time.Since(started)
		c.stats.record(latency, err)
		slog.Debug("asr: transcription round-trip",
			"duration_ms", latency.Milliseconds(), "audio_bytes", len(req.Audio),
			"text_chars", len(text), "error", err)
	}()
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

	target := c.base.JoinPath(c.path).String()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "asr.request", "asr: build request")
	}
	httpReq.Header.Set("Content-Type", contentType)
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, release, err := providerRoundTrip(c.client, httpReq)
	if err != nil {
		return "", err
	}
	defer release()

	return decodeTranscript(resp.Body)
}
