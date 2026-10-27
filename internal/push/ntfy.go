package push

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/larsartmann/go-error-family"
)

// ntfy wire: JSON publish to the server root — one round trip, the
// topic inside the body (the device's unguessable topic is its wake
// address; the server never subscribes to anything).
const (
	// ntfyRequestTimeout bounds one publish. A push the server cannot
	// ship inside this window is dropped — the ring lane logs and moves
	// on; the caller's phone must never wait on ntfy.
	ntfyRequestTimeout = 10 * time.Second
	// ntfyPriorityMax is ntfy's max/urgent priority (the ring wake);
	// ring_end rides the default priority (a stand-down is not urgent).
	ntfyPriorityMax     = 5
	ntfyPriorityDefault = 3
)

// NtfySender publishes ring facts to a self-hosted ntfy server.
type NtfySender struct {
	base   *url.URL
	token  string
	client *http.Client
}

// NewNtfySender builds the sender; baseURL like "https://ntfy.example.org".
// The token is optional (a public self-hosted server needs none) and is
// sent as a Bearer when set. An empty baseURL means "not configured".
func NewNtfySender(baseURL, token string) (*NtfySender, error) {
	if baseURL == "" {
		return &NtfySender{}, nil
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "config.push.ntfy.url", "parse ntfy url %q", baseURL)
	}
	return &NtfySender{
		base:   parsed,
		token:  token,
		client: &http.Client{Timeout: ntfyRequestTimeout},
	}, nil
}

// Enabled reports whether the transport is configured.
func (s *NtfySender) Enabled() bool { return s != nil && s.base != nil }

// Describe names the transport for logs.
func (s *NtfySender) Describe() string { return TransportNtfy }

// ntfyMessage is the JSON publish body. The message field carries the
// ring fact as compact JSON — the shell's receiver parses it, while
// title/priority/tags keep the stock ntfy app's notification shade
// honest when the shell is not installed.
type ntfyMessage struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags,omitempty"`
}

// Send publishes one ring fact to the device's topic.
func (s *NtfySender) Send(ctx context.Context, topic string, ring Ring) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	priority := ntfyPriorityDefault
	title := "Call ended"
	if ring.Event == EventRing {
		priority = ntfyPriorityMax
		title = "Incoming call"
	}
	body, err := json.Marshal(ntfyMessage{
		Topic:    topic,
		Title:    title,
		Message:  string(mustMarshalRing(ring)),
		Priority: priority,
		Tags:     []string{"telephone"},
	})
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.ntfy_send", "encode ntfy message")
	}
	return s.publish(ctx, body, topic)
}

// SendDraft publishes one draft-ready fact. Good news is not urgent:
// the default priority (the ring_end lane's level) — the shell renders
// the fact, and the stock app shows an ordinary shade entry.
func (s *NtfySender) SendDraft(ctx context.Context, topic string, draft Draft) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	body, err := json.Marshal(ntfyMessage{
		Topic:    topic,
		Title:    "Draft ready",
		Message:  string(mustMarshalDraft(draft)),
		Priority: ntfyPriorityDefault,
		Tags:     []string{"telephone"},
	})
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.ntfy_send", "encode ntfy message")
	}
	return s.publish(ctx, body, topic)
}

// publish is the ONE ntfy round trip: JSON to the server root, the
// status mapping every lane shares.
func (s *NtfySender) publish(ctx context.Context, body []byte, topic string) error {
	target := s.base.JoinPath("/").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.ntfy_send", "build ntfy request")
	}
	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "push.ntfy_send", "publish to ntfy")
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // best-effort socket return
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))

	if statusOK(resp.StatusCode) {
		return nil
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return errorfamily.WrapRejectionf(ErrUnauthorized, "push.ntfy_send", "ntfy rejected the credential (%d)", resp.StatusCode)
	case http.StatusNotFound:
		return errorfamily.WrapInfrastructuref(ErrUnregistered, "push.ntfy_send", "ntfy topic %q not found", topic)
	default:
		return errorfamily.Newf(errorfamily.Infrastructure, "push.ntfy_send", "ntfy answered %d for topic %q", resp.StatusCode, topic)
	}
}

// mustMarshalRing renders the ring fact as compact JSON for the ntfy
// message field. A Ring carries only strings; marshaling cannot fail —
// but a panic-free contract beats an impossible-error branch.
func mustMarshalRing(ring Ring) []byte {
	body, err := json.Marshal(ring)
	if err != nil {
		return []byte(`{"event":"ring"}`)
	}
	return body
}

// mustMarshalDraft is the draft fact's marshal twin.
func mustMarshalDraft(draft Draft) []byte {
	body, err := json.Marshal(draft)
	if err != nil {
		return []byte(`{"event":"draft_ready"}`)
	}
	return body
}

// ntfyTopicShape bounds what the registry may store as an ntfy
// address: the shell generates 32 lowercase hex characters; the bound
// admits reasonable operator-chosen topics without admitting junk.
const ntfyTopicShape = 64

// ValidNtfyTopic reports whether a topic looks like an ntfy topic
// (the registration API's validation arm).
func ValidNtfyTopic(topic string) bool {
	if topic == "" || len(topic) > ntfyTopicShape || strings.ContainsAny(topic, " /\t\r\n") {
		return false
	}
	return true
}
