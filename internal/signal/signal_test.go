package signal

import (
	"bufio"
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// fakeDaemon is a scripted signal-cli: a unix socket answering
// JSON-RPC calls and able to push `receive` notifications.
type fakeDaemon struct {
	mu      sync.Mutex
	ln      net.Listener
	conns   []net.Conn
	sends   []map[string]any
	failRPC bool
}

func startFakeDaemon(t *testing.T) *fakeDaemon {
	t.Helper()
	path := filepath.Join(t.TempDir(), "socket")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	d := &fakeDaemon{ln: ln}
	t.Cleanup(func() {
		_ = ln.Close()
		d.closeConns()
	})
	go d.accept()
	return d
}

func (d *fakeDaemon) accept() {
	for {
		conn, err := d.ln.Accept()
		if err != nil {
			return
		}
		d.mu.Lock()
		d.conns = append(d.conns, conn)
		d.mu.Unlock()
		go d.serve(conn)
	}
}

func (d *fakeDaemon) serve(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 8<<20)
	for scanner.Scan() {
		var req struct {
			ID     string         `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		d.mu.Lock()
		fail := d.failRPC
		d.mu.Unlock()
		if fail {
			_, _ = fmt.Fprintf(conn, `{"jsonrpc":"2.0","error":{"code":-1,"message":"daemon says no"},"id":%q}`+"\n", req.ID)
			continue
		}
		switch req.Method {
		case "getVersion":
			_, _ = fmt.Fprintf(conn, `{"jsonrpc":"2.0","result":{"version":"0.14.9"},"id":%q}`+"\n", req.ID)
		case "send":
			d.mu.Lock()
			d.sends = append(d.sends, req.Params)
			d.mu.Unlock()
			_, _ = fmt.Fprintf(conn, `{"jsonrpc":"2.0","result":{"timestamp":1760000099999},"id":%q}`+"\n", req.ID)
		default:
			_, _ = fmt.Fprintf(conn, `{"jsonrpc":"2.0","error":{"code":-32601,"message":"unknown method"},"id":%q}`+"\n", req.ID)
		}
	}
}

// push writes one unsolicited receive notification to every client.
func (d *fakeDaemon) push(t *testing.T, envelope string) {
	t.Helper()
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, conn := range d.conns {
		if _, err := conn.Write([]byte(`{"jsonrpc":"2.0","method":"receive","params":{"envelope":` + envelope + `}}` + "\n")); err != nil {
			t.Fatalf("push: %v", err)
		}
	}
}

func (d *fakeDaemon) closeConns() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, conn := range d.conns {
		_ = conn.Close()
	}
}

func dialer(path string) *Client {
	c := New(path, "+4915112345678", slog.Default())
	return c
}

func TestPingAndSendOverTheSocket(t *testing.T) {
	d := startFakeDaemon(t)
	c := New(socketPathOf(d), "+4915112345678", slog.Default())
	if err := c.Ping(t.Context()); err != nil {
		t.Fatalf("ping: %v", err)
	}

	att := filepath.Join(t.TempDir(), "pic.png")
	if err := os.WriteFile(att, []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParseAddress(domain.NetworkSignal, "+491599887766"),
		Body:  "hello signal",
		Attachments: []gateway.OutboundAttachment{
			{Name: "pic.png", MimeType: "image/png", Path: att},
		},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if receipt.ProviderRef != "1760000099999" {
		t.Fatalf("receipt must be the send timestamp, got %q", receipt.ProviderRef)
	}
	if receipt.Resolution != gateway.ResolutionDeferred {
		t.Fatal("signal receipts are deferred — delivery arrives as receipt pushes")
	}

	d.mu.Lock()
	sends := append([]map[string]any{}, d.sends...)
	d.mu.Unlock()
	if len(sends) != 1 {
		t.Fatalf("one send must reach the daemon, got %d", len(sends))
	}
	params := sends[0]
	if params["account"] != "+4915112345678" || params["message"] != "hello signal" {
		t.Fatalf("params wrong: %+v", params)
	}
	recipients, _ := params["recipient"].([]any)
	if len(recipients) != 1 || recipients[0] != "+491599887766" {
		t.Fatalf("recipient wrong: %v", params["recipient"])
	}
	uris, _ := params["attachments"].([]any)
	if len(uris) != 1 || !strings.HasPrefix(uris[0].(string), "data:image/png;filename=pic.png;base64,") {
		t.Fatalf("attachment must ride a data URI, got %v", uris)
	}
}

func TestRPCErrorClassifiesRejection(t *testing.T) {
	d := startFakeDaemon(t)
	c := New(socketPathOf(d), "+4915112345678", slog.Default())
	if err := c.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	d.mu.Lock()
	d.failRPC = true
	d.mu.Unlock()
	_, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkSignal, "+491599887766"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	if !strings.Contains(err.Error(), "daemon says no") {
		t.Fatalf("the daemon's words must ride the error, got %v", err)
	}
}

// inSink captures inbound deliveries; inReceipts captures verdicts.
type inSink struct {
	mu  sync.Mutex
	got []domain.InboundMessage
}

func (s *inSink) Receive(_ context.Context, inbound domain.InboundMessage) (domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.got = append(s.got, inbound)
	return domain.Message{}, nil
}

// RemoteRead is the Sink's receipt leg: the signal receiver never emits
// read markers today, so the test fake proves the interface is
// satisfied without inventing protocol behavior.
func (*inSink) RemoteRead(context.Context, string) error { return nil }

type inReceipts struct {
	mu  sync.Mutex
	got []string
}

func (r *inReceipts) DeliveryReceipt(_ context.Context, ref string, _ domain.OutboundStatus, _ string) (domain.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.got = append(r.got, ref)
	return domain.Message{}, nil
}

func TestReceivePushBecomesInboundAndReceipts(t *testing.T) {
	d := startFakeDaemon(t)
	c := New(socketPathOf(d), "+4915112345678", slog.Default())
	got := &inSink{}
	receipts := &inReceipts{}
	r := NewReceiver(c, domain.MustParseExtension("1001"), got, receipts, slog.Default())
	r.client.OnEnvelope(r.handle)
	if err := c.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}

	d.push(t, `{"source":"+491599887766","timestamp":1760000000001,"dataMessage":{"timestamp":1760000000001,"message":"hi from signal"}}`)
	d.push(t, `{"source":"+491599887766","timestamp":1760000000002,"receiptMessage":{"when":1760000000002,"isDelivery":true,"isRead":false,"timestamps":[1760000099999]}}`)

	deadline := time.After(5 * time.Second)
	for {
		got.mu.Lock()
		inCount := len(got.got)
		got.mu.Unlock()
		receipts.mu.Lock()
		refCount := len(receipts.got)
		receipts.mu.Unlock()
		if inCount == 1 && refCount == 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("pushes must arrive: inbound=%d receipts=%d", inCount, refCount)
		case <-time.After(5 * time.Millisecond):
		}
	}

	got.mu.Lock()
	in := got.got[0]
	got.mu.Unlock()
	if in.From.String() != "+491599887766" || in.From.Network() != domain.NetworkSignal || in.Body != "hi from signal" {
		t.Fatalf("inbound wrong: %+v", in)
	}
	if in.Owner.String() != "1001" || in.ReceivedAt.UnixMilli() != 1760000000001 {
		t.Fatalf("owner/ts wrong: %s %v", in.Owner, in.ReceivedAt)
	}
	receipts.mu.Lock()
	ref := receipts.got[0]
	receipts.mu.Unlock()
	if ref != "1760000099999" {
		t.Fatalf("receipt must target the send timestamp ref, got %q", ref)
	}
}

func TestGroupMessagesAreSkippedLoudly(t *testing.T) {
	d := startFakeDaemon(t)
	c := New(socketPathOf(d), "+4915112345678", slog.Default())
	got := &inSink{}
	r := NewReceiver(c, domain.MustParseExtension("1001"), got, &inReceipts{}, slog.Default())
	r.client.OnEnvelope(r.handle)
	if err := c.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	d.push(t, `{"source":"+491599887766","timestamp":1,"dataMessage":{"timestamp":1,"message":"group hi","groupInfo":{"groupId":"abc="}}}`)
	deadline := time.After(2 * time.Second)
	for {
		got.mu.Lock()
		n := len(got.got)
		got.mu.Unlock()
		if n == 0 {
			select {
			case <-deadline:
				return // skipped by design — nothing may arrive
			case <-time.After(5 * time.Millisecond):
			}
			continue
		}
		t.Fatal("group messages must not become threads (v1)")
	}
}

func socketPathOf(d *fakeDaemon) string {
	return d.ln.Addr().String()
}
