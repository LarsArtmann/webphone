package session

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// openTestDB opens a fresh SQLite file for the session store tests.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "sessions-test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testExtension(t *testing.T) domain.Extension {
	t.Helper()
	ext, err := domain.ParseExtension("1001")
	if err != nil {
		t.Fatal(err)
	}
	return ext
}

// TestSQLiteSessionStoreSurvivesRestart pins the WHOLE point of the
// SQLite-backed store: a session minted before the database handle closed
// is still live on a fresh store over the same file — a service restart
// no longer signs every tab out.
func TestSQLiteSessionStoreSurvivesRestart(t *testing.T) {
	ext := testExtension(t)
	path := filepath.Join(t.TempDir(), "sessions.db")

	boot1, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	store1, err := NewSQLiteStore(boot1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := store1.Create(ext, "dir-pw")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store1.Get(token); !ok {
		t.Fatal("fresh session not live before restart")
	}
	// Simulate the restart: everything in-process dies, only the file
	// survives.
	if err := boot1.Close(); err != nil {
		t.Fatal(err)
	}

	boot2, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = boot2.Close() }()
	store2, err := NewSQLiteStore(boot2, time.Hour)
	if err != nil {
		t.Fatalf("reopen over existing table: %v", err)
	}
	got, ok := store2.Get(token)
	if !ok {
		t.Fatal("session did not survive the restart")
	}
	if got.Extension != ext {
		t.Errorf("session extension = %q, want %q", got.Extension.String(), ext.String())
	}
	if got.Password != "dir-pw" {
		t.Errorf("session password not restored (phone-api proxy would break)")
	}
	if got.ExpiresAt.IsZero() || got.CreatedAt.IsZero() {
		t.Error("timestamps lost across the restart")
	}
}

// TestSQLiteSessionTTLExpiryAndSweep pins the mem-store TTL parity: an
// expired row reads as dead AND is deleted on that read (a session that
// died while the server was down must not resurrect on first read after
// boot), while a session minted after the sweep stays live.
func TestSQLiteSessionTTLExpiryAndSweep(t *testing.T) {
	ext := testExtension(t)
	store, err := NewSQLiteStore(openTestDB(t), 40*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	dying, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(60 * time.Millisecond)

	if _, ok := store.Get(dying); ok {
		t.Fatal("expired session still live — TTL not honored")
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM sessions WHERE token = ?`, dying).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("expired row not deleted on read — it would resurrect on disk until the next sweep")
	}

	// A session minted after the sweep is untouched by the earlier expiry.
	fresh, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Get(fresh); !ok {
		t.Fatal("young session not live after another row was expired")
	}
}

// TestSQLiteSessionSweepOnCreate pins the Create-time sweep: rows that
// expired while the server was DOWN (no reads happened) are gone after
// the next mint.
func TestSQLiteSessionSweepOnCreate(t *testing.T) {
	ext := testExtension(t)
	store, err := NewSQLiteStore(openTestDB(t), 40*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	stale, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(60 * time.Millisecond)
	if _, err := store.Create(ext, "pw"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM sessions WHERE token = ?`, stale).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("expired row survived the Create-time sweep")
	}
}

// TestSQLiteSessionDeleteLogout pins logout: after Delete the token is
// dead and the row is gone.
func TestSQLiteSessionDeleteLogout(t *testing.T) {
	ext := testExtension(t)
	store, err := NewSQLiteStore(openTestDB(t), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}
	store.Delete(token)
	if _, ok := store.Get(token); ok {
		t.Fatal("deleted session still live")
	}
	if _, ok := store.Get("never-existed"); ok {
		t.Error("unknown token answered live")
	}
}

// TestNewSQLiteStoreRejectsBadConstruction pins the constructor guards:
// nil db and non-positive TTL fail loudly instead of panicking at first
// use.
func TestNewSQLiteStoreRejectsBadConstruction(t *testing.T) {
	if _, err := NewSQLiteStore(nil, time.Hour); err == nil {
		t.Error("nil db accepted")
	}
	db := openTestDB(t)
	if _, err := NewSQLiteStore(db, 0); err == nil {
		t.Error("zero TTL accepted")
	}
	if _, err := NewSQLiteStore(db, -time.Hour); err == nil {
		t.Error("negative TTL accepted")
	}
}

// TestSQLiteStoreSatisfiesStoreSeam pins that the SQLite store rides the
// same interface the server wires — the seam is the whole design.
func TestSQLiteStoreSatisfiesStoreSeam(t *testing.T) {
	store, err := NewSQLiteStore(openTestDB(t), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var _ Store = store
}
