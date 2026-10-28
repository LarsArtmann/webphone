// Package matrix is the OPTIONAL chat-network seam to the Matrix
// Client-Server API: a deliberately thin, stdlib-only client (three
// endpoints + media) behind the matrix config gate. No SDK, no E2EE
// machinery — encrypted rooms are skipped loudly (WARN), plaintext DMs
// only, exactly like the hand-rolled pbx/crm/asr HTTP seams this repo
// favors over heavyweight wrappers.
package matrix

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// downloadCap bounds one inbound media download: the homeserver accepts
// up to ~50 MiB, our blob use case is chat media, and an unbounded
// read is an unbounded memory claim.
const downloadCap = 25 << 20

// Client talks to one homeserver with one long-lived access token (the
// standard bot pattern: minted once via POST /login, stored server-side
// as a secret). It carries the DM room map synced from m.direct
// account data — the sender resolves peers through it, the receiver
// refreshes it on every sync, so one shared instance serves both.
// brandid-lint:ignore(BD003) the bot MXID is foreign matrix identity vocabulary, passed through verbatim (strong-id ruling)
type Client struct {
	homeserver string
	userID     string
	token      string
	http       *http.Client
	txn        atomic.Uint64

	mu      sync.Mutex
	dmRooms map[string]string // peer MXID -> DM room id
}

// New builds the client. Config validation has already failed closed on
// half configurations; the constructor stays mechanical.
// compile-time conformance: the client IS the message gateway of the
// matrix lane (the messaging service routes it by network).
var _ gateway.MessageGateway = (*Client)(nil)

// New builds the client. Config validation has already failed closed on
// half configurations; the constructor stays mechanical.
// brandid-lint:ignore(BD003) the bot MXID arrives pre-validated from config; foreign matrix vocabulary (strong-id ruling)
func New(homeserver, userID, token string) *Client { //nolint:branching-flow // foreign matrix vocabulary, pre-validated config (BD003 ruling at site)
	return &Client{
		homeserver: strings.TrimRight(homeserver, "/"),
		userID:     userID,
		token:      token,
		http:       &http.Client{Timeout: 60 * time.Second},
		dmRooms:    map[string]string{},
	}
}

// SendMessage implements gateway.MessageGateway: the outbound lane for
// matrix threads. The destination is the peer MXID (the thread
// identity); the DM room is resolved (or created) per send, attachments
// upload as media events, and the homeserver's event id is the receipt
// — the acceptance IS the sent verdict (no later status push exists).
func (c *Client) SendMessage(ctx context.Context, msg gateway.OutboundMessage) (gateway.Receipt, error) {
	if msg.To.Network() != domain.NetworkMatrix {
		return gateway.Receipt{}, errorfamily.Newf(errorfamily.Rejection, "matrix.network", "matrix: destination %s is not a matrix address", msg.To)
	}
	peer := msg.To.String()
	room, err := c.roomFor(ctx, peer)
	if err != nil {
		return gateway.Receipt{}, err
	}

	lastEvent := ""
	for _, att := range msg.Attachments {
		data, err := os.ReadFile(att.Path)
		if err != nil {
			return gateway.Receipt{}, errorfamily.WrapInfrastructuref(err, "matrix.attachment_read", "matrix: read attachment %s", att.Path)
		}
		mxc, err := c.Upload(ctx, att.Name, att.MimeType, data)
		if err != nil {
			return gateway.Receipt{}, err
		}
		eventID, err := c.sendEvent(ctx, room, fileMessageType(att.MimeType), map[string]any{
			"msgtype": fileMessageType(att.MimeType),
			"body":    att.Name,
			"url":     mxc,
			"info":    map[string]any{"mimetype": att.MimeType, "size": len(data)},
		})
		if err != nil {
			return gateway.Receipt{}, err
		}
		lastEvent = eventID
	}
	if msg.Body != "" {
		eventID, err := c.sendEvent(ctx, room, "m.room.message", map[string]any{
			"msgtype": "m.text",
			"body":    msg.Body,
		})
		if err != nil {
			return gateway.Receipt{}, err
		}
		lastEvent = eventID
	}
	if lastEvent == "" {
		return gateway.Receipt{}, errorfamily.NewRejection("matrix.empty", "matrix: nothing to send (no body, no attachments)")
	}
	return gateway.Receipt{ProviderRef: lastEvent, Resolution: gateway.ResolutionImmediate}, nil
}

