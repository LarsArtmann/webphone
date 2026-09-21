// Package pbx is the server-side client for the per-extension phone API
// (voicemail, CDR history) that the telephony stack serves. It rides the
// extension's own SIP credentials via Basic auth — the same contract the
// browser island uses through the /phone-api proxy.
package pbx

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
		return nil, fmt.Errorf("parse phone api url %q: %w", baseURL, err)
	}
	return &Client{
		base:   parsed.JoinPath("/phone-api"),
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// ErrDisabled is returned when no phone API is configured.
var ErrDisabled = fmt.Errorf("phone api not configured")

// ErrUnauthorized is returned when the phone API rejected the presented
// credentials (HTTP 401/403): the server-side proof that the
// extension/password pair is not valid in the PBX directory.
var ErrUnauthorized = fmt.Errorf("phone api rejected the credentials")

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
	var page HistoryPage
	if !c.Enabled() {
		return page, ErrDisabled
	}
	var body struct {
		Entries []CDR `json:"entries"`
	}
	if err := c.getJSON(ctx, creds, fmt.Sprintf("/history?limit=%d", limit), &body); err != nil {
		return page, err
	}
	return HistoryPage{Entries: body.Entries}, nil
}

// VoicemailSummary fetches the new/old message counts.
func (c *Client) VoicemailSummary(ctx context.Context, creds Credentials) (VoicemailSummary, error) {
	var summary VoicemailSummary
	if !c.Enabled() {
		return summary, ErrDisabled
	}
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
	if !c.Enabled() {
		return ErrDisabled
	}
	var summary VoicemailSummary
	return c.getJSON(ctx, creds, "/voicemail/"+creds.Extension+"/summary", &summary)
}

// VoicemailMessages lists the extension's voicemail.
func (c *Client) VoicemailMessages(ctx context.Context, creds Credentials) (VoicemailPage, error) {
	var page VoicemailPage
	if !c.Enabled() {
		return page, ErrDisabled
	}
	err := c.getJSON(ctx, creds, "/voicemail/"+creds.Extension+"/messages", &page)
	return page, err
}

// DeleteVoicemail removes one message.
func (c *Client) DeleteVoicemail(ctx context.Context, creds Credentials, uuid string) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	return c.do(ctx, creds, http.MethodDelete, "/voicemail/"+creds.Extension+"/messages/"+uuid, nil, nil)
}

func (c *Client) getJSON(ctx context.Context, creds Credentials, path string, out any) error {
	return c.do(ctx, creds, http.MethodGet, path, nil, out)
}

func (c *Client) do(
	ctx context.Context, creds Credentials, method, path string, in, out any,
) error {
	var bodyReader io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
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
		return fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(creds.Extension, creds.Password)
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("phone api call: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // read-side close on defer; nothing left to act on

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("phone api: HTTP %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.UnmarshalRead(resp.Body, out); err != nil {
			return fmt.Errorf("decode phone api response: %w", err)
		}
	}

	return nil
}
