package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func newTestDB(t *testing.T) (*Messages, *Faxes, *Contacts) {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMessages(db), NewFaxes(db), NewContacts(db)
}

func TestMessageLifecycle(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	now := time.Now()

	threadID, err := messages.FindThread(ctx, owner, remote, now)
	if err != nil {
		t.Fatal(err)
	}

	outbound := domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "out one", Status: domain.StatusQueued, CreatedAt: now,
	}
	if err := messages.AppendMessage(ctx, outbound); err != nil {
		t.Fatal(err)
	}
	inbound := domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "in one", CreatedAt: now.Add(time.Second),
	}
	if err := messages.AppendMessage(ctx, inbound); err != nil {
		t.Fatal(err)
	}

	threads, err := messages.ListThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 1 {
		t.Fatalf("threads: got %d, want 1", len(threads))
	}
	if threads[0].Thread.Unread != 1 {
		t.Fatalf("unread after inbound: got %d, want 1", threads[0].Thread.Unread)
	}
	if threads[0].LastBody != "in one" {
		t.Fatalf("preview: got %q", threads[0].LastBody)
	}

	msgs, err := messages.ListMessages(ctx, owner, threadID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0].Body != "out one" || msgs[1].Body != "in one" {
		t.Fatalf("transcript order wrong: %v", msgs)
	}

	if err := messages.UpdateOutboundStatus(ctx, outbound.ID, domain.StatusSent, "ref-1"); err != nil {
		t.Fatal(err)
	}
	msgs, _ = messages.ListMessages(ctx, owner, threadID, 10)
	if msgs[0].Status != domain.StatusSent || msgs[0].ProviderRef != "ref-1" {
		t.Fatalf("status update lost: %+v", msgs[0])
	}

	// The delivery webhook's lookup path: by provider ref, outbound only.
	byRef, err := messages.MessageByProviderRef(ctx, "ref-1")
	if err != nil {
		t.Fatal(err)
	}
	if byRef.ID != outbound.ID {
		t.Fatalf("provider-ref lookup: got %s, want %s", byRef.ID, outbound.ID)
	}
	if _, err := messages.MessageByProviderRef(ctx, inbound.ProviderRef); err == nil {
		t.Fatal("inbound message must not resolve by provider ref")
	}
	if _, err := messages.MessageByProviderRef(ctx, "missing"); err == nil {
		t.Fatal("unknown provider ref must not resolve")
	}
	if err := messages.UpdateOutboundStatus(ctx, byRef.ID, domain.StatusDelivered, byRef.ProviderRef); err != nil {
		t.Fatal(err)
	}
	msgs, _ = messages.ListMessages(ctx, owner, threadID, 10)
	if msgs[0].Status != domain.StatusDelivered {
		t.Fatalf("delivered verdict lost: %+v", msgs[0])
	}

	if err := messages.MarkThreadRead(ctx, owner, threadID); err != nil {
		t.Fatal(err)
	}
	threads, _ = messages.ListThreads(ctx, owner)
	if threads[0].Thread.Unread != 0 {
		t.Fatal("mark read did not zero unread")
	}

	// Owner scoping: another extension sees nothing.
	other := domain.MustParseExtension("1002")
	otherThreads, err := messages.ListThreads(ctx, other)
	if err != nil {
		t.Fatal(err)
	}
	if len(otherThreads) != 0 {
		t.Fatalf("owner scoping broken: %d threads for other owner", len(otherThreads))
	}
}

func TestAttachmentOwnerScoped(t *testing.T) {
	messages, _, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParsePhone("+441632960961")
	threadID, _ := messages.FindThread(ctx, owner, remote, time.Now())

	msg := domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelMMS,
		Body: "", CreatedAt: time.Now(),
		Attachments: []domain.Attachment{{
			ID:        domain.GenerateAttachmentID(),
			Name:      "pic.png",
			MimeType:  "image/png",
			SizeBytes: 8,
			Path:      "attachments/x.png",
		}},
	}
	msg.Attachments[0].MessageID = msg.ID
	if err := messages.AppendMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}

	got, err := messages.AttachmentByID(ctx, owner, msg.Attachments[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != "attachments/x.png" {
		t.Fatalf("attachment path: %q", got.Path)
	}

	other := domain.MustParseExtension("1002")
	if _, err := messages.AttachmentByID(ctx, other, msg.Attachments[0].ID); err != ErrNotFound {
		t.Fatalf("other owner must not read attachment, got err=%v", err)
	}
}

func TestFaxJobStatusMachine(t *testing.T) {
	_, faxes, _ := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	job := domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: owner, Remote: domain.MustParsePhone("+441632960961"),
		Direction: domain.FaxOutbound, Status: domain.FaxQueued, DocumentPath: "faxes/a.pdf",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := faxes.Create(ctx, job); err != nil {
		t.Fatal(err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxSending, "gw-9", "", 0); err != nil {
		t.Fatal(err)
	}
	if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxTransmitted, "", "", 3); err != nil {
		t.Fatal(err)
	}

	got, err := faxes.ByProviderRef(ctx, "gw-9")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.FaxTransmitted || got.Pages != 3 {
		t.Fatalf("job state: %+v", got)
	}
}

func TestContactUpsertByNumber(t *testing.T) {
	_, _, contacts := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	first := domain.Contact{ID: domain.GenerateContactID(), Owner: owner, Name: "Bob", Phone: domain.MustParsePhone("1002"), CreatedAt: time.Now()}
	if err := contacts.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	rename := first
	rename.ID = domain.GenerateContactID()
	rename.Name = "Robert"
	if err := contacts.Save(ctx, rename); err != nil {
		t.Fatal(err)
	}

	list, err := contacts.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "Robert" {
		t.Fatalf("upsert by number broken: %+v", list)
	}
}
