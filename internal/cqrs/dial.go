// Package cqrs owns webphone's command/event-sourced domain: deciders,
// commands, and the journal projections riding the system composition
// root (the SUPERB plan). Each aggregate lands here with its commands
// and folds; the package imports domain facts only — never server or
// the app — so the arch boundary (services stay framework-free) holds.
package cqrs

import (
	"context"
	"errors"
	"fmt"
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
	Remote   string
	Cooldown time.Duration
	Now      time.Time
}

// SaveOverride arms the one-shot bypass for (owner, remote): the NEXT
// AttemptDial spends it (even mid-window) and dials.
type SaveOverride struct {
	*command.BasicCommand
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
	owner := streamID.String()
	now := cmd.Now

	if slot, ok := s.Overrides[cmd.Remote]; ok && !slot.Consumed {
		consume, err := dialEvent(domain.DialOverrideSaved{
			Key: domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
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
	owner := streamID.String()

	attempt, err := dialEvent(domain.DialAttempted{
		Key: domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
		Owner: owner, Remote: cmd.Remote, At: cmd.Now,
	}, streamID, ver)
	if err != nil {
		return nil, err
	}
	events := []event.Event{attempt}

	windows := make(map[string]int64, len(s.Windows)+1)
	for remote, at := range s.Windows {
		windows[remote] = at
	}
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
	owner := streamID.String()

	arm, err := dialEvent(domain.DialOverrideSaved{
		Key: domain.DialWindowKey{Owner: owner, Remote: cmd.Remote},
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

// dialWindowsCollection / dialOverridesCollection mirror the readmodels
// engine's collection names so operators grepping either store see one
// vocabulary.
const (
	dialWindowsCollection   = "dial_windows"
	dialOverridesCollection = "dial_overrides"

	// dialWindowsGlobalCeiling mirrors the readmodels declaration's
	// collection-global safety net.
	dialWindowsGlobalCeiling = 100_000
)

// dialWindowKeyInput is the key-extraction input for the journal
// projections (same shape the readmodels engine declares).
type dialWindowKeyInput struct{ Key domain.DialWindowKey }

// dialWindowsProjection folds DialAttempted facts into the SAME view
// shape the server reads today (readmodels.DialWindowView).
func dialWindowsProjection() metaengine.QueryDecl[dialWindowKeyInput, readmodels.DialWindowView] {
	return metaengine.Query[dialWindowKeyInput, readmodels.DialWindowView](dialWindowsCollection,
		metaengine.OnRecord(domain.DialAttempted{},
			func(_ record.Record, e domain.DialAttempted) (domain.DialWindowKey, readmodels.DialWindowView) {
				return e.Key, readmodels.DialWindowView{Key: e.Key, Owner: e.Owner, Remote: e.Remote, At: e.At.Unix()}
			}),
		metaengine.OnRecord(domain.DialWindowExpunged{}, metaengine.Remove[readmodels.DialWindowView]()),
		metaengine.FilterOnField[readmodels.DialWindowView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.DialWindowView]("at", false),
		metaengine.WithCap(dialWindowsGlobalCeiling, "at"),
		metaengine.WithCapPerGroup("owner", dialWindowsMaxSlots),
	)
}

// dialOverridesProjection folds the one-shot bypass slots.
func dialOverridesProjection() metaengine.QueryDecl[dialWindowKeyInput, readmodels.DialOverrideView] {
	return metaengine.Query[dialWindowKeyInput, readmodels.DialOverrideView](dialOverridesCollection,
		metaengine.OnRecord(domain.DialOverrideSaved{},
			func(_ record.Record, e domain.DialOverrideSaved) (domain.DialWindowKey, readmodels.DialOverrideView) {
				return e.Key, readmodels.DialOverrideView{Key: e.Key, Owner: e.Owner, Remote: e.Remote, Consumed: e.Consumed, At: e.At.Unix()}
			}),
		metaengine.OnRecord(domain.DialOverrideExpunged{}, metaengine.Remove[readmodels.DialOverrideView]()),
		metaengine.FilterOnField[readmodels.DialOverrideView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.DialOverrideView]("at", false),
		metaengine.WithCap(dialWindowsGlobalCeiling, "at"),
		metaengine.WithCapPerGroup("owner", dialOverridesMaxSlots),
	)
}

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

// ─── the backend the server keeps calling ──────────────────────────

// DialWindows is the dispatch-backed dial-cooldown service: the SAME
// surface readmodels.DialWindows exposes (the server's DialGate
// interface), answering gate verdicts from the DECISION (read-your-
// writes without re-reading the projection) and reading slot state
// through the system's projection store. One process-wide mutex keeps
// double-click dispatches serialized per owner — the repository's
// optimistic per-stream concurrency already rejects a lost race, but
// the parity port keeps the seam's serialization local and loud.
type DialWindows struct {
	sys       *system.System
	windows   *metaengine.TypedReader[readmodels.DialWindowView]
	overrides *metaengine.TypedReader[readmodels.DialOverrideView]
	mu        sync.Mutex
}

// NewDialWindows builds the backend over a started-or-startable
// System.
func NewDialWindows(sys *system.System) *DialWindows {
	return &DialWindows{
		sys:       sys,
		windows:   metaengine.NewReader[readmodels.DialWindowView](sys.MetaEngine(), dialWindowsCollection),
		overrides: metaengine.NewReader[readmodels.DialOverrideView](sys.MetaEngine(), dialOverridesCollection),
	}
}

// timeNow is the dial seam's clock, overridable for deterministic
// tests (window expiry, cap-expunge ordering).
var timeNow = time.Now

// Attempt is the gate's check+record transaction (see AttemptDial).
// The verdict and stamp come from the decision itself; a deny carries
// the slot's last-at via *DialDeniedError.
func (s *DialWindows) Attempt(ctx context.Context, owner domain.Extension, remote string, cooldown time.Duration) (bool, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timeNow()
	streamID, err := id.ParseStreamID(owner.String())
	if err != nil {
		return false, now, errorfamily.WrapRejectionf(err, "cqrs.dial_stream", "parse owner stream id")
	}
	base, err := command.New("dial.attempt", streamID)
	if err != nil {
		return false, now, errorfamily.WrapInfrastructuref(err, "cqrs.dial_command", "build dial.attempt")
	}
	cmd := &AttemptDial{BasicCommand: base, Remote: remote, Cooldown: cooldown, Now: now}
	if err := s.sys.CommandDispatcher().Dispatch(ctx, cmd); err != nil {
		var denied *DialDeniedError
		if errors.As(err, &denied) {
			return false, denied.LastAt, nil
		}
		return false, now, errorfamily.WrapInfrastructuref(err, "cqrs.dial_attempt", "dispatch dial.attempt")
	}
	return true, now, nil
}

// ArmOverride arms the one-shot bypass for (owner, remote).
func (s *DialWindows) ArmOverride(ctx context.Context, owner domain.Extension, remote string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	streamID, err := id.ParseStreamID(owner.String())
	if err != nil {
		return errorfamily.WrapRejectionf(err, "cqrs.dial_stream", "parse owner stream id")
	}
	base, err := command.New("dial.override", streamID)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_command", "build dial.override")
	}
	cmd := &SaveOverride{BasicCommand: base, Remote: remote, Now: timeNow()}
	if err := s.sys.CommandDispatcher().Dispatch(ctx, cmd); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_override", "dispatch dial.override")
	}
	return nil
}

// Last returns the slot's last allowed dial (found=false when none).
func (s *DialWindows) Last(ctx context.Context, owner domain.Extension, remote string) (time.Time, bool, error) {
	win, found, err := s.windows.Get(ctx, domain.DialWindowKey{Owner: owner.String(), Remote: remote})
	if err != nil {
		return time.Time{}, false, errorfamily.WrapInfrastructuref(err, "cqrs.dial_window_load", "load window slot")
	}
	if !found {
		return time.Time{}, false, nil
	}
	return time.Unix(win.At, 0), true, nil
}

// Active returns the owner's cooldown windows still inside their
// window (newest first) — the typeahead's row decoration.
func (s *DialWindows) Active(ctx context.Context, owner domain.Extension, cooldown time.Duration) ([]readmodels.ActiveSlot, error) {
	wins, err := s.windows.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("at", true),
		metaengine.WithLimit(readmodels.DialSlotsMaxPerResponse),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.dial_window_load", "scan active windows")
	}
	now := time.Now()
	slots := make([]readmodels.ActiveSlot, 0, len(wins))
	windowSecs := int64(cooldown / time.Second)
	for _, v := range wins {
		last := time.Unix(v.At, 0)
		if now.Sub(last) >= cooldown {
			continue
		}
		slots = append(slots, readmodels.ActiveSlot{Remote: v.Remote, LastAt: v.At, Window: windowSecs})
	}
	return slots, nil
}
