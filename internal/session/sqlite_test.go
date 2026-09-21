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

// TestSQLiteSessionRenewExtendsAndCaps pins the persisted sliding
// renewal: a past-half-life row's new expiry lands in the DB (not just
// in the return value), the throttle suppresses young renewals, a
// session older than its absolute cap can never renew again, and
// expired/unknown tokens read dead.
func TestSQLiteSessionRenewExtendsAndCaps(t *testing.T) {
	ext := testExtension(t)
	store, err := NewSQLiteStore(openTestDB(t), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	const idle = time.Hour
	const maxAge = 30 * 24 * time.Hour

	token, err := store.Create(ext, "pw")
	if err != nil {
		t.Fatal(err)
	}

	// Backdate to 20m remaining: past the half-life, so the first renew
	// extends to (approximately) the full idle window — and persists it.
	if _, err := store.db.Exec(
		`UPDATE sessions SET expires_at = ? WHERE token = ?`,
		time.Now().Add(20*time.Minute).UnixMilli(), token,
	); err != nil {
		t.Fatal(err)
	}
	renewed, ok := store.Renew(token, idle, maxAge)
	if !ok {
		t.Fatal("past-half-life session did not renew")
	}
	if until := time.Until(renewed.ExpiresAt); until < 55*time.Minute || until > idle {
		t.Errorf("renewed expiry %s from now, want ~the full idle window", until)
	}
	persisted, ok := store.Get(token)
	if !ok {
		t.Fatal("session lost after renew")
	}
	// Stamps are unix millis: compare with that granularity.
	if drift := persisted.ExpiresAt.Sub(renewed.ExpiresAt); drift.Abs() > 2*time.Millisecond {
		t.Errorf("extension not persisted: got %s, want %s", persisted.ExpiresAt, renewed.ExpiresAt)
	}

	// Young again: the throttle keeps the immediate second renew off.
	if _, ok := store.Renew(token, idle, maxAge); ok {
		t.Error("young session renewed again — the write throttle is broken")
	}

	// A session whose absolute cap is already in the past (older than
	// created+max) can never renew: the cap is the unconditional bound.
	if _, err := store.db.Exec(
		`UPDATE sessions SET created_at = ?, expires_at = ? WHERE token = ?`,
		time.Now().Add(-40*24*time.Hour).UnixMilli(), time.Now().Add(20*time.Minute).UnixMilli(), token,
	); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Renew(token, idle, maxAge); ok {
		t.Error("session older than its absolute cap renewed")
	}

	// Expired rows read dead on Renew (and are deleted, like on Get).
	if _, err := store.db.Exec(
		`UPDATE sessions SET created_at = ?, expires_at = ? WHERE token = ?`,
		time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(-time.Minute).UnixMilli(), token,
	); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Renew(token, idle, maxAge); ok {
		t.Error("expired session renewed")
	}
	if _, ok := store.Get(token); ok {
		t.Error("expired row survived the Renew read")
	}

	if _, ok := store.Renew("never-existed", idle, maxAge); ok {
		t.Error("unknown token renewed")
	}
}
