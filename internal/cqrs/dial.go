// Package cqrs owns webphone's command/event-sourced domain: deciders,
// commands, and the journal projections riding the system composition
// root (the SUPERB plan). Each aggregate lands here with its commands
// and folds; the package imports domain facts only — never server or
// the app — so the arch boundary (services stay framework-free) holds.
package cqrs

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// DialStreamType keys every dial-window stream: ONE stream per owner
// (the cap unit — 500 slots per extension), the stream id being the
// owner's extension string.
const DialStreamType = "dial"

// dialWindowsMaxSlots mirrors readmodels.DialWindowsMaxPerExtension at
// the aggregate: the decider emits expunge events for the oldest slots
// beyond it, so the cap is enforced by the decision, not a post-write
// sweep.
const dialWindowsMaxSlots = readmodels.DialWindowsMaxPerExtension

// dialOverridesMaxSlots bounds the one-shot bypass slots the same way.
const dialOverridesMaxSlots = readmodels.DialWindowsMaxPerExtension

// dialWindowMaxPerOwner documented cap parity (readmodels' constant is
// the one home; this alias keeps the intent readable at the decider).
const dialWindowMaxPerOwner = dialWindowsMaxSlots

// ─── commands ──────────────────────────────────────────────────────

// AttemptDial is the gate's check+record transaction as ONE command:
// the decider either spends an unconsumed override and records the
// attempt (allowed), denies inside the cooldown window (nothing
// recorded — a retry loop must not extend the window it fights), or
// records the attempt. Cooldown and Now ride the command as POLICY
// params (the aggregate owns the facts, not the policy of the day).
type AttemptDial struct {
	*command.BasicCommand
	// Owner is the extension string — carried explicitly because the
	// stream id's String() renders a branded prefix ("StreamMarker:…")
	// the facts must NOT carry (they key the owner-scoped rows).
	Owner    string
	Remote   string
	Cooldown time.Duration
	Now      time.Time
}

// SaveOverride arms the one-shot bypass for (owner, remote): the NEXT
// AttemptDial spends it (even mid-window) and dials.
type SaveOverride struct {
	*command.BasicCommand
	// Owner as on AttemptDial (branded stream ids never reach facts).
	Owner  string
	Remote string
	Now    time.Time
}

// DialDeniedError is the decider's DENY verdict: a typed rejection
// carrying the slot's last-at stamp (for remaining-window math). The
// dial gate maps it to its 200 verdict — nothing was recorded.
type DialDeniedError struct {
	LastAt time.Time
}

func (e *DialDeniedError) Error() string {
	return "dial denied: inside the cooldown window"
}

// ─── aggregate state ───────────────────────────────────────────────

// dialOverrideSlot is one (owner, remote) one-shot bypass slot.
type dialOverrideSlot struct {
	At       time.Time
	Consumed bool
}

// dialState is the per-owner aggregate: every cooldown slot and
// override the owner holds. Stream-scoped (one owner), so the maps key
// on the remote alone.
type dialState struct {
	Windows   map[string]int64 // remote → last allowed attempt (unix)
	Overrides map[string]dialOverrideSlot
}

func (dialState) initial() dialState {
	return dialState{Windows: map[string]int64{}, Overrides: map[string]dialOverrideSlot{}}
}

func applyDialState(s dialState, evt event.Event) (dialState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.DialAttempted{})):
		fact, err := event.DecodePayloadAuto[domain.DialAttempted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.dial_fold", "decode DialAttempted")
		}
		s.Windows[fact.Remote] = fact.At.Unix()
	case event.Type(metaengine.EventTypeName(domain.DialOverrideSaved{})):
		fact, err := event.DecodePayloadAuto[domain.DialOverrideSaved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.dial_fold", "decode DialOverrideSaved")
		}
		s.Overrides[fact.Remote] = dialOverrideSlot{At: fact.At, Consumed: fact.Consumed}
	case event.Type(metaengine.EventTypeName(domain.DialWindowExpunged{})):
		fact, err := event.DecodePayloadAuto[domain.DialWindowExpunged](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.dial_fold", "decode DialWindowExpunged")
		}
		delete(s.Windows, fact.Key.Remote)
	case event.Type(metaengine.EventTypeName(domain.DialOverrideExpunged{})):
		fact, err := event.DecodePayloadAuto[domain.DialOverrideExpunged](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.dial_fold", "decode DialOverrideExpunged")
		}
		delete(s.Overrides, fact.Key.Remote)
	}
	return s, nil
}