// Whoami verifies the token at boot: a wrong or expired token fails
// HERE, with the homeserver's own words, instead of at the first send.
func (c *Client) Whoami(ctx context.Context) error {
	var out struct {
		UserID string `json:"user_id"`
	}
	if err := c.do(ctx, http.MethodGet, "/_matrix/client/v3/account/whoami", nil, &out); err != nil {
		return err
	}
	if out.UserID != c.userID {
		return errorfamily.Newf(errorfamily.Rejection, "matrix.token", "matrix: token belongs to %s, config says %s", out.UserID, c.userID)
	}
	return nil
}

// roomFor resolves the DM room for a peer, creating one (is_direct) on
// first contact. Room creation also updates m.direct account data so
// the peer's client renders the room as a DM.
func (c *Client) roomFor(ctx context.Context, peer string) (string, error) {
	c.mu.Lock()
	room, ok := c.dmRooms[peer]
	c.mu.Unlock()
	if ok {
		return room, nil
	}

	var out struct {
		RoomID string `json:"room_id"`
	}
	if err := c.do(ctx, http.MethodPost, "/_matrix/client/v3/createRoom", map[string]any{
		"is_direct": true,
		"invite":    []string{peer},
		"preset":    "private_chat",
	}, &out); err != nil {
		return "", err
	}
	if out.RoomID == "" {
		return "", errorfamily.NewInfrastructure("matrix.create_room", "matrix: homeserver created a room but named none")
	}
	c.mu.Lock()
	c.dmRooms[peer] = out.RoomID
	dm := make(map[string]string, len(c.dmRooms))
	maps.Copy(dm, c.dmRooms)
	c.mu.Unlock()
	if err := c.putDirect(ctx, dm); err != nil {
		// The room exists and works; the account-data nicety is
		// best-effort (logged, never fatal to the send).
		slog.Warn("matrix: record m.direct", "peer", peer, "error", err)
	}
	return out.RoomID, nil
}

// putDirect writes the merged m.direct account data.
func (c *Client) putDirect(ctx context.Context, dm map[string]string) error {
	content := make(map[string][]string, len(dm))
	for peer, room := range dm {
		content[peer] = []string{room}
	}
	return c.do(ctx, http.MethodPut, "/_matrix/client/v3/user/"+url.PathEscape(c.userID)+"/account_data/m.direct", content, nil)
}

// Upload stores one file in the homeserver content repository and
// returns its mxc:// URI.
func (c *Client) Upload(ctx context.Context, name, mimeType string, data []byte) (string, error) { //nolint:branching-flow // foreign matrix upload vocabulary (mxc URIs)
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.homeserver+"/_matrix/media/v3/upload?filename="+url.QueryEscape(name), bytes.NewReader(data))
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "matrix.request", "matrix: build upload %s", name)
	}
	c.authorize(req)
	req.Header.Set("Content-Type", mimeType)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "matrix.upload", "matrix: upload %s", name)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // close-after-use: nothing left to do on failure
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", classify(resp, "matrix.upload")
	}
	var out struct {
		ContentURI string `json:"content_uri"`
	}
	answer, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return "", errorfamily.WrapTransientf(err, "matrix.decode", "matrix: read upload answer")
	}
	if err := json.Unmarshal(answer, &out); err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "matrix.decode", "matrix: decode upload answer")
	}
	if out.ContentURI == "" {
		return "", errorfamily.NewInfrastructure("matrix.upload", "matrix: upload answered no content_uri")
	}
	return out.ContentURI, nil
}

