// The agent-audit aggregate (SUPERB Phase 3, A23): the durable
// read model behind the greppable journal line as a per-owner
// APPEND-ONLY stream — no update folds, the only mutation besides
// append is the cap trim (expunge-oldest, the transcripts posture;
// the JOURNAL keeps everything, per the D3 keep-forever ruling).
// Each audit event carries the ACTOR (user:<email> — the multi-user
// substrate's "who approved") and the CAUSATION (the append command's
// id) as event metadata, so journal archaeology answers "who" without
// decoding payloads. Settings' newest-20 reads the projection.
package cqrs

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// AgentAuditStreamType keys per-owner audit streams.
const AgentAuditStreamType = "agent_audit"

// agentAuditMaxRows mirrors readmodels.AgentAuditMaxPerExtension at
// the aggregate (a var so the cap test does not append 500 rows).
var agentAuditMaxRows = readmodels.AgentAuditMaxPerExtension

// agentAuditCommand names the append command (the causation token).
const agentAuditCommand = "agent_audit.append"

// AppendAgentAudit appends one audit row (append-only: the decision
// never rejects, it trims the oldest rows past the cap).
type AppendAgentAudit struct {
	*command.BasicCommand
	Owner   string
	Actor   string
	Event   string
	Subject string
	Outcome string
	Key     domain.AgentAuditKey
	At      time.Time
}

// agentAuditRow is one audit row's aggregate state.
type agentAuditRow struct {
	Actor   string
	Event   string
	Subject string
	Outcome string
	Seq     int64
}

// agentAuditState is the owner's audit window (bounded by the cap).
type agentAuditState struct {
	Rows map[string]agentAuditRow
}

func (agentAuditState) initial() agentAuditState {
	return agentAuditState{Rows: map[string]agentAuditRow{}}
}

func applyAgentAuditState(s agentAuditState, evt event.Event) (agentAuditState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.AgentAuditAppended{})):
		fact, err := event.DecodePayloadAuto[domain.AgentAuditAppended](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.agent_audit_fold", "decode AgentAuditAppended")
		}
		s.Rows[fact.Key.ID] = agentAuditRow{Actor: fact.Actor, Event: fact.Event, Subject: fact.Subject, Outcome: fact.Outcome, Seq: fact.Seq}
	case event.Type(metaengine.EventTypeName(domain.AgentAuditExpunged{})):
		fact, err := event.DecodePayloadAuto[domain.AgentAuditExpunged](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.agent_audit_fold", "decode AgentAuditExpunged")
		}
		delete(s.Rows, fact.Key.ID)
	}
	return s, nil
}

// decideAppendAgentAudit emits the append plus the expunges that hold
// the owner's window at the cap — the engine store's converging trim
// as ONE decision over in-memory state (the state is already bounded,
// no scan loop needed).
func decideAppendAgentAudit(s agentAuditState, ver event.Version, cmd AppendAgentAudit) ([]event.Event, error) {
	seq := cmd.At.UnixNano()
	facts := []any{domain.AgentAuditAppended{
		Key:     cmd.Key,
		Owner:   cmd.Owner,
		Actor:   cmd.Actor,
		Event:   cmd.Event,
		Subject: cmd.Subject,
		Outcome: cmd.Outcome,
		At:      cmd.At,
		Seq:     seq,
	}}
	if excess := len(s.Rows) + 1 - agentAuditMaxRows; excess > 0 {
		ids := make([]string, 0, len(s.Rows))
		for id := range s.Rows {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return s.Rows[ids[i]].Seq < s.Rows[ids[j]].Seq })
		for _, id := range ids[:excess] {
			facts = append(facts, domain.AgentAuditExpunged{Key: domain.AgentAuditKey{ID: id}})
		}
	}

	actor := actorForAudit(cmd.Actor)
	events := make([]event.Event, 0, len(facts))
	for _, fact := range facts {
		ver++
		opts := []event.Option{event.WithCausation(agentAuditCommand, cmd.ID())}
		if !actor.IsZero() {
			opts = append(opts, event.WithActor(actor))
		}
		evt, err := event.New(event.Type(metaengine.EventTypeName(fact)), cmd.StreamID(), AgentAuditStreamType, ver, fact, opts...)
		if err != nil {
			return nil, errorfamily.WrapInfrastructuref(err, "cqrs.agent_audit_append", "build audit fact")
		}
		events = append(events, evt)
	}
	return events, nil
}

// actorForAudit maps the audit vocabulary's actor (login email, or
// empty for the classic extension login) onto the event actor: a
// user-kind actor whose raw is the email — the journal answers "who"
// from metadata alone. (ParseActorID over "user:<email>", because a
// UserID itself must be a ULID and an email is the honest raw.)
func actorForAudit(actorEmail string) id.ActorID {
	if actorEmail == "" {
		return id.ActorID{}
	}
	actor, err := id.ParseActorID("user:" + actorEmail)
	if err != nil {
		return id.ActorID{}
	}
	return actor
}

