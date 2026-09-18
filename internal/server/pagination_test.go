package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func TestThreadPaginationLoadsOlderPages(t *testing.T) {
	server := newTestServer(t)
	c := signIn(t, server)

	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	ctx := context.Background()
	base := time.Now().Add(-time.Hour)

	threadID, err := server.messages.FindThread(ctx, owner, remote, base)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 205 {
		if err := server.messages.AppendMessage(ctx, domain.Message{
			ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
			Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
			Body: fmt.Sprintf("msg-%03d", i), CreatedAt: base.Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatal(err)
		}
	}

	path := "/partials/messages/" + threadID.String()

	// Page 0: the newest 200 (msg-005..msg-204) plus the load-older affordance.
	_, body := c.do(http.MethodGet, path, nil, "")
	page := string(body)
	if !strings.Contains(page, "Load older messages") {
		t.Error("page 0 must offer older messages")
	}
	if !strings.Contains(page, "msg-005") || !strings.Contains(page, "msg-204") {
		t.Error("page 0 must show the newest window")
	}
	if strings.Contains(page, "msg-004") {
		t.Error("page 0 must not leak older messages")
	}

	// Page 1: the remaining 5 oldest, no further pages.
	_, body = c.do(http.MethodGet, path+"?older=1", nil, "")
	page = string(body)
	if strings.Contains(page, "Load older messages") {
		t.Error("page 1 has no older messages left")
	}
	for _, want := range []string{"msg-000", "msg-004"} {
		if !strings.Contains(page, want) {
			t.Errorf("page 1 missing %q", want)
		}
	}
	if strings.Contains(page, "msg-005") {
		t.Error("page 1 must not repeat the newest window")
	}

	// A nonsense page number degrades to page 0.
	_, body = c.do(http.MethodGet, path+"?older=-3", nil, "")
	if !strings.Contains(string(body), "msg-204") {
		t.Error("invalid page must fall back to the newest window")
	}
}