// Download fetches one mxc:// media item and reports its bytes and
// honest Content-Type.
func (c *Client) Download(ctx context.Context, mxc string) ([]byte, string, error) {
	rest, ok := strings.CutPrefix(mxc, "mxc://")
	if !ok {
		return nil, "", errorfamily.Newf(errorfamily.Rejection, "matrix.mxc", "matrix: %q is not an mxc:// URI", mxc)
	}
	server, mediaID, ok := strings.Cut(rest, "/")
	if !ok || server == "" || mediaID == "" {
		return nil, "", errorfamily.Newf(errorfamily.Rejection, "matrix.mxc", "matrix: %q is not an mxc:// URI", mxc)
	}
	path := "/_matrix/client/v1/media/download/" + url.PathEscape(server) + "/" + url.PathEscape(mediaID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.homeserver+path, nil)
	if err != nil {
		return nil, "", errorfamily.WrapInfrastructuref(err, "matrix.request", "matrix: build download %s", mxc)
	}
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", errorfamily.WrapTransientf(err, "matrix.download", "matrix: download %s", mxc)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // close-after-use: nothing left to do on failure
	if resp.StatusCode != http.StatusOK {
		return nil, "", classify(resp, "matrix.download")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, downloadCap+1))
	if err != nil {
		return nil, "", errorfamily.WrapTransientf(err, "matrix.download", "matrix: read %s", mxc)
	}
	if len(data) > downloadCap {
		return nil, "", errorfamily.NewRejection("matrix.download", "matrix: media larger than 25 MiB")
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// sendEvent PUTs one room event with a fresh transaction id: a retried
// PUT with the SAME txn id is answered idempotently by the homeserver.
func (c *Client) sendEvent(ctx context.Context, room, eventType string, content map[string]any) (string, error) {
	txn := fmt.Sprintf("webphone-%d-%d", time.Now().UnixNano(), c.txn.Add(1))
	path := "/_matrix/client/v3/rooms/" + url.PathEscape(room) + "/send/" + url.PathEscape(eventType) + "/" + url.PathEscape(txn)
	var out struct {
		EventID string `json:"event_id"`
	}
	if err := c.do(ctx, http.MethodPut, path, content, &out); err != nil {
		return "", err
	}
	if out.EventID == "" {
		return "", errorfamily.NewInfrastructure("matrix.send", "matrix: homeserver accepted the event but named none")
	}
	return out.EventID, nil
}

// fileMessageType maps a mime type onto the matrix file message types.
func fileMessageType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "m.image"
	}
	return "m.file"
}

// do is the ONE JSON request home: bearer auth, JSON bodies, JSON
// answers, and error classification (auth/permission = Rejection,
// transport + rate limits = Transient — the send-failure arm offers
// retry on the latter only).
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return errorfamily.WrapRejectionf(err, "matrix.encode", "matrix: encode %s body", method)
		}
		payload = encoded
	}
	req, err := http.NewRequestWithContext(ctx, method, c.homeserver+path, bytes.NewReader(payload))
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "matrix.request", "matrix: build %s %s", method, path)
	}
	c.authorize(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return errorfamily.WrapTransientf(err, "matrix.request", "matrix: %s %s", method, path)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:erraudit // close-after-use: nothing left to do on failure
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return classify(resp, "matrix.request")
	}
	if out != nil {
		answer, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return errorfamily.WrapTransientf(err, "matrix.decode", "matrix: read %s %s answer", method, path)
		}
		if err := json.Unmarshal(answer, out); err != nil {
			return errorfamily.WrapInfrastructuref(err, "matrix.decode", "matrix: decode %s %s answer", method, path)
		}
	}
	return nil
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
}

// classify turns a non-2xx homeserver answer into an error family:
// auth/permission problems are Rejections (retrying changes nothing),
// rate limits and server trouble are Transient.
func classify(resp *http.Response, code string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096)) //nolint:erraudit // diagnostic context only; the status carries the verdict
	var shape struct {
		ErrCode string `json:"errcode"`
		Error   string `json:"error"`
	}
	_ = json.Unmarshal(body, &shape) //nolint:erraudit // a non-JSON error body still classifies by status below
	detail := shape.Error
	if detail == "" {
		detail = strconv.Itoa(resp.StatusCode)
	}
	if shape.ErrCode != "" {
		detail = shape.ErrCode + ": " + detail
	}
	// The message rides a variable so the registry scanner reads the
	// CODE argument, not the seam prefix inside the message literal.
	message := "matrix: " + detail
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return errorfamily.NewTransient(code, message)
	}
	return errorfamily.NewRejection(code, message)
}
