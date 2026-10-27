package readmodels_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func mustExtension(t *testing.T, raw string) domain.Extension {
	t.Helper()
	owner, err := domain.ParseExtension(raw)
	if err != nil {
		t.Fatal(err)
	}
	return owner
}

// TestNotesSetGetRoundTrip pins the one-slot-per-call contract: a save
// is visible via Get, a re-save REPLACES the body (no duplicate rows,
// the slot is the owner+call pair), and another extension's call id is
// a different slot entirely.
func TestNotesSetGetRoundTrip(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	notes := engine.Notes()
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	if view, err := notes.Get(ctx, owner, "call-1"); err != nil || view != nil {
		t.Fatalf("unknown call: %v %v", view, err)
	}

	if err := notes.Set(ctx, owner, "call-1", "rueckruf vereinbart"); err != nil {
		t.Fatalf("set: %v", err)
	}
	view, err := notes.Get(ctx, owner, "call-1")
	if err != nil || view == nil {
		t.Fatalf("get after set: %v %v", view, err)
	}
	if view.Body != "rueckruf vereinbart" || view.CallID != "call-1" {
		t.Fatalf("note mismatch: %+v", view)
	}

	// Replace keeps ONE slot and refreshes the body.
	if err := notes.Set(ctx, owner, "call-1", "rueckruf verschoben auf freitag"); err != nil {
		t.Fatalf("re-set: %v", err)
	}
	view, err = notes.Get(ctx, owner, "call-1")
	if err != nil || view == nil {
		t.Fatalf("get after re-set: %v %v", view, err)
	}
	if view.Body != "rueckruf verschoben auf freitag" {
		t.Fatalf("replace must overwrite the body: %+v", view)
	}

	// Owner scoping: the same call id under another extension is a
	// different (empty) slot.
	if view, err := notes.Get(ctx, mustExtension(t, "2002"), "call-1"); err != nil || view != nil {
		t.Fatalf("another owner's slot must be empty: %v %v", view, err)
	}

	// The empty body is a legal note (cleared state), not an absence.
	if err := notes.Set(ctx, owner, "call-1", ""); err != nil {
		t.Fatalf("clear: %v", err)
	}
	view, err = notes.Get(ctx, owner, "call-1")
	if err != nil || view == nil || view.Body != "" {
		t.Fatalf("empty body must stay a slot: %v %v", view, err)
	}
}

// TestNotesCapRejectsNewCalls pins the cap's honest shape: hand-written
// content is never silently expunged — a NEW call past the bound is a
// Rejection, while re-saving an EXISTING call's note stays open.
func TestNotesCapRejectsNewCalls(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	notes := engine.Notes()
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	for i := range readmodels.NotesMaxPerExtension {
		if err := notes.Set(ctx, owner, fmt.Sprintf("seed-%d", i), "note"); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	err = notes.Set(ctx, owner, "call-over-cap", "note")
	if !errors.Is(err, readmodels.ErrNoteListFull) {
		t.Fatalf("over-cap set: %v, want ErrNoteListFull", err)
	}
	// Re-saving an existing call's note stays open past the cap.
	if err := notes.Set(ctx, owner, "seed-0", "updated"); err != nil {
		t.Fatalf("re-save over cap must stay open: %v", err)
	}
}
