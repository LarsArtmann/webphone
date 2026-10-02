package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// seedThread creates a thread with one outbound message and returns its
// id; appendInbound/appendOutbound grow it for ordering and
// auto-unarchive scenarios.
func seedThread(t *testing.T, messages *Messages, owner domain.Extension, remote domain.Phone) domain.ThreadID {
	t.Helper()
	ctx := context.Background()
	threadID, err := messages.FindThread(ctx, owner, remote, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "seed", Status: domain.StatusSent, CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	return threadID
}

func appendMsg(t *testing.T, messages *Messages, owner domain.Extension, remote domain.Phone, threadID domain.ThreadID, direction domain.Direction, body string) {
	t.Helper()
	if err := messages.AppendMessage(context.Background(), domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: direction, Channel: domain.ChannelSMS,
		Body: body, Status: domain.StatusSent, CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestThreadFlagRoundTrips(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	threadID := seedThread(t, messages, owner, remote)

	for _, flag := range []ThreadFlag{FlagPinned, FlagArchived, FlagMuted} {
		if err := messages.SetThreadFlag(ctx, owner, threadID, flag, true); err != nil {
			t.Fatalf("set %s: %v", flag, err)
		}
	}
	thread, err := messages.GetThread(ctx, owner, threadID)
	if err != nil {
		t.Fatal(err)
	}
	if !thread.Pinned || !thread.Archived || !thread.Muted {
		t.Fatalf("flags not persisted: %+v", thread)
	}

	for _, flag := range []ThreadFlag{FlagPinned, FlagArchived, FlagMuted} {
		if err := messages.SetThreadFlag(ctx, owner, threadID, flag, false); err != nil {
			t.Fatalf("clear %s: %v", flag, err)
		}
	}
	thread, _ = messages.GetThread(ctx, owner, threadID)
	if thread.Pinned || thread.Archived || thread.Muted {
		t.Fatalf("flags not cleared: %+v", thread)
	}

	// Owner scoping and misses are ErrNotFound, not silent success.
	other := domain.MustParseExtension("1002")
	if err := messages.SetThreadFlag(ctx, other, threadID, FlagPinned, true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign owner flag: got %v, want ErrNotFound", err)
	}
	missing := domain.GenerateThreadID()
	if err := messages.SetThreadFlag(ctx, owner, missing, FlagPinned, true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing thread flag: got %v, want ErrNotFound", err)
	}

	// An unknown flag is a programming error surfaced as a rejection.
	if err := messages.SetThreadFlag(ctx, owner, threadID, ThreadFlag("nope"), true); err == nil ||
		!strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown flag: got %v, want unknown-flag rejection", err)
	}
}

func TestArchiveExclusionAndAutoUnarchive(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remoteA := domain.MustParsePhone("+441632960961")
	remoteB := domain.MustParsePhone("+441632960962")
	threadA := seedThread(t, messages, owner, remoteA)
	seedThread(t, messages, owner, remoteB)

	if err := messages.SetThreadFlag(ctx, owner, threadA, FlagArchived, true); err != nil {
		t.Fatal(err)
	}

	active, err := messages.ListThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].Thread.Remote != remoteB {
		t.Fatalf("active list must exclude the archived thread: %+v", active)
	}
	archived, err := messages.ListArchivedThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(archived) != 1 || archived[0].Thread.ID != threadA {
		t.Fatalf("archived list wrong: %+v", archived)
	}
	if count, err := messages.CountArchived(ctx, owner); err != nil || count != 1 {
		t.Fatalf("CountArchived: got %d err %v, want 1", count, err)
	}
	if hits, err := messages.SearchThreads(ctx, owner, "seed"); err != nil {
		t.Fatal(err)
	} else if len(hits) != 1 || hits[0].Thread.ID != threadIDFor(t, messages, owner, remoteB) {
		t.Fatalf("search must not resurrect archived threads: %+v", hits)
	}

	// D7: an INBOUND message auto-unarchives — the conversation is live
	// again; the count follows.
	appendMsg(t, messages, owner, remoteA, threadA, domain.DirectionInbound, "hello again")
	if count, err := messages.CountArchived(ctx, owner); err != nil || count != 0 {
		t.Fatalf("inbound must unarchive: count %d err %v", count, err)
	}
	archived, _ = messages.ListArchivedThreads(ctx, owner)
	if len(archived) != 0 {
		t.Fatalf("archived list after inbound: %+v", archived)
	}

	// Outbound never unarchives (you know you replied).
	if err := messages.SetThreadFlag(ctx, owner, threadA, FlagArchived, true); err != nil {
		t.Fatal(err)
	}
	appendMsg(t, messages, owner, remoteA, threadA, domain.DirectionOutbound, "reply while filed")
	if count, _ := messages.CountArchived(ctx, owner); count != 1 {
		t.Fatalf("outbound must keep the thread archived: count %d", count)
	}
}

// threadIDFor resolves the thread id for a remote (the archived-thread
// search assertion needs identity, not just count).
func threadIDFor(t *testing.T, messages *Messages, owner domain.Extension, remote domain.Phone) domain.ThreadID {
	t.Helper()
	threads, err := messages.ListThreads(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
	}
	for _, summary := range threads {
		if summary.Thread.Remote == remote {
			return summary.Thread.ID
		}
	}
	t.Fatalf("no thread for %s", remote)
	return domain.ThreadID{}
}

func TestPinnedOrdering(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remoteOld := domain.MustParsePhone("+441632960961")
	remoteNew := domain.MustParsePhone("+441632960962")
	oldID := seedThread(t, messages, owner, remoteOld)
	seedThread(t, messages, owner, remoteNew)

	// Pin the OLDER thread: pins jump the queue (D5, sort-first) — the
	// pre-pin order is deliberately unasserted (two seeds in the same
	// second tie on last_activity_at; only the pin flip is this test's
	// subject).
	if err := messages.SetThreadFlag(ctx, owner, oldID, FlagPinned, true); err != nil {
		t.Fatal(err)
	}
	threads, _ := messages.ListThreads(ctx, owner)
	if threads[0].Thread.Remote != remoteOld || !threads[0].Thread.Pinned {
		t.Fatalf("pinned thread must sort first: %+v", threads)
	}
	// Search follows the same order.
	hits, _ := messages.SearchThreads(ctx, owner, "seed")
	if hits[0].Thread.Remote != remoteOld {
		t.Fatalf("search must keep pinned first: %+v", hits)
	}
	// And the summary row carries the pin for the row's rendering.
	if !threads[0].Thread.Pinned {
		t.Fatal("summary must expose Pinned")
	}
}

// TestSummaryFlagsAndAttachmentPreview pins the summary-row surface
// M21/M22 added: the flag columns flow into every ThreadSummary and the
// last message's attachment flips LastAttachment (the 📎 preview).
func TestSummaryFlagsAndAttachmentPreview(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	threadID := seedThread(t, messages, owner, remote)

	if err := messages.SetThreadFlag(ctx, owner, threadID, FlagPinned, true); err != nil {
		t.Fatal(err)
	}
	withAttachment := domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelMMS,
		CreatedAt: time.Now(),
		Attachments: []domain.Attachment{{
			ID: domain.GenerateAttachmentID(), Name: "pic.png",
			MimeType: "image/png", SizeBytes: 2048, Path: t.TempDir() + "/pic.png",
		}},
	}
	if err := messages.AppendMessage(ctx, withAttachment); err != nil {
		t.Fatal(err)
	}

	threads, err := messages.ListThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 1 {
		t.Fatalf("threads: %d", len(threads))
	}
	summary := threads[0]
	if !summary.Thread.Pinned {
		t.Error("summary must carry Pinned")
	}
	if !summary.LastAttachment {
		t.Error("summary must carry LastAttachment for an MMS tail (the 📎 preview)")
	}
	if summary.LastDirection != domain.DirectionInbound {
		t.Errorf("last direction: %s", summary.LastDirection)
	}
}

