package domain

import "time"

// AgentAuditKey identifies one agent-action audit row in the read-model
// store. Audit rows are APPEND-ONLY facts (never upserted), so every
// append mints a fresh key — the transcript-segment precedent.
// brandid-lint:ignore(BD003) engine key minted by GenerateAgentAuditKey; read-model facts are wire shapes (strong-id wire ruling)
type AgentAuditKey struct {
	ID string
}

// GenerateAgentAuditKey mints a fresh audit-row key.
func GenerateAgentAuditKey() AgentAuditKey {
	return AgentAuditKey{ID: nanoid.Must().String()}
}

// AgentAuditAppended is the read-model fact inserting one agent-action
// audit row (D5, the durable read model the multi-actor substrate
// armed): WHO acted (the session's login email when it has one, else
// the extension), on which line, which agent event (approve, discard,
// mode pin, auto_review, …), the draft/call it named, and the outcome.
// Seq carries the append order (nanos at the seam) so Recent sorts by
// happening time, not by opaque key.
//
// branching-flow COMPOSITION_mixin (baselined): the fact mirrors the
// AgentAuditView row on purpose (the fold is a plain insert of the wire
// shape).
// brandid-lint:ignore(BD003) audit vocabulary: raw strings are the journal-line tokens (the greppable-line contract, verbatim passthrough)
type AgentAuditAppended struct {
	Key     AgentAuditKey
	Owner   string
	Actor   string
	Event   string
	Subject string
	Outcome string
	At      time.Time
	Seq     int64
}

// AgentAuditExpunged removes one audit row (the cap trim).
// brandid-lint:ignore(BD003) engine key passthrough (strong-id wire ruling)
type AgentAuditExpunged struct {
	Key AgentAuditKey
}
