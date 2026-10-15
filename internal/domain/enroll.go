package domain

import "time"

// EnrollTokenKey identifies one enrollment token in the read-model
// store: the SHA-256 hex of the URL secret (the raw secret is never
// stored — the key IS the hash). A distinct struct type mirrors the
// other read-model keys so the metaengine fold key extraction stays
// unambiguous.
type EnrollTokenKey struct {
	TokenHash string
}

// EnrollTokenMinted is the read-model fact inserting one enrollment
// token (the hash, the account it enrolls, and its expiry).
//
// branching-flow COMPOSITION_mixin (baselined): read-model facts mirror
// their engine row on purpose (strong-id wire ruling) — the fold is a
// plain insert of the wire shape.
//brandid-lint:ignore(BD003) read-model fact: wire shape; identitymodel.UserID parses strictly at the userauth door (strong-id wire ruling)
type EnrollTokenMinted struct {
	Key       EnrollTokenKey
	Email     string
	UserID    string
	ExpiresAt time.Time
}

// EnrollTokenBurned is the read-model fact removing one enrollment
// token — the consume-once burn and the expiry sweep ride it.
type EnrollTokenBurned struct {
	Key EnrollTokenKey
}
