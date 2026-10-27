package signal

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/larsartmann/webphone/internal/chatloop"
	"github.com/larsartmann/webphone/internal/domain"
)

// ReceiptSink applies delivery verdicts to outbound rows correlated by
// provider_ref — the signal send timestamps. The messaging service
// satisfies it.
type ReceiptSink interface {
	DeliveryReceipt(ctx context.Context, providerRef string, status domain.OutboundStatus, errMsg string) (domain.Message, error)
}

// Receiver is the signal inbound lane: envelope pushes become inbound
// messages, receipt pushes become delivery verdicts. Groups and
// inbound attachments are v1 fences, skipped LOUDLY — a skipped
// message is a documented limit, never a silent lie.
type Receiver struct {
	client   *Client
	owner    domain.Extension
	sink     chatloop.Sink
	receipts ReceiptSink
	log      *slog.Logger
}

// NewReceiver builds the receiver.
func NewReceiver(client *Client, owner domain.Extension, sink chatloop.Sink, receipts ReceiptSink, log *slog.Logger) *Receiver {
	return &Receiver{client: client, owner: owner, sink: sink, receipts: receipts, log: chatloop.Default(log)}
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
	if _, err := r.sink.Receive(ctx, domain.InboundMessage{
		Owner:      r.owner,
		From:       from,
		Body:       env.DataMessage.Message,
		ReceivedAt: time.UnixMilli(ts),
	}); err != nil {
		r.log.Error("signal: ingest inbound message", "from", from.String(), "error", err)
	}
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
