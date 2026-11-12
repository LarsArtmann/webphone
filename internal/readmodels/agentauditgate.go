// The agent-action audit seam's GATE + VIEW surface: the interface the
// server handlers consume (the dual-write target of the agentAudit
// journal line and Settings' newest-20 read) and the projection row the
// journal backend folds. The engine-backed store that used to implement
// the gate here retired with the readmodels engine (2026-11-11
// storage-ladder closeout) — internal/cqrs/agentaudit.go is the one
// implementation now.
package readmodels

import (
	"context"

	"github.com/larsartmann/webphone/internal/domain"
)

// AgentAuditMaxPerExtension bounds one extension's audit rows. Audit
// rows are SYSTEM facts, not user content: the trim policy is the
// transcripts posture (expunge the oldest — Recent renders the newest
// window), never the notes posture (reject).
const AgentAuditMaxPerExtension = 500

// AgentAuditView is the projection row ONE journal fold produces for an
// agent-action audit row (D5): actor, line, event, subject, outcome,
// stamp.
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type AgentAuditView struct {
	ID      string `json:"id"`
	Owner   string `json:"owner"`
	Actor   string `json:"actor"`
	Event   string `json:"event"`
	Subject string `json:"subject"`
	Outcome string `json:"outcome"`
	At      int64  `json:"at"`
	Seq     int64  `json:"seq"`
}

// AgentAuditGate is the agent-action audit seam as the server sees it
// (the dual-write target of the agentAudit journal line and Settings'
// newest-20 read). The journal-backed service (internal/cqrs,
// SUPERB Phase 3) is the implementation.
type AgentAuditGate interface {
	Append(ctx context.Context, owner domain.Extension, actor, event, subject, outcome string) error
	Recent(ctx context.Context, owner domain.Extension, limit int) ([]AgentAuditView, error)
}
