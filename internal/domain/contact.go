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

// SharedContact is an operator-provided contact from the server config,
// rendered alongside personal ones. The json tags ARE the
// window.PBX_CONFIG wire contract: the island's dial typeahead reads
// lowercase name/number, and the config-file contacts list uses the
// same keys.
type SharedContact struct { //nolint:branching-flow // domain layer of the boundary trio (vcard.Card, apiSharedContact); the json tags ARE the island/config wire contract (registry 2026-10-05)
	Name   string `json:"name"`
	Number string `json:"number"`
}
