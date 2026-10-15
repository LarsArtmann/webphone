package domain

import (
	"time"

	"github.com/sixafter/nanoid"
)

// CallTranscriptSegment is one live-transcribed chunk of a call. The
// island's capture loop posts each recorder segment as its own row, so a
// call's transcript is the ordered sequence of its segments — durable in
// the owner-scoped store even when the call card (or the whole tab) dies.
// branching-flow COMPOSITION_mixin (baselined): segment (the storage
// grain) vs reassembled transcript (the view grain) — same identity
// fields, different lifecycles by design.
type CallTranscriptSegment struct {
	Owner     Extension
	CallID    string //nolint:branching-flow // the island-owned SIP Call-ID: opaque foreign token passed through verbatim (the 09-18 pbx-UUID class), never joined to CDR rows
	Direction Direction
	Remote    string
	StartedAt time.Time
	Text      string
}

// CallTranscript is one call's reassembled transcript for the History
// view: the call identity plus its segments in spoken order.
// branching-flow COMPOSITION_mixin (baselined): domain vs export-DTO
// boundary layers (exportTranscript mirrors this shape for the zip).
type CallTranscript struct {
	Owner     Extension
	CallID    string //nolint:branching-flow // island-owned SIP Call-ID, verbatim passthrough (see CallTranscriptSegment.CallID)
	Direction Direction
	Remote    string
	StartedAt time.Time
	Lines     []string
}

// TranscriptSegmentKey identifies one transcript segment in the
// read-model store. SQL rows carried no id (the autoincrement rowid
// was private); the engine requires a key, so the seam mints one per
// append — same uniqueness role, now durable and explicit.
type TranscriptSegmentKey struct {
	ID string
}

// TranscriptAppended is the read-model fact inserting one transcript
// segment. Seq carries the append order (nanos at the seam) — the
// rowid tiebreak's engine equivalent; sorting by seq reproduces the
// SQL `started_at DESC, rowid DESC` stream because live segments
// arrive in spoken order.
type TranscriptAppended struct {
	Key       TranscriptSegmentKey
	Owner     string
	CallID    string //nolint:branching-flow // island-owned SIP Call-ID, verbatim passthrough (see CallTranscriptSegment.CallID)
	Direction Direction
	Remote    string
	StartedAt time.Time
	Seq       int64
	Text      string
}

// TranscriptExpunged is the read-model fact removing one transcript
// segment — the caps' trim arm, the end-user delete-call leg, and the
// retention sweep all ride it.
type TranscriptExpunged struct {
	Key TranscriptSegmentKey
}

// GenerateTranscriptSegmentKey mints a fresh segment key: the engine
// row identity (internal — never rendered on the wire).
func GenerateTranscriptSegmentKey() TranscriptSegmentKey {
	return TranscriptSegmentKey{ID: nanoid.Must().String()}
}
