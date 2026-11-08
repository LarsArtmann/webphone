package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/id/v4"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// seedLegacyThread appends one message through the LEGACY write path
// (the store the pre-swap product used) with caller-chosen stamps.
func seedLegacyThread(t *testing.T, legacy *store.Messages, owner domain.Extension, remote domain.Address, now time.Time) domain.ThreadID {
	t.Helper()
	ctx := context.Background()
	id, err := legacy.FindThread(ctx, owner, remote, now)
	if err != nil {
		t.Fatalf("seed FindThread: %v", err)
	}
	return id
}

func seedLegacyMessage(t *testing.T, legacy *store.Messages, msg domain.Message) {
	t.Helper()
	if err := legacy.AppendMessage(context.Background(), msg); err != nil {
		t.Fatalf("seed AppendMessage: %v", err)
	}
}

// TestThreadsImportDualRunParity is A35's gate: the legacy
// threads/messages/attachments history survives the journal port
// row-for-row — same threads, same flags, same unread state, same
// transcript order, same attachments, same provider-ref lookup — the
// marker latches, and a second boot imports nothing new. Message-less
// legacy threads (the old empty-row insert) are DELIBERATELY dropped:
// the A32 design rules no thread row exists until a message lands.
func TestThreadsImportDualRunParity(t *testing.T) {
	cfg := testConfig(t, false)
	ctx := context.Background()

	// Seed the LEGACY store exactly as the old product wrote it.
	db, err := store.OpenDatabase(filepath.Join(cfg.DataDir, "webphone.db"))
	if err != nil {
		t.Fatalf("open legacy store: %v", err)
	}
	legacy := store.NewMessages(db.SQL())
	ownerA := mustExtension(t, "1001")
	ownerB := mustExtension(t, "1002")
	remoteA := domain.MustParseAddress(domain.NetworkSMS, "+491511111111")
	remoteB := domain.MustParseAddress(domain.NetworkSMS, "+491519999999")
	remoteC := domain.MustParseAddress(domain.NetworkSMS, "+491512222222")
	base := time.Now().Add(-48 * time.Hour)

	// Thread A: three messages whose INSERTION order deliberately
	// disagrees with created_at (rowid order m1, m0, m2) — the import
	// must follow created_at, the legacy list's own order. The outbound
	// row carries a settled delivery state + provider ref (the webhook
	// lane's lookup key), and the thread was read after the last
	// inbound (unread 0: row state NOT derivable from the sequence).
	threadA := seedLegacyThread(t, legacy, ownerA, remoteA, base)
	msgA0 := domain.NewInboundMessage(domain.GenerateMessageID(), threadA, ownerA, remoteA, "oldest", 0, base)
	msgA1 := domain.NewOutboundMessage(domain.GenerateMessageID(), threadA, ownerA, remoteA, "reply", 0, base.Add(time.Minute))
	msgA1.Status = domain.StatusDelivered
	msgA1.ProviderRef = "legacy-ref-a1"
	msgA2 := domain.NewInboundMessage(domain.GenerateMessageID(), threadA, ownerA, remoteA, "newest", 0, base.Add(2*time.Minute))
	for _, m := range []domain.Message{msgA1, msgA0, msgA2} { // insertion order ≠ created_at order
		seedLegacyMessage(t, legacy, m)
	}
	if err := legacy.MarkThreadRead(ctx, ownerA, threadA); err != nil {
		t.Fatalf("seed MarkThreadRead: %v", err)
	}

	// Thread B: flagged on all three axes with one attachment-bearing
	// inbound message (the MMS row's blob path must survive verbatim).
	threadB := seedLegacyThread(t, legacy, ownerA, remoteB, base)
	attID := domain.GenerateAttachmentID()
	msgB := domain.NewInboundMessage(domain.GenerateMessageID(), threadB, ownerA, remoteB, "with picture", 0, base.Add(3*time.Minute))
	msgB.Attachments = []domain.Attachment{{
		ID: attID, MessageID: msgB.ID, Name: "pic.png", MimeType: "image/png",
		SizeBytes: 8, Path: "1001/pic.png",
	}}
	seedLegacyMessage(t, legacy, msgB)
	for _, flag := range []store.ThreadFlag{store.FlagPinned, store.FlagArchived, store.FlagMuted} {
		if err := legacy.SetThreadFlag(ctx, ownerA, threadB, flag, true); err != nil {
			t.Fatalf("seed SetThreadFlag %s: %v", flag, err)
		}
	}

	// Thread C on another owner (the scoping axis) + one EMPTY thread
	// (the legacy empty-row insert the import drops by design).
	threadC := seedLegacyThread(t, legacy, ownerB, remoteC, base)
	seedLegacyMessage(t, legacy, domain.NewInboundMessage(domain.GenerateMessageID(), threadC, ownerB, remoteC, "other line", 0, base.Add(4*time.Minute)))
	emptyThread := seedLegacyThread(t, legacy, ownerA, domain.MustParseAddress(domain.NetworkSMS, "+491590000000"), base)

	// The expected world, read through the legacy lane before the swap.
	// The legacy ACTIVE list carries the empty thread (dropped by
	// design at import); archived thread B lives in its own lane.
	wantAll, err := legacy.ListThreads(ctx, ownerA)
	if err != nil {
		t.Fatalf("legacy ListThreads: %v", err)
	}
	wantThreadsA := make([]store.ThreadSummary, 0, len(wantAll))
	for _, s := range wantAll {
		if s.Thread.ID != emptyThread {
			wantThreadsA = append(wantThreadsA, s)
		}
	}
	wantArchivedA, err := legacy.ListArchivedThreads(ctx, ownerA)
	if err != nil {
		t.Fatalf("legacy ListArchivedThreads: %v", err)
	}
	wantMsgsA, err := legacy.ListMessages(ctx, ownerA, threadA, 100)
	if err != nil {
		t.Fatalf("legacy ListMessages: %v", err)
	}
	if err := db.Shutdown(); err != nil {
		t.Fatalf("close legacy store: %v", err)
	}

	// First boot: the importer ports the rows into the journal.
	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	svc := mustResolveCqrs(t, application)
	gate := svc.ThreadsSvc()

	deadline := time.Now().Add(20 * time.Second)
	for {
		got, err := gate.ListThreads(ctx, ownerA)
		if err != nil {
			t.Fatalf("gate ListThreads: %v", err)
		}
		if len(got) == len(wantThreadsA) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("imported threads never reached the projection: got %d want %d (active, non-empty)", len(got), len(wantThreadsA))
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Row-by-row diff, owner A first: every summary field the list
	// renders, in the legacy list's own order.
	gotViews, err := gate.ListThreads(ctx, ownerA)
	if err != nil {
		t.Fatalf("gate ListThreads: %v", err)
	}
	gotThreads := readmodels.SummariesOfView(gotViews)
	if len(gotThreads) != len(wantThreadsA) {
		t.Fatalf("owner A threads: got %d want %d", len(gotThreads), len(wantThreadsA))
	}
	for i := range wantThreadsA {
		want, gotSummary := wantThreadsA[i], gotThreads[i]
		if gotSummary.Thread != want.Thread {
			t.Errorf("thread %d: got %+v want %+v", i, gotSummary.Thread, want.Thread)
		}
		if gotSummary.LastBody != want.LastBody || gotSummary.LastDirection != want.LastDirection ||
			gotSummary.LastChannel != want.LastChannel || gotSummary.LastAttachment != want.LastAttachment {
			t.Errorf("thread %d preview: got (%s/%s/%s/%v) want (%s/%s/%s/%v)",
				i, gotSummary.LastBody, gotSummary.LastDirection, gotSummary.LastChannel, gotSummary.LastAttachment,
				want.LastBody, want.LastDirection, want.LastChannel, want.LastAttachment)
		}
	}

	// The archived lane parity: thread B survives with all three flags.
	gotArchived, err := gate.ListArchivedThreads(ctx, ownerA)
	if err != nil {
		t.Fatalf("gate ListArchivedThreads: %v", err)
	}
	if len(gotArchived) != len(wantArchivedA) {
		t.Fatalf("owner A archived threads: got %d want %d", len(gotArchived), len(wantArchivedA))
	}
	for i := range wantArchivedA {
		want, gotSummary := wantArchivedA[i], readmodels.SummaryOfView(gotArchived[i])
		if gotSummary.Thread != want.Thread {
			t.Errorf("archived thread %d: got %+v want %+v", i, gotSummary.Thread, want.Thread)
		}
	}

	// Transcript parity: same rows in the same order, attachments and
	// delivery state included.
	gotMsgs, err := gate.ListMessages(ctx, ownerA, threadA, 0)
	if err != nil {
		t.Fatalf("gate ListMessages: %v", err)
	}
	if len(gotMsgs) != len(wantMsgsA) {
		t.Fatalf("thread A messages: got %d want %d", len(gotMsgs), len(wantMsgsA))
	}
	for i := range wantMsgsA {
		want, got := wantMsgsA[i], gotMsgs[i]
		if got.ID != want.ID || got.Body != want.Body || got.Direction != want.Direction ||
			got.Channel != want.Channel || got.Status != want.Status || got.ProviderRef != want.ProviderRef ||
			got.CreatedAt != want.CreatedAt {
			t.Errorf("message %d: got %+v want %+v", i, got, want)
		}
	}
	if err := gate.MarkThreadRead(ctx, ownerA, threadA); err != nil {
		t.Fatalf("gate MarkThreadRead: %v", err)
	}

	// The webhook lane's lookup key survives: the provider ref resolves
	// to the same settled outbound row.
	byRef, err := gate.MessageByProviderRef(ctx, "legacy-ref-a1")
	if err != nil || byRef.ID != msgA1.ID || byRef.Status != domain.StatusDelivered {
		t.Fatalf("provider-ref lookup: got %+v err %v, want message %s delivered", byRef, err, msgA1.ID)
	}

	// Attachment hydration: the MMS row's path resolves through the gate.
	att, err := gate.AttachmentByID(ctx, ownerA, attID)
	if err != nil || att.Path != "1001/pic.png" || att.Name != "pic.png" || att.SizeBytes != 8 {
		t.Fatalf("attachment lookup: got %+v err %v", att, err)
	}

	// Owner scoping: B sees exactly its own thread, and A's empty thread
	// is gone by design (no message, no stream).
	gotB, err := gate.ListThreads(ctx, ownerB)
	if err != nil {
		t.Fatalf("gate ListThreads B: %v", err)
	}
	if len(gotB) != 1 || gotB[0].ThreadID != threadC.String() {
		t.Fatalf("owner B threads: got %+v want exactly thread %s", gotB, threadC)
	}
	if _, err := gate.GetThread(ctx, ownerA, emptyThread); err != store.ErrNotFound {
		t.Fatalf("empty legacy thread must not survive the import: err %v", err)
	}

	// The journal holds the imported events; the marker latched.
	key := cqrs.ThreadStreamKey(ownerA, remoteA.Network(), remoteA)
	ref := mustThreadStreamRef(t, key)
	stream, err := svc.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("journal load: %v", err)
	}
	if len(stream) != 3 {
		t.Fatalf("thread A journal: %d events, want 3 (one per message)", len(stream))
	}
	flagDB, err := store.OpenDatabase(filepath.Join(cfg.DataDir, "webphone.db"))
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	imported, err := store.FlagPresent(ctx, flagDB.SQL(), store.FlagThreadsJournalImported)
	if err != nil || !imported {
		t.Fatalf("import marker: present %v err %v", imported, err)
	}
	if err := flagDB.Shutdown(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	// Second boot: the marker holds, nothing re-imports (the journal is
	// byte-stable), and a NEW message continues the imported stream.
	if err := application.Shutdown(); err != nil {
		t.Fatalf("shutdown app 1: %v", err)
	}
	application2, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New 2: %v", err)
	}
	if err := application2.Start(ctx); err != nil {
		t.Fatalf("Start 2: %v", err)
	}
	t.Cleanup(func() {
		if err := application2.Shutdown(); err != nil {
			t.Errorf("Shutdown 2: %v", err)
		}
	})
	svc2 := mustResolveCqrs(t, application2)
	stream2, err := svc2.Sys.EventStore().Load(ctx, ref)
	if err != nil {
		t.Fatalf("journal load 2: %v", err)
	}
	if len(stream2) != len(stream) {
		t.Fatalf("second boot re-imported: journal grew %d → %d events", len(stream), len(stream2))
	}
	live := domain.NewInboundMessage(domain.GenerateMessageID(), threadA, ownerA, remoteA, "post-import", 0, time.Now())
	if err := svc2.ThreadsSvc().AppendMessage(ctx, live); err != nil {
		t.Fatalf("post-import append: %v", err)
	}
	gotLive, err := svc2.ThreadsSvc().ListMessages(ctx, ownerA, threadA, 0)
	if err != nil {
		t.Fatalf("post-import list: %v", err)
	}
	if len(gotLive) != len(gotMsgs)+1 || gotLive[len(gotLive)-1].Body != "post-import" {
		t.Fatalf("post-import transcript: got %d rows, want %d with the live append on top", len(gotLive), len(gotMsgs)+1)
	}
}

func mustThreadStreamRef(t *testing.T, key string) id.StreamRef {
	t.Helper()
	streamID, err := id.ParseStreamID(key)
	if err != nil {
		t.Fatalf("parse stream key %q: %v", key, err)
	}
	return id.NewStreamRef(cqrs.ThreadsStreamType, streamID)
}
