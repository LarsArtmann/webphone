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

// NewContacts builds the contact store.
func NewContacts(db *sql.DB) *Contacts { return &Contacts{db: db} }

// Save upserts a contact by (owner, phone): a second save with the same
// number renames it.
func (s *Contacts) Save(ctx context.Context, contact domain.Contact) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contacts (id, owner, name, phone, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(owner, phone) DO UPDATE SET name = excluded.name
	`, contact.ID.String(), contact.Owner.String(), contact.Name,
		contact.Phone.String(), contact.CreatedAt.Unix())
	if err != nil {
		return fmt.Errorf("save contact: %w", err)
	}
	return nil
}

// List returns the owner's contacts, newest first.
func (s *Contacts) List(ctx context.Context, owner domain.Extension) ([]domain.Contact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner, name, phone, created_at
		FROM contacts WHERE owner = ?
		ORDER BY created_at DESC, rowid DESC
	`, owner.String())
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	contacts := make([]domain.Contact, 0)
	for rows.Next() {
		var (
			id, owner, name, phone string
			created                int64
		)
		if err := rows.Scan(&id, &owner, &name, &phone, &created); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, domain.Contact{
			ID:        domain.MustContactID(id),
			Owner:     domain.MustParseExtension(owner),
			Name:      name,
			Phone:     domain.MustParsePhone(phone),
			CreatedAt: time.Unix(created, 0),
		})
	}

	return contacts, rows.Err()
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
