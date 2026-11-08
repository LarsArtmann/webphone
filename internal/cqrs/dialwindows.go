package cqrs

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

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
	cmd := &AttemptDial{BasicCommand: base, Owner: owner.String(), Remote: remote, Cooldown: cooldown, Now: now}
	if err := s.sys.CommandDispatcher().Dispatch(ctx, cmd); err != nil {
		if denied, ok := errors.AsType[*DialDeniedError](err); ok {
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
	cmd := &SaveOverride{BasicCommand: base, Owner: owner.String(), Remote: remote, Now: timeNow()}
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
