// Package textagent is the OPTIONAL seam to the stack's supervised
// text-message agent (services.telephony.textagent): the loopback
// review API the Messages tab surfaces. The server only ever sees this
// seam — nil or disabled means the drafts panel and the compose
// affordances do not exist, exactly like the ASR and CRM seams.
package textagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client talks to the review API: one base URL, one bearer secret.
// The zero value is NOT usable; New validates and reads the secret
// file (the deployment shape the stack's secrets pipeline prefers —
// the value never lands in the rendered config JSON).
type Client struct {
	base   string
	secret string
	http   *http.Client
}

// New builds the client from the config pair. A missing or empty
// secret file is a Rejection: half a seam must fail at boot, not at
// the first button press (the same drift argument as the gateway
// secret pair).
func New(url, secretFile string) (*Client, error) {
	data, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("textagent: read secret file %s: %w", secretFile, err)
	}
	secret := strings.TrimSpace(string(data))
	if secret == "" {
		return nil, errors.New("textagent: secret file " + secretFile + " is empty")
	}
	return &Client{
		base:   strings.TrimRight(url, "/"),
		secret: secret,
		http:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Draft is one pending review item (the textagent's draft JSON shape).
type Draft struct {
	ID         string `json:"id"`
	Contact    string `json:"contact"`
	Trigger    string `json:"trigger"`
	Text       string `json:"text"`
	EnGloss    string `json:"en_gloss"`
	Language   string `json:"language"`
	Confidence int    `json:"confidence"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	SendError  string `json:"send_error"`
}

// Frozen is a conversation the kill-switch put on human-only review.
type Frozen struct {
	Contact string `json:"contact"`
	Reason  string `json:"reason"`
}

// Suppressed is a contact that opted out (STOP).
type Suppressed struct {
	Contact string `json:"contact"`
	Since   string `json:"since"`
	Keyword string `json:"keyword"`
}

// Reviews is the review-list payload: pending drafts, frozen
// conversations, opt-outs, and the autopilot mode for the header line.
type Reviews struct {
	Drafts     []Draft      `json:"drafts"`
	Frozen     []Frozen     `json:"frozen"`
	Suppressed []Suppressed `json:"suppressed"`
	Mode       string       `json:"mode"`
}

// Error carries the review API's honest failure: the HTTP status and
// the message the agent already phrased for a human (the fail-closed
// branches name the fix). Handlers map it through, never re-guess.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("textagent: HTTP %d: %s", e.Status, e.Message) }

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("textagent: encode request: %w", err)
		}
		reader = strings.NewReader(string(encoded))
	}
	request, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return fmt.Errorf("textagent: build request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.secret)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("textagent: reach %s: %w", c.base, err)
	}
	defer response.Body.Close() //nolint:errcheck // read-side close on a loopback JSON body
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("textagent: read response: %w", err)
	}
	if response.StatusCode >= 300 {
		message := strings.TrimSpace(string(raw))
		var structured struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &structured) == nil && structured.Error != "" {
			message = structured.Error
		}
		return &Error{Status: response.StatusCode, Message: message}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("textagent: decode response: %w", err)
	}
	return nil
}

// Reviews lists the pending review queue.
func (c *Client) Reviews(ctx context.Context) (*Reviews, error) {
	var reviews Reviews
	if err := c.do(ctx, http.MethodGet, "/reviews", nil, &reviews); err != nil {
		return nil, err
	}
	return &reviews, nil
}

// Approve sends a draft: empty text approves the draft verbatim, a
// non-empty text approves the HUMAN edit (the agent records the
// correction as a learned preference).
func (c *Client) Approve(ctx context.Context, id, text string) (*Draft, error) {
	payload := map[string]string{}
	if text != "" {
		payload["text"] = text
	}
	var draft Draft
	if err := c.do(ctx, http.MethodPost, "/reviews/"+id+"/approve", payload, &draft); err != nil {
		return nil, err
	}
	return &draft, nil
}

// Discard drops a pending draft without sending.
func (c *Client) Discard(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, "/reviews/"+id+"/discard", map[string]string{}, nil)
}

// DraftFor asks for an on-demand draft (the compose button): the
// contact number plus an optional hint the owner typed.
func (c *Client) DraftFor(ctx context.Context, to, hint string) (*Draft, error) {
	payload := map[string]string{"to": to}
	if hint != "" {
		payload["hint"] = hint
	}
	var draft Draft
	if err := c.do(ctx, http.MethodPost, "/draft", payload, &draft); err != nil {
		return nil, err
	}
	return &draft, nil
}
