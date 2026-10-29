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

// DialWindowsMaxPerExtension bounds one extension's cooldown slots.
// Cooldown rows are SYSTEM facts, not user content: the trim policy is
// the transcripts posture (expunge the oldest), not the notes posture
// (reject) — nothing the owner authored is ever lost to this cap.
const DialWindowsMaxPerExtension = 500

// DialWindowView is the engine-side projection of one (owner, remote)
// cooldown slot: the LAST allowed dial's timestamp.
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type DialWindowView struct {
	Owner  string `json:"owner"`
	Remote string `json:"remote"`
	At     int64  `json:"at"`
}

// DialOverrideView is the engine-side projection of the one-shot
// bypass slot (issue #7): Consumed=true marks an override the gate
// already spent.
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type DialOverrideView struct {
	Owner    string `json:"owner"`
	Remote   string `json:"remote"`
	Consumed bool   `json:"consumed"`
	At       int64  `json:"at"`
}

const (
	dialWindowsCollection   = "dial_windows"
	dialOverridesCollection = "dial_overrides"
)

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type dialWindowInput struct{ Key domain.DialWindowKey }

// dialWindowsQuery declares the cooldown read model: ONE slot per
// (owner, remote), each allowed dial replacing the row.
var dialWindowsQuery = metaengine.Query[dialWindowInput, DialWindowView](dialWindowsCollection,
	metaengine.OnRecord(domain.DialAttempted{},
		func(_ record.Record, e domain.DialAttempted) (domain.DialWindowKey, DialWindowView) {
			return e.Key, DialWindowView{
				Owner:  e.Owner,
				Remote: e.Remote,
				At:     e.At.Unix(),
			}
		}),
	metaengine.OnRecord(domain.DialWindowExpunged{}, metaengine.Remove[DialWindowView]()),
	metaengine.FilterOnField[DialWindowView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[DialWindowView]("at", false),
	metaengine.Volume(humanScaleVolume),
)

// dialOverridesQuery declares the one-shot bypass read model.
var dialOverridesQuery = metaengine.Query[dialWindowInput, DialOverrideView](dialOverridesCollection,
	metaengine.OnRecord(domain.DialOverrideSaved{},
		func(_ record.Record, e domain.DialOverrideSaved) (domain.DialWindowKey, DialOverrideView) {
			return e.Key, DialOverrideView{
				Owner:    e.Owner,
				Remote:   e.Remote,
				Consumed: e.Consumed,
				At:       e.At.Unix(),
			}
		}),
	metaengine.OnRecord(domain.DialOverrideExpunged{}, metaengine.Remove[DialOverrideView]()),
	metaengine.FilterOnField[DialOverrideView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[DialOverrideView]("at", false),
	metaengine.Volume(humanScaleVolume),
)

// DialWindows is the engine-backed dial-cooldown store (issue #7):
// per-(owner, remote) last-attempt windows plus one-shot overrides.
// The gate decision and the record ride ONE mutex section — the
// check+record transaction — so a racing double-click cannot pass two
// checks before either record lands. Rides readmodels.db (the backup
// set covers it).
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Notes/Reviews) — deliberate symmetry,
// not a mixin candidate.
type DialWindows struct {
	models    *metaengine.Store
	windows   *metaengine.TypedReader[DialWindowView]
	overrides *metaengine.TypedReader[DialOverrideView]
	mu        sync.Mutex
}

func newDialWindows(models *metaengine.Store) *DialWindows {
	return &DialWindows{
		models:    models,
		windows:   metaengine.NewReader[DialWindowView](models, dialWindowsCollection),
		overrides: metaengine.NewReader[DialOverrideView](models, dialOverridesCollection),
	}
}

// Attempt is the gate's check+record transaction: an unconsumed
// override is spent and the attempt allowed; an attempt inside the
// cooldown window is DENIED (and never recorded — the window must not
// grow by retrying); anything else records the attempt and allows.
// The returned stamp is the new attempt's time (allowed) or the
// slot's last-at time (denied, for remaining-window math).
func (s *DialWindows) Attempt(ctx context.Context, owner domain.Extension, remote string, cooldown time.Duration) (allowed bool, stamp time.Time, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	key := domain.DialWindowKey{Owner: owner.String(), Remote: remote}
	if ov, found, err := s.overrides.Get(ctx, key); err != nil {
		return false, now, errorfamily.WrapInfrastructuref(err, "readmodels.dial_override_load", "load override slot")
	} else if found && !ov.Consumed {
		if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.DialOverrideSaved{}), domain.DialOverrideSaved{
			Key: key, Owner: ov.Owner, Remote: ov.Remote, Consumed: true, At: now,
		}); err != nil {
			return false, now, errorfamily.WrapInfrastructuref(err, "readmodels.dial_override_consume", "consume override")
		}
		return true, now, s.recordAttempt(ctx, owner, remote, now)
	}

	if win, found, err := s.windows.Get(ctx, key); err != nil {
		return false, now, errorfamily.WrapInfrastructuref(err, "readmodels.dial_window_load", "load window slot")
	} else if found && cooldown > 0 && now.Sub(time.Unix(win.At, 0)) < cooldown {
		return false, time.Unix(win.At, 0), nil
	}

	return true, now, s.recordAttempt(ctx, owner, remote, now)
}

