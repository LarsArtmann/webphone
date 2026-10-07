package domain

import "time"

// CallTranscriptSegment is one live-transcribed chunk of a call. The
// island's capture loop posts each recorder segment as its own row, so a
// call's transcript is the ordered sequence of its segments — durable in
// the owner-scoped store even when the call card (or the whole tab) dies.
type CallTranscriptSegment struct {
	Owner     Extension
	CallID    string
	Direction Direction
	Remote    string
	StartedAt time.Time
	Text      string
}

// CallTranscript is one call's reassembled transcript for the History
// view: the call identity plus its segments in spoken order.
type CallTranscript struct {
	Owner     Extension
	CallID    string
	Direction Direction
	Remote    string
	StartedAt time.Time
	Lines     []string
}
