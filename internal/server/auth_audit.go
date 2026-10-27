package server

import "log/slog"

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
// HAS sent) with the actor extension, the draft id when the action
// names one, and the outcome. Identifiers only, like authAudit.
//
// Durable-journal decision (F107): the usermgmt auth-audit journal is
// the IDENTITY layer's store — agent approvals are not auth events, and
// writing them there would cross the seam; an app_flags window would be
// a lossy hand-rolled journal. The structured journal line IS the audit
// home (journald persists it with the rest of the unit's log); a
// durable read model becomes worth building only when an operator needs
// in-UI history — the trigger is recorded in the plan errata.
func agentAudit(event, actor, draftID, outcome string) {
	if draftID != "" {
		slog.Info("agent "+event, "extension", actor, "draft", draftID, "outcome", outcome)
		return
	}
	slog.Info("agent "+event, "extension", actor, "outcome", outcome)
}
