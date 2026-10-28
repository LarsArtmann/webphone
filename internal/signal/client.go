// Package signal is the OPTIONAL chat-network seam to a signal-cli
// daemon: JSON-RPC 2.0 over its unix socket (signal-cli daemon
// --socket). One persistent connection carries BOTH directions —
// method calls out, unsolicited `receive` notifications in — so the
// client is a demuxer: responses route to waiters by id, envelopes
// route to the receiver. Filesystem permissions on the socket are the
// trust boundary, the same loopback posture as the gateway.
package signal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// compile-time conformance: the client IS the message gateway of the
// signal lane.
var _ gateway.MessageGateway = (*Client)(nil)

// Client speaks JSON-RPC 2.0 to one signal-cli socket.
// brandid-lint:ignore(BD003) nextID is the JSON-RPC request counter: a sequence number, not an identity (strong-id ruling)
type Client struct {
	socketPath string
	account    string
	dial       func() (net.Conn, error)

	mu      sync.Mutex
	conn    net.Conn
	pending map[string]chan rpcResult
	nextID  uint64

	onEnvelope func(Envelope)

	log *slog.Logger
}

// Envelope is the slice of signal-cli's `receive` notification this
// seam consumes: data messages (text in, attachments noted but NOT
// fetched — they live in signal-cli's data dir, a documented v1 limit)
// and receipt messages (delivery/read acknowledgements of OUR send
// timestamps, which are exactly our provider refs).
type Envelope struct {
	Source      string `json:"source"`
	SourceUuid  string `json:"sourceUuid"`
	Timestamp   int64  `json:"timestamp"`
	DataMessage *struct {
		Message   string `json:"message"`
		Timestamp int64  `json:"timestamp"`
		GroupInfo *struct {
			GroupId string `json:"groupId"`
		} `json:"groupInfo"`
	} `json:"dataMessage"`
	ReceiptMessage *struct {
		When       int64   `json:"when"`
		IsDelivery bool    `json:"isDelivery"`
		IsRead     bool    `json:"isRead"`
		Timestamps []int64 `json:"timestamps"`
	} `json:"receiptMessage"`
}

// New builds the client. dial is injectable so tests drive a fake
// socket; nil dials the configured unix socket path.
func New(socketPath, account string, log *slog.Logger) *Client { //nolint:branching-flow // foreign signal-cli vocabulary, config-validated
	if log == nil {
		log = slog.Default()
	}
	c := &Client{
		socketPath: socketPath,
		account:    account,
		pending:    map[string]chan rpcResult{},
		log:        log,
	}
	c.dial = func() (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(context.Background(), "unix", socketPath)
	}
	return c
}

// OnEnvelope registers the push handler (the receiver). Call it before
// Connect.
func (c *Client) OnEnvelope(fn func(Envelope)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEnvelope = fn
}

// SendMessage implements gateway.MessageGateway: `send` with
// recipients and base64 data-URI attachments. The receipt is the send
// timestamp — DEFERRED: signal-cli accepted the job; delivery and read
// acknowledgements arrive later as receiptMessage notifications (the
// receiver feeds them to the delivery pipeline).
func (c *Client) SendMessage(ctx context.Context, msg gateway.OutboundMessage) (gateway.Receipt, error) {
	if msg.To.Network() != domain.NetworkSignal {
		return gateway.Receipt{}, errorfamily.Newf(errorfamily.Rejection, "signal.network", "signal: destination %s is not a signal address", msg.To)
	}
	params := map[string]any{
		"account":   c.account,
		"recipient": []string{msg.To.String()},
		"message":   msg.Body,
	}
	if len(msg.Attachments) > 0 {
		uris := make([]string, 0, len(msg.Attachments))
		for _, att := range msg.Attachments {
			uri, err := dataURI(att)
			if err != nil {
				return gateway.Receipt{}, err
			}
			uris = append(uris, uri)
		}
		params["attachments"] = uris
	}
	var result struct {
		Timestamp int64 `json:"timestamp"`
	}
	if err := c.call(ctx, "send", params, &result); err != nil {
		return gateway.Receipt{}, err
	}
	if result.Timestamp == 0 {
		return gateway.Receipt{}, errorfamily.NewInfrastructure("signal.send", "signal: send answered no timestamp")
	}
	return gateway.Receipt{
		ProviderRef: strconv.FormatInt(result.Timestamp, 10),
		Resolution:  gateway.ResolutionDeferred,
	}, nil
}

// Connect dials the socket and starts the read loop. Idempotent: an
// existing connection is kept.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connectLocked(ctx)
}

func (c *Client) connectLocked(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}
	conn, err := c.dial()
	if err != nil {
		return errorfamily.WrapTransientf(err, "signal.connect", "signal: dial %s", c.socketPath)
	}
	c.conn = conn
	go c.readLoop(conn)
	return nil
}

// Ping verifies the daemon answers (`getVersion`): the boot check the
// wiring uses to fail fast on a dead socket.
func (c *Client) Ping(ctx context.Context) error {
	var version struct {
		Version string `json:"version"`
	}
	return c.call(ctx, "getVersion", nil, &version)
}

