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

// TestDispositionSharesTheNoteSlot pins the operator-verdict seam
// (issue #6): the verdict lives on the SAME (owner, call) slot as the
// note, each save carries the other field forward (editing one never
// erases the other), another owner's call is a different slot, and the
// cap applies to NEW slots either way.
func TestDispositionSharesTheNoteSlot(t *testing.T) {
	engine, err := readmodels.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	notes := engine.Notes()
	ctx := context.Background()
	owner := mustExtension(t, "1001")

	// Verdict on a call with no note: the slot exists with an empty body.
	if err := notes.SetDisposition(ctx, owner, "call-9", domain.DispositionVoicemail); err != nil {
		t.Fatalf("set disposition on noteless call: %v", err)
	}
	view, err := notes.Get(ctx, owner, "call-9")
	if err != nil || view == nil {
		t.Fatalf("get after disposition: %v %v", view, err)
	}
	if view.Disposition != "voicemail" || view.Body != "" {
		t.Fatalf("disposition slot: %+v", view)
	}

	// A later note save must NOT erase the verdict.
	if err := notes.Set(ctx, owner, "call-9", "rueckruf vereinbart"); err != nil {
		t.Fatalf("set note: %v", err)
	}
	view, err = notes.Get(ctx, owner, "call-9")
	if err != nil || view == nil {
		t.Fatalf("get after note: %v %v", view, err)
	}
	if view.Body != "rueckruf vereinbart" || view.Disposition != "voicemail" {
		t.Fatalf("note save must carry the verdict forward: %+v", view)
	}

	// A later verdict must NOT erase the note (and a correction wins).
	if err := notes.SetDisposition(ctx, owner, "call-9", domain.DispositionAnswered); err != nil {
		t.Fatalf("correction: %v", err)
	}
	view, err = notes.Get(ctx, owner, "call-9")
	if err != nil || view == nil {
		t.Fatalf("get after correction: %v %v", view, err)
	}
	if view.Body != "rueckruf vereinbart" || view.Disposition != "answered" {
		t.Fatalf("correction must carry the note forward: %+v", view)
	}

	// Owner scoping: another extension's slot stays empty.
	if view, err := notes.Get(ctx, mustExtension(t, "2002"), "call-9"); err != nil || view != nil {
		t.Fatalf("another owner's verdict slot must be empty: %v %v", view, err)
	}

	// The cap counts NEW slots either way: a brand-new call over the
	// cap refuses the verdict too, while an existing slot stays open.
	for i := range readmodels.NotesMaxPerExtension - 1 {
		if err := notes.Set(ctx, owner, fmt.Sprintf("dseed-%d", i), "note"); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	if err := notes.SetDisposition(ctx, owner, "call-over-cap", domain.DispositionMissed); !errors.Is(err, readmodels.ErrNoteListFull) {
		t.Fatalf("over-cap disposition: %v, want ErrNoteListFull", err)
	}
	if err := notes.SetDisposition(ctx, owner, "call-9", domain.DispositionMissed); err != nil {
		t.Fatalf("existing-slot disposition over cap must stay open: %v", err)
	}
}
