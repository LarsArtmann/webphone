// The contacts seam's GATE + VIEW surface: the interface the server
// consumes and the projection row the journal backend folds. The
// engine-backed store that used to implement the gate here retired
// with the readmodels engine (2026-11-11 storage-ladder train C) —
// internal/cqrs/contacts.go is the one implementation now.
package readmodels

import (
	"context"
	"errors"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// ContactsMaxPerExtension bounds one extension's personal contacts
// (the notes-cap posture: a NEW contact past the bound is REJECTED,
// never trimmed — renames of existing rows always stay open).
const ContactsMaxPerExtension = 500

// ErrListFull is returned when a NEW contact would exceed the
// per-owner cap; renaming an existing contact stays open.
var ErrListFull = errors.New("contact list full")

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrListFull: errorfamily.Rejection,
	})
}

// ContactView is the projection row ONE journal fold produces for a
// personal contact.
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type ContactView struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	CreatedAt int64  `json:"created_at"`
}

// ContactGate is the personal-contacts seam as the server surfaces see
// it (upsert by (owner, phone), owner-scoped list, the metrics count,
// owner-scoped delete). The journal-backed service (internal/cqrs) is
// the implementation.
type ContactGate interface {
	Save(ctx context.Context, contact domain.Contact) error
	List(ctx context.Context, owner domain.Extension) ([]domain.Contact, error)
	Count(ctx context.Context) (int64, error)
	Delete(ctx context.Context, owner domain.Extension, id domain.ContactID) error
}
