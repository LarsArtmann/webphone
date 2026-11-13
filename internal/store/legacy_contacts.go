// Package store — the app-flag seam: the app_flags table is the ONE
// webphone.db home for cursor-style state the journal does not own —
// the matrix/signal sync cursors persist here (app/chat.go) beside the
// retired import markers. The one-time legacy-table importers died with
// the 2026-11-09 unification (7122de92 deleted internal/app/legacy.go),
// so the markers have no reader; the tables themselves stay until the
// storage ladder's final DROP migration
// (docs/planning/2026-11-09_storage-ladder-retirement.md §5,
// operator-gated).
package store

import (
	"context"
	"database/sql"

	"github.com/larsartmann/go-error-family"
)

// FlagLegacyContactsImported marked the one-time legacy-contacts import
// as complete; its importer died with the same unification. The const
// stays for operators inspecting stamped flags.
const FlagLegacyContactsImported = "legacy_contacts_imported"

// FlagThreadsJournalImported marked the one-time threads journal import
// (SUPERB A35) as complete; its importer died with the same
// unification. The const stays for the same reason.
const FlagThreadsJournalImported = "threads_journal_imported"

// FlagPresent reports whether an app_flags row exists for key.
func FlagPresent(ctx context.Context, db *sql.DB, key string) (bool, error) {
	var one int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM app_flags WHERE key = ?`, key).Scan(&one)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errorfamily.WrapInfrastructuref(err, "store.flag_read", "read app flag %s", key)
	}
	return true, nil
}

// FlagValue reads one app_flags row's value (false = absent). The
// matrix sync cursor rides this beside the one-time import markers.
func FlagValue(ctx context.Context, db *sql.DB, key string) (string, bool, error) {
	var value string
	err := db.QueryRowContext(ctx, `SELECT value FROM app_flags WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, errorfamily.WrapInfrastructuref(err, "store.flag_read", "read app flag %s", key)
	}
	return value, true, nil
}

// SetFlag stamps an app_flags row (insert-or-replace).
func SetFlag(ctx context.Context, db *sql.DB, key, value string) error { //nolint:branching-flow // key/value app-flag pair; one call site, literal args visible together (legacy import path, retiring with the window)
	if _, err := db.ExecContext(ctx, `INSERT INTO app_flags (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
		return errorfamily.WrapInfrastructuref(err, "store.flag_write", "set app flag %s", key)
	}
	return nil
}
