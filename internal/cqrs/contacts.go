// The contacts aggregate (SUPERB Phase 4, A26): the personal-contacts
// list as a per-owner stream. The 500-per-owner cap for NEW numbers,
// the rename-keeps-id-and-age semantics, and the delete-by-id
// resolution now answer from the DECISION (the old store pre-read rows
// under a mutex; the decider simply knows its state). The never-sweep
// doctrine (D3) is structural: contacts emit no trim events of any
// kind. ErrListFull survives the swap verbatim, and Save/Delete await
// projection visibility (D7) because the Contacts tab partial
// re-renders from the projection in the same request.
package cqrs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// ContactsStreamType keys per-owner contact streams.
const ContactsStreamType = "contacts"

// contactsMaxPerOwner mirrors readmodels.ContactsMaxPerExtension at the
// aggregate.
const contactsMaxPerOwner = readmodels.ContactsMaxPerExtension

// SaveContact upserts a contact by (owner, phone): the decision keeps
// an existing slot's ID and creation time (rename) and refuses a NEW
// number past the cap.
type SaveContact struct {
	*command.BasicCommand
	Contact domain.Contact
}

// DeleteContact removes one contact by id, owner-scoped; an unknown id
// is the missing rejection (the handler's 404 mapping is unchanged).
type DeleteContact struct {
	*command.BasicCommand
	Owner     string
	ContactID domain.ContactID
}

// contactSlot is one contact's aggregate state.
type contactSlot struct {
	ID        string
	Name      string
	CreatedAt int64 // unix nanos — the list's exact numeric order
}

// contactsState is the owner's contact list keyed by phone (the upsert
// axis — a rename rewrites the same slot).
type contactsState struct {
	Slots map[string]contactSlot
}

func (contactsState) initial() contactsState { return contactsState{Slots: map[string]contactSlot{}} }

// contactStreamEventTypes is the aggregate's stream vocabulary: the
// domain decl's Events entry AND the StrictApply guard list (an unknown
// type on the stream fails replay loudly instead of silently skipping).
var contactStreamEventTypes = []event.Type{
	event.Type(metaengine.EventTypeName(domain.ContactSaved{})),
	event.Type(metaengine.EventTypeName(domain.ContactDeleted{})),
}

func applyContactsState(s contactsState, evt event.Event) (contactsState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.ContactSaved{})):
		fact, err := event.DecodePayloadAuto[domain.ContactSaved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.contact_fold", "decode ContactSaved")
		}
		s.Slots[fact.Key.Phone] = contactSlot{ID: fact.ID, Name: fact.Name, CreatedAt: fact.CreatedAt.UnixNano()}
	case event.Type(metaengine.EventTypeName(domain.ContactDeleted{})):
		fact, err := event.DecodePayloadAuto[domain.ContactDeleted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.contact_fold", "decode ContactDeleted")
		}
		delete(s.Slots, fact.Key.Phone)
	}
	return s, nil
}

func decideSaveContact(s contactsState, ver event.Version, cmd SaveContact) ([]event.Event, error) {
	owner := cmd.Contact.Owner.String()
	phone := cmd.Contact.Phone.String()
	id := cmd.Contact.ID.String()
	createdAt := cmd.Contact.CreatedAt
	if prev, ok := s.Slots[phone]; ok {
		// Rename: identity (id) and age stay with the slot.
		id = prev.ID
		createdAt = time.Unix(0, prev.CreatedAt)
	} else if len(s.Slots) >= contactsMaxPerOwner {
		return nil, errorfamily.WrapRejectionf(readmodels.ErrListFull, "cqrs.contact_list_full",
			"%s at %d contacts", owner, contactsMaxPerOwner)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ContactSaved{})), cmd.StreamID(), ContactsStreamType, ver.Increment(), domain.ContactSaved{
		Key:       domain.ContactKey{Owner: owner, Phone: phone},
		ID:        id,
		Name:      cmd.Contact.Name,
		CreatedAt: createdAt,
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.contact_save", "build ContactSaved")
	}
	return []event.Event{evt}, nil
}

func decideDeleteContact(s contactsState, ver event.Version, cmd DeleteContact) ([]event.Event, error) {
	for phone, slot := range s.Slots {
		if slot.ID != cmd.ContactID.String() {
			continue
		}
		evt, err := event.New(event.Type(metaengine.EventTypeName(domain.ContactDeleted{})), cmd.StreamID(), ContactsStreamType, ver.Increment(), domain.ContactDeleted{
			Key: domain.ContactKey{Owner: cmd.Owner, Phone: phone},
		}, event.WithSchemaVersion(1))
		if err != nil {
			return nil, errorfamily.WrapInfrastructuref(err, "cqrs.contact_delete", "build ContactDeleted")
		}
		return []event.Event{evt}, nil
	}
	return nil, errorfamily.NewRejection("cqrs.contact_missing", fmt.Sprintf("contact %s", cmd.ContactID.String()))
}

