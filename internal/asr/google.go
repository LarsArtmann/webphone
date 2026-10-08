// GoogleClient is the second Provider kind: Google Cloud Speech-to-Text
// V2 synchronous recognition (the "telephony"-tuned hosted model behind
// an EU-resident regional endpoint). Unlike the OpenAI-compatible wire
// this API is JSON-only — the audio rides as base64 inside the request
// body — and it has no per-request minimum billing (1-second rounding),
// which is what makes interactive 4-second island segments affordable.
//
// Chosen defaults, all overridable: region europe-west3 (Frankfurt — the
// whole point of the hosted path is EU residency for caller audio),
// model "telephony" (8 kHz-tuned, matches phone audio end to end), and
// language fallback de-DE (V2 REQUIRES an explicit language code — there
// is no auto-detect — so an untagged segment needs a default; the
// per-request island language always wins when present).
package asr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/larsartmann/go-error-family"
)

const (
	// googleDefaultLocation is the EU residency default (see above).
	googleDefaultLocation = "europe-west3"
	// googleDefaultModel is the telephony model: 8 kHz phone audio is
	// its native domain, and it is the cheapest interactive tier.
	googleDefaultModel = "telephony"
	// googleDefaultLanguage is the language-code floor for requests
	// that carry no hint (V2 has no auto-detect).
	googleDefaultLanguage = "de-DE"
)

// GoogleConfig carries everything the Google provider needs. The zero
// value is the disabled provider (no project); validation upstream
// rejects a half-configured kind before it reaches the constructor.
type GoogleConfig struct {
	// Endpoint overrides the regional API base for tests and
	// key-injecting proxies; empty derives it from Location.
	Endpoint string
	// APIKey is sent as x-goog-api-key when set. Optional so a local
	// proxy can own the credential instead.
	APIKey string
	// Project is the Google Cloud project id — part of the recognizer
	// resource path, so it is required even with an Endpoint override.
	Project string
	// Location is the region (e.g. "europe-west3"); empty keeps the
	// EU default.
	Location string
	// Model names the V2 model variant; empty keeps "telephony".
	Model string
	// Language is the fallback BCP-47 code for requests without a
	// per-request hint; empty keeps "de-DE".
	Language string
}

// googleEndpoint derives the regional v2 API base. Regional hosts keep
// the audio inside the region (the "global" host would not).
func googleEndpoint(location string) string {
	return "https://speech." + location + ".rep.googleapis.com"
}

// GoogleClient talks to Google Cloud Speech-to-Text V2.
type GoogleClient struct {
	base       *url.URL
	apiKey     string
	recognizer string
	model      string
	language   string
	client     *http.Client
}

// NewGoogleClient builds the client. A config without a project is the
// disabled provider: every call returns ErrDisabled and the server
// treats the whole seam as off (mirrors NewClient's empty-URL rule).
func NewGoogleClient(cfg GoogleConfig) (*GoogleClient, error) {
	if cfg.Project == "" {
		return &GoogleClient{}, nil
	}

	location := cfg.Location
	if location == "" {
		location = googleDefaultLocation
	}
	model := cfg.Model
	if model == "" {
		model = googleDefaultModel
	}
	language := cfg.Language
	if language == "" {
		language = googleDefaultLanguage
	}

	base := cfg.Endpoint
	if base == "" {
		base = googleEndpoint(location)
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "asr.url", "parse asr url %q", base)
	}

	return &GoogleClient{
		base:       parsed,
		apiKey:     cfg.APIKey,
		recognizer: "/v2/projects/" + cfg.Project + "/locations/" + location + "/recognizers/_:recognize",
		model:      model,
		language:   language,
		client:     &http.Client{Timeout: requestTimeout},
	}, nil
}

// Enabled reports whether a provider is wired up.
func (c *GoogleClient) Enabled() bool { return c != nil && c.base != nil }

// Describe names the wire kind and the configured model.
func (c *GoogleClient) Describe() string { return "google · " + c.model }

// googleRecognizeRequest is the v2 recognize wire shape. The "_" in the
// recognizer path is the ad-hoc implicit recognizer: no
// recognizers.create round-trip, the inline config is the whole truth.
// autoDecodingConfig lets Google sniff the container (webm/opus, wav,
// mp3 — everything the island and stored voicemail produce), so
// Request.Filename and Request.ContentType are deliberately unused on
// this wire.
type googleRecognizeRequest struct {
	Config  googleRecognitionConfig `json:"config"`
	Content string                  `json:"content"`
}

type googleRecognitionConfig struct {
	Model              string   `json:"model,omitempty"`
	LanguageCodes      []string `json:"languageCodes"`
	AutoDecodingConfig struct{} `json:"autoDecodingConfig"`
}

// googleRecognizeResponse keeps only the transcript-bearing fields;
// confidence is intentionally not plumbed (0.0 is Google's "unset"
// sentinel, which would render as "low confidence" — see the M22
// decision note in the SUPERB plan).
type googleRecognizeResponse struct {
	Results []struct {
		Alternatives []struct {
			Transcript string `json:"transcript"`
		} `json:"alternatives"`
	} `json:"results"`
}

// Transcribe sends one audio file to the provider and returns the text.
func (c *GoogleClient) Transcribe(ctx context.Context, req Request) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	if len(req.Audio) == 0 {
		return "", errorfamily.NewRejection("asr.empty", "no audio to transcribe")
	}

	language := req.Language
	if language == "" {
		language = c.language
	}
	payload := googleRecognizeRequest{
		Config: googleRecognitionConfig{
			Model:              c.model,
			LanguageCodes:      []string{language},
			AutoDecodingConfig: struct{}{},
		},
		Content: base64.StdEncoding.EncodeToString(req.Audio),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "asr.encode", "asr: encode google request")
	}

	target := c.base.JoinPath(c.recognizer).String()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "asr.request", "asr: build request")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("x-goog-api-key", c.apiKey)
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
	case resp.StatusCode == http.StatusTooManyRequests:
		// A rate-limit answer is load, not misconfiguration: the
		// caller may retry after a beat.
		return "", errorfamily.Newf(errorfamily.Transient, "asr.http", "asr: provider answered 429")
	case resp.StatusCode != http.StatusOK:
		family := errorfamily.Transient
		if resp.StatusCode < 500 {
			family = errorfamily.Rejection
		}
		return "", errorfamily.Newf(family, "asr.http", "asr: provider answered %d", resp.StatusCode)
	}

	return c.decodeTranscript(resp.Body)
}

// decodeTranscript joins the per-result top alternatives into one line
// — the same contract the OpenAI wire offers with its single text
// field.
func (c *GoogleClient) decodeTranscript(r io.Reader) (string, error) {
	raw, err := io.ReadAll(io.LimitReader(r, 1<<20))
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.read", "asr: read response")
	}
	var parsed googleRecognizeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errorfamily.WrapTransientf(err, "asr.decode", "asr: decode response")
	}
	parts := make([]string, 0, len(parsed.Results))
	for _, result := range parsed.Results {
		if len(result.Alternatives) > 0 && strings.TrimSpace(result.Alternatives[0].Transcript) != "" {
			parts = append(parts, strings.TrimSpace(result.Alternatives[0].Transcript))
		}
	}
	return strings.Join(parts, " "), nil
}
