package domain

import "time"

// SnippetKey identifies one reply-snippet slot in the read-model
// store: the snippet id a save upserts on. A distinct struct type (not
// a bare string) mirrors ContactKey so the metaengine fold key
// extraction stays unambiguous — exactly one field of this type per
// read-model fact.
type SnippetKey struct {
	ID string
}

// SnippetSaved is the read-model fact for a reply-snippet upsert. The
// store seam computes Key from the typed snippet and preserves the
// existing creation time on replace BEFORE emitting the fact, so the
// fold stays a plain insert. Raw string id by the strong-id wire
// ruling: read-model facts are wire shapes, not domain types.
type SnippetSaved struct {
	Key       SnippetKey
	Owner     string
	Body      string
	Quick     bool
	CreatedAt time.Time
}

// SnippetDeleted is the read-model fact removing one snippet slot.
type SnippetDeleted struct {
	Key SnippetKey
}
