package cqrs

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// threadBattery boots the thread aggregate over a real system and answers
// the dispatch-backed Threads backend (the code-complete definition: this
// battery is green or the aggregate does not exist).
func threadBattery(t *testing.T) (*system.System, *Threads) {
	t.Helper()
	sys := bootDomainSystem(t, t.TempDir(), ThreadsDomainDecl())
	return sys, NewThreads(sys)
}

func TestThreadAppendCreatesAndAdvances(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	id, err := backend.FindThread(ctx, owner, remote, time.Now())
	if err != nil {
		t.Fatalf("FindThread: %v", err)
	}

	first := domain.NewInboundMessage(domain.GenerateMessageID(), id, owner, remote, "hello", 0, time.Now())
	if err := backend.AppendMessage(ctx, first); err != nil {
		t.Fatalf("AppendMessage first: %v", err)
	}

	threads, err := backend.ListThreads(ctx, owner)
	if err != nil {
		t.Fatalf("ListThreads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("after first append: %d threads, want 1", len(threads))
	}
	got := threads[0]
	if got.ThreadID != id.String() || got.Unread != 1 || got.LastBody != "hello" ||
		got.LastDirection != string(domain.DirectionInbound) || got.Archived {
		t.Fatalf("summary = %+v", got)
	}

	secondAt := time.Now().Add(time.Second)
	second := domain.NewOutboundMessage(domain.GenerateMessageID(), id, owner, remote, "hi there", 0, secondAt)
	if err := backend.AppendMessage(ctx, second); err != nil {
		t.Fatalf("AppendMessage second: %v", err)
	}

	threads, err = backend.ListThreads(ctx, owner)
	if err != nil {
		t.Fatalf("ListThreads: %v", err)
	}
	if threads[0].Unread != 1 || threads[0].LastBody != "hi there" {
		t.Fatalf("after second append summary = %+v", threads[0])
	}

	msgs, err := backend.ListMessages(ctx, owner, id, 0)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 2 || msgs[0].Body != "hello" || msgs[1].Body != "hi there" {
		t.Fatalf("transcript order = %+v", msgs)
	}
}

func TestThreadDuplicateMessageRejected(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	id, _ := backend.FindThread(ctx, owner, remote, time.Now())
	msg := domain.NewInboundMessage(domain.GenerateMessageID(), id, owner, remote, "once", 0, time.Now())
	if err := backend.AppendMessage(ctx, msg); err != nil {
		t.Fatalf("append: %v", err)
	}
	err := backend.AppendMessage(ctx, msg)
	if err == nil {
		t.Fatal("duplicate append accepted")
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
}

func TestThreadDeliveryMachine(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
	msg := domain.NewOutboundMessage(domain.GenerateMessageID(), threadID, owner, remote, "out", 0, time.Now())
	if err := backend.AppendMessage(ctx, msg); err != nil {
		t.Fatalf("append: %v", err)
	}

	// queued → sent (send accepted, provider ref correlates)
	if err := backend.AdvanceOutbound(ctx, msg.ID, domain.SendAccepted("ref-1").Step()); err != nil {
		t.Fatalf("advance accepted: %v", err)
	}
	byRef, err := backend.MessageByProviderRef(ctx, "ref-1")
	if err != nil {
		t.Fatalf("MessageByProviderRef: %v", err)
	}
	if byRef.Status != domain.StatusSent || byRef.ProviderRef != "ref-1" {
		t.Fatalf("after accepted = %+v", byRef)
	}

	// sent → delivered (verdict; failure story cleared)
	if err := backend.AdvanceOutbound(ctx, msg.ID, domain.ProviderVerdict(domain.VerdictDelivered, "").Step()); err != nil {
		t.Fatalf("advance verdict: %v", err)
	}
	byRef, err = backend.MessageByProviderRef(ctx, "ref-1")
	if err != nil {
		t.Fatalf("MessageByProviderRef: %v", err)
	}
	if byRef.Status != domain.StatusDelivered {
		t.Fatalf("after verdict = %+v", byRef)
	}

	// delivered → failed (out-of-order verdicts are legal, failure recorded)
	if err := backend.AdvanceOutbound(ctx, msg.ID, domain.ProviderVerdict(domain.VerdictFailed, "boom").Step()); err != nil {
		t.Fatalf("advance late verdict: %v", err)
	}
	byRef, _ = backend.MessageByProviderRef(ctx, "ref-1")
	if byRef.Status != domain.StatusFailed || byRef.FailureKind != domain.FailureProvider || byRef.FailureDetail != "boom" {
		t.Fatalf("after late verdict = %+v", byRef)
	}

	// a send-refused move from a non-queued state is INERT (the settled
	// state answers, no event, no error)
	if err := backend.AdvanceOutbound(ctx, msg.ID, domain.SendRefused(domain.FailureTransient, "late refuse").Step()); err != nil {
		t.Fatalf("inert advance errored: %v", err)
	}
	byRef, _ = backend.MessageByProviderRef(ctx, "ref-1")
	if byRef.Status != domain.StatusFailed || byRef.FailureDetail != "boom" {
		t.Fatalf("after inert advance = %+v", byRef)
	}
}

func TestThreadFlagsReadSearchArchive(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
	if err := backend.AppendMessage(ctx, domain.NewInboundMessage(
		domain.GenerateMessageID(), threadID, owner, remote, "needle body", 0, time.Now())); err != nil {
		t.Fatalf("append: %v", err)
	}

	if err := backend.SetThreadFlag(ctx, owner, threadID, "pinned", true); err != nil {
		t.Fatalf("pin: %v", err)
	}
	if err := backend.SetThreadFlag(ctx, owner, threadID, "muted", true); err != nil {
		t.Fatalf("mute: %v", err)
	}
	thread, err := backend.GetThread(ctx, owner, threadID)
	if err != nil {
		t.Fatalf("GetThread: %v", err)
	}
	if !thread.Pinned || !thread.Muted || thread.Unread != 1 {
		t.Fatalf("flags thread = %+v", thread)
	}

	if err := backend.MarkThreadRead(ctx, owner, threadID); err != nil {
		t.Fatalf("MarkThreadRead: %v", err)
	}
	thread, _ = backend.GetThread(ctx, owner, threadID)
	if thread.Unread != 0 {
		t.Fatalf("after mark-read unread = %d", thread.Unread)
	}

	// search matches the message body and the remote
	for _, needle := range []string{"needle", "1701"} {
		hits, err := backend.SearchThreads(ctx, owner, needle)
		if err != nil {
			t.Fatalf("SearchThreads(%q): %v", needle, err)
		}
		if len(hits) != 1 {
			t.Fatalf("search %q: %d hits, want 1", needle, len(hits))
		}
	}
	misses, err := backend.SearchThreads(ctx, owner, "zzz")
	if err != nil || len(misses) != 0 {
		t.Fatalf("search miss: (%d, %v)", len(misses), err)
	}

	// archived leaves the default list, joins the archived view
	if err := backend.SetThreadFlag(ctx, owner, threadID, "archived", true); err != nil {
		t.Fatalf("archive: %v", err)
	}
	active, _ := backend.ListThreads(ctx, owner)
	archived, _ := backend.ListArchivedThreads(ctx, owner)
	count, _ := backend.CountArchived(ctx, owner)
	if len(active) != 0 || len(archived) != 1 || count != 1 {
		t.Fatalf("archived split: active=%d archived=%d count=%d", len(active), len(archived), count)
	}

	// an inbound message un-archives
	if err := backend.AppendMessage(ctx, domain.NewInboundMessage(
		domain.GenerateMessageID(), threadID, owner, remote, "again", 0, time.Now())); err != nil {
		t.Fatalf("append after archive: %v", err)
	}
	active, _ = backend.ListThreads(ctx, owner)
	if len(active) != 1 || active[0].Unread != 1 {
		t.Fatalf("after un-archiving append: %+v", active)
	}
}

func TestThreadAttachmentsRideAndHydrate(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
	msg := domain.NewInboundMessage(domain.GenerateMessageID(), threadID, owner, remote, "mms", 2, time.Now())
	msg.Attachments = []domain.Attachment{
		{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "b.png", MimeType: "image/png", SizeBytes: 3, Path: "blob/b"},
		{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "a.png", MimeType: "image/png", SizeBytes: 2, Path: "blob/a"},
	}
	if err := backend.AppendMessage(ctx, msg); err != nil {
		t.Fatalf("append mms: %v", err)
	}

	msgs, err := backend.ListMessages(ctx, owner, threadID, 0)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Channel != domain.ChannelMMS || len(msgs[0].Attachments) != 2 {
		t.Fatalf("mms message = %+v", msgs)
	}
	if msgs[0].Attachments[0].Name != "a.png" || msgs[0].Attachments[1].Name != "b.png" {
		t.Fatalf("attachments not name-ordered: %+v", msgs[0].Attachments)
	}

	threads, _ := backend.ListThreads(ctx, owner)
	if !threads[0].LastAttachment || threads[0].LastChannel != string(domain.ChannelMMS) {
		t.Fatalf("summary preview = %+v", threads[0])
	}

	aID := msg.Attachments[0].ID
	if msg.Attachments[0].Name == "b.png" {
		aID = msg.Attachments[1].ID
	}
	att, err := backend.AttachmentByID(ctx, owner, aID)
	if err != nil || att.Path != "blob/a" {
		t.Fatalf("AttachmentByID = (%+v, %v)", att, err)
	}
	other := mustExt(t, "200")
	if _, err := backend.AttachmentByID(ctx, other, aID); err != store.ErrNotFound {
		t.Fatalf("cross-owner attachment read: err=%v, want ErrNotFound", err)
	}
}

func TestThreadOwnerScopingAndAdminSeams(t *testing.T) {
	_, backend := threadBattery(t)
	ctx := context.Background()
	a := mustExt(t, "100")
	b := mustExt(t, "200")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491701")

	for _, owner := range []domain.Extension{a, b} {
		threadID, _ := backend.FindThread(ctx, owner, remote, time.Now())
		if err := backend.AppendMessage(ctx, domain.NewInboundMessage(
			domain.GenerateMessageID(), threadID, owner, remote, "mine", 0, time.Now())); err != nil {
			t.Fatalf("append for %s: %v", owner, err)
		}
	}

	aThreads, _ := backend.ListThreads(ctx, a)
	if len(aThreads) != 1 || aThreads[0].Owner != a.String() {
		t.Fatalf("owner list bleed: %+v", aThreads)
	}

	// the admin seams see both
	all, err := backend.ListAllThreads(ctx, 0)
	if err != nil || len(all) != 2 {
		t.Fatalf("ListAllThreads = (%d, %v)", len(all), err)
	}
	anyThread, err := backend.GetThreadAny(ctx, domain.MustThreadID(aThreads[0].ThreadID))
	if err != nil || anyThread.Owner != a {
		t.Fatalf("GetThreadAny = (%+v, %v)", anyThread, err)
	}

	// a missing thread is the store's ErrNotFound (the handler 404 map)
	if _, err := backend.GetThread(ctx, b, domain.MustThreadID(aThreads[0].ThreadID)); err != store.ErrNotFound {
		t.Fatalf("cross-owner thread read: err=%v, want ErrNotFound", err)
	}
}
