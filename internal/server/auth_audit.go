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
