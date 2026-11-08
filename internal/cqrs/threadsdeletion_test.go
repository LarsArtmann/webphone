// Thread deletion/retention battery (SUPERB Phase 5, A33/A36): the
// tombstone's row removals, the journal byte-trim, the sweep gate
// (journal empty, views gone, fresh replay without the content),
// and the restart reload across them.
package cqrs

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

func TestThreadDeleteRemovesRowsAndJournal(t *testing.T) {
	sys, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
	msg := domain.NewInboundMessage(domain.GenerateMessageID(), threadID, owner, remote, "bye", 1, time.Now())
	msg.Attachments = []domain.Attachment{
		{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "x.png", MimeType: "image/png", Path: "blob/x"},
	}
	if err := backend.AppendMessage(ctx, msg); err != nil {
		t.Fatalf("append: %v", err)
	}

	if err := backend.DeleteThread(ctx, owner, threadID, "retention"); err != nil {
		t.Fatalf("DeleteThread: %v", err)
	}

	if threads, _ := backend.ListThreads(ctx, owner); len(threads) != 0 {
		t.Fatalf("threads after delete: %+v", threads)
	}
	if msgs, _ := backend.ListMessages(ctx, owner, threadID, 0); len(msgs) != 0 {
		t.Fatalf("messages after delete: %+v", msgs)
	}
	if _, err := backend.AttachmentByIDAny(ctx, msg.Attachments[0].ID); err != store.ErrNotFound {
		t.Fatalf("attachment after delete: err=%v, want ErrNotFound", err)
	}

	// THE deletion invariant (A36): the journal bytes are GONE and a
	// replay rebuilds without the deleted content.
	ref := streamRefForTest(t, owner, remote)
	if loaded, err := sys.EventStore().Load(ctx, ref); err != nil || len(loaded) != 0 {
		t.Fatalf("journal after trim: (%d events, %v), want (0, nil)", len(loaded), err)
	}
}

func TestThreadRestartReloadsState(t *testing.T) {
	dir := t.TempDir()
	sys := bootDomainSystem(t, dir, ThreadsDomainDecl())
	backend := NewThreads(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
	if err := backend.AppendMessage(ctx, domain.NewInboundMessage(
		domain.GenerateMessageID(), threadID, owner, remote, "persist", 0, time.Now())); err != nil {
		t.Fatalf("append: %v", err)
	}

	if err := sys.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	sys2 := bootDomainSystem(t, dir, ThreadsDomainDecl())
	defer func() { _ = sys2.Close() }()
	backend2 := NewThreads(sys2)
	threads, err := backend2.ListThreads(ctx, owner)
	if err != nil || len(threads) != 1 || threads[0].LastBody != "persist" {
		t.Fatalf("after restart = (%+v, %v)", threads, err)
	}
}

func streamRefForTest(t *testing.T, owner domain.Extension, remote domain.Address) id.StreamRef {
	t.Helper()
	ref, err := streamRef(ThreadsStreamType, ThreadStreamKey(owner, remote.Network(), remote))
	if err != nil {
		t.Fatalf("streamRef: %v", err)
	}
	return ref
}

// TestThreadSweepDeletesJournalViewsAndBlobs is the A36 deletion gate:
// after SweepBefore, an expired thread is GONE from the journal (bytes,
// not just rows), from every projection, from the returned blob set
// (and ONLY its own blobs — a sibling thread's files are not this
// sweep's), and a FRESH replay over the same directory rebuilds without
// the deleted content while the live thread survives untouched.
func TestThreadSweepDeletesJournalViewsAndBlobs(t *testing.T) {
	dir := t.TempDir()
	sys := bootDomainSystem(t, dir, ThreadsDomainDecl())
	backend := NewThreads(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	oldRemote := domain.MustParseAddress(domain.NetworkSMS, "+4917old")
	freshRemote := domain.MustParseAddress(domain.NetworkSMS, "+491new")

	oldAt := time.Now().Add(-10 * 24 * time.Hour)
	oldID, err := backend.FindThread(ctx, owner, oldRemote, oldAt)
	if err != nil {
		t.Fatalf("FindThread old: %v", err)
	}
	oldMsg := domain.NewInboundMessage(domain.GenerateMessageID(), oldID, owner, oldRemote, "ancient", 0, oldAt)
	oldMsg.Attachments = []domain.Attachment{{
		ID: domain.GenerateAttachmentID(), MessageID: oldMsg.ID, Name: "old.png",
		MimeType: "image/png", SizeBytes: 4, Path: "100/old.png",
	}}
	if err := backend.AppendMessage(ctx, oldMsg); err != nil {
		t.Fatalf("append old: %v", err)
	}
	freshID, err := backend.FindThread(ctx, owner, freshRemote, time.Now())
	if err != nil {
		t.Fatalf("FindThread fresh: %v", err)
	}
	if err := backend.AppendMessage(ctx, domain.NewInboundMessage(
		domain.GenerateMessageID(), freshID, owner, freshRemote, "live", 0, time.Now())); err != nil {
		t.Fatalf("append fresh: %v", err)
	}

	paths, removed, err := backend.SweepBefore(ctx, time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("SweepBefore: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1 (only the expired thread)", removed)
	}
	if len(paths) != 1 || paths[0] != "100/old.png" {
		t.Fatalf("blob paths = %v, want exactly the expired thread's attachment", paths)
	}

	// Views: the expired thread and its rows are gone; the live one stays.
	if _, err := backend.GetThread(ctx, owner, oldID); err != store.ErrNotFound {
		t.Fatalf("expired thread must be gone from the views: err %v", err)
	}
	threads, err := backend.ListThreads(ctx, owner)
	if err != nil || len(threads) != 1 || threads[0].ThreadID != freshID.String() {
		t.Fatalf("post-sweep list = (%+v, %v)", threads, err)
	}
	if _, err := backend.AttachmentByID(ctx, owner, oldMsg.Attachments[0].ID); err != store.ErrNotFound {
		t.Fatalf("expired attachment must be gone from the views: err %v", err)
	}

	// Journal: the trimmed stream folds to NOTHING — the bytes are gone.
	events, err := sys.EventStore().Load(ctx, streamRefForTest(t, owner, oldRemote))
	if err != nil {
		t.Fatalf("journal load: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("trimmed stream still holds %d events, want 0", len(events))
	}

	// Replay consistency: a fresh system over the same directory rebuilds
	// WITHOUT the deleted content (no partial resurrection) and keeps the
	// live thread whole.
	if err := sys.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	sys2 := bootDomainSystem(t, dir, ThreadsDomainDecl())
	defer func() { _ = sys2.Close() }()
	backend2 := NewThreads(sys2)
	replayed, err := backend2.ListThreads(ctx, owner)
	if err != nil || len(replayed) != 1 || replayed[0].ThreadID != freshID.String() {
		t.Fatalf("post-replay list = (%+v, %v), want only the live thread", replayed, err)
	}
	replayedMsgs, err := backend2.ListMessages(ctx, owner, freshID, 0)
	if err != nil || len(replayedMsgs) != 1 || replayedMsgs[0].Body != "live" {
		t.Fatalf("post-replay transcript = (%+v, %v)", replayedMsgs, err)
	}
}
