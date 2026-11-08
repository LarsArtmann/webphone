package cqrs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func bootReviews(t *testing.T, dir string) *Reviews {
	t.Helper()
	return NewReviews(bootDomainSystem(t, dir, ReviewsDomainDecl()))
}

// awaitReview polls the projection until the owner's slot holds the
// wanted state.
func awaitReview(t *testing.T, s *Reviews, owner domain.Extension, callID, wantState string) *readmodels.CallReviewView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := s.Get(context.Background(), owner, callID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.State == wantState {
			return view
		}
		if time.Now().After(deadline) {
			t.Fatalf("review for %s never reached %q: %+v", callID, wantState, view)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestReviewReplaceSemantics pins the wholesale-replace save: pending →
// present rides the same door and REPLACES the row (the old dims never
// linger), and an absent slot reads nil.
func TestReviewReplaceSemantics(t *testing.T) {
	s := bootReviews(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if view, err := s.Get(ctx, owner, "call-a"); err != nil || view != nil {
		t.Fatalf("absent review must read nil: %v %+v", err, view)
	}

	if err := s.Set(ctx, owner, "call-a", domain.ReviewPending, "", "", "", 0); err != nil {
		t.Fatalf("Set pending: %v", err)
	}
	awaitReview(t, s, owner, "call-a", domain.ReviewPending.String())

	if err := s.Set(ctx, owner, "call-a", domain.ReviewPresent, "draft-7", "solid call, follow up Friday", "en", 87); err != nil {
		t.Fatalf("Set present: %v", err)
	}
	view := awaitReview(t, s, owner, "call-a", domain.ReviewPresent.String())
	if view.DraftID != "draft-7" || view.Confidence != 87 || view.Language != "en" {
		t.Fatalf("the present review must carry its dims: %+v", view)
	}

	if err := s.Set(ctx, owner, "call-a", domain.ReviewFailed, "", "", "", 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	view = awaitReview(t, s, owner, "call-a", domain.ReviewFailed.String())
	if view.DraftID != "" || view.Text != "" || view.Confidence != 0 {
		t.Fatalf("a replace must clear the previous dims (wholesale, not merge): %+v", view)
	}

	// Owner scoping: a second owner never sees the slot.
	if view, err := s.Get(ctx, mustExt(t, "1002"), "call-a"); err != nil || view != nil {
		t.Fatalf("owner 1002 must not read 1001's review: %v %+v", err, view)
	}
}

// TestReviewCapIsRejectionAtTheDecision pins the 2000-cap boundary at
// the DECISION: a NEW call past the cap refuses with the
// ErrReviewListFull REJECTION contract; re-reviewing an existing call
// at the cap stays open.
func TestReviewCapIsRejectionAtTheDecision(t *testing.T) {
	full := reviewsState{Slots: make(map[string]reviewSlot, reviewsMaxPerOwner)}
	for i := 0; i < reviewsMaxPerOwner; i++ {
		full.Slots[fmt.Sprintf("call-%05d", i)] = reviewSlot{State: domain.ReviewPresent}
	}

	_, err := decideSaveReview(full, event.Version(reviewsMaxPerOwner), SaveCallReview{BasicCommand: basicCommand(t, "review.save", "1001"), Owner: "1001", CallID: "call-new", State: domain.ReviewPending})
	if err == nil || !errors.Is(err, readmodels.ErrReviewListFull) {
		t.Fatalf("a NEW call past the cap must refuse with the ErrReviewListFull contract: %v", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)

	events, err := decideSaveReview(full, event.Version(reviewsMaxPerOwner), SaveCallReview{BasicCommand: basicCommand(t, "review.save", "1001"), Owner: "1001", CallID: "call-00000", State: domain.ReviewFailed})
	if err != nil || len(events) != 1 {
		t.Fatalf("re-reviewing an existing call at the cap must stay open: %v %d", err, len(events))
	}
}

// TestReviewStateSurvivesRestart pins the journal-reload fold.
func TestReviewStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s := bootReviews(t, dir)
	if err := s.Set(ctx, owner, "call-a", domain.ReviewPresent, "draft-1", "text", "en", 50); err != nil {
		t.Fatalf("Set: %v", err)
	}

	s2 := bootReviews(t, dir)
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := s2.Get(ctx, owner, "call-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.State == domain.ReviewPresent.String() && view.DraftID == "draft-1" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("review did not survive the restart: %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
