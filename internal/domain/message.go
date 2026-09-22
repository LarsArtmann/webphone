package domain

import "time"

// Channel distinguishes SMS (plain text) from MMS (carries attachments).
// The channel of a message is derived: attachments present means MMS.
type Channel string

const (
	// ChannelSMS is a text-only message.
	ChannelSMS Channel = "sms"
	// ChannelMMS is a message with attachments.
	ChannelMMS Channel = "mms"
)

// Direction tells which way a message traveled.
type Direction string

const (
	// DirectionInbound traveled from the remote party to the extension.
	DirectionInbound Direction = "in"
	// DirectionOutbound traveled from the extension to the remote party.
	DirectionOutbound Direction = "out"
)

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

// Thread is one conversation between an extension and a remote number,
// keyed by (Owner, Remote). It exists only as a read model: rows are
// created on first message and updated on every append.
type Thread struct {
	ID             ThreadID
	Owner          Extension
	Remote         Phone
	LastActivityAt time.Time
	Unread         int //nolint:branching-flow // display counter, not identity
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
	ID          MessageID
	ThreadID    ThreadID
	Owner       Extension
	Remote      Phone
	Direction   Direction
	Channel     Channel
	Body        string
	Status       OutboundStatus // zero for inbound (see OutboundStatus)
	ProviderRef  string         // gateway correlation id, "" when none
	FailureKind  string         // "", "transient", "rejected" or "provider"; only transient is retryable
	FailureDetail string        // raw failure reason, rendered verbatim (operator English)
	Attachments  []Attachment
	CreatedAt    time.Time
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