// ─── the decision ──────────────────────────────────────────────────

// decideDialAttempt is the gate: override → spend + record; inside the
// window → deny (nothing recorded); else record. Cap overflow emits
// expunge events for the oldest slots beyond the bound.
func decideDialAttempt(s dialState, ver event.Version, cmd AttemptDial) ([]event.Event, error) {
	streamID := cmd.StreamID()
	owner := cmd.Owner
	now := cmd.Now

	if slot, ok := s.Overrides[cmd.Remote]; ok && !slot.Consumed {
		consume, err := dialEvent(domain.DialOverrideSaved{
			Key:   domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
			Owner: owner, Remote: cmd.Remote, Consumed: true, At: now,
		}, streamID, ver)
		if err != nil {
			return nil, err
		}
		attempt, err := dialAttemptEvents(s, cmd, ver+1)
		if err != nil {
			return nil, err
		}
		return append([]event.Event{consume}, attempt...), nil
	}

	if at, ok := s.Windows[cmd.Remote]; ok && cmd.Cooldown > 0 &&
		now.Sub(time.Unix(at, 0)) < cmd.Cooldown {
		return nil, &DialDeniedError{LastAt: time.Unix(at, 0)}
	}

	return dialAttemptEvents(s, cmd, ver)
}

// dialAttemptEvents emits the attempt fact plus the cap trims it
// caused (oldest slots beyond the per-owner bound).
func dialAttemptEvents(s dialState, cmd AttemptDial, ver event.Version) ([]event.Event, error) {
	streamID := cmd.StreamID()
	owner := cmd.Owner

	attempt, err := dialEvent(domain.DialAttempted{
		Key:   domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
		Owner: owner, Remote: cmd.Remote, At: cmd.Now,
	}, streamID, ver)
	if err != nil {
		return nil, err
	}
	events := []event.Event{attempt}

	windows := make(map[string]int64, len(s.Windows)+1)
	maps.Copy(windows, s.Windows)
	windows[cmd.Remote] = cmd.Now.Unix()
	events = append(events, dialExpungeEvents(windows, owner, streamID, event.Type(metaengine.EventTypeName(domain.DialWindowExpunged{})), ver+1)...)

	return events, nil
}

// dialExpungeEvents returns expunge facts for the oldest slots beyond
// the per-owner cap (the map is this owner's; deterministic order by
// stamp, then name).
func dialExpungeEvents(slots map[string]int64, owner string, streamID id.StreamID, expungeType event.Type, startVer event.Version) []event.Event {
	if len(slots) <= dialWindowMaxPerOwner {
		return nil
	}
	type stamped struct {
		remote string
		at     int64
	}
	all := make([]stamped, 0, len(slots))
	for remote, at := range slots {
		all = append(all, stamped{remote, at})
	}
	// Oldest first; ties broken by remote for determinism.
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && (all[j].at < all[j-1].at || (all[j].at == all[j-1].at && all[j].remote < all[j-1].remote)); j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	overflow := all[:len(all)-dialWindowMaxPerOwner]
	facts := make([]event.Event, 0, len(overflow))
	ver := startVer
	for _, old := range overflow {
		fact := domain.DialWindowExpunged{Key: domain.DialWindowKey{Owner: owner, Remote: old.remote}}
		evt, err := dialEvent(fact, streamID, ver)
		if err != nil {
			continue // an unemittable expunge must never block the attempt
		}
		facts = append(facts, evt)
		ver++
	}
	return facts
}

