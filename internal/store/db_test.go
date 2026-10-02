package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"modernc.org/sqlite"
)

// Contract pins for the listRows extraction (2026-09-22 dedup train;
// families added by the 2026-09-30 family-adoption train): one home
// for the query→close→scan→rows.Err() lifecycle, with the documented
// error shape — query AND iteration failures wrap the op and classify
// Infrastructure under the store.query code, scan failures pass the
// scan function's own (already contextual) error through untouched.
func TestListRowsErrorShapes(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE nums (n INTEGER)`); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := db.ExecContext(ctx, `INSERT INTO nums (n) VALUES (7)`); err != nil {
			t.Fatal(err)
		}
	}

	scanNum := func(row rowScanner) (int, error) {
		var n int
		return n, row.Scan(&n)
	}

	t.Run("query failure wraps the op", func(t *testing.T) {
		_, err := listRows(ctx, db, "walk nums", `SELECT * FROM no_such_table`, nil, scanNum)
		if err == nil || !strings.Contains(err.Error(), "walk nums: ") {
			t.Fatalf("query failure must wrap the op, got %v", err)
		}
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.query")
	})

	t.Run("iteration failure wraps the op rows shape", func(t *testing.T) {
		// A scalar UDF that answers the first call and fails the second
		// makes the ITERATION fail deterministically (the query itself
		// succeeds): the failure must surface through the rows.Err()
		// wrap with the op shape, not the scan pass-through.
		calls := 0
		sqlite.MustRegisterScalarFunction("wp_test_boom_second_call", 0,
			func(_ *sqlite.FunctionContext, _ []driver.Value) (driver.Value, error) {
				calls++
				if calls > 1 {
					return nil, errors.New("boom on second call")
				}
				return int64(1), nil
			})

		udfDB, err := Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = udfDB.Close() })
		if _, err := udfDB.ExecContext(ctx, `CREATE TABLE nums (n INTEGER)`); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if _, err := udfDB.ExecContext(ctx, `INSERT INTO nums (n) VALUES (7)`); err != nil {
				t.Fatal(err)
			}
		}

		_, err = listRows(ctx, udfDB, "walk nums", `SELECT wp_test_boom_second_call() FROM nums`, nil, scanNum)
		if err == nil || !strings.Contains(err.Error(), "walk nums rows:") {
			t.Fatalf("iteration failure must carry the op-rows wrap, got %v", err)
		}
		if !strings.Contains(err.Error(), "boom on second call") {
			t.Fatalf("iteration failure must wrap the driver cause, got %v", err)
		}
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
		errorfamilytest.AssertCode(t, err, "store.query")
	})

	t.Run("scan failure passes through unwrapped", func(t *testing.T) {
		_, err := listRows(ctx, db, "walk nums", `SELECT n FROM nums`, nil, func(rowScanner) (int, error) {
			return 0, errScanBoom
		})
		if !errors.Is(err, errScanBoom) {
			t.Fatalf("scan failure must pass the scan error through, got %v", err)
		}
		if strings.Contains(err.Error(), "walk nums") {
			t.Fatalf("scan pass-through must not gain the op wrap, got %q", err.Error())
		}
	})

	t.Run("empty result set is a non-nil empty slice", func(t *testing.T) {
		rows, err := listRows(ctx, db, "walk nums", `SELECT n FROM nums WHERE n = 1`, nil, scanNum)
		if err != nil {
			t.Fatal(err)
		}
		if rows == nil || len(rows) != 0 {
			t.Fatalf("want non-nil empty slice, got %#v", rows)
		}
	})
}

var errScanBoom = errors.New("scan boom")

// staticResult is a sql.Result whose RowsAffected answer is fixed —
// enough to drive updatedOrNotFound without a live statement.
type staticResult int

func (r staticResult) LastInsertId() (int64, error) { return 0, nil }
func (r staticResult) RowsAffected() (int64, error) { return int64(r), nil }

// TestUpdatedOrNotFoundShapes pins the write-side twin of listRows: a
// status-advance UPDATE that matched no row is the caller-visible
// ErrNotFound miss, a matched row is a plain success — so callers can
// map store misses to 404s without string matching.
func TestUpdatedOrNotFoundShapes(t *testing.T) {
	if err := updatedOrNotFound(staticResult(0)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("zero rows affected: %v (want ErrNotFound)", err)
	}
	if err := updatedOrNotFound(staticResult(2)); err != nil {
		t.Fatalf("matched rows: %v (want nil)", err)
	}
}

// TestDatabaseAdapterHealthAndShutdown pins the container adapter's
// contract: HealthCheck pings a live handle (and reports a closed one
// as an Infrastructure failure under the store.ping code), SQL hands
// back the underlying handle, and Shutdown is idempotent — the
// container's shutdown cascade may run it twice.
func TestDatabaseAdapterHealthAndShutdown(t *testing.T) {
	db, err := OpenDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Shutdown() })

	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		t.Fatalf("live database HealthCheck: %v", err)
	}
	if db.SQL() == nil {
		t.Fatal("SQL() returned nil")
	}

	if err := db.Shutdown(); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	if err := db.Shutdown(); err != nil {
		t.Fatalf("second Shutdown must be inert, got: %v", err)
	}

	err = db.HealthCheck(ctx)
	if err == nil {
		t.Fatal("closed database must fail HealthCheck")
	}
	var classified errorfamily.Classified
	if !errors.As(err, &classified) || classified.ErrorFamily() != errorfamily.Infrastructure {
		t.Errorf("closed-database HealthCheck family = %v, want Infrastructure (%v)", classified, err)
	}
}

// The versioned migration chain (T18 design note): fresh and legacy
// databases converge on the same shape-history, the version row is the
// single apply-guard, and reruns never duplicate columns. v1 is the
// duplicate-tolerant baseline (it must absorb pre-versioning databases);
// v2+ are strict and run exactly once.
func TestVersionedMigrations(t *testing.T) {
	t.Run("fresh database walks the whole chain to latest", func(t *testing.T) {
		db, err := Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		assertSchemaVersion(t, db, schemaVersion)
		assertColumns(t, db, "threads", "pinned", "archived", "muted")
		assertColumns(t, db, "messages", "failure_kind", "failure_detail")
		assertTable(t, db, "snippets")
	})

	t.Run("legacy pre-versioning database converges and stamps", func(t *testing.T) {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		// A database from before versioning: the v1-era shape INCLUDING
		// the ad-hoc failure columns, but no schema_version row.
		ctx := context.Background()
		for _, stmt := range []string{
			`CREATE TABLE threads (
				id TEXT PRIMARY KEY, owner TEXT NOT NULL, remote TEXT NOT NULL,
				last_activity_at INTEGER NOT NULL, unread INTEGER NOT NULL DEFAULT 0,
				UNIQUE(owner, remote))`,
			`CREATE TABLE messages (
				id TEXT PRIMARY KEY, thread_id TEXT NOT NULL, owner TEXT NOT NULL,
				remote TEXT NOT NULL, direction TEXT NOT NULL, channel TEXT NOT NULL,
				body TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT '',
				provider_ref TEXT NOT NULL DEFAULT '',
				failure_kind TEXT NOT NULL DEFAULT '', failure_detail TEXT NOT NULL DEFAULT '',
				created_at INTEGER NOT NULL)`,
		} {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				t.Fatal(err)
			}
		}
		if err := migrate(ctx, db); err != nil {
			t.Fatalf("migrate legacy: %v", err)
		}
		assertSchemaVersion(t, db, schemaVersion)
		assertColumns(t, db, "threads", "pinned", "archived", "muted")
		assertTable(t, db, "snippets")
	})

	t.Run("already-latest database is a no-op rerun", func(t *testing.T) {
		db, err := Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		if err := migrate(context.Background(), db); err != nil {
			t.Fatalf("rerun: %v", err)
		}
		assertSchemaVersion(t, db, schemaVersion)
	})

	t.Run("the chain stays consistent with schemaVersion", func(t *testing.T) {
		if len(migrations) == 0 || migrations[len(migrations)-1].version != schemaVersion {
			t.Fatalf("last migration version %d != schemaVersion %d", migrations[len(migrations)-1].version, schemaVersion)
		}
		for i, m := range migrations {
			if i > 0 && m.version != migrations[i-1].version+1 {
				t.Fatalf("migration versions must be contiguous: %d after %d", m.version, migrations[i-1].version)
			}
		}
	})
}

func assertSchemaVersion(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&got); err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if got != want {
		t.Fatalf("schema_version = %d, want %d", got, want)
	}
}

func assertColumns(t *testing.T, db *sql.DB, table string, wants ...string) {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}
	defer rows.Close() //nolint:erraudit // read-only probe; nothing to report on close
	have := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		have[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("table_info iteration: %v", err)
	}
	for _, want := range wants {
		if !have[want] {
			t.Errorf("table %s lacks column %q (has %v)", table, want, have)
		}
	}
}

func assertTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	var name string
	if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name); err != nil {
		t.Fatalf("table %s missing: %v", table, err)
	}
}
