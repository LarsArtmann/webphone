package domain

import "time"

// CallTranscriptSegment is one live-transcribed chunk of a call. The
// island's capture loop posts each recorder segment as its own row, so a
// call's transcript is the ordered sequence of its segments — durable in
// the owner-scoped store even when the call card (or the whole tab) dies.
type CallTranscriptSegment struct { //nolint:branching-flow // segment (the storage grain) vs reassembled transcript (the view grain): same identity fields, different lifecycles by design
	Owner     Extension
	CallID    string //nolint:branching-flow // the island-owned SIP Call-ID: opaque foreign token passed through verbatim (the 09-18 pbx-UUID class), never joined to CDR rows
	Direction Direction
	Remote    string
	StartedAt time.Time
	Text      string
}

// CallTranscript is one call's reassembled transcript for the History
// view: the call identity plus its segments in spoken order.
type CallTranscript struct { //nolint:branching-flow // domain vs export-DTO boundary layers (exportTranscript mirrors this shape for the zip)
	Owner     Extension
	CallID    string //nolint:branching-flow // island-owned SIP Call-ID, verbatim passthrough (see CallTranscriptSegment.CallID)
	Direction Direction
	Remote    string
	StartedAt time.Time
	Lines     []string
}
