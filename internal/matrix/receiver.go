package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/larsartmann/webphone/internal/chatloop"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// syncResponse is the slice of GET /sync this seam consumes: the
// cursor, global account data (m.direct DM map), and the joined rooms'
// timelines.
type syncResponse struct {
	NextBatch   string `json:"next_batch"`
	AccountData struct {
		Events []accountDataEvent `json:"events"`
	} `json:"account_data"`
	Rooms struct {
		Join map[string]roomSync `json:"join"`
	} `json:"rooms"`
}

type accountDataEvent struct {
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
}

type roomSync struct {
	Timeline struct {
		Events []roomEvent `json:"events"`
	} `json:"timeline"`
	// Ephemeral carries the room's read receipts (m.receipt): the
	// peer's m.read marker on OUR event id is the outbound message's
	// strongest delivery verdict (lane v2).
	Ephemeral struct {
		Events []ephemeralEvent `json:"events"`
	} `json:"ephemeral"`
}

// ephemeralEvent is one room-ephemeral event; only m.receipt is
// consumed, so the content stays raw until the type is known.
type ephemeralEvent struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// receiptContent is m.receipt's nested map: event id → receipt type →
// reader → metadata. Only the reader KEYS matter (self-reads are
// skipped); the ts is parsed for shape honesty and unused.
type receiptContent map[string]map[string]map[string]struct {
	TS int64 `json:"ts"`
}

// brandid-lint:ignore(BD003) homeserver wire shape: matrix event ids are foreign tokens (strong-id wire ruling)
type roomEvent struct {
	Type           string `json:"type"`
	Sender         string `json:"sender"`
	EventID        string `json:"event_id"`
	OriginServerTS int64  `json:"origin_server_ts"`
	Content        struct {
		MsgType string `json:"msgtype"`
		Body    string `json:"body"`
		URL     string `json:"url"`
		Info    struct {
			MimeType string `json:"mimetype"`
		} `json:"info"`
	} `json:"content"`
}

// Sync long-polls the homeserver: since="" performs the immediate
// initial snapshot, otherwise the call holds up to timeout for new
// events. The returned cursor (next_batch) is the caller's resume
// point.
func (c *Client) Sync(ctx context.Context, since string, timeout time.Duration) (syncResponse, error) {
	query := url.Values{}
	if since != "" {
		query.Set("since", since)
	}
	query.Set("timeout", strconv.Itoa(int(timeout.Milliseconds())))
	var out syncResponse
	if err := c.do(ctx, http.MethodGet, "/_matrix/client/v3/sync?"+query.Encode(), nil, &out); err != nil {
		return syncResponse{}, err
	}
	return out, nil
}

// peerRooms returns the current room→peer view of the DM map (the
// inverse of m.direct): the receiver resolves which conversation a
// timeline event belongs to through it.
func (c *Client) peerRooms() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	inverse := make(map[string]string, len(c.dmRooms))
	for peer, room := range c.dmRooms {
		if _, taken := inverse[room]; !taken {
			inverse[room] = peer
		}
	}
	return inverse
}

// applyDirect folds one m.direct account-data event into the DM map.
func (c *Client) applyDirect(ev accountDataEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for peer, rooms := range ev.Content {
		list, ok := rooms.([]any)
		if !ok || len(list) == 0 {
			continue
		}
		if room, ok := list[0].(string); ok && room != "" {
			c.dmRooms[peer] = room
		}
	}
}

// Cursor persists the sync resume point across restarts. The
// composition root backs it with the app_flags table (one home).
type Cursor interface {
	Load(ctx context.Context) (string, error)
	Save(ctx context.Context, token string) error
}

// Receiver is the matrix inbound lane: a sync loop that turns DM
// timeline events into inbound messages. Encrypted rooms and non-DM
// rooms are skipped LOUDLY (WARN) — a skipped message is a documented
// limit, never a silent lie. One limit, documented here instead of a
// comment at every skip: a per-event store failure logs ERROR and the
// loop advances (a poison message must not wedge the receiver forever;
// the homeserver keeps the history as the recovery source).
type Receiver struct {
	client *Client
	owner  domain.Extension
	sink   chatloop.Sink
	cursor Cursor
	log    *slog.Logger
}

// NewReceiver builds the receiver; all dependencies are injected, so
// tests drive it with a fake cursor and an in-memory sink.
func NewReceiver(client *Client, owner domain.Extension, sink chatloop.Sink, cursor Cursor, log *slog.Logger) *Receiver {
	return &Receiver{client: client, owner: owner, sink: sink, cursor: cursor, log: chatloop.Default(log)}
}

// Run blocks on the sync loop until ctx is done. Errors back off
// (1s..30s doubling); every successful round persists the cursor
// BEFORE the next long-poll, so a crash loses at most the live batch.
func (r *Receiver) Run(ctx context.Context) error {
	backoff := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		since, err := r.cursor.Load(ctx)
		if err != nil {
			r.log.Error("matrix: load sync cursor", "error", err)
			var alive bool
			backoff, alive = chatloop.Backoff(ctx, r.log, "matrix: retrying cursor load", err, backoff, 30*time.Second)
			if !alive {
				return nil
			}
			continue
		}
		timeout := 30 * time.Second
		if since == "" {
			timeout = 0 // initial snapshot returns immediately
		}
		resp, err := r.client.Sync(ctx, since, timeout)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			var alive bool
			backoff, alive = chatloop.Backoff(ctx, r.log, "matrix: sync", err, backoff, 30*time.Second)
			if !alive {
				return nil
			}
			continue
		}
		backoff = time.Second
		r.ingest(ctx, resp)
		if err := r.cursor.Save(ctx, resp.NextBatch); err != nil {
			r.log.Error("matrix: save sync cursor", "error", err)
		}
	}
}