// ArmOverride arms the one-shot bypass for (owner, remote): the NEXT
// Attempt spends it (even mid-window) and dials.
func (s *DialWindows) ArmOverride(ctx context.Context, owner domain.Extension, remote string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.DialOverrideSaved{}), domain.DialOverrideSaved{
		Key:    domain.DialWindowKey{Owner: owner.String(), Remote: remote},
		Owner:  owner.String(),
		Remote: remote,
		At:     time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.dial_override_arm", "arm override")
	}
	return s.trimWindows(ctx, owner.String())
}

// Last returns the slot's last allowed dial (found=false when none).
func (s *DialWindows) Last(ctx context.Context, owner domain.Extension, remote string) (time.Time, bool, error) {
	win, found, err := s.windows.Get(ctx, domain.DialWindowKey{Owner: owner.String(), Remote: remote})
	if err != nil {
		return time.Time{}, false, errorfamily.WrapInfrastructuref(err, "readmodels.dial_window_load", "load window slot")
	}
	if !found {
		return time.Time{}, false, nil
	}
	return time.Unix(win.At, 0), true, nil
}

// DialSlotsMaxPerResponse bounds the ACTIVE-slot listing (the typeahead's
// cooldown decoration) — the owner's cap is 500 rows, but a page only
// phrases the freshest ones; oldest-first beyond this bound is invisible
// state, exactly like the trim posture.
const DialSlotsMaxPerResponse = 50

// ActiveSlot is one live cooldown window: the remote is inside its
// cooldown until LastAt+Window.
//
// branching-flow COMPOSITION_mixin (baselined): mirrors DialWindowView
// plus the window length the verdict needs (the engine row stores only
// the stamp). brandid-lint:ignore(BD003) wire DTO: the remote is the dialable string (strong-id wire ruling)
type ActiveSlot struct {
	Remote string `json:"remote"`
	LastAt int64  `json:"last_at"`
	Window int64  `json:"window"`
}

// Active returns the owner's cooldown windows still inside their window
// (newest first), for the typeahead's row decoration. Expired slots are
// skipped lazily — the Attempt path is what expunges/replaces rows; this
// read never writes.
func (s *DialWindows) Active(ctx context.Context, owner domain.Extension, cooldown time.Duration) ([]ActiveSlot, error) {
	wins, err := s.windows.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("at", true),
		metaengine.WithLimit(DialSlotsMaxPerResponse),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.dial_window_load", "scan active windows")
	}
	now := time.Now()
	slots := make([]ActiveSlot, 0, len(wins))
	windowSecs := int64(cooldown / time.Second)
	for _, v := range wins {
		last := time.Unix(v.At, 0)
		if now.Sub(last) >= cooldown {
			continue
		}
		slots = append(slots, ActiveSlot{Remote: v.Remote, LastAt: v.At, Window: windowSecs})
	}
	return slots, nil
}

func (s *DialWindows) recordAttempt(ctx context.Context, owner domain.Extension, remote string, at time.Time) error {
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.DialAttempted{}), domain.DialAttempted{
		Key:    domain.DialWindowKey{Owner: owner.String(), Remote: remote},
		Owner:  owner.String(),
		Remote: remote,
		At:     at,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.dial_attempt", "record dial attempt")
	}
	return s.trimWindows(ctx, owner.String())
}

// trimWindows expunges the owner's oldest rows past the cap on BOTH
// collections (the transcripts posture: system facts are trimmed,
// never user content).
func (s *DialWindows) trimWindows(ctx context.Context, owner string) error {
	wins, err := s.windows.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner),
		metaengine.WithSort("at", true),
		metaengine.WithLimit(DialWindowsMaxPerExtension+1),
	)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.dial_trim", "scan window trim window")
	}
	for _, v := range safeTail(wins) {
		if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.DialWindowExpunged{}), domain.DialWindowExpunged{
			Key: domain.DialWindowKey{Owner: v.Owner, Remote: v.Remote},
		}); err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.dial_trim", "expunge window row")
		}
	}
	ovs, err := s.overrides.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner),
		metaengine.WithSort("at", true),
		metaengine.WithLimit(DialWindowsMaxPerExtension+1),
	)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.dial_trim", "scan override trim window")
	}
	for _, v := range safeTail(ovs) {
		if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.DialOverrideExpunged{}), domain.DialOverrideExpunged{
			Key: domain.DialWindowKey{Owner: v.Owner, Remote: v.Remote},
		}); err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.dial_trim", "expunge override row")
		}
	}
	return nil
}

// safeTail returns the rows past the cap (an empty slice for
// at-cap scans).
func safeTail[T any](views []T) []T {
	if len(views) <= DialWindowsMaxPerExtension {
		return nil
	}
	return views[DialWindowsMaxPerExtension:]
}
