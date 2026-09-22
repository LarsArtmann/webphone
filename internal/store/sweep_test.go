package store

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

func TestSweepDeletesOnlyExpiredContent(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	messages := NewMessages(db)
	faxes := NewFaxes(db)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	now := time.Now()
	old := now.Add(-48 * time.Hour)

	oldThread, err := messages.FindThread(ctx, owner, domain.MustParsePhone("+441632960961"), old)
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: oldThread, Owner: owner,
		Remote:    domain.MustParsePhone("+441632960961"),
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "old one", Status: domain.StatusSent, ProviderRef: "old-ref",
		CreatedAt: old,
		Attachments: []domain.Attachment{{
			ID: domain.GenerateAttachmentID(), Name: "old.png", MimeType: "image/png",
			SizeBytes: 4, Path: "old-blob",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	liveThread, err := messages.FindThread(ctx, owner, domain.MustParsePhone("+441632960977"), now)
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: liveThread, Owner: owner,
		Remote:    domain.MustParsePhone("+441632960977"),
		Direction: domain.DirectionOutbound, Channel: domain.ChannelSMS,
		Body: "keep me", Status: domain.StatusSent,
		CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := faxes.Create(ctx, domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: owner, Remote: domain.MustParsePhone("+441632960961"),
		Status: domain.FaxTransmitted, DocumentPath: "old-fax-doc", CreatedAt: old, UpdatedAt: old,
	}); err != nil {
		t.Fatal(err)
	}
	if err := faxes.Create(ctx, domain.FaxJob{
		ID: domain.GenerateFaxID(), Owner: owner, Remote: domain.MustParsePhone("+441632960977"),
		Status: domain.FaxTransmitted, DocumentPath: "live-fax-doc", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	res, err := Sweep(ctx, db, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res.Messages != 1 || res.Faxes != 1 || res.Threads != 1 {
		t.Fatalf("sweep counts wrong: %+v", res)
	}
	if len(res.BlobPaths) != 2 {
		t.Fatalf("expected the old attachment + old fax doc paths, got %v", res.BlobPaths)
	}

	// The live thread and its message survive; the swept provider ref
	// is gone (idempotency guard: a second sweep is a no-op).
	if msgs, err := messages.ListMessages(ctx, owner, liveThread, 10); err != nil || len(msgs) != 1 {
		t.Fatalf("live message lost: %v %v", msgs, err)
	}
	if _, err := messages.MessageByProviderRef(ctx, "old-ref"); err == nil {
		t.Fatal("old provider ref must not resolve after the sweep")
	}
	if res2, err := Sweep(ctx, db, now.Add(-24*time.Hour)); err != nil || res2.Messages != 0 || res2.Threads != 0 {
		t.Fatalf("second sweep must be a no-op: %+v %v", res2, err)
	}

	// Empty OLD threads go; empty RECENT threads stay (the thread row
	// is the conversation handle, not content).
	if _, err := db.ExecContext(ctx,
		`INSERT INTO threads (id, owner, remote, last_activity_at, unread) VALUES ('t-old','1001','+441632960999',?,0)`,
		old.Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO threads (id, owner, remote, last_activity_at, unread) VALUES ('t-new','1001','+441632960988',?,0)`,
		now.Unix()); err != nil {
		t.Fatal(err)
	}
	res3, err := Sweep(ctx, db, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res3.Threads != 1 {
		t.Fatalf("only the old empty thread must go: %+v", res3)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM threads WHERE id = 't-new'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("recent empty thread lost: %d %v", count, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM threads WHERE id = 't-old'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("old empty thread survived: %d %v", count, err)
	}
}
