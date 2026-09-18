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
// rendered alongside personal ones.
type SharedContact struct {
	Name   string
	Number string
}
