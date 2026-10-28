package readmodels

import (
	"context"
	"errors"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// ReviewsMaxPerExtension bounds one extension's reviewed calls (the
// notes-cap posture: a NEW call's review past the bound is REJECTED,
// never trimmed — reviews are the durable record of what the AI made
// of a call; re-reviewing an existing call stays open).
const ReviewsMaxPerExtension = 2000

// ErrReviewListFull is returned when a NEW call's review would exceed
// the per-owner cap; re-reviewing an existing call stays open.
var ErrReviewListFull = errors.New("call review list full")

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrReviewListFull: errorfamily.Rejection,
	})
}

// CallReviewView is the engine-side projection of one call's review
// slot: the automatic post-call review's state plus the present-state
// draft dimensions (meaningful only when State is present).
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the domain fact verbatim (the fold is a plain insert).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema; CallID is the opaque island token (strong-id wire ruling)
type CallReviewView struct {
	Owner      string `json:"owner"`
	CallID     string `json:"call_id"`
	State      string `json:"state"`
	DraftID    string `json:"draft_id"`
	Text       string `json:"text"`
	Language   string `json:"language"`
	Confidence int    `json:"confidence"`
	UpdatedAt  int64  `json:"updated_at"`
}

// reviewsCollection names the engine collection (query name).
const reviewsCollection = "call_reviews"

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type callReviewInput struct{ Key domain.CallReviewKey }

// reviewsQuery declares the call-reviews read model: ONE slot per
// (owner, call), upserted by CallReviewSaved facts that snapshot the
// whole slot (replace semantics — the next dispatch or re-review
// replaces the row), filtered by owner, sorted by update time.
var reviewsQuery = metaengine.Query[callReviewInput, CallReviewView](reviewsCollection,
	metaengine.OnRecord(domain.CallReviewSaved{},
		func(_ record.Record, e domain.CallReviewSaved) (domain.CallReviewKey, CallReviewView) {
			return e.Key, CallReviewView{
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
	metaengine.FilterOnField[CallReviewView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[CallReviewView]("updated_at", false),
	metaengine.Volume(humanScaleVolume),
)

// Reviews is the engine-backed post-call-review store: the durable
// record of what the AI made of one finished call, one slot per
// (owner, call id). The automatic job and the manual summarize path
// both write through Set; History renders the states. Rides
// readmodels.db (user data — the backup set covers it since the
// readmodels train).
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Notes/Contacts) — deliberate symmetry,
// not a mixin candidate.
type Reviews struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[CallReviewView]
	mu     sync.Mutex
}

func newReviews(models *metaengine.Store) *Reviews {
	return &Reviews{
		models: models,
		reader: metaengine.NewReader[CallReviewView](models, reviewsCollection),
	}
}

// Get returns the owner's review for one call, or nil when none exists
// (the Notes.Get precedent: the caller owns the absent-vs-empty
// vocabulary).
func (s *Reviews) Get(ctx context.Context, owner domain.Extension, callID string) (*CallReviewView, error) {
	view, found, err := s.reader.Get(ctx, domain.CallReviewKey{Owner: owner.String(), CallID: callID})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.review_get", "load review slot for %s", callID)
	}
	if !found || view.Owner != owner.String() {
		return nil, nil
	}
	return &view, nil
}

// Set upserts the owner's review for one call: the slot is the
// (owner, call) pair, so a save REPLACES the row wholesale (pending →
// present → failed all ride the same door). A NEW call's review past
// the per-owner cap is refused with ErrReviewListFull.
func (s *Reviews) Set(ctx context.Context, owner domain.Extension, callID string, state domain.ReviewState, draftID, text, language string, confidence int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.claimSlot(ctx, owner, callID); err != nil {
		return err
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.CallReviewSaved{}), domain.CallReviewSaved{
		Key:        domain.CallReviewKey{Owner: owner.String(), CallID: callID},
		Owner:      owner.String(),
		CallID:     callID,
		State:      state,
		DraftID:    draftID,
		Text:       text,
		Language:   language,
		Confidence: confidence,
		UpdatedAt:  time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.review_save", "apply review save")
	}
	return nil
}

// claimSlot enforces the per-owner cap for a NEW call (existing slots
// always stay open) under the store mutex; the caller emits its fact
// while still holding the mutex.
func (s *Reviews) claimSlot(ctx context.Context, owner domain.Extension, callID string) error {
	key := domain.CallReviewKey{Owner: owner.String(), CallID: callID}
	_, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.review_save", "load review slot")
	}
	if found {
		return nil
	}
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithLimit(ReviewsMaxPerExtension+1),
	)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.review_save", "count reviews for %s", owner.String())
	}
	if len(views) >= ReviewsMaxPerExtension {
		return errorfamily.WrapRejectionf(ErrReviewListFull, "readmodels.review_list_full",
			"%s at %d reviewed calls", owner.String(), ReviewsMaxPerExtension)
	}
	return nil
}