// RegisterContacts wires the contacts aggregate into a System.
func RegisterContacts(sys *system.System) error {
	if err := system.RegisterDecider[contactsState](sys, ContactsStreamType, decider.Decider[contactsState]{
		Initial: contactsState{}.initial(),
		Apply:   decider.StrictApply(applyContactsState, contactStreamEventTypes),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.contact_register", "register contacts decider")
	}
	if err := system.RegisterCommand[*SaveContact, contactsState](sys, "contact.save",
		func(_ context.Context, cmd *SaveContact) system.Op[contactsState] {
			return system.Execute[contactsState](context.Background(), cmd.StreamID(), ContactsStreamType,
				func(state contactsState, ver event.Version) ([]event.Event, error) {
					return decideSaveContact(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.contact_register", "register contact.save")
	}
	if err := system.RegisterCommand[*DeleteContact, contactsState](sys, "contact.delete",
		func(_ context.Context, cmd *DeleteContact) system.Op[contactsState] {
			return system.Execute[contactsState](context.Background(), cmd.StreamID(), ContactsStreamType,
				func(state contactsState, ver event.Version) ([]event.Event, error) {
					return decideDeleteContact(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.contact_register", "register contact.delete")
	}
	return nil
}

// contactsKeyInput is the projection's key-extraction input (the same
// shape the readmodels engine declares).
type contactsKeyInput struct {
	Owner string
	Phone string
}

// contactsProjection folds contact facts into the SAME view shape the
// server reads today (readmodels.ContactView).
func contactsProjection() metaengine.QueryDecl[contactsKeyInput, readmodels.ContactView] {
	return metaengine.Query[contactsKeyInput, readmodels.ContactView](contactsCollection,
		metaengine.OnRecord(domain.ContactSaved{},
			func(_ record.Record, e domain.ContactSaved) (domain.ContactKey, readmodels.ContactView) {
				return e.Key, readmodels.ContactView{
					ID:        e.ID,
					Owner:     e.Key.Owner,
					Name:      e.Name,
					Phone:     e.Key.Phone,
					CreatedAt: e.CreatedAt.UnixNano(),
				}
			}),
		metaengine.OnRecord(domain.ContactDeleted{}, metaengine.Remove[readmodels.ContactView]()),
		metaengine.FilterOnField[readmodels.ContactView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.ContactView]("created_at", true),
	)
}

// contactsCollection names the system engine's projection (the same
// name the readmodels engine declares — different databases, one
// shape).
const contactsCollection = "contacts"

// ContactsDomainDecl is the contacts aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and
// panics loudly).
func ContactsDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterContacts(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register contacts aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(contactsProjection())},
		Events: contactStreamEventTypes,
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Contacts is the dispatch-backed personal-contacts store: the same
// contract the engine-backed store exposed (Save upsert keeping id and
// age on rename, newest-first owner-scoped List, owner-scoped Delete
// by id with the missing rejection, engine-native Count).
type Contacts struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.ContactView]
	vers   streamVersions
	mu     sync.Mutex
}

var _ readmodels.ContactGate = (*Contacts)(nil)

// NewContacts builds the backend over a System.
func NewContacts(sys *system.System) *Contacts {
	return &Contacts{sys: sys, reader: metaengine.NewReader[readmodels.ContactView](sys.MetaEngine(), contactsCollection), vers: streamVersions{}}
}

// awaitOwnWrite waits until the contact projection has folded the
// just-dispatched batch (D7): the Contacts tab partial re-renders from
// the projection in the SAME request. Best-effort — a tail-read
// failure WARNs (the write committed) and never fails the request.
func (s *Contacts) awaitOwnWrite(ctx context.Context, owner string) {
	awaitOwnWrite(ctx, s.sys, ContactsStreamType, owner, contactsCollection, s.vers)
}

// Save upserts a contact by (owner, phone): an existing slot is renamed
// keeping its ID and creation time; a NEW number past the per-owner cap
// is refused with the readmodels ErrListFull contract.
func (s *Contacts) Save(ctx context.Context, contact domain.Contact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "contact.save", contact.Owner.String(),
		func(base *command.BasicCommand) *SaveContact {
			return &SaveContact{BasicCommand: base, Contact: contact}
		}); err != nil {
		if errors.Is(err, readmodels.ErrListFull) {
			return err
		}
		return errorfamily.WrapInfrastructuref(err, "cqrs.contact_save", "dispatch contact.save")
	}
	s.awaitOwnWrite(ctx, contact.Owner.String())
	return nil
}

// List returns the owner's contacts, newest first.
func (s *Contacts) List(ctx context.Context, owner domain.Extension) ([]domain.Contact, error) {
	// WithLimit is load-bearing: the whole list renders and the cap
	// bounds this scan at 500 rows.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(contactsMaxPerOwner+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.contact_list", "list contacts for %s", owner.String())
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

// Count reports the total stored contacts across all owners (the
// metrics surface's webphone_contacts_total) via the engine's native
// COUNT aggregate.
func (s *Contacts) Count(ctx context.Context) (int64, error) {
	n, err := engineCount(ctx, s.sys, contactsCollection)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.contact_count", "count contacts")
	}
	return n, nil
}

// Delete removes one contact scoped to its owner; an unknown id
// answers the missing rejection (the handler's 404 mapping).
func (s *Contacts) Delete(ctx context.Context, owner domain.Extension, id domain.ContactID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "contact.delete", owner.String(),
		func(base *command.BasicCommand) *DeleteContact {
			return &DeleteContact{BasicCommand: base, Owner: owner.String(), ContactID: id}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, owner.String())
	return nil
}
