package readmodels_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// TestReviewsLifecycleRoundTrip pins the review-slot contract (issue
// #8): one slot per (owner, call), the state machine walks pending →
// present (and failed on outage), every save REPLACES the row (a retry
// overwrites a stale failed state), and the present-state dims ride
// only the present save.
func TestReviewsLifecycleRoundTrip(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	reviews := engine.Reviews()
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	if view, err := reviews.Get(ctx, owner, "call-1"); err != nil || view != nil {
		t.Fatalf("unknown call: %v %v", view, err)
	}

	if err := reviews.Set(ctx, owner, "call-1", domain.ReviewPending, "", "", "", 0); err != nil {
		t.Fatalf("set pending: %v", err)
	}
	view, err := reviews.Get(ctx, owner, "call-1")
	if err != nil || view == nil || view.State != "pending" {
		t.Fatalf("pending slot: %v %v %+v", view, err, view)
	}

	if err := reviews.Set(ctx, owner, "call-1", domain.ReviewPresent, "d-7", "Guten Morgen", "de", 90); err != nil {
		t.Fatalf("set present: %v", err)
	}
	view, err = reviews.Get(ctx, owner, "call-1")
	if err != nil || view == nil {
		t.Fatalf("get after present: %v %v", view, err)
	}
	if view.State != "present" || view.DraftID != "d-7" || view.Text != "Guten Morgen" || view.Language != "de" || view.Confidence != 90 {
		t.Fatalf("present dims must ride the slot: %+v", view)
	}

	// Replace semantics: a later outage REPLACES the present row — the
	// slot always carries the LATEST assessment.
	if err := reviews.Set(ctx, owner, "call-1", domain.ReviewFailed, "", "", "", 0); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	view, err = reviews.Get(ctx, owner, "call-1")
	if err != nil || view == nil || view.State != "failed" || view.DraftID != "" || view.Text != "" {
		t.Fatalf("failed save must replace wholesale: %v %v %+v", view, err, view)
	}

	// Owner scoping: the same call id under another extension is a
	// different (empty) slot.
	if view, err := reviews.Get(ctx, mustExtension(t, "2002"), "call-1"); err != nil || view != nil {
		t.Fatalf("another owner's slot must be empty: %v %v", view, err)
	}
}

// TestReviewsCapRejectsNewCalls pins the cap posture: a NEW call past
// the bound is a Rejection, re-reviewing an EXISTING call stays open.
func TestReviewsCapRejectsNewCalls(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	reviews := engine.Reviews()
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	for i := range readmodels.ReviewsMaxPerExtension {
		if err := reviews.Set(ctx, owner, fmt.Sprintf("seed-%d", i), domain.ReviewPresent, "d", "t", "de", 50); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	err = reviews.Set(ctx, owner, "call-over-cap", domain.ReviewPending, "", "", "", 0)
	if !errors.Is(err, readmodels.ErrReviewListFull) {
		t.Fatalf("over-cap set: %v, want ErrReviewListFull", err)
	}
	if err := reviews.Set(ctx, owner, "seed-0", domain.ReviewFailed, "", "", "", 0); err != nil {
		t.Fatalf("re-review over cap must stay open: %v", err)
	}
}
