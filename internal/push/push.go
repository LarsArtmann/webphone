// Package push is the server-side push seam behind ring-while-closed:
// the registered-device registry (readmodels.PushDeviceGate, answered
// by the journal backend) holds one
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
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
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
// Transport names one wake wire (the fanout's map key and the device
// row's vocabulary). ParseTransport is the door — raw strings never
// become Transport values, so a client typo dies at the registration
// boundary instead of stranding a device behind a lookup that can
// never hit.
type Transport string

const (
	TransportNtfy Transport = "ntfy"
	TransportFcm  Transport = "fcm"
)

// ParseTransport resolves the wire vocabulary; anything outside it is a
// Rejection (the registration API answers 400, the family test pins the
// shape).
func ParseTransport(raw string) (Transport, error) {
	switch t := Transport(raw); t {
	case TransportNtfy, TransportFcm:
		return t, nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "push.transport_parse",
			"%q is not a push transport (ntfy or fcm)", raw)
	}
}

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

// Fanout routes a fact to the sender its transport names and owns the
// whole delivery policy: detach from the caller's context (a hook's
// client hangup must never kill a fan mid-flight), one bounded budget,
// prune-on-unregistered. The zero Fanout (and a nil pointer) is the
// disabled seam.
type Fanout struct {
	senders  map[Transport]Sender
	registry DeviceRegistry
	fans     sync.WaitGroup
}

// Device is one registered wake address, flattened for the fanout (the
// read model's view carries owner/updatedAt the fanout never needs).
type Device struct {
	DeviceID  string
	Transport Transport
	Token     string
}

// DeviceRegistry is the fanout's prune seam: the registered-device
// store it lists from and prunes dead rows out of. Satisfied by an
// adapter over readmodels.PushDeviceGate at the composition root — the
// fanout stays dependency-light (domain only).
type DeviceRegistry interface {
	List(ctx context.Context, owner domain.Extension) ([]Device, error)
	Remove(ctx context.Context, owner domain.Extension, deviceID string) error
}

// NewFanout builds the fanout from transport name to sender; nil and
// disabled senders are dropped (a half-built transport never silently
// receives traffic). The registry may be nil — Fan then refuses (a
// fan without a registry has nothing to fan to); the composition root
// arms it once the device store resolves.
func NewFanout(senders map[Transport]Sender, registry DeviceRegistry) *Fanout {
	live := make(map[Transport]Sender, len(senders))
	for transport, sender := range senders {
		if sender != nil && sender.Enabled() {
			live[transport] = sender
		}
	}
	return &Fanout{senders: live, registry: registry}
}

// WithRegistry arms (or replaces) the prune seam — the composition
// root's hook for the registry that resolves after the senders do.
func (f *Fanout) WithRegistry(registry DeviceRegistry) {
	if f != nil {
		f.registry = registry
	}
}

// Enabled reports whether any transport is live.
func (f *Fanout) Enabled() bool { return f != nil && len(f.senders) > 0 }

// Knows reports whether a transport would actually carry — the device
// registration API's honest signal (registering a transport the server
// cannot send to would strand the device).
func (f *Fanout) Knows(transport Transport) bool {
	return f != nil && f.senders[transport] != nil
}

// Describe names the live transports, sorted (log-stable).
func (f *Fanout) Describe() string {
	if !f.Enabled() {
		return "off"
	}
	names := make([]string, 0, len(f.senders))
	for name := range f.senders {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return strings.Join(names, "+")
}

// fanBudget bounds one whole fan (list + every device's send): the
// transports' own per-request timeouts sit under it, and the callers'
// answers never wait — Fan detaches before the budget even starts.
const fanBudget = 15 * time.Second

// Fan fans one fact (Ring or Draft) to every registered device of the
// owner — the ONE home of the delivery policy. Fire-and-forget: the
// caller's answer never waits on push health (the ring hook's
// 202-inert doctrine), so Fan validates synchronously (disabled seam,
// unknown fact, missing registry — the caller's own bugs, answered
// immediately) and detaches everything else: WithoutCancel so a hook's
// client hangup cannot kill rings mid-flight, one bounded budget, the
// transports' own timeouts underneath. A transport reporting a device
// dead prunes the registration on the spot; every other per-device
// failure is a greppable Warn, never a caller error.
func (f *Fanout) Fan(ctx context.Context, owner domain.Extension, fact any) error {
	if !f.Enabled() {
		return ErrDisabled
	}
	switch fact.(type) {
	case Ring, Draft:
	default:
		return errorfamily.Newf(errorfamily.Rejection, "push.fact", "fan: unsupported fact %T", fact)
	}
	if f.registry == nil {
		return errorfamily.Newf(errorfamily.Rejection, "push.registry", "fan: no device registry wired")
	}
	detached := context.WithoutCancel(ctx)
	f.fans.Go(func() {
		// Goroutine containment (owner ruling 2026-10-27): RecoveryMiddleware
		// cannot see this goroutine, so a panicking transport or registry
		// would take the whole process down. Recover, log greppably, survive.
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("push: fan-out PANIC recovered; goroutine contained, process survives",
					"panic", rec, "extension", owner.String())
			}
		}()
		fctx, cancel := context.WithTimeout(detached, fanBudget)
		defer cancel()
		devices, err := f.registry.List(fctx, owner)
		if err != nil {
			slog.Warn("push: fan-out aborted; devices unreachable",
				"error", err, "extension", owner.String())
			return
		}
		for _, device := range devices {
			if err := f.route(fctx, device, fact); err != nil {
				if errors.Is(err, ErrUnregistered) {
					if rmErr := f.registry.Remove(fctx, owner, device.DeviceID); rmErr != nil {
						slog.Warn("push: prune dead device failed", "device", device.DeviceID, "err", rmErr)
					} else {
						slog.Info("push: pruned dead device", "device", device.DeviceID, "transport", device.Transport)
					}
					continue
				}
				slog.Warn("push: fan send failed",
					"device", device.DeviceID, "transport", device.Transport, "err", err)
			}
		}
	})
	return nil
}

// Settle awaits every detached fan this Fanout launched. Production
// callers never wait (the 202-inert doctrine: the answer never waits
// on push health), but tests that hand Fan a store living in a TempDir
// must settle before its cleanup removes the directory, or the fan's
// store read races the removal (the 2026-11-13 ring-hook CI flake).
// Bounded by ctx.
func (f *Fanout) Settle(ctx context.Context) error {
	if f == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		f.fans.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// route sends one fact to one device through the sender its transport
// names. An unknown transport is a Rejection (the registry and the
// fanout disagreed — an operator problem, not a transport outage).
func (f *Fanout) route(ctx context.Context, device Device, fact any) error {
	sender, ok := f.senders[device.Transport]
	if !ok {
		return errorfamily.Newf(errorfamily.Rejection, "push.transport", "no sender configured for transport %q", device.Transport)
	}
	switch fact := fact.(type) {
	case Ring:
		return sender.Send(ctx, device.Token, fact)
	case Draft:
		return sender.SendDraft(ctx, device.Token, fact)
	default:
		return errorfamily.Newf(errorfamily.Rejection, "push.fact", "route: unsupported fact %T", fact)
	}
}
