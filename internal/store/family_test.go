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
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if _, err := store.CountSessions(ctx, db); err == nil {
		t.Fatal("CountSessions on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	if _, _, err := store.FlagValue(ctx, db, "any-key"); err == nil {
		t.Fatal("FlagValue on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.flag_read")
	}
}
