// Store family pins (family-adoption train, 2026-09-30): persistence
// failures classify Infrastructure, the absence and snippet-cap
// sentinels classify Rejection via registration. The contacts
// ErrListFull pins moved home to internal/readmodels with the contacts
// read model (the SQL contacts store is retired).
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
	errorfamilytest.AssertFamily(t, store.ErrSnippetListFull, errorfamily.Rejection)
	errorfamilytest.AssertRetryable(t, store.ErrNotFound, false)
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
