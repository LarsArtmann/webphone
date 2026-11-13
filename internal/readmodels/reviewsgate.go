// The call-reviews seam's GATE + VIEW surface: the interface the
// server consumes and the projection row the journal backend folds.
// The engine-backed store that used to implement the gate here retired
// with the readmodels engine (2026-11-11 storage-ladder train B) —
// internal/cqrs/callreviews.go is the one implementation now.
package readmodels

import (
	"context"
	"errors"

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

// CallReviewView is the projection row ONE journal fold produces for a
// call's review slot: the automatic post-call review's state plus the
// present-state draft dimensions (meaningful only when State is
// present).
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the domain fact verbatim (the fold is a plain insert).
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

// ReviewsGate is the post-call-review seam as the server surfaces see
// it (slot read, wholesale-replace upsert). The journal-backed service
// (internal/cqrs) is the implementation.
type ReviewsGate interface {
	Get(ctx context.Context, owner domain.Extension, callID string) (*CallReviewView, error)
	Set(ctx context.Context, owner domain.Extension, callID string, state domain.ReviewState, draftID, text, language string, confidence int) error
}
