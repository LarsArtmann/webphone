package domain

import "time"

// DialWindowKey identifies one (owner, remote) dial-window slot: the
// cooldown is per OWNER-REMOTE pair, never global — one extension
// dialing a number says nothing about another extension's right to
// dial it.
// brandid-lint:ignore(BD003) read-model key: the remote is the dialable string the island dials verbatim (strong-id wire ruling)
type DialWindowKey struct {
	Owner  string
	Remote string
}

// DialAttempted is the read-model fact for one allowed dial: the slot
// is the (owner, remote) pair, so a new attempt REPLACES the row (the
// window always carries the LAST attempt). Denied attempts never emit
// this fact — a retry loop must not extend the cooldown it is fighting.
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// DialWindowView row on purpose (the fold is a plain insert).
// brandid-lint:ignore(BD003) dialable remote passthrough (strong-id wire ruling)
type DialAttempted struct {
	Key    DialWindowKey
	Owner  string
	Remote string
	At     time.Time
}

// DialOverrideSaved is the read-model fact for the one-shot bypass
// (issue #7): setting it arms the override, consuming it flips
// Consumed=true on the SAME slot (upsert, replace semantics — the next
// check reads one row, never a scan).
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// DialOverrideView row on purpose (the fold is a plain insert).
// brandid-lint:ignore(BD003) dialable remote passthrough (strong-id wire ruling)
type DialOverrideSaved struct {
	Key      DialWindowKey
	Owner    string
	Remote   string
	Consumed bool
	At       time.Time
}
