package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// Legacy-contacts seam (2026-10-15): the personal-contacts truth moved
// to the metaengine read models (readmodels.db). The SQL contacts table
// stays untouched in webphone.db (the migration ladder never rewrites
// shipped shape), so an upgraded deployment's contacts ride this ONE
// transitional read at boot: the composition root imports them into the
// engine store once and stamps the app_flags row. This file exists to
// be deleted when the transitional window closes.

// FlagLegacyContactsImported marks the one-time legacy-contacts import
// as complete (value carries the import date for operators).
const FlagLegacyContactsImported = "legacy_contacts_imported"

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

// LegacyContact is one row of the retired SQL contacts table, kept as
// plain strings: import-time parsing decides what survives.
//
// branching-flow DUPLICATE_TYPE (baselined): the parse-layer row of the
// RETIRED table — typing it would dress dead storage as domain (delete
// with the legacy import window, TODO_LIST).
// brandid-lint:ignore(BD003) retired SQL row kept as plain strings; import-time parsing decides what survives
type LegacyContact struct {
	ID        string
	Owner     string
	Name      string
	Phone     string
	CreatedAt int64
}

// ReadLegacyContacts reads the whole retired table oldest-first (the
// engine keeps newest-first lists; importing in order preserves the
// legacy ordering semantics).
func ReadLegacyContacts(ctx context.Context, db *sql.DB) ([]LegacyContact, error) {
	return listRows(ctx, db, "read legacy contacts", `
		SELECT id, owner, name, phone, created_at
		FROM contacts
		ORDER BY created_at ASC, rowid ASC
	`, nil, func(row rowScanner) (LegacyContact, error) {
		var c LegacyContact
		return c, row.Scan(&c.ID, &c.Owner, &c.Name, &c.Phone, &c.CreatedAt)
	})
}

// ToContact validates one legacy row into a domain.Contact. An
// unparseable owner, phone, or id demotes to a freshly minted id where
// possible; a row whose owner or phone does not parse is skipped (the
// caller logs it) — a corrupt legacy row must not brick a boot.
func (c LegacyContact) ToContact() (domain.Contact, bool) {
	owner, err := domain.ParseExtension(c.Owner)
	if err != nil {
		return domain.Contact{}, false
	}
	phone, err := domain.ParsePhone(c.Phone)
	if err != nil {
		return domain.Contact{}, false
	}
	id, err := domain.ParseContactID(c.ID)
	if err != nil {
		id = domain.GenerateContactID()
	}
	return domain.Contact{
		ID:        id,
		Owner:     owner,
		Name:      c.Name,
		Phone:     phone,
		CreatedAt: time.Unix(c.CreatedAt, 0),
	}, true
}
