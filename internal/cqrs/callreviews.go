// The call-reviews aggregate (SUPERB Phase 4, A28): the automatic
// post-call review slot as a per-owner stream (one slot per (owner,
// call); the stream is per OWNER for the same reason as notes — the
// 2000 cap REJECTS from the DECISION, never trims). Save semantics are
// wholesale replace: pending → present → failed all ride the same
// door, and the decider snapshots the whole row (the system-written
// doctrine: NO owner PUT exists on this surface — the command shape
// carries no ownership ceremony because nothing user-authored writes
// it). ErrReviewListFull stays a Rejection (the 422 contract).
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
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// ReviewsStreamType keys per-owner call-review streams.
const ReviewsStreamType = "call_reviews"

// reviewsMaxPerOwner mirrors readmodels.ReviewsMaxPerExtension at the
// aggregate.
const reviewsMaxPerOwner = readmodels.ReviewsMaxPerExtension

// SaveCallReview replaces the owner's review slot for one call
// wholesale (the automatic job and the manual summarize path share the
// door). A NEW call's review past the cap answers the 422 contract.
type SaveCallReview struct {
	*command.BasicCommand
	Owner      string
	CallID     string
	State      domain.ReviewState
	DraftID    string
	Text       string
	Language   string
	Confidence int
	//cqrs-lint:ignore(C013) command INPUT, not a journaled payload — the detector's shape guess misnames it a projection view; the journaled fact stores UpdatedAt and the fold reads .Unix()
	At time.Time // stamped at the service seam; the decision stays deterministic
}

// reviewSlot is one call's review-slot state.
type reviewSlot struct {
	State      domain.ReviewState
	DraftID    string
	Text       string
	Language   string
	Confidence int
	UpdatedAt  int64 // unix seconds — the view's numeric order
}

// reviewsState is the owner's review slots keyed by call id.
type reviewsState struct {
	Slots map[string]reviewSlot
}

func (reviewsState) initial() reviewsState { return reviewsState{Slots: map[string]reviewSlot{}} }

// reviewStreamEventTypes is the aggregate's stream vocabulary: the
// domain decl's Events entry AND the StrictApply guard list (an unknown
// type on the stream fails replay loudly instead of silently skipping).
var reviewStreamEventTypes = []event.Type{
	event.Type(metaengine.EventTypeName(domain.CallReviewSaved{})),
}

func applyReviewsState(s reviewsState, evt event.Event) (reviewsState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.CallReviewSaved{})):
		fact, err := event.DecodePayloadAuto[domain.CallReviewSaved](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.review_fold", "decode CallReviewSaved")
		}
		s.Slots[fact.CallID] = reviewSlot{
			State:      fact.State,
			DraftID:    fact.DraftID,
			Text:       fact.Text,
			Language:   fact.Language,
			Confidence: fact.Confidence,
			UpdatedAt:  fact.UpdatedAt.Unix(),
		}
	}
	return s, nil
}

func decideSaveReview(s reviewsState, ver event.Version, cmd SaveCallReview) ([]event.Event, error) {
	if _, ok := s.Slots[cmd.CallID]; !ok && len(s.Slots) >= reviewsMaxPerOwner {
		return nil, errorfamily.WrapRejectionf(readmodels.ErrReviewListFull, "cqrs.review_list_full",
			"%s at %d reviewed calls", cmd.Owner, reviewsMaxPerOwner)
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.CallReviewSaved{})), cmd.StreamID(), ReviewsStreamType, ver.Increment(), domain.CallReviewSaved{
		Key:        domain.CallReviewKey{Owner: cmd.Owner, CallID: cmd.CallID},
		Owner:      cmd.Owner,
		CallID:     cmd.CallID,
		State:      cmd.State,
		DraftID:    cmd.DraftID,
		Text:       cmd.Text,
		Language:   cmd.Language,
		Confidence: cmd.Confidence,
		UpdatedAt:  cmd.At,
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.review_save", "build CallReviewSaved")
	}
	return []event.Event{evt}, nil
}

