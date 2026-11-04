package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// The cross-owner admin seams (ListAllThreads, GetThreadAny,
// AttachmentByIDAny) are the ONLY reads allowed to skip the owner
// filter — these pins hold them to that contract: they exist, they
// cross lines, and they stay read-shaped (nothing here mutates).
func newAdminReadsFixture(t *testing.T) (*Messages, domain.Extension, domain.Extension, domain.ThreadID, domain.AttachmentID) {
	t.Helper()
	messages := newTestDB(t)
	ctx := context.Background()
	first := domain.MustParseExtension("1001")
	second := domain.MustParseExtension("1002")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
	now := time.Now()

	firstThread, err := messages.FindThread(ctx, first, remote, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: firstThread, Owner: first, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "first line body", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	secondThread, err := messages.FindThread(ctx, second, remote, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	attachmentID := domain.GenerateAttachmentID()
	attachmentMessage := domain.GenerateMessageID()
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: attachmentMessage, ThreadID: secondThread, Owner: second, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "second line body", CreatedAt: now.Add(2 * time.Minute),
		Attachments: []domain.Attachment{{
			ID: attachmentID, MessageID: attachmentMessage, Name: "probe.png",
			MimeType: "image/png", SizeBytes: 4, Path: "probe/probe.png",
		}},
	}); err != nil {
		t.Fatal(err)
	}
	return messages, first, second, secondThread, attachmentID
}

func TestListAllThreadsCrossesOwnersNewestFirst(t *testing.T) {
	messages, first, second, _, _ := newAdminReadsFixture(t)
	threads, err := messages.ListAllThreads(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 2 {
		t.Fatalf("threads = %d, want 2 (one per owner)", len(threads))
	}
	// The second line's thread was touched last: newest first.
	if threads[0].Thread.Owner != second || threads[1].Thread.Owner != first {
		t.Errorf("owner order = [%s %s], want newest-activity first", threads[0].Thread.Owner, threads[1].Thread.Owner)
	}
	if threads[0].LastBody != "second line body" {
		t.Errorf("last body = %q, want the newest thread's preview", threads[0].LastBody)
	}
	if threads, err = messages.ListAllThreads(context.Background(), 1); err != nil || len(threads) != 1 {
		t.Errorf("limit must bound the inventory: %d, %v", len(threads), err)
	}
}

func TestGetThreadAnyCrossesOwnersAndCarriesOwner(t *testing.T) {
	messages, first, second, secondThread, _ := newAdminReadsFixture(t)
	ctx := context.Background()

	thread, err := messages.GetThreadAny(ctx, secondThread)
	if err != nil {
		t.Fatal(err)
	}
	if thread.Owner != second {
		t.Errorf("owner = %s, want %s (the row carries its line)", thread.Owner, second)
	}
	// The owner-scoped sibling still refuses the foreign read.
	if _, err = messages.GetThread(ctx, first, secondThread); !errors.Is(err, ErrNotFound) {
		t.Errorf("owner-scoped GetThread must refuse the foreign thread: %v", err)
	}
	if _, err = messages.GetThreadAny(ctx, domain.GenerateThreadID()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id must answer ErrNotFound: %v", err)
	}
}

func TestAttachmentByIDAnyCrossesOwners(t *testing.T) {
	messages, first, second, _, attachmentID := newAdminReadsFixture(t)
	ctx := context.Background()

	attachment, err := messages.AttachmentByIDAny(ctx, attachmentID)
	if err != nil {
		t.Fatal(err)
	}
	if attachment.Name != "probe.png" {
		t.Errorf("attachment name = %q", attachment.Name)
	}
	// The owner-scoped sibling still refuses the foreign read.
	if _, err = messages.AttachmentByID(ctx, first, attachmentID); !errors.Is(err, ErrNotFound) {
		t.Errorf("owner-scoped AttachmentByID must refuse the foreign row: %v", err)
	}
	if _, err = messages.AttachmentByIDAny(ctx, domain.GenerateAttachmentID()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id must answer ErrNotFound: %v", err)
	}
}
