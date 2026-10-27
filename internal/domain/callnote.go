package domain

import "time"

// CallNoteKey identifies one call's note slot in the read-model store:
// owner plus the call id the note is ABOUT (the transcript section's
// correlation handle — CDR rows carry no call uuid, so this pair is
// the only key the History surface has). A distinct struct type mirrors
// ContactKey/PushDeviceKey so the metaengine fold key extraction stays
// unambiguous.
// brandid-lint:ignore(BD003) read-model key: the island Call-ID is an opaque foreign token (strong-id wire ruling)
type CallNoteKey struct {
	Owner  string
	CallID string
}

// CallNoteSaved is the read-model fact for a call-review-slot save:
// the owner's freeform note PLUS the operator's disposition verdict for
// one call, snapshotted together (the slot is the (owner, call) pair —
// a save replaces the row; the fold is a plain insert). The seam
// pre-reads the slot and fills BOTH fields so each event carries the
// full row state. Disposition is the HUMAN verdict (issue #6) — the one
// authority that may claim answered; notes are user-authored prose and
// an empty body is a legal note.
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// CallNoteView row on purpose (the fold is a plain insert of the wire
// shape).
// brandid-lint:ignore(BD003) island-owned SIP Call-ID: opaque foreign token, verbatim passthrough (strong-id wire ruling)
type CallNoteSaved struct {
	Key         CallNoteKey
	Owner       string
	CallID      string
	Body        string
	Disposition Disposition
	UpdatedAt   time.Time
}
