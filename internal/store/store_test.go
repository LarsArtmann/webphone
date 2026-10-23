package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestListMessagesFansAttachmentsOutInOnePass pins the grouped IN-query
// fan-out: every message gets exactly its own attachments, in name
// order, and a message with none stays empty — the N+1 loop this
// replaced could attach (or duplicate) per-query rows by accident.
func TestListMessagesFansAttachmentsOutInOnePass(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
	now := time.Now()

	threadID, err := messages.FindThread(ctx, owner, remote, now)
	if err != nil {
		t.Fatal(err)
	}
	for i, body := range []string{"m1", "m2", "m3"} {
		msg := domain.NewOutboundMessage(
			domain.GenerateMessageID(), threadID, owner, remote,
			body, 0, now.Add(time.Duration(i)*time.Second),
		)
		switch i {
		case 0:
			msg.Attachments = []domain.Attachment{
				{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "b.png", MimeType: "image/png", SizeBytes: 2, Path: "attachments/b.png"},
				{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "a.png", MimeType: "image/png", SizeBytes: 1, Path: "attachments/a.png"},
			}
		case 2:
			msg.Attachments = []domain.Attachment{
				{ID: domain.GenerateAttachmentID(), MessageID: msg.ID, Name: "c.pdf", MimeType: "application/pdf", SizeBytes: 3, Path: "attachments/c.pdf"},
			}
		}
		if err := messages.AppendMessage(ctx, msg); err != nil {
			t.Fatal(err)
		}
	}

	msgs, err := messages.ListMessages(ctx, owner, threadID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 3 {
		t.Fatalf("messages: %d (want 3)", len(msgs))
	}
	if got := msgs[0].Attachments; len(got) != 2 || got[0].Name != "a.png" || got[1].Name != "b.png" {
		t.Errorf("m1 attachments must be exactly its own, in name order: %+v", got)
	}
	if len(msgs[1].Attachments) != 0 {
		t.Errorf("m2 must carry no attachments: %+v", msgs[1].Attachments)
	}
	if got := msgs[2].Attachments; len(got) != 1 || got[0].Name != "c.pdf" {
		t.Errorf("m3 attachments must be exactly its own: %+v", got)
	}
}

func newTestDB(t *testing.T) *Messages {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMessages(db)
}

func TestMessageLifecycle(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
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

	if err := messages.UpdateOutboundStatus(ctx, outbound.ID, domain.Delivery{Status: domain.StatusSent, ProviderRef: "ref-1"}); err != nil {
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
	// Failure story (send-failure plan D): a failed verdict persists
	// kind + detail; a later delivered verdict clears them.
	if err := messages.UpdateOutboundStatus(ctx, byRef.ID, domain.Delivery{
		Status: domain.StatusFailed, ProviderRef: byRef.ProviderRef,
		FailureKind: "provider", FailureDetail: "217022: not a valid SMS destination",
	}); err != nil {
		t.Fatal(err)
	}
	msgs, _ = messages.ListMessages(ctx, owner, threadID, 10)
	if msgs[0].FailureKind != "provider" || msgs[0].FailureDetail != "217022: not a valid SMS destination" {
		t.Fatalf("failure story lost: %+v", msgs[0])
	}
	if err := messages.UpdateOutboundStatus(ctx, byRef.ID, domain.Delivery{Status: domain.StatusDelivered, ProviderRef: byRef.ProviderRef}); err != nil {
		t.Fatal(err)
	}
	msgs, _ = messages.ListMessages(ctx, owner, threadID, 10)
	if msgs[0].Status != domain.StatusDelivered || msgs[0].FailureKind != "" || msgs[0].FailureDetail != "" {
		t.Fatalf("delivered verdict lost or failure not cleared: %+v", msgs[0])
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

// TestAppendWithoutPriorThread pins AppendMessage's creation fallback:
// the service normally pre-creates the row via FindThread, but an
// append into a fresh (owner, remote) pair must still create the thread
// (adopting the message's thread id) and count the inbound as unread.
func TestAppendWithoutPriorThread(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
	msg := domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: domain.GenerateThreadID(),
		Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "cold open", Status: domain.StatusQueued, CreatedAt: time.Now(),
	}

	if err := messages.AppendMessage(ctx, msg); err != nil {
		t.Fatal(err)
	}

	threads, err := messages.ListThreads(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 1 || threads[0].Thread.ID != msg.ThreadID {
		t.Fatalf("creation fallback must adopt the message's thread id: %+v", threads)
	}
	if threads[0].Thread.Unread != 1 || threads[0].Thread.Archived {
		t.Fatalf("creation fallback invariants: unread = %d archived = %v", threads[0].Thread.Unread, threads[0].Thread.Archived)
	}
}

func TestAttachmentOwnerScoped(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
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

func TestMessageProviderRefUniqueness(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
	now := time.Now()

	threadID, err := messages.FindThread(ctx, owner, remote, now)
	if err != nil {
		t.Fatal(err)
	}
	mk := func(ref string) domain.Message {
		return domain.Message{
			ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
			Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
			Body: "claiming " + ref, Status: domain.StatusSent, ProviderRef: ref, CreatedAt: now,
		}
	}

	if err := messages.AppendMessage(ctx, mk("ref-dup")); err != nil {
		t.Fatal(err)
	}
	// The UNIQUE partial index is the last line of defense against a
	// replayed status webhook double-claiming one ref.
	if err := messages.AppendMessage(ctx, mk("ref-dup")); err == nil {
		t.Fatal("second message with the same provider_ref must be rejected")
	}
	// Empty refs (queued, inbound) are exempt: many messages may wait.
	if err := messages.AppendMessage(ctx, mk("")); err != nil {
		t.Fatalf("empty provider_ref rows must not collide: %v", err)
	}
}

func TestSearchThreads(t *testing.T) {
	messages := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	other := domain.MustParseExtension("2002")

	// Thread one: the match lives in an OLDER message body, not the
	// preview — the search must scan every message, not just the last.
	first := domain.MustParseAddress(domain.NetworkSMS, "+441632960961")
	firstID, err := messages.FindThread(ctx, owner, first, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: firstID, Owner: owner, Remote: first,
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "the quarterly revenue is 50%", Status: domain.StatusQueued,
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: firstID, Owner: owner, Remote: first,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "thanks, call me later", CreatedAt: time.Now().Add(-time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	// Thread two: matches only by remote number.
	second := domain.MustParseAddress(domain.NetworkSMS, "+491601234567")
	secondID, err := messages.FindThread(ctx, owner, second, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: secondID, Owner: owner, Remote: second,
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "ping", Status: domain.StatusQueued, CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	// A different owner's thread with the same body must stay invisible.
	otherID, err := messages.FindThread(ctx, other, first, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: otherID, Owner: other, Remote: first,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "the quarterly revenue is 50%", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"body match in an older message", "quarterly revenue", []string{firstID.String()}},
		{"remote number match", "60961", []string{firstID.String()}},
		{"other remote match", "01234567", []string{secondID.String()}},
		{"case-insensitive ASCII", "QUARTERLY", []string{firstID.String()}},
		{"percent is literal, not a wildcard", "50%", []string{firstID.String()}},
		{"underscore is literal, not a wildcard", "n_revenue", []string{}},
		{"no match", "voicemail greeting", []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.SearchThreads(ctx, owner, tc.query)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]string, 0, len(got))
			for _, summary := range got {
				ids = append(ids, summary.Thread.ID.String())
			}
			if len(ids) != len(tc.want) {
				t.Fatalf("query %q: got %v, want %v", tc.query, ids, tc.want)
			}
			for i, id := range tc.want {
				if ids[i] != id {
					t.Fatalf("query %q: got %v, want %v", tc.query, ids, tc.want)
				}
			}
		})
	}

	// Owner scoping: the same query against the other owner returns only
	// their thread.
	others, err := messages.SearchThreads(ctx, other, "quarterly")
	if err != nil {
		t.Fatal(err)
	}
	if len(others) != 1 || others[0].Thread.Owner.String() != other.String() {
		t.Fatalf("owner scoping: got %+v", others)
	}

	// Ordering: most recently active first (thread two was touched last).
	// A shared digit hits both threads; the empty-query routing decision
	// (list everything, no pattern) lives in the panel handler.
	list, err := messages.SearchThreads(ctx, owner, "0")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("shared-digit query should hit both threads: got %d", len(list))
	}
	if list[0].Thread.ID.String() != secondID.String() {
		t.Fatalf("ordering: want newest thread first, got %s then %s", list[0].Thread.ID, list[1].Thread.ID)
	}
}
