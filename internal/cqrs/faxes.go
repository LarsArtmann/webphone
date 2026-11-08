// Fax aggregate (SUPERB Phase 6, A39): one stream per fax job — the
// job's whole life is a handful of facts, so the per-job key keeps
// state restores trivial (no snapshot strategy; the streams are tiny
// and terminal). faxMoves stays the ONE legality home: the decider
// judges every lifecycle move against the table (the service's own
// FaxAdvance drive was and remains the caller-side sight of the same
// rule); the inbound "received" fact is terminal — no move starts or
// ends there.
package cqrs

import (
	"context"
	"fmt"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// FaxesStreamType is the fax-job stream family; the key is the fax id
// itself (a valid StreamID; never parsed back — the facts carry the
// readable fields).
const FaxesStreamType = "fax"

// FaxStreamKey builds one fax job's stream key from its id.
func FaxStreamKey(id domain.FaxID) string { return id.String() }

// RecordFax persists a fax job's birth row: an outbound send born
// queued (faxMoves' only entry state) or an inbound received fact. The
// service owns WHICH (NewOutboundFaxJob/NewInboundFaxJob have no
// cross-over construction path); the decider owns "a stream holds
// exactly one job".
type RecordFax struct {
	*command.BasicCommand
	Job domain.FaxJob
}

// AdvanceFaxStatus moves one recorded job's lifecycle by its target
// status. Legality is faxMoves' (from → to) sight; an illegal move is
// a zero-event dispatch (the settled state answers — the same-status
// verdict refresh stays LEGAL by design: a later page count rides it).
// The At stamp is the caller's clock: facts carry it as AdvancedAt.
type AdvanceFaxStatus struct {
	*command.BasicCommand
	FaxID    domain.FaxID
	Delivery domain.FaxDelivery
	At       time.Time
}

// ExpungeFax is the retention tombstone: the decision emits the keyed
// removal fact; the caller (the sweep) trims the stream's bytes after
// the fold is visible.
type ExpungeFax struct {
	*command.BasicCommand
	FaxID domain.FaxID
}

// faxState is the job row the decider folds; exists distinguishes a
// trimmed/never-recorded stream from a live one.
type faxState struct {
	exists bool
	job    domain.FaxJob
}

func (faxState) initial() faxState { return faxState{} }

// applyFaxState folds the wire facts into the typed row. Journal-
// trusted: only parsed domain values were ever appended.
func applyFaxState(s faxState, evt event.Event) (faxState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.FaxReceipted{})):
		fact, err := event.DecodePayloadAuto[domain.FaxReceipted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.fax_fold", "decode FaxReceipted")
		}
		s.job = domain.FaxJob{
			ID:           domain.MustFaxID(fact.Key.ID),
			Owner:        domain.MustParseExtension(fact.Owner),
			Remote:       domain.MustParsePhone(fact.Remote),
			Direction:    fact.Direction,
			Status:       fact.Status,
			Pages:        fact.Pages,
			DocumentPath: fact.DocumentPath,
			ProviderRef:  fact.ProviderRef,
			Error:        fact.Error,
			CreatedAt:    fact.CreatedAt,
			UpdatedAt:    fact.UpdatedAt,
		}
		s.exists = true
	case event.Type(metaengine.EventTypeName(domain.FaxStatusAdvanced{})):
		fact, err := event.DecodePayloadAuto[domain.FaxStatusAdvanced](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.fax_fold", "decode FaxStatusAdvanced")
		}
		s.job.Status = fact.Status
		if fact.ProviderRef != "" {
			s.job.ProviderRef = fact.ProviderRef
		}
		s.job.Error = fact.Error
		s.job.Pages = fact.Pages
		s.job.UpdatedAt = fact.AdvancedAt
	case event.Type(metaengine.EventTypeName(domain.FaxExpunged{})):
		s = faxState{}.initial()
	}
	return s, nil
}

