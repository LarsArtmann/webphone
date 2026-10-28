package readmodels

import (
	"context"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// AgentAuditMaxPerExtension bounds one extension's audit rows. Audit
// rows are SYSTEM facts, not user content: the trim policy is the
// transcripts posture (expunge the oldest — Recent renders the newest
// window), never the notes posture (reject).
const AgentAuditMaxPerExtension = 500

// AgentAuditView is the engine-side projection of one agent-action
// audit row (D5): actor, line, event, subject, outcome, stamp.
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

const agentAuditCollection = "agent_audit"

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type agentAuditInput struct{ Key domain.AgentAuditKey }

// agentAuditQuery declares the durable agent-action read model (D5):
// append-only rows, filtered by owner, sorted by happening order.
var agentAuditQuery = metaengine.Query[agentAuditInput, AgentAuditView](agentAuditCollection,
	metaengine.OnRecord(domain.AgentAuditAppended{},
		func(_ record.Record, e domain.AgentAuditAppended) (domain.AgentAuditKey, AgentAuditView) {
			return e.Key, AgentAuditView{
				ID:      e.Key.ID,
				Owner:   e.Owner,
				Actor:   e.Actor,
				Event:   e.Event,
				Subject: e.Subject,
				Outcome: e.Outcome,
				At:      e.At.Unix(),
				Seq:     e.Seq,
			}
		}),
	metaengine.OnRecord(domain.AgentAuditExpunged{}, metaengine.Remove[AgentAuditView]()),
	metaengine.FilterOnField[AgentAuditView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[AgentAuditView]("seq", false),
	metaengine.Volume(humanScaleVolume),
)

// AgentAudit is the engine-backed agent-action audit store (D5): the
// durable read model behind the greppable journal line. The line stays
// the operator-grep surface; this store answers "who approved what,
// when" in the UI (Settings' recent-actions section) — restart-proof,
// owner-scoped, newest window first.
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Notes/Reviews) — deliberate symmetry,
// not a mixin candidate.
type AgentAudit struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[AgentAuditView]
	mu     sync.Mutex
}

func newAgentAudit(models *metaengine.Store) *AgentAudit {
	return &AgentAudit{
		models: models,
		reader: metaengine.NewReader[AgentAuditView](models, agentAuditCollection),
	}
}

// Append stores one audit row and trims the owner's oldest rows past
// the cap (the transcripts converging trim: scan the newest cap+1 by
// seq, expunge the excess, repeat — one bounded scan never materializes
// the whole store).
//
// brandid-lint:ignore(BD003) audit vocabulary: raw strings are the journal-line tokens (the greppable-line contract)
func (s *AgentAudit) Append(ctx context.Context, owner domain.Extension, actor, event, subject, outcome string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.AgentAuditAppended{}), domain.AgentAuditAppended{
		Key:     domain.GenerateAgentAuditKey(),
		Owner:   owner.String(),
		Actor:   actor,
		Event:   event,
		Subject: subject,
		Outcome: outcome,
		At:      now,
		Seq:     now.UnixNano(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.agent_audit_append", "append audit row")
	}
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
			metaengine.WithSort("seq", true),
			metaengine.WithLimit(AgentAuditMaxPerExtension+1),
		)
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.agent_audit_trim", "scan audit trim window")
		}
		if len(views) <= AgentAuditMaxPerExtension {
			return nil
		}
		for _, v := range views[AgentAuditMaxPerExtension:] {
			if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.AgentAuditExpunged{}), domain.AgentAuditExpunged{
				Key: domain.AgentAuditKey{ID: v.ID},
			}); err != nil {
				return errorfamily.WrapInfrastructuref(err, "readmodels.agent_audit_trim", "expunge audit row")
			}
		}
	}
}

// Recent returns the owner's newest audit rows (newest first), or nil
// when none exist — the caller owns the absent-vs-empty vocabulary.
func (s *AgentAudit) Recent(ctx context.Context, owner domain.Extension, limit int) ([]AgentAuditView, error) {
	if limit <= 0 {
		limit = 1
	}
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(limit),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.agent_audit_recent", "scan audit rows")
	}
	if len(views) == 0 {
		return nil, nil
	}
	return views, nil
}
