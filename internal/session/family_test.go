// Session family pins (family-adoption train, 2026-09-30): store
// wiring and persistence failures classify Infrastructure — a session
// problem is never the user's fault and never retryable by them.
package session_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/session"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

func TestSessionStoreFailuresAreInfrastructure(t *testing.T) {
	if _, err := session.NewSQLiteStore(nil, time.Hour); err == nil {
		t.Fatal("nil db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "session.nil_db")
	}

	freshDB, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "family-test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = freshDB.Close() })
	if _, err := session.NewSQLiteStore(freshDB, 0); err == nil {
		t.Fatal("zero ttl: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "session.ttl")
	}

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "family-live.db"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.NewSQLiteStore(db, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(domain.MustParseExtension("1001"), "pw"); err == nil {
		t.Fatal("Create on closed db: want error")
	} else {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}
}
