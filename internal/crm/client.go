// Package crm is the server-side client for the OPTIONAL Ledger CRM
// integration: caller-name enrichment (lookup by phone number) and
// call-activity logging (one call_logged event on the contact's journal).
// A client without a configured base URL is "disabled": every call returns
// ErrDisabled and the server renders raw numbers instead of names.
package crm

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
)

// ErrDisabled is returned when no CRM is configured.
var ErrDisabled = errors.New("crm not configured")

// ErrUnauthorized is returned when the CRM rejected the bearer token
// (HTTP 401/403): the integration is half-configured and must surface.
var ErrUnauthorized = errors.New("crm rejected the bearer token")

// ErrNotFound is returned when the CRM holds no contact for the number.
var ErrNotFound = errors.New("crm has no contact for this number")

// requestTimeout bounds every CRM round-trip. The integration is
// best-effort decoration around the phone: a slow CRM must never make a
// tab render hang.
const requestTimeout = 3 * time.Second

// Client talks to one Ledger CRM machine API (mounted with -api-token).
type Client struct {
	base   *url.URL
	token  string
	client *http.Client
}

// NewClient builds the client; baseURL like "http://127.0.0.1:8080". An
// empty baseURL means "disabled": every call returns ErrDisabled.
func NewClient(baseURL, token string) (*Client, error) {
	if baseURL == "" {
		return &Client{}, nil
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse crm url %q: %w", baseURL, err)
	}

	return &Client{
		base:   parsed,
		token:  token,
		client: &http.Client{Timeout: requestTimeout},
	}, nil
}

// Enabled reports whether a CRM is wired up.
func (c *Client) Enabled() bool { return c != nil && c.base != nil }

// Match is one contact hit for a phone number.
type Match struct {
	ID   string
	Name string
}

// do is the single disabled-policy and request home (the same chokepoint
// shape as pbx.Client.do): every CRM call funnels through here, so an
// unconfigured CRM fails every method with ErrDisabled before any URL or
// request is built, and no caller can forget the bearer header.
func (c *Client) do(
	ctx context.Context, method, path string, query url.Values, body []byte,
) (*http.Response, error) {
	if !c.Enabled() {
		return nil, ErrDisabled
	}

	target := c.base.JoinPath(path).String()
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("crm: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crm: request: %w", err)
	}

	return resp, nil
}

// lookupResponse is the CRM's GET /api/contacts/by-phone wire shape.
type lookupResponse struct {
	Results []struct {
		ID        string   `json:"id"`
		FirstName string   `json:"first_name,omitempty"`
		LastName  string   `json:"last_name,omitempty"`
		Email     string   `json:"email"`
		Phones    []string `json:"phones,omitempty"`
	} `json:"results"`
}

// LookupByPhone resolves a phone number to the first matching CRM contact.
// ErrNotFound on zero matches; the number travels verbatim because the
// CRM owns the whole matching contract (digit normalization, trunk and
// country-code suffix tolerance).
func (c *Client) LookupByPhone(ctx context.Context, number string) (Match, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/contacts/by-phone",
		url.Values{"number": {number}}, nil)
	if err != nil {
		return Match{}, err
	}
	defer drainClose(resp)

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return Match{}, ErrUnauthorized
	case resp.StatusCode != http.StatusOK:
		return Match{}, fmt.Errorf("crm: lookup failed with status %d", resp.StatusCode)
	}

	var buf bytes.Buffer

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return Match{}, fmt.Errorf("crm: read response: %w", err)
	}

	var parsed lookupResponse
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		return Match{}, fmt.Errorf("crm: decode lookup response: %w", err)
	}

	if len(parsed.Results) == 0 {
		return Match{}, ErrNotFound
	}

	first := parsed.Results[0]
	name := strings.TrimSpace(first.FirstName + " " + first.LastName)
	if name == "" {
		name = first.Email
	}

	return Match{ID: first.ID, Name: name}, nil
}

// LogCall appends one call activity to the CRM contact (204 expected).
func (c *Client) LogCall(ctx context.Context, contactID string, direction, number string, seconds int, outcome string) error {
	payload := struct {
		Direction string `json:"direction"`
		Number    string `json:"number"`
		Seconds   int    `json:"seconds"`
		Outcome   string `json:"outcome,omitempty"`
	}{Direction: direction, Number: number, Seconds: seconds, Outcome: outcome}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("crm: encode call log (contact %s): %w", contactID, err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/contacts/"+contactID+"/calls", nil, encoded)
	if err != nil {
		return err
	}
	defer drainClose(resp)

	switch {
	case resp.StatusCode == http.StatusNoContent:
		return nil
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return ErrUnauthorized
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		// 404 (contact gone) and 400/409 (domain rejections) are CRM-side
		// answers, not our bugs — surface them wrapped, not as errors.
		return fmt.Errorf("crm: call log rejected (contact %s, status %d)", contactID, resp.StatusCode)
	default:
		return fmt.Errorf("crm: call log failed (contact %s, status %d)", contactID, resp.StatusCode)
	}
}

// drainClose empties then closes the body so the connection re-enters the
// pool; a read error here is harmless (best-effort reuse).
func drainClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:erraudit // best-effort drain; reuse beats the read error
	_ = resp.Body.Close()                 //nolint:erraudit // best-effort pool return
}
