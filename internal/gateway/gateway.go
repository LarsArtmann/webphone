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
type OutboundMessage struct {
	Owner       domain.Extension
	To          domain.Phone
	Body        string
	Attachments []OutboundAttachment
}

// Receipt is the gateway's acceptance answer.
type Receipt struct {
	ProviderRef string
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

// NewMessageGateway picks the message gateway by config mode.
func NewMessageGateway(cfg config.Gateway, client *http.Client) MessageGateway {
	switch cfg.Mode {
	case config.GatewayWebhook:
		return &Webhook{cfg: cfg, client: client}
	default:
		return &Loopback{prefix: "loopback-msg"}
	}
}

// NewFaxGateway picks the fax gateway by config mode.
func NewFaxGateway(cfg config.Gateway, client *http.Client) FaxGateway {
	switch cfg.Mode {
	case config.GatewayWebhook:
		return &FaxWebhook{cfg: cfg, client: client}
	default:
		return &Loopback{prefix: "loopback-fax"}
	}
}

// Loopback accepts everything instantly. It exists so the whole product
// runs with zero external dependencies: messages are marked "sent" and
// faxes "transmitted" immediately. Nothing leaves the machine.
type Loopback struct {
	prefix string
}

// SendMessage accepts the message instantly.
func (l *Loopback) SendMessage(_ context.Context, _ OutboundMessage) (Receipt, error) {
	return Receipt{ProviderRef: fmt.Sprintf("%s-%d", l.prefix, time.Now().UnixNano())}, nil
}

// SendFax accepts the fax instantly.
func (l *Loopback) SendFax(_ context.Context, _ OutboundFax) (Receipt, error) {
	return Receipt{ProviderRef: fmt.Sprintf("%s-%d", l.prefix, time.Now().UnixNano())}, nil
}
