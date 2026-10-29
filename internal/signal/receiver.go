package signal

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/larsartmann/webphone/internal/chatloop"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/errorfamily"
)

// ReceiptSink applies delivery verdicts to outbound rows correlated by
// provider_ref — the signal send timestamps. The messaging service
// satisfies it.
type ReceiptSink interface {
	DeliveryReceipt(ctx context.Context, providerRef string, status domain.OutboundStatus, errMsg string) (domain.Message, error)
}

// Receiver is the signal inbound lane: envelope pushes become inbound
// messages, receipt pushes become delivery verdicts. Groups are a v1
// fence, skipped LOUDLY. Inbound attachments arm when the receiver was
// built with signal-cli's attachments directory (config
// `signal.attachments_dir`): each file is read from that directory by
// its envelope id and lands in Messages like MMS. Without the
// directory they stay fenced — a skipped attachment is a documented
// limit, never a silent lie.
type Receiver struct {
	client         *Client
	owner          domain.Extension
	sink           chatloop.Sink
	receipts       ReceiptSink
	attachmentsDir string
	log            *slog.Logger
}

// Inbound-attachment bounds. Aligned with the messaging seam's
// attachment policy (5 per message, 10 MiB each); a message carrying
// more keeps its TEXT, the excess attachments are skipped with a WARN —
// the envelope stays a delivered message, not a dropped one.
const (
	maxInboundAttachments = 5
	maxInboundAttachment  = 10 << 20 // 10 MiB
)

// NewReceiver builds the receiver; attachmentsDir is signal-cli's
// received-attachment directory ("" keeps the attachment fence).

// NewReceiver builds the receiver.
func NewReceiver(client *Client, owner domain.Extension, sink chatloop.Sink, receipts ReceiptSink, attachmentsDir string, log *slog.Logger) *Receiver {
	return &Receiver{client: client, owner: owner, sink: sink, receipts: receipts, attachmentsDir: attachmentsDir, log: chatloop.Default(log)}
}

// Run blocks until ctx is done: it registers the envelope handler and
// keeps the connection honest with a periodic ping (a dead connection
// is re-dialed by the next ping's call path; signal-cli re-fetches
// anything the socket missed while down).
func (r *Receiver) Run(ctx context.Context) error {
	r.client.OnEnvelope(r.handle)
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := r.client.Ping(ctx); err != nil {
			var alive bool
			backoff, alive = chatloop.Backoff(ctx, r.log, "signal: daemon unreachable", err, backoff, 60*time.Second)
			if !alive {
				return nil
			}
			continue
		}
		backoff = time.Second
		if !chatloop.Sleep(ctx, 30*time.Second) {
			return nil
		}
	}
}

// handle maps one envelope onto the pipelines.
func (r *Receiver) handle(env Envelope) {
	ctx := context.Background()
	switch {
	case env.DataMessage != nil:
		r.handleData(ctx, env)
	case env.ReceiptMessage != nil:
		r.handleReceipt(ctx, env)
	default:
		// typing indicators, sync echoes from other devices, calls...
		r.log.Debug("signal: skipped non-DM envelope", "source", env.Source)
	}
}

func (r *Receiver) handleData(ctx context.Context, env Envelope) {
	if env.DataMessage.GroupInfo != nil {
		r.log.Warn("signal: skipped a group message (v1: DMs only)", "source", env.Source)
		return
	}
	source := env.Source
	if source == "" {
		r.log.Warn("signal: envelope without a source")
		return
	}
	from, err := domain.ParseAddress(domain.NetworkSignal, source)
	if err != nil {
		r.log.Error("signal: source address", "source", source, "error", err)
		return
	}
	ts := env.DataMessage.Timestamp
	if ts == 0 {
		ts = env.Timestamp
	}
	attachments, err := r.readAttachments(env)
	if err != nil {
		r.log.Error("signal: read inbound attachments", "from", source, "error", err)
	}
	if _, err := r.sink.Receive(ctx, domain.InboundMessage{
		Owner:       r.owner,
		From:        from,
		Body:        env.DataMessage.Message,
		Attachments: attachments,
		ReceivedAt:  time.UnixMilli(ts),
	}); err != nil {
		r.log.Error("signal: ingest inbound message", "from", from.String(), "error", err)
	}
}

// readAttachments loads the envelope's attachment files from
// signal-cli's directory. Per-attachment problems (unknown id, oversized,
// over-count) skip THAT attachment with a WARN and keep the rest — the
// message text lands regardless; only a directory-level failure (unset
// dir with attachments present, or an unreadable directory) returns an
// error for the caller's log while the text still delivers.
func (r *Receiver) readAttachments(env Envelope) ([]domain.AttachmentContent, error) {
	atts := env.DataMessage.Attachments
	if len(atts) == 0 {
		return nil, nil
	}
	if r.attachmentsDir == "" {
		r.log.Warn("signal: skipped inbound attachment(s) — signal.attachments_dir is not configured", "count", len(atts))
		return nil, nil
	}
	var (
		out []domain.AttachmentContent
		err error
	)
	for i, att := range atts {
		if i >= maxInboundAttachments {
			r.log.Warn("signal: skipped excess inbound attachment", "count", len(atts)-maxInboundAttachments, "cap", maxInboundAttachments)
			break
		}
		if att.ID == "" {
			r.log.Warn("signal: skipped inbound attachment without an id")
			continue
		}
		if att.Size > maxInboundAttachment {
			r.log.Warn("signal: skipped oversized inbound attachment", "id", att.ID, "size", att.Size, "cap", maxInboundAttachment)
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(r.attachmentsDir, filepath.Base(att.ID)))
		if readErr != nil {
			err = errors.Join(err, errorfamily.WrapInfrastructuref(readErr, "signal.attachment_read", "signal: read attachment %s", att.ID))
			continue
		}
		name := att.Filename
		if name == "" {
			name = "signal-attachment"
		}
		out = append(out, domain.AttachmentContent{Name: name, MimeType: att.ContentType, Bytes: data})
	}
	return out, err
}

func (r *Receiver) handleReceipt(ctx context.Context, env Envelope) {
	if !env.ReceiptMessage.IsDelivery && !env.ReceiptMessage.IsRead {
		return // viewed receipts carry no extra state worth surfacing
	}
	for _, ts := range env.ReceiptMessage.Timestamps {
		ref := strconv.FormatInt(ts, 10)
		if ref == "0" {
			continue
		}
		// deliveryMoves: every verdict is legal from every state, so a
		// stale or replayed receipt can only refresh, never corrupt.
		if _, err := r.receipts.DeliveryReceipt(ctx, ref, domain.StatusDelivered, ""); err != nil {
			r.log.Warn("signal: apply receipt", "ref", ref, "error", err)
		}
	}
}
