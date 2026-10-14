package domain

import (
	"testing"
	"time"
)

// AppendToThread parity pins (SUPERB declarative train C1, 2026-10-14):
// the function replaces the thread-upsert CASE logic; these tables pin
// the historical semantics BEFORE the store rewires onto it.
func newThreadFixture() (Thread, Message) {
	older := time.Unix(1_700_000_000, 0)
	newer := time.Unix(1_700_000_100, 0)
	thread := Thread{
		ID:             GenerateThreadID(),
		Owner:          MustParseExtension("100"),
		Remote:         MustParsePhone("+15550001111"),
		LastActivityAt: older,
		Unread:         2,
		Archived:       true,
	}
	msg := Message{
		ID:        GenerateMessageID(),
		ThreadID:  thread.ID,
		Owner:     thread.Owner,
		Remote:    thread.Remote,
		Direction: DirectionInbound,
		CreatedAt: newer,
	}
	return thread, msg
}

func TestAppendToThreadInbound(t *testing.T) {
	thread, msg := newThreadFixture()

	next := AppendToThread(thread, msg)

	if next.Unread != thread.Unread+1 {
		t.Fatalf("inbound must increment unread: %d -> %d", thread.Unread, next.Unread)
	}
	if next.Archived {
		t.Fatal("inbound must un-archive the thread")
	}
	if !next.LastActivityAt.Equal(msg.CreatedAt) {
		t.Fatalf("last activity must follow the message: %v", next.LastActivityAt)
	}
	if next.ID != thread.ID {
		t.Fatalf("append must never mint a new thread id: %s -> %s", thread.ID, next.ID)
	}
}

func TestAppendToThreadOutbound(t *testing.T) {
	thread, msg := newThreadFixture()
	msg.Direction = DirectionOutbound

	next := AppendToThread(thread, msg)

	if next.Unread != thread.Unread {
		t.Fatalf("outbound must not increment unread: %d -> %d", thread.Unread, next.Unread)
	}
	if !next.Archived {
		t.Fatal("outbound must leave the archived flag untouched")
	}
	if !next.LastActivityAt.Equal(msg.CreatedAt) {
		t.Fatalf("outbound still advances last activity: %v", next.LastActivityAt)
	}
}

func TestAppendToThreadCreation(t *testing.T) {
	_, msg := newThreadFixture()

	inbound := AppendToThread(Thread{}, msg)
	if inbound.ID != msg.ThreadID || inbound.Owner != msg.Owner || inbound.Remote != msg.Remote {
		t.Fatalf("creation must adopt the message identity: %+v", inbound)
	}
	if inbound.Unread != 1 || inbound.Archived {
		t.Fatalf("inbound creation: unread = %d archived = %v, want 1/false", inbound.Unread, inbound.Archived)
	}

	msg.Direction = DirectionOutbound
	outbound := AppendToThread(Thread{}, msg)
	if outbound.Unread != 0 || outbound.Archived {
		t.Fatalf("outbound creation: unread = %d archived = %v, want 0/false", outbound.Unread, outbound.Archived)
	}
}

// The historical upsert overwrites last_activity_at unconditionally —
// even an out-of-order older stamp wins. Pinned so the rewire cannot
// silently "fix" it; changing this is an owner decision.
func TestAppendToThreadOlderStampStillOverwrites(t *testing.T) {
	thread, msg := newThreadFixture()
	msg.CreatedAt = thread.LastActivityAt.Add(-time.Hour)
	msg.Direction = DirectionOutbound

	next := AppendToThread(thread, msg)

	if !next.LastActivityAt.Equal(msg.CreatedAt) {
		t.Fatalf("parity pin: the upsert overwrote with the older stamp; got %v", next.LastActivityAt)
	}
}

// Value semantics: the caller's thread is never mutated.
func TestAppendToThreadDoesNotMutateInput(t *testing.T) {
	thread, msg := newThreadFixture()
	before := thread.Unread
	beforeArchived := thread.Archived

	_ = AppendToThread(thread, msg)

	if thread.Unread != before || thread.Archived != beforeArchived {
		t.Fatal("AppendToThread must not mutate its input")
	}
}
