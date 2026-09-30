// Package pbx is the server-side client for the per-extension phone API
// (voicemail, CDR history) that the telephony stack serves. It rides the
// extension's own SIP credentials via Basic auth — the same contract the
// browser island uses through the /phone-api proxy.
package pbx

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
)

// Client talks to one phone API base URL.
type Client struct {
	base   *url.URL
	client *http.Client
}

// NewClient builds the client; baseURL like "https://pbx.example.com" (the
// API is mounted at /phone-api there). An empty baseURL means "disabled":
// every call returns ErrDisabled.
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return &Client{}, nil
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "pbx.url", "parse phone api url %q", baseURL)
	}
	return &Client{
		base:   parsed.JoinPath("/phone-api"),
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Sentinel families (family-adoption train, 2026-09-30): an
// unconfigured phone API cannot serve (Infrastructure, nil-deps
// class), and rejected credentials are the operator's/user's to fix
// (Rejection). The sentinels keep their errors.New identity.
func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrDisabled:     errorfamily.Infrastructure,
		ErrUnauthorized: errorfamily.Rejection,
	})
}

// ErrDisabled is returned when no phone API is configured.
var ErrDisabled = errors.New("phone api not configured")

// ErrUnauthorized is returned when the phone API rejected the presented
// credentials (HTTP 401/403): the server-side proof that the
// extension/password pair is not valid in the PBX directory.
var ErrUnauthorized = errors.New("phone api rejected the credentials")

// Enabled reports whether a phone API is wired up.
func (c *Client) Enabled() bool { return c != nil && c.base != nil }

// ResolvePath builds the absolute upstream URL for a /phone-api-relative
// path (without the "/phone-api" prefix), used by the proxy.
func (c *Client) ResolvePath(rest, rawQuery string) string {
	target := c.base.JoinPath(rest).String()
	if rawQuery != "" {
		target += "?" + rawQuery
	}
	return target
}

// HTTPClient exposes the client the /phone-api proxy must ride on, so
// proxied calls share the same timeouts as direct pbx calls instead of
// the timeout-less http.DefaultClient.
func (c *Client) HTTPClient() *http.Client {
	if c.client != nil {
		return c.client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// Credentials are the extension's SIP credentials (Basic auth).
type Credentials struct {
	Extension string
	Password  string
}

// CDR is one call detail record row, field names exactly as the upstream
// API returns them (see the island's panels.js).
type CDR struct {
	Context           string `json:"context"`
	CallerIDNumber    string `json:"caller_id_number"`
	CallerIDName      string `json:"caller_id_name"`
	DestinationNumber string `json:"destination_number"`
	Start             string `json:"start"`
	Billsec           int    `json:"billsec"`
}

// HistoryPage is the /history response shape.
type HistoryPage struct {
	Entries []CDR `json:"entries"`
}

// VoicemailSummary is the /voicemail/{ext}/summary response shape.
type VoicemailSummary struct {
	New int `json:"new"`
	Old int `json:"old"`
}

// VoicemailMessage is one row of the /voicemail/{ext}/messages response.
type VoicemailMessage struct {
	UUID      string `json:"uuid"`
	CIDNumber string `json:"cid_number"`
	CIDName   string `json:"cid_name"`
	Seconds   int    `json:"seconds"`
	Created   int64  `json:"created"`
	Read      bool   `json:"read"`
	AudioURL  string `json:"audio_url"`
}

// VoicemailPage is the /voicemail/{ext}/messages response shape.
type VoicemailPage struct {
	Messages []VoicemailMessage `json:"messages"`
}

// History fetches the extension's recent calls.
func (c *Client) History(ctx context.Context, creds Credentials, limit int) (HistoryPage, error) {
	var body struct {
		Entries []CDR `json:"entries"`
	}
	if err := c.getJSON(ctx, creds, fmt.Sprintf("/history?limit=%d", limit), &body); err != nil {
		return HistoryPage{}, err
	}
	return HistoryPage{Entries: body.Entries}, nil
}

// VoicemailSummary fetches the new/old message counts.
func (c *Client) VoicemailSummary(ctx context.Context, creds Credentials) (VoicemailSummary, error) {
	var summary VoicemailSummary
	err := c.getJSON(ctx, creds, "/voicemail/"+creds.Extension+"/summary", &summary)
	return summary, err
}

// VerifyCredentials proves the extension/password pair against the PBX
// directory with the cheapest authenticated phone-api call (the voicemail
// summary: same Basic-auth directory credentials the SIP REGISTER
// checks). Session creation rides this instead of trusting the island's
// claim that a REGISTER succeeded — a forged POST /api/session must not
// open a tab session scoped to someone else's extension. Returns nil on
// success, ErrUnauthorized on bad credentials, and a wrapped error for
// anything else (PBX unreachable, 5xx). ErrDisabled means no phone API
// is configured and the caller decides the policy for that mode.
func (c *Client) VerifyCredentials(ctx context.Context, creds Credentials) error {
	_, err := c.VoicemailSummary(ctx, creds)
	return err
}

// VoicemailMessages lists the extension's voicemail.
func (c *Client) VoicemailMessages(ctx context.Context, creds Credentials) (VoicemailPage, error) {
	var page VoicemailPage
	err := c.getJSON(ctx, creds, "/voicemail/"+creds.Extension+"/messages", &page)
	return page, err
}

// DeleteVoicemail removes one message.
func (c *Client) DeleteVoicemail(ctx context.Context, creds Credentials, uuid string) error {
	return c.do(ctx, creds, http.MethodDelete, "/voicemail/"+creds.Extension+"/messages/"+uuid, nil, nil)
}

func (c *Client) getJSON(ctx context.Context, creds Credentials, path string, out any) error {
	return c.do(ctx, creds, http.MethodGet, path, nil, out)
}

func (c *Client) do(
	ctx context.Context, creds Credentials, method, path string, in, out any,
) error {
	// The single disabled-policy home: every pbx call funnels through
	// here, so an unconfigured phone API fails every method with
	// ErrDisabled before any request is built.
	if !c.Enabled() {
		return ErrDisabled
	}

	var bodyReader io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "pbx.encode", "encode request")
		}
		bodyReader = bytes.NewReader(encoded)
	}

	// JoinPath escapes the whole input, so a "?query" suffix would reach
	// the upstream percent-encoded ("%3F"). Split it off and re-attach it
	// as a real query string.
	pathOnly, query, _ := strings.Cut(path, "?")
	target := c.base.JoinPath(pathOnly).String()
	if query != "" {
		target += "?" + query
	}
	req, err := http.NewRequestWithContext(ctx, method, target, bodyReader)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "pbx.request", "build request")
	}
	req.SetBasicAuth(creds.Extension, creds.Password)
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errorfamily.WrapTransientf(err, "pbx.transport", "phone api call")
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // read-side close on defer; nothing left to act on

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 4xx is a refusal (Rejection), 5xx the phone API failing
		// (Transient) — the same split the gateway seam pins.
		family := errorfamily.Transient
		if resp.StatusCode < 500 {
			family = errorfamily.Rejection
		}
		return errorfamily.Newf(family, "pbx.http", "phone api: HTTP %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.UnmarshalRead(resp.Body, out); err != nil {
			return errorfamily.WrapTransientf(err, "pbx.decode", "decode phone api response")
		}
	}

	return nil
}