func decideRecordFax(s faxState, ver event.Version, cmd RecordFax) ([]event.Event, error) {
	if s.exists {
		return nil, errorfamily.NewRejection("cqrs.fax_duplicate",
			fmt.Sprintf("fax %s already recorded", cmd.Job.ID))
	}
	fact := domain.FaxReceipted{
		Key:          domain.FaxKey{ID: cmd.Job.ID.String()},
		Owner:        cmd.Job.Owner.String(),
		Remote:       cmd.Job.Remote.String(),
		Direction:    cmd.Job.Direction,
		Status:       cmd.Job.Status,
		Pages:        cmd.Job.Pages,
		DocumentPath: cmd.Job.DocumentPath,
		ProviderRef:  cmd.Job.ProviderRef,
		Error:        cmd.Job.Error,
		CreatedAt:    cmd.Job.CreatedAt,
		UpdatedAt:    cmd.Job.UpdatedAt,
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.FaxReceipted{})),
		cmd.StreamID(), FaxesStreamType, ver+1, fact)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_record", "build FaxReceipted")
	}
	return []event.Event{evt}, nil
}

func decideAdvanceFaxStatus(s faxState, ver event.Version, cmd AdvanceFaxStatus) ([]event.Event, error) {
	if !s.exists {
		return nil, errorfamily.NewRejection("cqrs.fax_missing",
			fmt.Sprintf("fax %s", cmd.FaxID))
	}
	// The inbound received fact has no lifecycle: every move against it
	// is inert (faxMoves has no row touching "received" as a from).
	if !domain.FaxMayMoveTo(s.job.Status, cmd.Delivery.Status) {
		return nil, nil
	}
	fact := domain.FaxStatusAdvanced{
		Key:         domain.FaxKey{ID: cmd.FaxID.String()},
		Status:      cmd.Delivery.Status,
		ProviderRef: cmd.Delivery.ProviderRef,
		Error:       cmd.Delivery.Error,
		Pages:       cmd.Delivery.Pages,
		AdvancedAt:  cmd.At,
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.FaxStatusAdvanced{})),
		cmd.StreamID(), FaxesStreamType, ver+1, fact)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_advance", "build FaxStatusAdvanced")
	}
	return []event.Event{evt}, nil
}

func decideExpungeFax(s faxState, ver event.Version, cmd ExpungeFax) ([]event.Event, error) {
	if !s.exists {
		return nil, nil
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.FaxExpunged{})),
		cmd.StreamID(), FaxesStreamType, ver+1, domain.FaxExpunged{Key: domain.FaxKey{ID: cmd.FaxID.String()}})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_expunge", "build FaxExpunged")
	}
	return []event.Event{evt}, nil
}

// RegisterFaxes wires the fax aggregate into a System. No snapshot
// strategy: fax streams are born, move a few times, and end — the
// EveryNEvents machinery would only add writes.
func RegisterFaxes(sys *system.System) error {
	if err := system.RegisterDecider[faxState](sys, FaxesStreamType, decider.Decider[faxState]{
		Initial: faxState{}.initial(),
		Apply:   applyFaxState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.fax_register", "register fax decider")
	}
	if err := system.RegisterCommand[*RecordFax, faxState](sys, "fax.record",
		func(_ context.Context, cmd *RecordFax) system.Op[faxState] {
			return system.Execute[faxState](context.Background(), cmd.StreamID(), FaxesStreamType,
				func(state faxState, ver event.Version) ([]event.Event, error) {
					return decideRecordFax(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.fax_register", "register fax.record")
	}
	if err := system.RegisterCommand[*AdvanceFaxStatus, faxState](sys, "fax.advance_status",
		func(_ context.Context, cmd *AdvanceFaxStatus) system.Op[faxState] {
			return system.Execute[faxState](context.Background(), cmd.StreamID(), FaxesStreamType,
				func(state faxState, ver event.Version) ([]event.Event, error) {
					return decideAdvanceFaxStatus(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.fax_register", "register fax.advance_status")
	}
	if err := system.RegisterCommand[*ExpungeFax, faxState](sys, "fax.expunge",
		func(_ context.Context, cmd *ExpungeFax) system.Op[faxState] {
			return system.Execute[faxState](context.Background(), cmd.StreamID(), FaxesStreamType,
				func(state faxState, ver event.Version) ([]event.Event, error) {
					return decideExpungeFax(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.fax_register", "register fax.expunge")
	}
	return nil
}
