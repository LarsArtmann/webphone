package domain

import "time"

// Contact is a personal contact saved by one extension. Shared contacts
// come from server config and never live in the store.
type Contact struct {
	ID        ContactID
	Owner     Extension
	Name      string
	Phone     Phone
	CreatedAt time.Time
}

// ContactKey identifies one personal-contact slot in the read-model
// store: the (owner, phone) pair a save upserts on. A distinct struct
// type (not a stringly composite) so the metaengine fold key
// extraction is unambiguous — exactly one field of this type per
// read-model fact — and so the pairing stays a named concept instead
// of a separator convention.
type ContactKey struct {
	Owner string
	Phone string
}

// ContactSaved is the read-model fact for a personal-contact upsert.
// The store seam computes Key from the typed contact and preserves the
// existing ID and creation time on rename BEFORE emitting the fact, so
// the fold stays a plain insert. Raw string id by the strong-id wire
// ruling: read-model facts are wire shapes, not domain types.
//brandid-lint:ignore(BD003) read-model fact: wire shape, not domain (strong-id wire ruling)
type ContactSaved struct {
	Key       ContactKey
	ID        string
	Name      string
	CreatedAt time.Time
}

// ContactDeleted is the read-model fact removing one contact slot.
type ContactDeleted struct {
	Key ContactKey
}

// SharedContact is an operator-provided contact from the server config,
// rendered alongside personal ones. The json tags ARE the
// window.PBX_CONFIG wire contract: the island's dial typeahead reads
// lowercase name/number, and the config-file contacts list uses the
// same keys.
// branching-flow DUPLICATE_TYPE (baselined): domain layer of the
// boundary trio (vcard.Card, apiSharedContact); the json tags ARE the
// island/config wire contract (registry 2026-10-05).
type SharedContact struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}
