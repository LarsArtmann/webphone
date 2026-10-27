// Package gateway defines the outbound seam for messages and faxes. The
// webphone never talks to a carrier or the PBX directly on this path — it
// hands traffic to a MessageGateway/FaxGateway implementation chosen by
// config:
//
//   - loopback: accepts everything instantly (development, demos, tests)
//   - webhook: forwards to a provider URL (FreeSWITCH bridge or any SMS/MMS/
//     fax HTTP provider)
//
// Inbound traffic arrives on the /hooks/* HTTP endpoints (see server).
package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
)

// OutboundAttachment references a stored attachment by path.
type OutboundAttachment struct {
	Name     string
	MimeType string
	Path     string
}

// OutboundMessage is one message handed to a gateway for delivery.
// The address's network selects the gateway: the messaging service
// routes by it (see messaging.Service.RegisterSender).
type OutboundMessage struct {
	Owner       domain.Extension
	To          domain.Address
	Body        string
	Attachments []OutboundAttachment
}

// Resolution says whether a Receipt is the final verdict or only an
// acceptance: loopback gateways resolve synchronously (immediate), while
// webhook providers merely accepted the job and the verdict arrives later
// on the status webhook (deferred). Deferred is the zero value on purpose:
// a Receipt constructed without an explicit resolution promises nothing.
type Resolution int

const (
	// ResolutionDeferred: the provider accepted the job; the outcome
	// arrives on the status webhook (or never, if it stays silent).
	ResolutionDeferred Resolution = iota
	// ResolutionImmediate: the receipt IS the verdict (loopback mode).
	ResolutionImmediate
)

// Receipt is the gateway's acceptance answer.
type Receipt struct {
	ProviderRef string
	// Resolution tells the caller whether this receipt already settles
	// the delivery (immediate) or only acknowledges acceptance
	// (deferred — wait for the status webhook).
	Resolution Resolution
}

// MessageGateway delivers outbound messages.
type MessageGateway interface {
	SendMessage(ctx context.Context, msg OutboundMessage) (Receipt, error)
}

// OutboundFax is one fax handed to a gateway for transmission.
type OutboundFax struct {
	Owner   domain.Extension
	To      domain.Phone
	PDFPath string
}

// FaxGateway transmits outbound faxes.
type FaxGateway interface {
	SendFax(ctx context.Context, fax OutboundFax) (Receipt, error)
}

// ForwardError is the ONE home for a provider-lane forward failure in
// the services (fax.Service.drive, messaging.Service.Send): the gateway
// refused locally (self-send) or the send failed, and the job/message
// row already carries the failure story. Family-neutral by construction
// — the wrapped cause owns its verdict (Rejection for a provider answer
// or a self-send, Transient for transport) — while the "gateway: "
// prefix keeps the outcome greppable in logs exactly as before.
type ForwardError struct {
	Err error
}

func (e *ForwardError) Error() string { return "gateway: " + e.Err.Error() }

func (e *ForwardError) Unwrap() error { return e.Err }

// NewMessageGateway picks the message gateway by config mode. The
// switch is exhaustive on purpose: an unknown mode PANICS here instead
// of silently degrading to loopback — config.validate is the rejection
// home, so an unvalidated mode reaching this constructor is a wiring
// defect, not a user error, and swallowing the traffic into a pretend
// gateway would hide it.
func NewMessageGateway(cfg config.Gateway, client *http.Client) MessageGateway {
	switch cfg.Mode {
	case config.GatewayWebhook:
		return &Webhook{provider{cfg: cfg, client: client}}
	case config.GatewayLoopback:
		return &Loopback{prefix: "loopback-msg"}
	default:
		panic(fmt.Sprintf("gateway: message mode %q is not one of loopback|webhook — config.validate should have rejected it", cfg.Mode))
	}
}

// NewFaxGateway picks the fax gateway by config mode.
func NewFaxGateway(cfg config.Gateway, client *http.Client) FaxGateway {
	switch cfg.Mode {
	case config.GatewayWebhook:
		return &FaxWebhook{provider{cfg: cfg, client: client}}
	case config.GatewayLoopback:
		return &Loopback{prefix: "loopback-fax"}
	default:
		panic(fmt.Sprintf("gateway: fax mode %q is not one of loopback|webhook — config.validate should have rejected it", cfg.Mode))
	}
}

// Loopback accepts everything instantly. It exists so the whole product
// runs with zero external dependencies: messages are marked "sent" and
// faxes "transmitted" immediately. Nothing leaves the machine.
type Loopback struct {
	prefix string
}

// SendMessage accepts the message instantly — and that acceptance IS the
// verdict (nothing outside this machine will ever say more about it).
func (l *Loopback) SendMessage(_ context.Context, _ OutboundMessage) (Receipt, error) {
	return Receipt{
		ProviderRef: fmt.Sprintf("%s-%d", l.prefix, time.Now().UnixNano()),
		Resolution:  ResolutionImmediate,
	}, nil
}

// SendFax accepts the fax instantly — resolution immediate, same story as
// SendMessage.
func (l *Loopback) SendFax(_ context.Context, _ OutboundFax) (Receipt, error) {
	return Receipt{
		ProviderRef: fmt.Sprintf("%s-%d", l.prefix, time.Now().UnixNano()),
		Resolution:  ResolutionImmediate,
	}, nil
}
