// Store family pins (family-adoption train, 2026-09-30): persistence
// failures classify Infrastructure, the absence and list-cap sentinels
// classify Rejection via registration, and the ErrListFull WRAP keeps
// its errors.Is identity (the contacts_api 422 branch depends on it).
package store_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSentinelsClassifyRejection(t *testing.T) {
	errorfamilytest.AssertFamily(t, store.ErrNotFound, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, store.ErrListFull, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, store.ErrSnippetListFull, errorfamily.Rejection)
	errorfamilytest.AssertRetryable(t, store.ErrNotFound, false)
}

func TestListCapWrapKeepsIdentityAndFamily(t *testing.T) {
	owner := domain.MustParseExtension("1001")
	full := errorfamily.WrapRejectionf(store.ErrListFull, "store.contact_list_full", "%s at %d contacts", owner, 500)
	errorfamilytest.AssertFamily(t, full, errorfamily.Rejection)
	if !errors.Is(full, store.ErrListFull) {
		t.Fatal("the list-full wrap must keep errors.Is identity")
	}
	wrapped := fmt.Errorf("save contact: %w", full)
	if !errors.Is(wrapped, store.ErrListFull) {
		t.Fatal("identity must survive service wraps")
	}
	errorfamilytest.AssertFamily(t, wrapped, errorfamily.Rejection)
}

func TestNotFoundJoinsStayNotFound(t *testing.T) {
	// Delete of a missing row answers errors.Join(ErrNotFound, context).
	contacts := store.NewContacts(openTestDB(t))

	err := contacts.Delete(context.Background(), domain.MustParseExtension("1001"), domain.GenerateContactID())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("delete of missing row: want ErrNotFound, got %v", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
}

func TestPersistenceFailuresAreInfrastructure(t *testing.T) {
	// A closed database makes every query fail with the driver's own
	// error; the store's wraps must classify those as Infrastructure.
	db := openTestDB(t)
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	snippets := store.NewSnippets(db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")

	if _, err := messages.ListThreads(ctx, owner); err == nil {
		t.Fatal("ListThreads on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.query")
	}

	if err := messages.MarkThreadRead(ctx, owner, domain.GenerateThreadID()); err == nil {
		t.Fatal("MarkThreadRead on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	if err := faxes.UpdateStatus(ctx, domain.GenerateFaxID(), domain.FaxSending, "", "", 0); err == nil {
		t.Fatal("UpdateStatus on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	if _, err := store.ReadCounts(ctx, db); err == nil {
		t.Fatal("ReadCounts on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	// Reply snippets (M21) joined the persistence surface later; pin the
	// save/delete codes so a refactor cannot drift them out of the family
	// contract the erraudit tier-2 gate reads.
	if err := snippets.Save(ctx, domain.Snippet{ID: domain.GenerateSnippetID(), Owner: owner}); err == nil {
		t.Fatal("Save snippet on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.snippet_save")
	}

	if err := snippets.Delete(ctx, owner, domain.GenerateSnippetID()); err == nil {
		t.Fatal("Delete snippet on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.snippet_delete")
	}

	// Thread organization (M22, T18): the flag toggle + archived count
	// joined later than the file's original pins; same closed-db posture.
	if err := messages.SetThreadFlag(ctx, owner, domain.GenerateThreadID(), store.FlagPinned, true); err == nil {
		t.Fatal("SetThreadFlag on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.thread_flag")
	}

	if _, err := messages.CountArchived(ctx, owner); err == nil {
		t.Fatal("CountArchived on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.count_archived")
	}
}