// RegisterReviews wires the call-reviews aggregate into a System.
func RegisterReviews(sys *system.System) error {
	if err := system.RegisterDecider[reviewsState](sys, ReviewsStreamType, decider.Decider[reviewsState]{
		Initial: reviewsState{}.initial(),
		Apply:   decider.StrictApply(applyReviewsState, reviewStreamEventTypes),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.review_register", "register reviews decider")
	}
	if err := system.RegisterCommand[*SaveCallReview, reviewsState](sys, "review.save",
		func(_ context.Context, cmd *SaveCallReview) system.Op[reviewsState] {
			return system.Execute[reviewsState](context.Background(), cmd.StreamID(), ReviewsStreamType,
				func(state reviewsState, ver event.Version) ([]event.Event, error) {
					return decideSaveReview(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.review_register", "register review.save")
	}
	return nil
}

// reviewsKeyInput is the projection's key-extraction input.
type reviewsKeyInput struct {
	Key domain.CallReviewKey
}

// reviewsCollection names the system engine's review projection.
const reviewsCollection = "call_reviews"

// reviewsProjection folds review facts into the SAME view shape the
// server reads today (readmodels.CallReviewView).
func reviewsProjection() metaengine.QueryDecl[reviewsKeyInput, readmodels.CallReviewView] {
	return metaengine.Query[reviewsKeyInput, readmodels.CallReviewView](reviewsCollection,
		metaengine.OnRecord(domain.CallReviewSaved{},
			func(_ record.Record, e domain.CallReviewSaved) (domain.CallReviewKey, readmodels.CallReviewView) {
				return e.Key, readmodels.CallReviewView{
					Owner:      e.Owner,
					CallID:     e.CallID,
					State:      e.State.String(),
					DraftID:    e.DraftID,
					Text:       e.Text,
					Language:   e.Language,
					Confidence: e.Confidence,
					UpdatedAt:  e.UpdatedAt.Unix(),
				}
			}),
		metaengine.FilterOnField[readmodels.CallReviewView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.CallReviewView]("updated_at", false),
		metaengine.Volume(humanScaleVolume),
	)
}

// ReviewsDomainDecl is the call-reviews aggregate's share of the
// system's DomainConfig (a registration failure is a boot-time miswire
// and panics loudly).
func ReviewsDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterReviews(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register call-reviews aggregate: %v", err)) //cqrs-lint:ignore(C009) boot-time fail-fast on a composition miswire (DO-1; the decl comment owns the why)
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(reviewsProjection())},
		Events:      reviewStreamEventTypes,
	}
}

// ─── the backend the server keeps calling ──────────────────────────

// Reviews is the dispatch-backed post-call-review store: the same
// contract the engine-backed store exposed (nil-when-absent
// owner-scoped Get, wholesale-replace Set, per-owner cap for NEW
// calls as the Rejection).
type Reviews struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.CallReviewView]
	vers   streamVersions
	mu     sync.Mutex
}

var _ readmodels.ReviewsGate = (*Reviews)(nil)

// NewReviews builds the backend over a System.
func NewReviews(sys *system.System) *Reviews {
	return &Reviews{sys: sys, reader: metaengine.NewReader[readmodels.CallReviewView](sys.MetaEngine(), reviewsCollection), vers: streamVersions{}}
}

// Get returns the owner's review for one call, or nil when none
// exists.
func (s *Reviews) Get(ctx context.Context, owner domain.Extension, callID string) (*readmodels.CallReviewView, error) {
	view, found, err := s.reader.Get(ctx, domain.CallReviewKey{Owner: owner.String(), CallID: callID})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.review_get", "load review slot for %s", callID)
	}
	if !found || view.Owner != owner.String() {
		return nil, nil
	}
	return &view, nil
}

// Set replaces the owner's review for one call wholesale. A NEW call
// past the cap refuses with the ErrReviewListFull contract.
func (s *Reviews) Set(ctx context.Context, owner domain.Extension, callID string, state domain.ReviewState, draftID, text, language string, confidence int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := dispatchOwnerCommand(ctx, s.sys, "review.save", owner.String(),
		func(base *command.BasicCommand) *SaveCallReview {
			return &SaveCallReview{
				BasicCommand: base,
				Owner:        owner.String(),
				CallID:       callID,
				State:        state,
				DraftID:      draftID,
				Text:         text,
				Language:     language,
				Confidence:   confidence,
				At:           time.Now(),
			}
		}); err != nil {
		if errors.Is(err, readmodels.ErrReviewListFull) {
			return err
		}
		return errorfamily.WrapInfrastructuref(err, "cqrs.review_save", "dispatch review.save")
	}
	awaitOwnWrite(ctx, s.sys, ReviewsStreamType, owner.String(), reviewsCollection, s.vers)
	return nil
}
