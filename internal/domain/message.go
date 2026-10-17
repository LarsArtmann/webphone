package domain

import (
	"time"

	"github.com/larsartmann/go-error-family"
)

// Channel distinguishes SMS (plain text) from MMS (carries attachments).
// The channel of a message is derived: attachments present means MMS.
type Channel string

const (
	// ChannelSMS is a text-only message.
	ChannelSMS Channel = "sms"
	// ChannelMMS is a message with attachments.
	ChannelMMS Channel = "mms"
)

// ParseChannel validates a stored or wire channel string.
func ParseChannel(raw string) (Channel, error) {
	switch Channel(raw) {
	case ChannelSMS, ChannelMMS:
		return Channel(raw), nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "domain.channel", "channel %q is not one of sms|mms", raw)
	}
}

// MustChannel is ParseChannel for row-scan sites: the writer only ever
// stores parsed channels, so an unparsable value is a corrupt row the
// operator must see at first read, not a mystery chip rendered to a user.
func MustChannel(raw string) Channel { return must(ParseChannel(raw)) }

// Direction tells which way a message traveled.
type Direction string

const (
	// DirectionInbound traveled from the remote party to the extension.
	DirectionInbound Direction = "in"
	// DirectionOutbound traveled from the extension to the remote party.
	DirectionOutbound Direction = "out"
)

// ParseDirection validates a stored or wire direction string.
func ParseDirection(raw string) (Direction, error) {
	switch Direction(raw) {
	case DirectionInbound, DirectionOutbound:
		return Direction(raw), nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "domain.direction", "direction %q is not one of in|out", raw)
	}
}

// MustDirection is ParseDirection for row-scan and importer sites: a
// corrupt row fails fast at first read instead of rendering as a lying
// arrow icon.
func MustDirection(raw string) Direction { return must(ParseDirection(raw)) }

// OutboundStatus is the delivery lifecycle of an outbound message. The zero
// value is "not applicable" and is the only legal status for inbound
// messages — inbound delivery is implicit.
type OutboundStatus string

const (
	// StatusQueued: accepted locally, not yet handed to a gateway.
	StatusQueued OutboundStatus = "queued"
	// StatusSent: the gateway accepted the message.
	StatusSent OutboundStatus = "sent"
	// StatusDelivered: the provider confirmed delivery to the handset.
	StatusDelivered OutboundStatus = "delivered"
	// StatusFailed: the gateway or provider rejected the message.
	StatusFailed OutboundStatus = "failed"
)

// ParseOutboundStatus validates a stored or wire status string; the zero
// value is legal (inbound rows carry no lifecycle).
func ParseOutboundStatus(raw string) (OutboundStatus, error) {
	switch OutboundStatus(raw) {
	case "", StatusQueued, StatusSent, StatusDelivered, StatusFailed:
		return OutboundStatus(raw), nil
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "domain.outbound_status", "status %q is not one of queued|sent|delivered|failed (or empty for inbound)", raw)
	}
}

// MustOutboundStatus is ParseOutboundStatus for row-scan sites (see
// MustChannel).
func MustOutboundStatus(raw string) OutboundStatus { return must(ParseOutboundStatus(raw)) }

// Thread is one conversation between an extension and a remote number,
// keyed by (Owner, Remote). Rows are created on first message and
// updated on every append — AppendToThread is that rule's one home.
type Thread struct {
	ID             ThreadID
	Owner          Extension
	Remote         Phone
	LastActivityAt time.Time
	Unread         int //nolint:branching-flow // display counter, not identity
	// Organization flags (M22): three independent axes, never an enum.
	// Pinned sorts first; Archived leaves the default list (an inbound
	// message clears it — the conversation is live again); Muted
	// suppresses only the attention surfaces (row badge, nav unread
	// total) — the unread count itself stays honest.
	Pinned   bool
	Archived bool
	Muted    bool
}

// Snippet is a canned reply text owned by one extension. Quick snippets
// additionally render as chips above the reply composer.
//
// branching-flow COMPOSITION_mixin (baselined): mirrors its storage row
// by the parity-port ruling — the shape IS the contract both stores
// (SQL, engine) served verbatim.
type Snippet struct {
	ID        SnippetID
	Owner     Extension
	Body      string
	Quick     bool
	CreatedAt time.Time
}

// Attachment is a stored MMS attachment. Content lives on disk at Path
// (relative to the data dir); the database keeps metadata only.
type Attachment struct {
	ID        AttachmentID
	MessageID MessageID
	Name      string
	MimeType  string
	SizeBytes int64
	Path      string
}

// Message is a single SMS/MMS inside a thread.
type Message struct {
	ID            MessageID
	ThreadID      ThreadID
	Owner         Extension
	Remote        Phone
	Direction     Direction
	Channel       Channel
	Body          string
	Status        OutboundStatus // zero for inbound (see OutboundStatus)
	ProviderRef   string         // gateway correlation id, "" when none
	FailureKind   FailureKind    // retry vocabulary; only transient is retryable (see delivery.go)
	FailureDetail string         // raw failure reason, rendered verbatim (operator English)
	Attachments   []Attachment
	CreatedAt     time.Time
}

// ChannelOf derives the channel from a message's shape: any attachment
// makes it an MMS.
func ChannelOf(attachmentCount int) Channel {
	if attachmentCount > 0 {
		return ChannelMMS
	}
	return ChannelSMS
}

// InboundMessage is an inbound message as it crosses the gateway boundary
// (webhook or dev loopback). It carries no status: inbound delivery is a
// fact, not a lifecycle.
// branching-flow COMPOSITION_mixin (baselined): OutboundMessage shares
// fields by ROLE, not duplication — inbound is a fact, outbound is a
// lifecycle (registry 2026-10-05).
type InboundMessage struct {
	Owner       Extension
	From        Phone
	Body        string
	Attachments []AttachmentContent
	ReceivedAt  time.Time
}

// AttachmentContent is raw attachment content as it crosses a boundary
// (upload form or inbound gateway) before being spooled to disk.
type AttachmentContent struct {
	Name     string
	MimeType string
	Bytes    []byte
}
