package domain

import "time"

// Thread-stream facts (SUPERB Phase 5): the journal payloads for the
// threads/messages/attachments core. Every field is a PRIMITIVE mirror —
// Address and the branded id types carry unexported/branded state the
// journal codec cannot round-trip, so the facts carry exactly what the
// folds need and the gates re-parse through the Must* doors (the journal
// is trusted: only parsed domain values are ever appended).

// AttachmentFact is one MMS attachment inside a MessageAppended fact.
type AttachmentFact struct {
	ID        string
	MessageID string
	Name      string
	MimeType  string
	SizeBytes int64
	Path      string
}

// MessageAppended is the ONE append fact: a first message on an empty
// stream is thread creation (the fold adopts the fact's thread identity
// exactly as the legacy AppendToThread upsert did), every later message
// advances the thread. Attachments ride inside — one event, three
// projection folds.
type MessageAppended struct {
	MessageID     string
	ThreadID      string
	Owner         string
	Network       string
	Remote        string
	Direction     string
	Channel       string
	Body          string
	Status        string
	ProviderRef   string
	FailureKind   string
	FailureDetail string
	Attachments   []AttachmentFact
	CreatedAt     time.Time
}

// OutboundAdvanced records the delivery machine's OUTCOME for one
// outbound message (legality was deliveryMoves' call at the decision; the
// fact carries the advanced slice so replay never re-judges it).
type OutboundAdvanced struct {
	MessageID     string
	Status        string
	ProviderRef   string
	FailureKind   string
	FailureDetail string
}

// ThreadFlagSet sets or clears one organization axis (pinned |
// archived | muted) to the DESIRED state.
type ThreadFlagSet struct {
	Owner    string
	ThreadID string
	Flag     string
	On       bool
}

// ThreadUnreadCleared zeroes the unread counter.
type ThreadUnreadCleared struct {
	Owner    string
	ThreadID string
}

// ThreadDeleted is the retention/owner tombstone: the decision gathered
// every id the stream ever carried, so each projection fold is a pure
// keyed Remove and replay is exact. Reason is "retention" or "owner".
type ThreadDeleted struct {
	Owner         string
	ThreadID      string
	Network       string
	Remote        string
	MessageIDs    []string
	AttachmentIDs []string
	Reason        string
}

// DeliveryStep is the wire form of a DeliveryEvent (the command payload):
// DeliveryEvent's fields are deliberately unexported — the constructors
// are the door — so the command seam carries the step and the decider
// rebuilds the event through the same constructors. Kind is "accepted" |
// "refused" | "verdict".
type DeliveryStep struct {
	Kind    string
	Ref     string
	Failure FailureKind
	Detail  string
	Verdict Verdict
}

// DeliveryStepKind constants (stringly on purpose: the step is wire data,
// not a domain closure).
const (
	DeliveryStepAccepted = "accepted"
	DeliveryStepRefused  = "refused"
	DeliveryStepVerdict  = "verdict"
)

// Step renders the event as its wire form.
func (e DeliveryEvent) Step() DeliveryStep {
	switch e.kind {
	case eventSendAccepted:
		return DeliveryStep{Kind: DeliveryStepAccepted, Ref: e.ref}
	case eventSendRefused:
		return DeliveryStep{Kind: DeliveryStepRefused, Failure: e.failure, Detail: e.detail}
	default:
		return DeliveryStep{Kind: DeliveryStepVerdict, Verdict: e.verdict, Detail: e.detail}
	}
}

// Event rebuilds the DeliveryEvent from its wire form (the same
// constructors the send lane uses — no second legality door).
func (s DeliveryStep) Event() DeliveryEvent {
	switch s.Kind {
	case DeliveryStepAccepted:
		return SendAccepted(s.Ref)
	case DeliveryStepRefused:
		return SendRefused(s.Failure, s.Detail)
	default:
		return ProviderVerdict(s.Verdict, s.Detail)
	}
}

// MessageRemoved is one keyed row removal of the transcript projection
// (the delete batch's per-message fact — a fold Remove is keyed, so a
// thread-wide clear is a batch of these, journal-replayable).
type MessageRemoved struct {
	MessageID string
}

// AttachmentStored is one attachment row's own fact (an MMS appends one
// MessageAppended plus one AttachmentStored per attachment — map folds
// are one row per event, so the attachment collection gets its own).
type AttachmentStored struct {
	ID        string
	MessageID string
	Owner     string
	Name      string
	MimeType  string
	SizeBytes int64
	Path      string
}

// AttachmentRemoved is the attachment projection's keyed removal twin.
type AttachmentRemoved struct {
	ID string
}
