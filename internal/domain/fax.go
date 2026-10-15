package domain

import "time"

// FaxDirection tells which way a fax traveled.
type FaxDirection string

const (
	// FaxOutbound was sent by the extension.
	FaxOutbound FaxDirection = "out"
	// FaxInbound arrived at the extension.
	FaxInbound FaxDirection = "in"
)

// FaxStatus is the fax job state. Two shapes share the enum: the
// OUTBOUND lifecycle (queued → sending → transmitted | failed, the
// faxMoves table is its one home) and the INBOUND fact (received — a
// document that landed, with no lifecycle at all; no move touches it).
// A late verdict may still overwrite a settled state (see faxMoves),
// and a same-status verdict refreshes its payload; only events
// against an inbound "received" fact are inert.
type FaxStatus string

const (
	// FaxQueued: PDF stored locally, gateway not yet called.
	FaxQueued FaxStatus = "queued"
	// FaxSending: gateway accepted the job; waiting for the verdict.
	FaxSending FaxStatus = "sending"
	// FaxTransmitted: the far end confirmed all pages.
	FaxTransmitted FaxStatus = "transmitted"
	// FaxReceived: an inbound fax document landed in the inbox.
	FaxReceived FaxStatus = "received"
	// FaxFailed: transmission failed; Error carries the reason.
	FaxFailed FaxStatus = "failed"
)