// ingest turns one sync response into inbound messages.
func (r *Receiver) ingest(ctx context.Context, resp syncResponse) {
	for _, ev := range resp.AccountData.Events {
		if ev.Type == "m.direct" {
			r.client.applyDirect(ev)
		}
	}
	roomPeer := r.client.peerRooms()
	for room, joined := range resp.Rooms.Join {
		peer, isDM := roomPeer[room]
		for _, ev := range joined.Timeline.Events {
			if ev.Sender == r.client.userID {
				continue // our own sends echo back; the outbound row is the truth
			}
			if !isDM {
				r.log.Warn("matrix: skipped event in a non-DM room (v1: DMs only)", "room", room, "event", ev.EventID)
				continue
			}
			r.ingestEvent(ctx, peer, ev)
		}
		if isDM {
			r.ingestReceipts(ctx, joined.Ephemeral.Events)
		}
	}
}

// ingestReceipts turns m.receipt ephemera into outbound delivery
// advances: a peer's m.read on an event id names OUR message in that
// room (the sender records the homeserver event id as provider_ref),
// so the receipt is a Delivered verdict through the same machine a
// provider webhook rides. Receipts for events webphone never
// correlated (inbound messages, other clients' history) answer
// store.ErrNotFound and are benign by construction.
func (r *Receiver) ingestReceipts(ctx context.Context, events []ephemeralEvent) {
	for _, ev := range events {
		if ev.Type != "m.receipt" {
			continue
		}
		var content receiptContent
		if err := json.Unmarshal(ev.Content, &content); err != nil {
			r.log.Warn("matrix: unreadable m.receipt content", "error", err)
			continue
		}
		for eventID, byType := range content {
			for reader := range byType["m.read"] {
				if reader == r.client.userID {
					continue // our own read echo is not a delivery fact
				}
				if err := r.sink.RemoteRead(ctx, eventID); err != nil {
					if errors.Is(err, store.ErrNotFound) {
						continue
					}
					r.log.Warn("matrix: apply read receipt", "event", eventID, "error", err)
				}
			}
		}
	}
}

// ingestEvent maps one timeline event onto the inbound pipeline.
func (r *Receiver) ingestEvent(ctx context.Context, peer string, ev roomEvent) {
	switch ev.Type {
	case "m.room.encrypted":
		r.log.Warn("matrix: skipped an encrypted event (v1: plaintext rooms only)", "peer", peer, "event", ev.EventID)
	case "m.room.message":
		switch ev.Content.MsgType {
		case "m.text":
			from, err := domain.ParseAddress(domain.NetworkMatrix, peer)
			if err != nil {
				r.log.Error("matrix: peer address", "peer", peer, "error", err)
				return
			}
			inbound := domain.InboundMessage{
				Owner:      r.owner,
				From:       from,
				Body:       ev.Content.Body,
				ReceivedAt: time.UnixMilli(ev.OriginServerTS),
			}
			r.deliver(ctx, inbound)
		case "m.image", "m.file":
			r.ingestMedia(ctx, peer, ev)
		default:
			// notices, emotes, locations...: not conversation text
			r.log.Debug("matrix: skipped non-text message", "peer", peer, "msgtype", ev.Content.MsgType)
		}
	}
}

// ingestMedia downloads one media event's content and delivers it as
// an attachment-bearing inbound message.
func (r *Receiver) ingestMedia(ctx context.Context, peer string, ev roomEvent) {
	from, err := domain.ParseAddress(domain.NetworkMatrix, peer)
	if err != nil {
		r.log.Error("matrix: peer address", "peer", peer, "error", err)
		return
	}
	data, mimeType, err := r.client.Download(ctx, ev.Content.URL)
	if err != nil {
		r.log.Error("matrix: download media", "peer", peer, "event", ev.EventID, "error", err)
		return
	}
	if mimeType == "" {
		mimeType = ev.Content.Info.MimeType
	}
	name := ev.Content.Body
	if name == "" {
		name = "matrix-attachment"
	}
	inbound := domain.InboundMessage{
		Owner:      r.owner,
		From:       from,
		Body:       "",
		ReceivedAt: time.UnixMilli(ev.OriginServerTS),
		Attachments: []domain.AttachmentContent{{
			Name:     name,
			MimeType: mimeType,
			Bytes:    data,
		}},
	}
	r.deliver(ctx, inbound)
}

func (r *Receiver) deliver(ctx context.Context, inbound domain.InboundMessage) {
	if _, err := r.sink.Receive(ctx, inbound); err != nil {
		r.log.Error("matrix: ingest inbound message", "peer", inbound.From.String(), "error", err)
	}
}
