// Package push is the server-side push seam behind ring-while-closed:
// the registered-device registry (readmodels.PushDevices) holds one
// wake address per install, and this package ships the inbound-call
// fact to it over every configured transport. Two transports exist by
// owner decision (2026-10-23): ntfy (self-hosted, no Google, the
// de-Googled device's lane) and FCM HTTP v1 (service-account
// credentials, the GMS device's lane). The Android shell registers the
// transport it can receive on; the server never assumes one.
//
// A fanout with no configured sender is "disabled": every send returns
// ErrDisabled, the /api/push-devices routes answer 404 and the product
// stays byte-identical — the same optional-seam doctrine as
// asr/crm/paperless.
package push

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/larsartmann/go-error-family"
)

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrDisabled:     errorfamily.Infrastructure,
		ErrUnauthorized: errorfamily.Rejection,
		ErrUnregistered: errorfamily.Infrastructure,
	})
}

// Transport names — the vocabulary the device registry and the fanout
// route on. Foreign protocol vocabulary, deliberately raw strings on
// the wire (the strong-id ruling's foreign-token carve-out).
const (
	TransportNtfy = "ntfy"
	TransportFcm  = "fcm"
)

// statusOK reports a 2xx transport answer — the one success every push
// provider promises (the bodies differ beyond that).
func statusOK(code int) bool {
	return code >= 200 && code < 300
}

// Event is the inbound-call lifecycle fact a ring message carries.
// "ring" wakes the device with a CallStyle notification; "ring_end"
// tells it to stand down (answered elsewhere, caller gave up).
type Event string

const (
	EventRing       Event = "ring"
	EventRingEnd    Event = "ring_end"
	EventDraftReady Event = "draft_ready"
)

// Ring is the inbound-call fact: which event, the caller, and
// the PBX call id (the dedup/ordering key the shell correlates a
// ring_end against its ring by).
// brandid-lint:ignore(BD003) shell wire contract: the PBX call id is an opaque foreign token (strong-id wire ruling)
type Ring struct {
	Event  Event  `json:"event"`
	From   string `json:"from"`
	CallID string `json:"call_id"`
}

// Draft is the draft-ready fact: a follow-up draft (post-call summarize
// or the agent's own drafting) now waits in the review queue for the
// owner. The contact names the conversation the draft addresses, so the
// shell can offer a jump straight into it.
// brandid-lint:ignore(BD003) shell wire contract: the draft id is a foreign textagent token (strong-id wire ruling)
type Draft struct {
	Event   Event  `json:"event"`
	Contact string `json:"contact"`
	DraftID string `json:"draft_id"`
}

// ErrDisabled is returned when no transport is configured (or the
// whole seam is nil).
var ErrDisabled = errors.New("push not configured")

// ErrUnauthorized is returned when a transport rejected the credential
// (HTTP 401/403): the integration is half-configured and must surface.
var ErrUnauthorized = errors.New("push transport rejected the credential")

// ErrUnregistered is returned when the transport reports the device
// address as dead (FCM UNREGISTERED, ntfy's 404 lane): the caller
// should prune the registration — shipping to it again is waste.
var ErrUnregistered = errors.New("push device no longer registered")

// Sender is one transport's wire: one device address, one ring fact.
// The ntfy and FCM clients implement it; tests inject fakes.
type Sender interface {
	// Enabled reports whether the transport is configured.
	Enabled() bool
	// Describe names the transport for logs.
	Describe() string
	// Send ships one ring fact to one device address.
	Send(ctx context.Context, target string, ring Ring) error
	// SendDraft ships one draft-ready fact to one device address.
	SendDraft(ctx context.Context, target string, draft Draft) error
}

// Fanout routes a ring to the sender its transport names. The
// zero Fanout (and a nil pointer) is the disabled seam.
type Fanout struct {
	senders map[string]Sender
}

// NewFanout builds the fanout from transport name to sender; nil and
// disabled senders are dropped (a half-built transport never silently
// receives traffic).
func NewFanout(senders map[string]Sender) *Fanout {
	live := make(map[string]Sender, len(senders))
	for transport, sender := range senders {
		if sender != nil && sender.Enabled() {
			live[transport] = sender
		}
	}
	return &Fanout{senders: live}
}

// Enabled reports whether any transport is live.
func (f *Fanout) Enabled() bool { return f != nil && len(f.senders) > 0 }

// Knows reports whether a transport would actually carry — the device
// registration API's honest signal (registering a transport the server
// cannot send to would strand the device).
func (f *Fanout) Knows(transport string) bool {
	return f != nil && f.senders[transport] != nil
}

// Describe names the live transports, sorted (log-stable).
func (f *Fanout) Describe() string {
	if !f.Enabled() {
		return "off"
	}
	names := make([]string, 0, len(f.senders))
	for name := range f.senders {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, "+")
}

// Send routes one ring fact to the sender its transport names. An
// unknown transport is a Rejection (the registry and the fanout
// disagreed — an operator problem, not a transport outage).
func (f *Fanout) Send(ctx context.Context, transport, target string, ring Ring) error { //nolint:branching-flow // push.Transport defined type is a recorded open-roadmap item (2026-10-25 review)
	if !f.Enabled() {
		return ErrDisabled
	}
	sender, ok := f.senders[transport]
	if !ok {
		return errorfamily.Newf(errorfamily.Rejection, "push.transport", "no sender configured for transport %q", transport)
	}
	return sender.Send(ctx, target, ring)
}

// SendDraft routes one draft-ready fact to the sender its transport
// names (the same routing contract as Send).
func (f *Fanout) SendDraft(ctx context.Context, transport, target string, draft Draft) error { //nolint:branching-flow // push.Transport defined type is a recorded open-roadmap item (2026-10-25 review)
	if !f.Enabled() {
		return ErrDisabled
	}
	sender, ok := f.senders[transport]
	if !ok {
		return errorfamily.Newf(errorfamily.Rejection, "push.transport", "no sender configured for transport %q", transport)
	}
	return sender.SendDraft(ctx, target, draft)
}
