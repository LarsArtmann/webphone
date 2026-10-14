package domain

// FailureKind is the retry vocabulary of a failed outbound message.
// The wire, database and export DTOs carry the raw strings; this type
// is the Go-side closure of the set.
type FailureKind string

const (
	// FailureNone records no failure (the zero value).
	FailureNone FailureKind = ""
	// FailureTransient: transport or infrastructure failure — a later
	// attempt may succeed, so the bubble offers a Retry.
	FailureTransient FailureKind = "transient"
	// FailureRejected: deterministic refusal (invalid number, provider
	// policy) — a retry would fail identically, so none is offered.
	FailureRejected FailureKind = "rejected"
	// FailureProvider: the provider's asynchronous verdict said failed.
	FailureProvider FailureKind = "provider"
)

// IsRetryable reports whether the UI offers a retry for this kind:
// only transient failures are worth another attempt. Views key the
// retry affordance on this method, never on a string literal.
func (k FailureKind) IsRetryable() bool { return k == FailureTransient }

// Verdict is a provider's asynchronous delivery answer (the webhook
// lane). The two verdict strings are also the statuses they produce —
// the webhook status vocabulary maps one-to-one onto the lifecycle.
type Verdict string

const (
	// VerdictDelivered: the provider confirmed delivery to the handset.
	VerdictDelivered Verdict = "delivered"
	// VerdictFailed: the provider's asynchronous answer said failed.
	VerdictFailed Verdict = "failed"
)

// VerdictOf maps an OutboundStatus onto the verdict set. ok is false
// for every non-verdict status (queued, sent, zero) — the webhook seam
// refuses those before the machine ever sees them, and the service
// re-checks with its own pinned rejection.
func VerdictOf(status OutboundStatus) (Verdict, bool) {
	switch status {
	case StatusDelivered:
		return VerdictDelivered, true
	case StatusFailed:
		return VerdictFailed, true
	default:
		return "", false
	}
}

// Status is the OutboundStatus a verdict produces.
func (v Verdict) Status() OutboundStatus { return OutboundStatus(v) }

// Delivery is the machine-owned slice of an outbound message: status,
// provider correlation and failure story. Value semantics — DeliveryAdvance
// returns the next state, it never mutates.
type Delivery struct {
	Status        OutboundStatus
	ProviderRef   string
	FailureKind   FailureKind
	FailureDetail string
}

// eventKind discriminates the machine's triggers.
type eventKind uint8

const (
	eventSendAccepted eventKind = iota + 1
	eventSendRefused
	eventVerdict
)

// DeliveryEvent is one trigger of the delivery machine. Construct via
// SendAccepted, SendRefused or ProviderVerdict.
type DeliveryEvent struct {
	kind    eventKind   // which trigger
	ref     string      // send accepted: gateway correlation id
	failure FailureKind // send refused: retry vocabulary
	detail  string      // refused / verdict-failed: raw reason, rendered verbatim
	verdict Verdict     // verdict: the provider's answer
}

// SendAccepted: the gateway took the message; ref correlates its later
// verdicts.
func SendAccepted(providerRef string) DeliveryEvent {
	return DeliveryEvent{kind: eventSendAccepted, ref: providerRef}
}

// SendRefused: the send failed synchronously; kind decides the retry
// affordance, detail is the raw reason.
func SendRefused(kind FailureKind, detail string) DeliveryEvent {
	return DeliveryEvent{kind: eventSendRefused, failure: kind, detail: detail}
}

// ProviderVerdict: the provider's asynchronous answer. detail carries
// the failure reason of a failed verdict ("" when the provider sent
// none).
func ProviderVerdict(v Verdict, detail string) DeliveryEvent {
	return DeliveryEvent{kind: eventVerdict, verdict: v, detail: detail}
}

// deliveryMove is one legal from→to transition.
type deliveryMove struct {
	from OutboundStatus
	on   eventKind
	to   OutboundStatus
}

// deliveryMoves is the ONE home for delivery-transition legality.
//
// Send events fire only from queued — Send is synchronous: the row is
// persisted queued, then the gateway answers inside the same call.
//
// Verdicts are legal from EVERY state. That is today's out-of-order
// semantics encoded as data: a delivered verdict arriving after failed
// wins and clears the failure story; a failed verdict arriving after
// delivered overwrites back; a duplicate verdict is a legal same-state
// move whose write is value-identical (the webhook lane's 1h
// idempotency window replays those as 202s). Tightening any of this is
// owner call A1.7 — delete rows here and nowhere else.
var deliveryMoves = map[deliveryMove]bool{
	{StatusQueued, eventSendAccepted, StatusSent}:  true,
	{StatusQueued, eventSendRefused, StatusFailed}: true,

	{StatusQueued, eventVerdict, StatusDelivered}:    true,
	{StatusSent, eventVerdict, StatusDelivered}:      true,
	{StatusFailed, eventVerdict, StatusDelivered}:    true,
	{StatusDelivered, eventVerdict, StatusDelivered}: true,

	{StatusQueued, eventVerdict, StatusFailed}:    true,
	{StatusSent, eventVerdict, StatusFailed}:      true,
	{StatusFailed, eventVerdict, StatusFailed}:    true,
	{StatusDelivered, eventVerdict, StatusFailed}: true,
}

// DeliveryAdvance applies one event to a delivery state and returns
// the advanced copy. An event with no legal move from the current
// state is inert: the input state returns unchanged with ok=false —
// the caller keeps the settled state instead of fabricating one.
// Unreachable for verdicts while every verdict row stands (see
// deliveryMoves); it guards the send events, which only move from
// queued.
func DeliveryAdvance(d Delivery, e DeliveryEvent) (Delivery, bool) {
	if !deliveryMoves[deliveryMove{from: d.Status, on: e.kind, to: e.target()}] {
		return d, false
	}
	next := d
	switch e.kind {
	case eventSendAccepted:
		next.Status = StatusSent
		next.ProviderRef = e.ref
		next.FailureKind, next.FailureDetail = FailureNone, ""
	case eventSendRefused:
		next.Status = StatusFailed
		next.FailureKind, next.FailureDetail = e.failure, e.detail
	case eventVerdict:
		next.Status = e.verdict.Status()
		if e.verdict == VerdictFailed {
			next.FailureKind, next.FailureDetail = FailureProvider, e.detail
		} else {
			next.FailureKind, next.FailureDetail = FailureNone, ""
		}
	}
	return next, true
}

// target is the to-status the event proposes.
func (e DeliveryEvent) target() OutboundStatus {
	switch e.kind {
	case eventSendAccepted:
		return StatusSent
	case eventSendRefused:
		return StatusFailed
	default:
		return e.verdict.Status()
	}
}

// DeliveryOf projects the machine-owned fields of a message.
func (m Message) DeliveryOf() Delivery {
	return Delivery{
		Status:        m.Status,
		ProviderRef:   m.ProviderRef,
		FailureKind:   m.FailureKind,
		FailureDetail: m.FailureDetail,
	}
}

// WithDelivery returns the message carrying an advanced delivery state.
func (m Message) WithDelivery(d Delivery) Message {
	m.Status, m.ProviderRef, m.FailureKind, m.FailureDetail = d.Status, d.ProviderRef, d.FailureKind, d.FailureDetail
	return m
}