func TestSnippetsCRUDAndCap(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	snippets := NewSnippets(db)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	if list, err := snippets.List(ctx, owner); err != nil || len(list) != 0 {
		t.Fatalf("fresh list: %v %+v", err, list)
	}

	first := domain.Snippet{ID: domain.GenerateSnippetID(), Owner: owner, Body: "Thanks, on it!", Quick: true, CreatedAt: time.Now()}
	if err := snippets.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	second := domain.Snippet{ID: domain.GenerateSnippetID(), Owner: owner, Body: "Please call back later", CreatedAt: time.Now().Add(time.Second)}
	if err := snippets.Save(ctx, second); err != nil {
		t.Fatal(err)
	}
	list, err := snippets.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != first.ID || !list[0].Quick || list[1].Quick {
		t.Fatalf("list order/quick wrong: %+v", list)
	}

	// Replace-by-id edits in place (body + quick flip) without consuming
	// cap or changing order.
	edited := first
	edited.Body = "Thanks — updated"
	edited.Quick = false
	if err := snippets.Save(ctx, edited); err != nil {
		t.Fatal(err)
	}
	list, _ = snippets.List(ctx, owner)
	if len(list) != 2 || list[0].Body != "Thanks — updated" || list[0].Quick {
		t.Fatalf("replace-by-id wrong: %+v", list)
	}

	// The cap: fill to SnippetsMaxPerExtension; a NEW id answers
	// ErrSnippetListFull while replacing an existing id still works.
	for i := len(list); i < SnippetsMaxPerExtension; i++ {
		if err := snippets.Save(ctx, domain.Snippet{
			ID: domain.GenerateSnippetID(), Owner: owner, Body: "filler", CreatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("fill %d: %v", i, err)
		}
	}
	overflow := domain.Snippet{ID: domain.GenerateSnippetID(), Owner: owner, Body: "one too many", CreatedAt: time.Now()}
	if err := snippets.Save(ctx, overflow); !errors.Is(err, ErrSnippetListFull) {
		t.Fatalf("cap: got %v, want ErrSnippetListFull", err)
	}
	if err := snippets.Save(ctx, edited); err != nil {
		t.Fatalf("replace at cap must stay open: %v", err)
	}

	// Delete is owner-scoped; a miss is ErrNotFound.
	if err := snippets.Delete(ctx, domain.MustParseExtension("1002"), second.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign delete: got %v, want ErrNotFound", err)
	}
	if err := snippets.Delete(ctx, owner, second.ID); err != nil {
		t.Fatal(err)
	}
	if err := snippets.Delete(ctx, owner, second.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("double delete: got %v, want ErrNotFound", err)
	}

	// Owner scoping on list: the other extension never sees snippets.
	if list, _ := snippets.List(ctx, domain.MustParseExtension("1002")); len(list) != 0 {
		t.Fatalf("owner scoping broken: %+v", list)
	}
}
