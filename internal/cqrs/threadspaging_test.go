// Thread paging battery (SUPERB Phase 5): the transcript window
// contract past the engine scan's default ceiling — page 0 is the
// newest window, hasMore offers older pages, and every row counts.
package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestThreadPagesBeyondTheEngineCeiling pins the paging contract past the
// engine scan's default 100-row ceiling: a thread with more messages than
// one bounded scan returns must still page newest-window-first and count
// every row (the A34 regression — the default ceiling silently truncated
// threadMessages to the newest 100 and collapsed hasMore).
func TestThreadPagesBeyondTheEngineCeiling(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")
	id, err := backend.FindThread(ctx, owner, remote, time.Now())
	if err != nil {
		t.Fatalf("FindThread: %v", err)
	}
	base := time.Now().Add(-time.Hour)
	for i := range 205 {
		msg := domain.NewInboundMessage(domain.GenerateMessageID(), id, owner, remote,
			fmt.Sprintf("msg-%03d", i), 0, base.Add(time.Duration(i)*time.Second))
		if err := backend.AppendMessage(ctx, msg); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	const pageSize = 200
	page0, hasMore, err := backend.ListMessagesPage(ctx, owner, id, 0, pageSize)
	if err != nil {
		t.Fatalf("page 0: %v", err)
	}
	if len(page0) != pageSize || !hasMore {
		t.Fatalf("page 0 = %d msgs, hasMore %v; want %d, true", len(page0), hasMore, pageSize)
	}
	if page0[0].Body != "msg-005" || page0[pageSize-1].Body != "msg-204" {
		t.Fatalf("page 0 window = %s..%s, want msg-005..msg-204", page0[0].Body, page0[pageSize-1].Body)
	}
	page1, hasMore, err := backend.ListMessagesPage(ctx, owner, id, 1, pageSize)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(page1) != 5 || hasMore {
		t.Fatalf("page 1 = %d msgs, hasMore %v; want 5, false", len(page1), hasMore)
	}
	if page1[0].Body != "msg-000" || page1[4].Body != "msg-004" {
		t.Fatalf("page 1 window = %s..%s, want msg-000..msg-004", page1[0].Body, page1[4].Body)
	}
	all, err := backend.ListMessages(ctx, owner, id, 0)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(all) != 205 {
		t.Fatalf("unbounded transcript = %d msgs, want 205", len(all))
	}
	total, err := backend.CountMessages(ctx)
	if err != nil {
		t.Fatalf("CountMessages: %v", err)
	}
	if total != 205 {
		t.Fatalf("CountMessages = %d, want 205", total)
	}
	threadsTotal, err := backend.CountThreads(ctx)
	if err != nil {
		t.Fatalf("CountThreads: %v", err)
	}
	if threadsTotal != 1 {
		t.Fatalf("CountThreads = %d, want 1", threadsTotal)
	}
}