// call performs one JSON-RPC request on the shared connection,
// reconnecting once when the stored connection died.
func (c *Client) call(ctx context.Context, method string, params any, out any) error {
	id := strconv.FormatUint(c.incID(), 10)
	ch := make(chan rpcResult, 1)

	c.mu.Lock()
	if err := c.connectLocked(ctx); err != nil {
		c.mu.Unlock()
		return err
	}
	c.pending[id] = ch
	req := map[string]any{"jsonrpc": "2.0", "method": method, "id": id}
	if params != nil {
		req["params"] = params
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		delete(c.pending, id)
		c.mu.Unlock()
		return errorfamily.WrapRejectionf(err, "signal.encode", "signal: encode %s", method)
	}
	err = c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		delete(c.pending, id)
		c.mu.Unlock()
		return errorfamily.WrapInfrastructuref(err, "signal.write", "signal: write deadline for %s", method)
	}
	_, err = c.conn.Write(append(encoded, '\n'))
	if err != nil {
		delete(c.pending, id)
		c.dropConnLocked(err)
		conn := c.conn
		c.mu.Unlock()
		_ = conn.Close() //nolint:erraudit // close-after-use: the connection is already dead
		return errorfamily.WrapTransientf(err, "signal.write", "signal: write %s", method)
	}
	c.mu.Unlock()

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return errorfamily.WrapTransientf(ctx.Err(), "signal.timeout", "signal: %s canceled", method)
	case res := <-ch:
		if res.err != nil {
			return res.err
		}
		if out == nil {
			return nil
		}
		if err := json.Unmarshal([]byte(res.result), out); err != nil {
			return errorfamily.WrapInfrastructuref(err, "signal.decode", "signal: decode %s answer", method)
		}
		return nil
	}
}

// readLoop demuxes the shared connection until it dies: answers route
// to waiters by id, `receive` notifications route to the envelope
// handler.
func (c *Client) readLoop(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 4<<20)
	// Goroutine containment: a panicking dispatch must not kill the
	// process; the deferred cleanup still drops the connection so the
	// next call reconnects.
	defer func() {
		if rec := recover(); rec != nil {
			c.log.Error("signal: read loop PANIC recovered; connection dropped, reconnect on next call", "panic", rec)
		}
		c.mu.Lock()
		c.dropConnLocked(scanner.Err())
		c.mu.Unlock()
	}()
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var msg struct {
			ID     string         `json:"id"`
			Result jsontext.Value `json:"result"`
			Error  *struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
			Method string         `json:"method"`
			Params jsontext.Value `json:"params"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			c.log.Warn("signal: undecodable line", "error", err)
			continue
		}
		switch {
		case msg.Method == "receive":
			c.dispatch(msg.Params)
		case msg.ID != "":
			c.resolve(msg.ID, msg.Result, msg.Error)
		default:
			c.log.Debug("signal: unhandled notification", "method", msg.Method)
		}
	}
}

// dispatch unpacks one receive notification into an Envelope.
func (c *Client) dispatch(params jsontext.Value) {
	c.mu.Lock()
	handler := c.onEnvelope
	c.mu.Unlock()
	if handler == nil {
		return
	}
	var wrapped struct {
		Envelope Envelope `json:"envelope"`
		Params   *struct {
			Envelope Envelope `json:"envelope"`
		} `json:"params"`
	}
	if err := json.Unmarshal(params, &wrapped); err != nil {
		c.log.Warn("signal: undecodable envelope", "error", err)
		return
	}
	env := wrapped.Envelope
	if wrapped.Params != nil && env.Timestamp == 0 {
		env = wrapped.Params.Envelope
	}
	handler(env)
}

// brandid-lint:ignore(BD003) JSON-RPC correlation: the id keys the in-flight pending map, echoed verbatim (strong-id ruling)
func (c *Client) resolve(id string, result jsontext.Value, rpcErr *struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}) {
	c.mu.Lock()
	ch, ok := c.pending[id]
	if ok {
		delete(c.pending, id)
	}
	c.mu.Unlock()
	if !ok {
		return
	}
	if rpcErr != nil {
		//nolint:branching-flow // ch is buffered(1) and never closed — the send can neither block nor panic
		ch <- rpcResult{err: errorfamily.Newf(errorfamily.Rejection, "signal.rpc", "signal: %s", rpcErr.Message)}
		return
	}
	//nolint:branching-flow // ch is buffered(1) and never closed — the send can neither block nor panic
	ch <- rpcResult{result: result}
}

// dropConnLocked forgets a dead connection (mu held); the next call
// reconnects.
func (c *Client) dropConnLocked(cause error) {
	if c.conn != nil {
		if cause != nil {
			c.log.Warn("signal: connection lost", "error", cause)
		}
		c.conn = nil
	}
}

func (c *Client) incID() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	return c.nextID
}

type rpcResult struct {
	result jsontext.Value
	err    error
}

// dataURI reads one stored attachment into signal-cli's inline
// attachment form: data:<mime>;filename=<name>;base64,<payload>.
func dataURI(att gateway.OutboundAttachment) (string, error) {
	data, err := os.ReadFile(att.Path)
	if err != nil {
		return "", errorfamily.WrapInfrastructuref(err, "signal.attachment_read", "signal: read attachment %s", att.Path)
	}
	mime := att.MimeType
	if mime == "" {
		mime = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;filename=%s;base64,%s", mime, att.Name, base64.StdEncoding.EncodeToString(data)), nil
}

// CloseConn tears the connection down (shutdown path).
func (c *Client) CloseConn() {
	c.mu.Lock()
	conn := c.conn
	c.dropConnLocked(nil)
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close() //nolint:erraudit // close-after-use: shutdown path
	}
}

// ensure the interface is used (io kept for future streaming reads)
var _ io.Closer = (*Client)(nil)

// Close implements io.Closer over CloseConn.
func (c *Client) Close() error {
	c.CloseConn()
	return nil
}
