package store

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"modernc.org/sqlite"
)

// Contract pins for the listRows extraction (2026-09-22 dedup train):
// one home for the query→close→scan→rows.Err() lifecycle, with the
// documented error shape — query AND iteration failures wrap the op,
// scan failures pass the scan function's own (already contextual)
// error through untouched.
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
		if err == nil || !strings.HasPrefix(err.Error(), "walk nums: ") {
			t.Fatalf("query failure must wrap the op, got %v", err)
		}
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
