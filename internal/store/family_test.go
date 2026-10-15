// Store family pins (family-adoption train, 2026-09-30): persistence
// failures classify Infrastructure, the absence sentinel classifies
// Rejection via registration. The ErrListFull pins moved home to
// internal/readmodels with the contacts read model and the
// ErrSnippetListFull pins with the snippets read model (both SQL
// stores are retired).
package store_test

import (
	"context"
	"database/sql"
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
	errorfamilytest.AssertRetryable(t, store.ErrNotFound, false)
}

func TestPersistenceFailuresAreInfrastructure(t *testing.T) {
	// A closed database makes every query fail with the driver's own
	// error; the store's wraps must classify those as Infrastructure.
	db := openTestDB(t)
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
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
