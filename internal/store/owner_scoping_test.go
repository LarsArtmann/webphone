// Owner-scoping behavior suite (plan T15, 2026-09-20): EVERY query that
// takes an owner must filter by it — a scoping regression in this
// package leaks one extension's messages, faxes or contacts to another.
// Webhook lookups (MessageByProviderRef, fax ByProviderRef,
// UpdateOutboundStatus/UpdateStatus) are deliberately NOT owner-scoped:
// the provider callback carries no extension context, they are scoped by
// the provider ref instead (unguessable, direction-pinned).
package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

type ownerScopeFixture struct {
	messages *Messages
	faxes    *Faxes
	contacts *Contacts
	owner    domain.Extension
	foreign  domain.Extension
	threadID domain.ThreadID
	faxID    domain.FaxID
	contact  domain.Contact
}

func newOwnerScopeFixture(t *testing.T) *ownerScopeFixture {
	t.Helper()
	messages, faxes, contacts := newTestDB(t)
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	foreign := domain.MustParseExtension("1002")
	remote := domain.MustParsePhone("+441632960961")
	now := time.Now()

	threadID, err := messages.FindThread(ctx, owner, remote, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := messages.AppendMessage(ctx, domain.Message{
		ID: domain.GenerateMessageID(), ThreadID: threadID, Owner: owner, Remote: remote,
		Direction: domain.DirectionInbound, Channel: domain.ChannelSMS,
		Body: "scoped body", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	faxID := domain.GenerateFaxID()
	if err := faxes.Create(ctx, domain.FaxJob{
		ID: faxID, Owner: owner, Remote: remote, Direction: domain.FaxOutbound,
		Status: domain.FaxQueued, Pages: 1, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	contact := domain.Contact{
		ID: domain.GenerateContactID(), Owner: owner, Name: "Scoped",
		Phone: remote, CreatedAt: now,
	}
	if err := contacts.Save(ctx, contact); err != nil {
		t.Fatal(err)
	}

	return &ownerScopeFixture{
		messages: messages, faxes: faxes, contacts: contacts,
		owner: owner, foreign: foreign, threadID: threadID, faxID: faxID, contact: contact,
	}
}

// TestOwnerScopedReads pins every READ path: the legitimate owner sees
// the row, a foreign extension gets not-found or empty — never the row.
func TestOwnerScopedReads(t *testing.T) {
	f := newOwnerScopeFixture(t)
	ctx := context.Background()

	t.Run("GetThread", func(t *testing.T) {
		if _, err := f.messages.GetThread(ctx, f.owner, f.threadID); err != nil {
			t.Fatal(err)
		}
		if _, err := f.messages.GetThread(ctx, f.foreign, f.threadID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("foreign GetThread: %v (want ErrNotFound)", err)
		}
	})

	t.Run("ListThreads", func(t *testing.T) {
		ownerThreads, err := f.messages.ListThreads(ctx, f.owner)
		if err != nil || len(ownerThreads) != 1 {
			t.Fatalf("owner ListThreads: %d threads, err %v", len(ownerThreads), err)
		}
		foreignThreads, err := f.messages.ListThreads(ctx, f.foreign)
		if err != nil || len(foreignThreads) != 0 {
			t.Fatalf("foreign ListThreads leaked %d threads (err %v)", len(foreignThreads), err)
		}
	})

	t.Run("ListMessages", func(t *testing.T) {
		msgs, err := f.messages.ListMessages(ctx, f.owner, f.threadID, 10)
		if err != nil || len(msgs) != 1 {
			t.Fatalf("owner ListMessages: %d msgs, err %v", len(msgs), err)
		}
		foreignMsgs, err := f.messages.ListMessages(ctx, f.foreign, f.threadID, 10)
		if err != nil || len(foreignMsgs) != 0 {
			t.Fatalf("foreign ListMessages leaked %d msgs (err %v)", len(foreignMsgs), err)
		}
	})

	t.Run("ListMessagesPage", func(t *testing.T) {
		foreignMsgs, _, err := f.messages.ListMessagesPage(ctx, f.foreign, f.threadID, 0, 10)
		if err != nil || len(foreignMsgs) != 0 {
			t.Fatalf("foreign ListMessagesPage leaked %d msgs (err %v)", len(foreignMsgs), err)
		}
	})

	t.Run("FaxGet", func(t *testing.T) {
		if _, err := f.faxes.Get(ctx, f.owner, f.faxID); err != nil {
			t.Fatal(err)
		}
		if _, err := f.faxes.Get(ctx, f.foreign, f.faxID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("foreign fax Get: %v (want ErrNotFound)", err)
		}
	})

	t.Run("FaxList", func(t *testing.T) {
		foreignFaxes, err := f.faxes.List(ctx, f.foreign, 10)
		if err != nil || len(foreignFaxes) != 0 {
			t.Fatalf("foreign fax List leaked %d jobs (err %v)", len(foreignFaxes), err)
		}
	})

	t.Run("ContactList", func(t *testing.T) {
		foreignContacts, err := f.contacts.List(ctx, f.foreign)
		if err != nil || len(foreignContacts) != 0 {
			t.Fatalf("foreign contact List leaked %d rows (err %v)", len(foreignContacts), err)
		}
	})
}

// TestOwnerScopedMutations pins every WRITE path that takes an owner:
// a foreign extension's mutation is an inert no-op — it must neither
// change the row nor error in a way that leaks existence.
func TestOwnerScopedMutations(t *testing.T) {
	f := newOwnerScopeFixture(t)
	ctx := context.Background()

	t.Run("MarkThreadRead foreign is inert", func(t *testing.T) {
		if err := f.messages.MarkThreadRead(ctx, f.foreign, f.threadID); err != nil {
			t.Fatal(err)
		}
		thread, err := f.messages.GetThread(ctx, f.owner, f.threadID)
		if err != nil {
			t.Fatal(err)
		}
		if thread.Unread != 1 {
			t.Fatalf("foreign mark-read mutated the owner's row: unread=%d", thread.Unread)
		}
	})

	t.Run("ContactDelete foreign is not-found", func(t *testing.T) {
		err := f.contacts.Delete(ctx, f.foreign, f.contact.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("foreign contact Delete: %v (want ErrNotFound)", err)
		}
		if _, err := f.contacts.List(ctx, f.owner); err != nil {
			t.Fatal(err)
		}
		ownerContacts, _ := f.contacts.List(ctx, f.owner)
		if len(ownerContacts) != 1 {
			t.Fatal("owner's contact vanished without a scoped delete")
		}
	})
}
