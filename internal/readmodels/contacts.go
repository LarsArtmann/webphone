package readmodels

import (
	"context"
	"errors"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// ContactsMaxPerExtension bounds one extension's personal list (same
// contract the SQL store enforced): renames never consume a slot, only
// a NEW number does; the bound is the durability story, not result
// windowing.
const ContactsMaxPerExtension = 500

// ErrListFull is returned when a NEW number would exceed the per-owner
// cap — the same contract the SQL store held before the read-model
// swap homed the sentinel here (the store's contacts table left with
// its file, the contract stays).
var ErrListFull = errors.New("contact list full")

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrListFull: errorfamily.Rejection,
	})
}

// ContactView is the engine-side projection of a personal contact.
// Unix-nanos created_at keeps list ordering EXACTLY numeric in sqlite
// (the SQL store compared unix seconds with a rowid tiebreak; nanos
// make ties virtually impossible and the order stable).
//
// branching-flow DUPLICATE_TYPE (baselined): the engine projection of
// the retired SQL row (parity port); the json tags ARE the engine's
// storage schema, not a DTO to consolidate with the domain type.
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type ContactView struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	CreatedAt int64  `json:"created_at"`
}

// contactsCollection names the engine collection (query name).
const contactsCollection = "contacts"

type contactsInput struct{ Owner string }

// contactsQuery declares the contacts read model: a keyed projection
// upserted by ContactSaved facts, removed by ContactDeleted facts,
// filtered by owner and sorted newest-first. The insert fold is the
// single ContactSaved fold on purpose — the upsert-keep-id semantics
// (rename preserves ID and creation time) live in the store seam,
// which pre-reads the existing view under its mutex before emitting.
var contactsQuery = metaengine.Query[contactsInput, ContactView](contactsCollection,
	metaengine.OnRecord(domain.ContactSaved{},
		func(_ record.Record, e domain.ContactSaved) (domain.ContactKey, ContactView) {
			return e.Key, ContactView{
				ID:        e.ID,
				Owner:     e.Key.Owner,
				Name:      e.Name,
				Phone:     e.Key.Phone,
				CreatedAt: e.CreatedAt.UnixNano(),
			}
		}),
	metaengine.OnRecord(domain.ContactDeleted{}, metaengine.Remove[ContactView]()),
	metaengine.FilterOnField[ContactView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[ContactView]("created_at", true),
	metaengine.Volume(humanScaleVolume),
)

// Contacts is the engine-backed personal-contacts store: the same
// contract the SQL store held (Save upsert by (owner, phone) keeping
// the old id and creation time on rename; owner-scoped newest-first
// List; owner-scoped Delete by id; per-owner cap for NEW numbers).
//
// mu serializes check-then-apply sequences (cap check, rename
// pre-read): the binary is single-process by doctrine, so a
// process-wide mutex restores the atomicity the SQL single-statement
// cap used to provide. Engine writes themselves are already
// serialized by the single sqlite connection.
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+agg+mu, see Faxes/Snippets) — deliberate
// symmetry, not a mixin candidate.
type Contacts struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[ContactView]
	agg    metaengine.AggregateReader
	mu     sync.Mutex
}

func newContacts(models *metaengine.Store, agg metaengine.AggregateReader) *Contacts {
	return &Contacts{
		models: models,
		reader: metaengine.NewReader[ContactView](models, contactsCollection),
		agg:    agg,
	}
}

// Save upserts a contact by (owner, phone): an existing slot is
// renamed keeping its ID and creation time; a NEW number past the
// per-owner cap is refused with this package's ErrListFull contract.
func (s *Contacts) Save(ctx context.Context, contact domain.Contact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := domain.ContactKey{Owner: contact.Owner.String(), Phone: contact.Phone.String()}

	prev, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.contact_save", "load contact slot")
	}
	if found {
		// Rename: identity (id) and age stay with the slot.
		contact.ID = domain.MustContactID(prev.ID)
		contact.CreatedAt = time.Unix(0, prev.CreatedAt)
	} else if err := s.ownerHasRoom(ctx, contact.Owner); err != nil {
		return err
	}

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.ContactSaved{}), domain.ContactSaved{
		Key:       key,
		ID:        contact.ID.String(),
		Name:      contact.Name,
		CreatedAt: contact.CreatedAt,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.contact_save", "apply contact save")
	}
	return nil
}

// ownerHasRoom rejects a NEW number when the owner is at the cap.
func (s *Contacts) ownerHasRoom(ctx context.Context, owner domain.Extension) error {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return err
	}
	if len(views) >= ContactsMaxPerExtension {
		return errorfamily.WrapRejectionf(ErrListFull, "readmodels.contact_list_full",
			"%s at %d contacts", owner.String(), ContactsMaxPerExtension)
	}
	return nil
}

// List returns the owner's contacts, newest first.
func (s *Contacts) List(ctx context.Context, owner domain.Extension) ([]domain.Contact, error) {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Contact, 0, len(views))
	for _, v := range views {
		out = append(out, domain.Contact{
			ID:        domain.MustContactID(v.ID),
			Owner:     domain.MustParseExtension(v.Owner),
			Name:      v.Name,
			Phone:     domain.MustParsePhone(v.Phone),
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	return out, nil
}

func (s *Contacts) ownerViews(ctx context.Context, owner domain.Extension) ([]ContactView, error) {
	// WithLimit is load-bearing: an unlimited engine scan silently
	// ceilings at the planner's default (100 rows), and both the
	// user-facing list (the whole list renders) and the cap check need
	// every row the owner owns — the per-owner cap bounds this at 500.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(ContactsMaxPerExtension+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.contact_list", "list contacts for %s", owner.String())
	}
	return views, nil
}

// Count reports the total stored contacts across all owners (the
// metrics surface's webphone_contacts_total). Served by the engine's
// native COUNT aggregate — no row materialization.
func (s *Contacts) Count(ctx context.Context) (int64, error) {
	n, err := s.agg.Aggregate(ctx, contactsCollection, metaengine.AggregateCount, "", nil)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "readmodels.contact_count", "count contacts")
	}
	return int64(n), nil
}

// Delete removes one contact scoped to its owner. The id-to-slot
// resolution rides the owner scan (the cap bounds it at 500 rows).
func (s *Contacts) Delete(ctx context.Context, owner domain.Extension, id domain.ContactID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return err
	}
	for _, v := range views {
		if v.ID != id.String() {
			continue
		}
		if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.ContactDeleted{}), domain.ContactDeleted{
			Key: domain.ContactKey{Owner: v.Owner, Phone: v.Phone},
		}); err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.contact_delete", "apply contact delete")
		}
		return nil
	}
	return notFound(errorfamily.Newf(errorfamily.Rejection, "readmodels.contact_missing", "contact %s", id.String()))
}
