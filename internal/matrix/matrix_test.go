package matrix

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// homeserver is a scripted Matrix CS-API fake: it records requests and
// answers from per-endpoint handlers the test installs.
type homeserver struct {
	mu        sync.Mutex
	sends     []sentEvent
	creations int
	direct    map[string][]string
	t         *testing.T
}

type sentEvent struct {
	Room    string
	Type    string
	Content map[string]any
}

func newHomeserver(t *testing.T) (*homeserver, *httptest.Server) {
	t.Helper()
	hs := &homeserver{t: t, direct: map[string][]string{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			http.Error(w, `{"errcode":"M_UNKNOWN_TOKEN","error":"bad token"}`, http.StatusUnauthorized)
			return
		}
		hs.mu.Lock()
		defer hs.mu.Unlock()
		switch {
		case r.URL.Path == "/_matrix/client/v3/account/whoami":
			_, _ = fmt.Fprint(w, `{"user_id":"@webphone:example.org"}`)
		case r.URL.Path == "/_matrix/client/v3/createRoom" && r.Method == http.MethodPost:
			hs.creations++
			_, _ = fmt.Fprint(w, `{"room_id":"!newroom:example.org"}`)
		case strings.HasPrefix(r.URL.Path, "/_matrix/client/v3/user/") && strings.HasSuffix(r.URL.Path, "/account_data/m.direct"):
			raw, _ := io.ReadAll(r.Body)
			var body map[string][]string
			if err := json.Unmarshal(raw, &body); err != nil {
				http.Error(w, "bad body", 400)
				return
			}
			hs.direct = body
			w.WriteHeader(http.StatusOK)
		case strings.HasPrefix(r.URL.Path, "/_matrix/client/v3/rooms/") && strings.Contains(r.URL.Path, "/send/"):
			parts := strings.Split(r.URL.Path, "/")
			room, eventType := parts[5], parts[7]
			raw, _ := io.ReadAll(r.Body)
			var content map[string]any
			if err := json.Unmarshal(raw, &content); err != nil {
				http.Error(w, "bad body", 400)
				return
			}
			hs.sends = append(hs.sends, sentEvent{Room: room, Type: eventType, Content: content})
			_, _ = fmt.Fprint(w, fmt.Sprintf(`{"event_id":"$ev%d"}`, len(hs.sends)))
		case r.URL.Path == "/_matrix/media/v3/upload":
			data, _ := io.ReadAll(r.Body)
			if len(data) == 0 {
				http.Error(w, "empty", 400)
				return
			}
			fmt.Fprint(w, `{"content_uri":"mxc://example.org/media1"}`)
		case strings.HasPrefix(r.URL.Path, "/_matrix/client/v1/media/download/"):
			w.Header().Set("Content-Type", "image/png")
			fmt.Fprint(w, "pngbytes")
		case r.URL.Path == "/_matrix/client/v3/sync":
			fmt.Fprint(w, `{"next_batch":"s1"}`)
		default:
			http.Error(w, `{"errcode":"M_UNRECOGNIZED","error":"nope"}`, http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return hs, server
}

func newTestClient(t *testing.T) (*homeserver, *Client) {
	t.Helper()
	hs, server := newHomeserver(t)
	return hs, New(server.URL, "@webphone:example.org", "tok")
}

func TestWhoamiVerifiesTheTokenAndIdentity(t *testing.T) {
	_, c := newTestClient(t)
	if err := c.Whoami(t.Context()); err != nil {
		t.Fatalf("whoami: %v", err)
	}
	wrong := New(c.homeserver, "@other:example.org", "tok")
	if err := wrong.Whoami(t.Context()); err == nil {
		t.Fatal("a token for another user must fail the boot check")
	}
}

func TestSendCreatesTheDMAndSendsText(t *testing.T) {
	hs, c := newTestClient(t)
	receipt, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		Owner: domain.MustParseExtension("1001"),
		To:    domain.MustParseAddress(domain.NetworkMatrix, "@peer:example.org"),
		Body:  "hello matrix",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if receipt.ProviderRef == "" || receipt.Resolution != gateway.ResolutionImmediate {
		t.Fatalf("receipt must be an immediate event id, got %+v", receipt)
	}
	if hs.creations != 1 {
		t.Fatalf("unknown peer must create exactly one DM room, got %d", hs.creations)
	}
	if rooms := hs.direct["@peer:example.org"]; len(rooms) != 1 || rooms[0] != "!newroom:example.org" {
		t.Fatalf("m.direct must record the new room, got %v", hs.direct)
	}
	if len(hs.sends) != 1 || hs.sends[0].Content["body"] != "hello matrix" {
		t.Fatalf("text event wrong: %+v", hs.sends)
	}
}

func TestSendUploadsAttachmentsAsMediaEvents(t *testing.T) {
	hs, c := newTestClient(t)
	att := t.TempDir() + "/pic.png"
	if err := os.WriteFile(att, []byte("filebytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To:          domain.MustParseAddress(domain.NetworkMatrix, "@peer:example.org"),
		Body:        "",
		Attachments: []gateway.OutboundAttachment{{Name: "pic.png", MimeType: "image/png", Path: att}},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if receipt.ProviderRef == "" {
		t.Fatal("attachment-only send still needs its event id receipt")
	}
	if len(hs.sends) != 1 || hs.sends[0].Type != "m.image" {
		t.Fatalf("image attachment must ride an m.image event, got %+v", hs.sends)
	}
	if hs.sends[0].Content["url"] != "mxc://example.org/media1" {
		t.Fatalf("media event must carry the upload uri, got %+v", hs.sends[0].Content)
	}
}

func TestSendRefusesForeignNetworks(t *testing.T) {
	_, c := newTestClient(t)
	_, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkSMS, "+441632960961"),
	})
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "matrix.network")
}

func TestHomeserverVerdictsClassify(t *testing.T) {
	rateLimited := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"errcode":"M_LIMIT_EXCEEDED","error":"slow down"}`)
	}))
	defer rateLimited.Close()
	c := New(rateLimited.URL, "@w:example.org", "tok")
	_, err := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkMatrix, "@peer:example.org"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	if !strings.Contains(err.Error(), "M_LIMIT_EXCEEDED") {
		t.Fatalf("the errcode must ride the message for the operator, got %v", err)
	}
}

// sink captures inbound deliveries.
type sink struct {
	mu       sync.Mutex
	received []domain.InboundMessage
}

func (s *sink) Receive(_ context.Context, inbound domain.InboundMessage) (domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.received = append(s.received, inbound)
	return domain.Message{}, nil
}

// cursor is an in-memory Cursor.
type cursor struct {
	mu    sync.Mutex
	token string
	saved []string
}

func (c *cursor) Load(context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.token, nil
}

func (c *cursor) Save(_ context.Context, token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.saved = append(c.saved, token)
	return nil
}

// saveCount reports the persisted-cursor count race-safely (the Run
// goroutine writes while the test polls).
func (c *cursor) saveCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.saved)
}

func TestIngestDeliversDMTextAndSkipsEncrypted(t *testing.T) {
	hs, server := newHomeserver(t)
	_ = hs
	c := New(server.URL, "@webphone:example.org", "tok")
	cur := &cursor{}
	got := &sink{}
	r := NewReceiver(c, domain.MustParseExtension("1001"), got, cur, slog.Default())

	resp := syncResponse{}
	payload := `{
		"next_batch": "s2",
		"account_data": {"events": [
			{"type": "m.direct", "content": {"@peer:example.org": ["!dm1:example.org"]}}
		]},
		"rooms": {"join": {"!dm1:example.org": {"timeline": {"events": [
			{"type": "m.room.message", "sender": "@webphone:example.org", "event_id": "$own", "origin_server_ts": 1, "content": {"msgtype": "m.text", "body": "echo of my own send"}},
			{"type": "m.room.message", "sender": "@peer:example.org", "event_id": "$e1", "origin_server_ts": 1760000000123, "content": {"msgtype": "m.text", "body": "hi from matrix"}},
			{"type": "m.room.encrypted", "sender": "@peer:example.org", "event_id": "$e2", "origin_server_ts": 1760000000124, "content": {"algorithm": "m.megolm.v1.aes-sha2"}}
		]}}}}
	}`
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatal(err)
	}
	r.ingest(t.Context(), resp)

	if len(got.received) != 1 {
		t.Fatalf("exactly the peer text must land (own echo + encrypted skipped), got %d", len(got.received))
	}
	in := got.received[0]
	if in.From.String() != "@peer:example.org" || in.From.Network() != domain.NetworkMatrix {
		t.Fatalf("from wrong: %v", in.From)
	}
	if in.Body != "hi from matrix" || in.ReceivedAt.UnixMilli() != 1760000000123 {
		t.Fatalf("body/ts wrong: %q %v", in.Body, in.ReceivedAt)
	}
	if in.Owner.String() != "1001" {
		t.Fatalf("owner wrong: %s", in.Owner)
	}
}

func TestIngestDownloadsDMMedia(t *testing.T) {
	_, server := newHomeserver(t)
	c := New(server.URL, "@webphone:example.org", "tok")
	got := &sink{}
	r := NewReceiver(c, domain.MustParseExtension("1001"), got, &cursor{}, slog.Default())

	payload := `{
		"next_batch": "s3",
		"account_data": {"events": [{"type": "m.direct", "content": {"@peer:example.org": ["!dm1:example.org"]}}]},
		"rooms": {"join": {"!dm1:example.org": {"timeline": {"events": [
			{"type": "m.room.message", "sender": "@peer:example.org", "event_id": "$img", "origin_server_ts": 1760000001000,
			 "content": {"msgtype": "m.image", "body": "photo.png", "url": "mxc://example.org/media1", "info": {"mimetype": "image/png"}}}
		]}}}}
	}`
	var resp syncResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatal(err)
	}
	r.ingest(t.Context(), resp)

	if len(got.received) != 1 || len(got.received[0].Attachments) != 1 {
		t.Fatalf("media must arrive as one attachment, got %+v", got.received)
	}
	att := got.received[0].Attachments[0]
	if att.Name != "photo.png" || att.MimeType != "image/png" || string(att.Bytes) != "pngbytes" {
		t.Fatalf("attachment wrong: %+v", att)
	}
}

func TestRunPersistsCursorAndExitsOnCancel(t *testing.T) {
	_, server := newHomeserver(t)
	c := New(server.URL, "@webphone:example.org", "tok")
	cur := &cursor{}
	r := NewReceiver(c, domain.MustParseExtension("1001"), &sink{}, cur, slog.Default())

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		_ = r.Run(ctx)
		close(done)
	}()
	deadline := time.After(5 * time.Second)
	for cur.saveCount() == 0 {
		select {
		case <-deadline:
			t.Fatal("receiver never persisted a cursor")
		case <-time.After(5 * time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run must exit on ctx cancel")
	}
}