// FaxJob is one fax transmission attempt (outbound) or one received
// document (inbound). The PDF lives on disk at DocumentPath.
//
// branching-flow COMPOSITION_mixin (baselined): the aggregate mirrors
// its storage row by the parity-port ruling — the SQL→engine swap
// served this shape verbatim, and accessors would be indirection
// without meaning.
type FaxJob struct {
	ID           FaxID
	Owner        Extension
	Remote       Phone
	Direction    FaxDirection
	Status       FaxStatus
	Pages        int
	DocumentPath string
	ProviderRef  string
	Error        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InboundFax is a received fax document crossing the gateway boundary.
type InboundFax struct {
	Owner       Extension
	From        Phone
	Pages       int
	PDFBytes    []byte
	Received    time.Time
	ProviderRef string
}

// faxEventKind discriminates the fax machine's triggers.
type faxEventKind uint8

const (
	faxAccepted faxEventKind = iota + 1
	faxSettled
	faxSubmitFailed
	faxVerdict
)

// FaxEvent is one trigger of the outbound fax machine. Construct via
// FaxAccepted, FaxSettled, FaxSubmitFailed or FaxVerdict.
type FaxEvent struct {
	kind   faxEventKind
	status FaxStatus // verdict: the provider's answer (transmitted or failed)
}

// FaxAccepted: the gateway took the job; its verdict arrives later on
// the status webhook (a deferred receipt).
func FaxAccepted() FaxEvent { return FaxEvent{kind: faxAccepted} }

// FaxSettled: an immediate receipt — the acceptance IS the verdict
// (loopback mode). Composes after FaxAccepted: drive applies both, so
// the provider ref lands on the sending row exactly as before.
func FaxSettled() FaxEvent { return FaxEvent{kind: faxSettled} }

// FaxSubmitFailed: the synchronous submit failed (gateway error or
// self-send refusal); Error carries the reason on the row.
func FaxSubmitFailed() FaxEvent { return FaxEvent{kind: faxSubmitFailed} }

// FaxVerdict: the provider's asynchronous answer. Only transmitted and
// failed reach the machine (the webhook seam rejects everything else
// and remaps received → transmitted for outbound rows).
func FaxVerdict(status FaxStatus) FaxEvent { return FaxEvent{kind: faxVerdict, status: status} }

// faxMove is one legal from→to transition.
type faxMove struct {
	from FaxStatus
	on   faxEventKind
	to   FaxStatus
}

// faxMoves is the ONE home for fax-transition legality. The outbound
// lifecycle is queued → sending → transmitted | failed; "received" is
// an inbound FACT, not a lifecycle state — no move starts or ends there.
//
// Submit events fire only from queued: Send and Resend drive fresh
// queued jobs synchronously. Settled fires only from sending — drive
// applies accepted first.
//
// Verdicts are legal between every pair of outbound states INCLUDING
// same-status: a verdict may arrive with a richer payload than the
// first one (a later page count), and the row refreshes — the same
// doctrine deliveryMoves encodes. Replay protection is NOT the
// machine's job: the hooks seam dedups by provider ref (hooksIdem,
// 1h TTL, refutes contradictory replays before the service runs).
// What stays inert: every event against "received" — an inbound fact
// has no lifecycle — and settle/accept/submit off their lanes.
var faxMoves = map[faxMove]bool{
	{FaxQueued, faxAccepted, FaxSending}:     true,
	{FaxQueued, faxSubmitFailed, FaxFailed}:  true,
	{FaxSending, faxSettled, FaxTransmitted}: true,

	{FaxQueued, faxVerdict, FaxTransmitted}:      true,
	{FaxSending, faxVerdict, FaxTransmitted}:     true,
	{FaxFailed, faxVerdict, FaxTransmitted}:      true,
	{FaxTransmitted, faxVerdict, FaxTransmitted}: true,

	{FaxQueued, faxVerdict, FaxFailed}:      true,
	{FaxSending, faxVerdict, FaxFailed}:     true,
	{FaxTransmitted, faxVerdict, FaxFailed}: true,
	{FaxFailed, faxVerdict, FaxFailed}:      true,
}

// FaxAdvance applies one event to an outbound fax status and returns
// the advanced status. An event with no legal move from the current
// status is inert: the input returns unchanged with ok=false — the
// caller keeps the settled state instead of writing a phantom one
// (reachable today only for events against an inbound "received"
// fact; the seam rejects off-lane submits before the machine runs).
func FaxAdvance(from FaxStatus, e FaxEvent) (FaxStatus, bool) {
	if !faxMoves[faxMove{from: from, on: e.kind, to: e.target()}] {
		return from, false
	}
	return e.target(), true
}

// target is the to-status the event proposes.
func (e FaxEvent) target() FaxStatus {
	switch e.kind {
	case faxAccepted:
		return FaxSending
	case faxSettled:
		return FaxTransmitted
	case faxSubmitFailed:
		return FaxFailed
	default:
		return e.status
	}
}

// FaxKey identifies one fax job row in the read-model store. A distinct
// struct type (not a bare string) mirrors ContactKey so the metaengine
// fold key extraction stays unambiguous — exactly one field of this
// type per read-model fact.
// brandid-lint:ignore(BD003) read-model key: wire shape (strong-id wire ruling)
type FaxKey struct {
	ID string
}

// FaxReceipted is the read-model fact inserting one fax job row: an
// outbound send at its initial status, or an inbound received document.
// Raw string ids by the strong-id wire ruling (read-model facts are
// wire shapes); the lifecycle enums stay typed — they ARE the faxMoves
// vocabulary.
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// FaxView row on purpose (the fold is a plain insert of the wire
// shape).
type FaxReceipted struct {
	Key          FaxKey
	Owner        string
	Remote       string
	Direction    FaxDirection
	Status       FaxStatus
	Pages        int
	DocumentPath string
	ProviderRef  string
	Error        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FaxStatusAdvanced is the read-model fact advancing a recorded job's
// lifecycle fields — exactly the SQL UPDATE's payload: the new status,
// the provider correlation id (empty = keep the existing one, the fold
// merges), the verbatim refusal/error text, and the page count.
type FaxStatusAdvanced struct {
	Key         FaxKey
	Status      FaxStatus
	ProviderRef string
	Error       string
	Pages       int
	AdvancedAt  time.Time
}

// FaxExpunged is the read-model fact removing one fax job row — the
// retention pass's engine arm (fax content shares the retention
// window; nothing else deletes fax rows).
type FaxExpunged struct {
	Key FaxKey
}
