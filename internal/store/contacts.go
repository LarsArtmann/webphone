package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// Contacts persists per-extension personal contacts.
type Contacts struct {
	db *sql.DB
}

// ContactsMaxPerExtension bounds one extension's personal list. Renames
// (same number) never consume a slot — only a NEW number does; the UI
// renders the whole list, so pagination is deliberately absent (the
// bound is the durability story, not result-windowing).
const ContactsMaxPerExtension = 500

// ErrListFull is returned when a NEW number would exceed the per-owner
// cap; the rename path (upsert on an existing number) stays open.
var ErrListFull = errors.New("contact list full")

// NewContacts builds the contact store.
func NewContacts(db *sql.DB) *Contacts { return &Contacts{db: db} }

// Save upserts a contact by (owner, phone): a second save with the same
// number renames it. A NEW number past ContactsMaxPerExtension inserts
// nothing and returns ErrListFull — one statement, so the cap holds even
// under concurrent saves (no read-then-write race).
func (s *Contacts) Save(ctx context.Context, contact domain.Contact) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO contacts (id, owner, name, phone, created_at)
		SELECT ?, ?, ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM contacts WHERE owner = ? AND phone = ?)
		   OR (SELECT COUNT(*) FROM contacts WHERE owner = ?) < ?
		ON CONFLICT(owner, phone) DO UPDATE SET name = excluded.name
	`, contact.ID.String(), contact.Owner.String(), contact.Name,
		contact.Phone.String(), contact.CreatedAt.Unix(),
		contact.Owner.String(), contact.Phone.String(),
		contact.Owner.String(), ContactsMaxPerExtension)
	if err != nil {
		return fmt.Errorf("save contact: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 { //nolint:erraudit // best-effort read right after the exec whose error is handled above
		return fmt.Errorf("%w: %s at %d contacts", ErrListFull, contact.Owner.String(), ContactsMaxPerExtension)
	}
	return nil
}

// List returns the owner's contacts, newest first.
func (s *Contacts) List(ctx context.Context, owner domain.Extension) ([]domain.Contact, error) {
	return listRows(ctx, s.db, "list contacts", `
		SELECT id, owner, name, phone, created_at
		FROM contacts WHERE owner = ?
		ORDER BY created_at DESC, rowid DESC
	`, []any{owner.String()}, scanContact)
}

func scanContact(row rowScanner) (domain.Contact, error) {
	var (
		id, owner, name, phone string
		created                int64
	)
	if err := row.Scan(&id, &owner, &name, &phone, &created); err != nil {
		return domain.Contact{}, fmt.Errorf("scan contact: %w", err)
	}
	return domain.Contact{
		ID:        domain.MustContactID(id),
		Owner:     domain.MustParseExtension(owner),
		Name:      name,
		Phone:     domain.MustParsePhone(phone),
		CreatedAt: time.Unix(created, 0),
	}, nil
}

// Delete removes one contact scoped to its owner.
func (s *Contacts) Delete(ctx context.Context, owner domain.Extension, id domain.ContactID) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM contacts WHERE id = ? AND owner = ?
	`, id.String(), owner.String())
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 { //nolint:erraudit // best-effort write; the response is already committed
		return errors.Join(ErrNotFound, fmt.Errorf("contact %s", id.String()))
	}
	return nil
}
