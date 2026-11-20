package store

import (
	"context"
	"database/sql"
	"testing"
)

// The legacy tables stay in webphone.db (the migration ladder never
// rewrites shipped shape; docs/planning/2026-11-09_storage-ladder-retirement.md
// §4 rows the whole frozen family for the operator-gated final DROP).
// The transitional BOOT importers are gone since the 2026-11-09
// unification (7122de92 deleted internal/app/legacy.go) and the sibling
// legacy readers (contacts/faxes/snippets/threads/transcripts) are
// DELETED 2026-11-14 (no production caller; the export-or-confirm tool
// reads raw sqlite per its design memo) — the remaining pins cover the
// app-flag helpers the sync cursors ride.

func seedAppFlag(t *testing.T, db *sql.DB, key, value string) {
	t.Helper()
	if err := SetFlag(context.Background(), db, key, value); err != nil {
		t.Fatalf("SetFlag(%q): %v", key, err)
	}
}

// openFlagDB boots the in-package test handle (family_test's
// openTestDB lives in the store_test package and is invisible here).
func openFlagDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestAppFlagsRoundTripAndAbsence(t *testing.T) {
	db := openFlagDB(t)
	ctx := context.Background()

	if present, ok, err := FlagValue(ctx, db, FlagLegacyContactsImported); err != nil || ok || present != "" {
		t.Fatalf("absent value: %q/%v/%v, want empty/false/nil", present, ok, err)
	}

	seedAppFlag(t, db, FlagLegacyContactsImported, "2026-11-07")
	if value, ok, err := FlagValue(ctx, db, FlagLegacyContactsImported); err != nil || !ok || value != "2026-11-07" {
		t.Fatalf("value = %q/%v/%v, want the stamped date", value, ok, err)
	}

	// SetFlag is insert-or-replace: the matrix sync cursor is rewritten every
	// poll, so a second write must UPDATE, not fail on the primary key.
	seedAppFlag(t, db, FlagLegacyContactsImported, "2026-11-08")
	if value, _, err := FlagValue(ctx, db, FlagLegacyContactsImported); err != nil || value != "2026-11-08" {
		t.Fatalf("after rewrite: %q/%v, want the newer value", value, err)
	}
}