// RegisterAgentAudit wires the agent-audit aggregate into a System.
func RegisterAgentAudit(sys *system.System) error {
	if err := system.RegisterDecider[agentAuditState](sys, AgentAuditStreamType, decider.Decider[agentAuditState]{
		Initial: agentAuditState{}.initial(),
		Apply:   applyAgentAuditState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.agent_audit_register", "register agent-audit decider")
	}
	if err := system.RegisterCommand[*AppendAgentAudit, agentAuditState](sys, agentAuditCommand,
		func(_ context.Context, cmd *AppendAgentAudit) system.Op[agentAuditState] {
			return system.Execute[agentAuditState](context.Background(), cmd.StreamID(), AgentAuditStreamType,
				func(state agentAuditState, ver event.Version) ([]event.Event, error) {
					return decideAppendAgentAudit(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.agent_audit_register", "register %s", agentAuditCommand)
	}
	return nil
}

// agentAuditProjection folds the audit facts into the SAME view shape
// Settings reads today (readmodels.AgentAuditView).
func agentAuditProjection() metaengine.QueryDecl[agentAuditInput, readmodels.AgentAuditView] {
	return metaengine.Query[agentAuditInput, readmodels.AgentAuditView](agentAuditCollection,
		metaengine.OnRecord(domain.AgentAuditAppended{},
			func(_ record.Record, e domain.AgentAuditAppended) (domain.AgentAuditKey, readmodels.AgentAuditView) {
				return e.Key, readmodels.AgentAuditView{
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
		metaengine.OnRecord(domain.AgentAuditExpunged{}, metaengine.Remove[readmodels.AgentAuditView]()),
		metaengine.FilterOnField[readmodels.AgentAuditView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.AgentAuditView]("seq", false),
	)
}

// agentAuditCollection names the system engine's projection (the same
// name the readmodels engine declares — different databases, one
// shape).
const agentAuditCollection = "agent_audit"

// agentAuditInput is the projection's key-extraction input (the same
// shape the readmodels engine declares).
type agentAuditInput struct{ Key domain.AgentAuditKey }

// AgentAuditDomainDecl is the agent-audit aggregate's share of the
// system's DomainConfig (a registration failure is a boot-time miswire
// and panics loudly).
func AgentAuditDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterAgentAudit(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register agent-audit aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(agentAuditProjection())},
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.AgentAuditAppended{})),
			event.Type(metaengine.EventTypeName(domain.AgentAuditExpunged{})),
		},
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// AgentAudit is the dispatch-backed agent-action audit store: the
// same contract the engine-backed store exposed (append with the cap
// trim, newest-first Recent window).
type AgentAudit struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.AgentAuditView]
	mu     sync.Mutex
}

var _ readmodels.AgentAuditGate = (*AgentAudit)(nil)

// NewAgentAudit builds the backend over a System.
func NewAgentAudit(sys *system.System) *AgentAudit {
	return &AgentAudit{sys: sys, reader: metaengine.NewReader[readmodels.AgentAuditView](sys.MetaEngine(), agentAuditCollection)}
}

// Append stores one audit row; the decision trims the owner's oldest
// rows past the cap in the same event batch.
//
// brandid-lint:ignore(BD003) audit vocabulary: raw strings are the journal-line tokens (the greppable-line contract)
func (s *AgentAudit) Append(ctx context.Context, owner domain.Extension, actor, event, subject, outcome string) error { //nolint:branching-flow // verbatim port of the retired engine store's Append signature (parity rule)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, agentAuditCommand, owner.String(),
		func(base *command.BasicCommand) *AppendAgentAudit {
			return &AppendAgentAudit{
				BasicCommand: base,
				Owner:        owner.String(),
				Actor:        actor,
				Event:        event,
				Subject:      subject,
				Outcome:      outcome,
				Key:          domain.GenerateAgentAuditKey(),
				At:           time.Now(),
			}
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.agent_audit_append", "dispatch %s", agentAuditCommand)
	}
	return nil
}

// Recent returns the owner's newest audit rows (newest first), or nil
// when none exist — the caller owns the absent-vs-empty vocabulary.
func (s *AgentAudit) Recent(ctx context.Context, owner domain.Extension, limit int) ([]readmodels.AgentAuditView, error) {
	if limit <= 0 {
		limit = 1
	}
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("seq", true),
		metaengine.WithLimit(limit),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.agent_audit_recent", "scan audit rows")
	}
	if len(views) == 0 {
		return nil, nil
	}
	return views, nil
}
