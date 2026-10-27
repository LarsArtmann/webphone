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

// CallNoteSaved is the read-model fact for a call-note upsert: the
// owner's freeform post-call note for one call. Notes are user-authored
// content — replacing a note REPLACES THE SLOT (no duplicate rows), and
// an empty body is a legal note (the UI renders "no note"). The store
// seam emits the fact; the fold is a plain insert.
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// CallNoteView row on purpose (the fold is a plain insert of the wire
// shape).
// brandid-lint:ignore(BD003) island-owned SIP Call-ID: opaque foreign token, verbatim passthrough (strong-id wire ruling)
type CallNoteSaved struct {
	Key       CallNoteKey
	Owner     string
	CallID    string
	Body      string
	UpdatedAt time.Time
}

// CallDispositionSaved is the read-model fact for an operator's
// post-call verdict (issue #6 confirm): the HUMAN answer to "how did
// this call end" — the one authority that may mint DispositionAnswered.
// It shares the note's (owner, call) slot but is a separate fact: the
// verdict drives the CRM correction write and the History filter, while
// the note stays freeform prose; editing one must never touch the
// other. The fold carries the slot's existing body forward.
// brandid-lint:ignore(BD003) island-owned SIP Call-ID: opaque foreign token, verbatim passthrough (strong-id wire ruling)
type CallDispositionSaved struct {
	Key         CallNoteKey
	Owner       string
	CallID      string
	Disposition Disposition
	UpdatedAt   time.Time
}
