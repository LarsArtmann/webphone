package server

import (
	"context"
	"log/slog"

	"github.com/larsartmann/webphone/internal/domain"
)

// authAudit is the ONE home of the semantic auth events (A3): one
// greppable slog line per login, denied login, logout, passkey
// enrollment and passkey revocation — the operator's audit trail over
// the journal. It logs IDENTIFIERS only (extension, email, credential
// name); never a password, a session token, a derived credential, or a
// ceremony body (the request log middleware never sees those either —
// bodies stay unparsed).
func authAudit(event string, key, value, outcome string) {
	slog.Info("auth "+event, key, value, "outcome", outcome)
}

// agentAudit is the agent-action branch of the same audit trail (M18):
// one greppable slog line per drafts-panel action (approve, discard,
// mode pin/reset — every path that changes what the agent may send or
// HAS sent) with the acting extension, the draft id when the action
// names one, and the outcome. Identifiers only, like authAudit.
//
// The multi-user substrate (2026-10-28): actorEmail names the HUMAN
// behind a passwordless session — with two people sharing one
// extension, the extension attr alone cannot say who approved. Empty
// (the classic extension login) keeps the historical line shape
// byte-identical; the actor attr only appears when there is one.
//
// Durable read model (D5, 2026-10-28 — the F107 trigger ARMED by the
// multi-actor substrate): the journal line stays the operator-grep
// surface, and the row DUAL-WRITES into the readmodels agent-audit
// store so Settings can answer "who approved what, when" without
// journal archaeology. The identity seam stays untouched (agent
// approvals are still not auth events); a nil store (test
// compositions) degrades to the line alone — the audit never blocks
// the action it records, and a store failure WARNs, never fails.
//
//nolint:branching-flow // journal-line builder: the raw strings ARE the audit-log vocabulary
func (h *handlers) agentAudit(event, actorEmail, extension, draftID, outcome string) {
	attrs := make([]any, 0, 8)
	attrs = append(attrs, "extension", extension)
	if actorEmail != "" && actorEmail != extension {
		attrs = append(attrs, "actor", actorEmail)
	}
	if draftID != "" {
		attrs = append(attrs, "draft", draftID)
	}
	attrs = append(attrs, "outcome", outcome)
	slog.Info("agent "+event, attrs...)
	if h.deps.AgentAudit == nil {
		return
	}
	owner, err := domain.ParseExtension(extension)
	if err != nil {
		return // the line logged what it could; a foreign shape never reaches the store
	}
	if err := h.deps.AgentAudit.Append(context.Background(), owner, actorEmail, event, draftID, outcome); err != nil {
		slog.Warn("agent audit row not stored (the journal line stands)", "error", err, "event", event, "extension", extension)
	}
}
