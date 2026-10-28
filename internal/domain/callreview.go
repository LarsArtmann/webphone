package domain

import "time"

// ReviewState is the automatic post-call review's lifecycle state
// (issue #8): pending while the detached job runs, present once the AI
// assessment landed, failed when the agent refused or was unreachable.
// A restart mid-pending leaves a stale pending row — the History render
// shows it honestly and the next dispatch replaces it (the slot is the
// (owner, call) pair, replace semantics).
type ReviewState string

const (
	ReviewPending ReviewState = "pending"
	ReviewPresent ReviewState = "present"
	ReviewFailed  ReviewState = "failed"
)

// String renders the state for engine rows and wire shapes.
func (s ReviewState) String() string { return string(s) }

// CallReviewKey identifies one call's review slot — the (owner, call)
// pair, mirroring CallNoteKey so the metaengine fold key extraction
// stays unambiguous.
// brandid-lint:ignore(BD003) read-model key: the island Call-ID is an opaque foreign token (strong-id wire ruling)
type CallReviewKey struct {
	Owner  string
	CallID string
}

// CallReviewSaved is the read-model fact for a call-review save: the
// AI's assessment of one finished call, snapshotted whole (a save
// replaces the row; the fold is a plain insert). DraftID/Text/Language/
// Confidence are the present-state dimensions — meaningful only when
// State is ReviewPresent; the summary/points/score dims from the plan
// stay optional-empty until the model-routing decision (T15) fixes the
// agent's score vocabulary. The agent's confidence is the ONE dimension
// it emits today (0 when absent).
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// CallReviewView row on purpose (the fold is a plain insert of the wire
// shape).
// brandid-lint:ignore(BD003) island-owned SIP Call-ID: opaque foreign token, verbatim passthrough (strong-id wire ruling)
type CallReviewSaved struct {
	Key        CallReviewKey
	Owner      string
	CallID     string
	State      ReviewState
	DraftID    string
	Text       string
	Language   string
	Confidence int
	UpdatedAt  time.Time
}
