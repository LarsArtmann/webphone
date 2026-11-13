package matrix

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"encoding/json/jsontext"
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
	Type    string         `json:"type"`
	Content jsontext.Value `json:"content"`
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