// decideSaveOverride arms the bypass (replace semantics) plus cap
// trims on the override slots.
func decideSaveOverride(s dialState, ver event.Version, cmd SaveOverride) ([]event.Event, error) {
	streamID := cmd.StreamID()
	owner := cmd.Owner

	arm, err := dialEvent(domain.DialOverrideSaved{
		Key:   domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
		Owner: owner, Remote: cmd.Remote, Consumed: false, At: cmd.Now,
	}, streamID, ver)
	if err != nil {
		return nil, err
	}
	events := []event.Event{arm}

	slots := make(map[string]int64, len(s.Overrides)+1)
	for remote, slot := range s.Overrides {
		slots[remote] = slot.At.Unix()
	}
	slots[cmd.Remote] = cmd.Now.Unix()
	if len(slots) > dialOverridesMaxSlots {
		type stamped struct {
			remote string
			at     int64
		}
		all := make([]stamped, 0, len(slots))
		for remote, at := range slots {
			all = append(all, stamped{remote, at})
		}
		for i := 1; i < len(all); i++ {
			for j := i; j > 0 && (all[j].at < all[j-1].at || (all[j].at == all[j-1].at && all[j].remote < all[j-1].remote)); j-- {
				all[j], all[j-1] = all[j-1], all[j]
			}
		}
		next := ver + 1
		for _, old := range all[:len(all)-dialOverridesMaxSlots] {
			evt, err := dialEvent(domain.DialOverrideExpunged{
				Key: domain.DialWindowKey{Owner: owner, Remote: old.remote},
			}, streamID, next)
			if err != nil {
				continue
			}
			events = append(events, evt)
			next++
		}
	}
	return events, nil
}

func dialEvent(fact any, streamID id.StreamID, ver event.Version) (event.Event, error) {
	return event.New(event.Type(metaengine.EventTypeName(fact)), streamID, DialStreamType, ver+1, fact)
}

// ─── registration + projections ────────────────────────────────────

// RegisterDial wires the dial aggregate into a System: the decider,
// both commands, and both projections. Called from the composition
// root's domain closure.
func RegisterDial(sys *system.System) error {
	if err := system.RegisterDecider[dialState](sys, DialStreamType, decider.Decider[dialState]{
		Initial: dialState{}.initial(),
		Apply:   applyDialState,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_register", "register dial decider")
	}
	if err := system.RegisterCommand[*AttemptDial, dialState](sys, "dial.attempt",
		func(_ context.Context, cmd *AttemptDial) system.Op[dialState] {
			return system.Execute[dialState](context.Background(), cmd.StreamID(), DialStreamType,
				func(state dialState, ver event.Version) ([]event.Event, error) {
					return decideDialAttempt(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_register", "register dial.attempt")
	}
	if err := system.RegisterCommand[*SaveOverride, dialState](sys, "dial.override",
		func(_ context.Context, cmd *SaveOverride) system.Op[dialState] {
			return system.Execute[dialState](context.Background(), cmd.StreamID(), DialStreamType,
				func(state dialState, ver event.Version) ([]event.Event, error) {
					return decideSaveOverride(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_register", "register dial.override")
	}
	return nil
}

// DialDomainDecl is the dial aggregate's share of the system's
// DomainConfig: projections, the declared fact universe, and the
// registration closure (a registration failure is a boot-time miswire
// — the composition-root fail-fast class — so the closure panics
// loudly instead of half-registering).
func DialDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterDial(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register dial aggregate: %v", err))
			}
		},
		Projections: []system.ProjectionDeclaration{
			system.RawQuery(dialWindowsProjection()),
			system.RawQuery(dialOverridesProjection()),
		},
		Events: []event.Type{
			event.Type(metaengine.EventTypeName(domain.DialAttempted{})),
			event.Type(metaengine.EventTypeName(domain.DialOverrideSaved{})),
			event.Type(metaengine.EventTypeName(domain.DialWindowExpunged{})),
			event.Type(metaengine.EventTypeName(domain.DialOverrideExpunged{})),
		},
	}
}
