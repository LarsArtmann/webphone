package cqrs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-cqrs-lite/event/v4"

	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func bootNotes(t *testing.T, dir string) *Notes {
	t.Helper()
	return NewNotes(bootDomainSystem(t, dir, NotesDomainDecl()))
}

func awaitNote(t *testing.T, s *Notes, owner domain.Extension, callID string) *readmodels.CallNoteView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := s.Get(context.Background(), owner, callID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil {
			return view
		}
		if time.Now().After(deadline) {
			t.Fatalf("note for %s never landed", callID)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestNoteMergeSemantics pins the slot-merge behavior now living in the
// decision: a note save keeps the verdict, a verdict save keeps the
// prose, and an absent slot reads nil.
func TestNoteMergeSemantics(t *testing.T) {
	s := bootNotes(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if view, err := s.Get(ctx, owner, "call-a"); err != nil || view != nil {
		t.Fatalf("absent note must read nil: %v %+v", err, view)
	}

	if err := s.Set(ctx, owner, "call-a", "discussed the invoice"); err != nil {
		t.Fatalf("Set note: %v", err)
	}
	view := awaitNote(t, s, owner, "call-a")
	if view.Body != "discussed the invoice" || view.Disposition != "" {
		t.Fatalf("note save must land body with empty disposition: %+v", view)
	}

	if err := s.SetDisposition(ctx, owner, "call-a", domain.DispositionAnswered); err != nil {
		t.Fatalf("SetDisposition: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := s.Get(ctx, owner, "call-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.Disposition == domain.DispositionAnswered.String() {
			if view.Body != "discussed the invoice" {
				t.Fatalf("verdict save must keep the prose: %+v", view)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("disposition never landed: %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err := s.Set(ctx, owner, "call-a", "v2"); err != nil {
		t.Fatalf("Set note v2: %v", err)
	}
	deadline = time.Now().Add(5 * time.Second)
	for {
		view, err := s.Get(ctx, owner, "call-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.Body == "v2" {
			if view.Disposition != domain.DispositionAnswered.String() {
				t.Fatalf("note save must keep the verdict: %+v", view)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("note v2 never landed: %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestNoteOwnerScopingKeepsSlotsApart pins the owner axis.
func TestNoteOwnerScopingKeepsSlotsApart(t *testing.T) {
	s := bootNotes(t, t.TempDir())
	ctx := context.Background()

	if err := s.Set(ctx, mustExt(t, "1001"), "call-a", "private"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if view, err := s.Get(ctx, mustExt(t, "1002"), "call-a"); err != nil || view != nil {
		t.Fatalf("owner 1002 must not read 1001's note: %v %+v", err, view)
	}
}

// TestNoteCapIsConflictAtTheDecision pins the 2000-cap boundary at the
// DECISION (2000 dispatches through the backend would be too slow to
// be the boundary pin): a NEW call past the cap refuses with the
// ErrNoteListFull CONFLICT contract, an existing call at the cap stays
// open, and both write doors share the cap.
func TestNoteCapIsConflictAtTheDecision(t *testing.T) {
	full := notesState{Slots: make(map[string]noteSlot, notesMaxPerOwner)}
	for i := range notesMaxPerOwner {
		full.Slots[fmt.Sprintf("call-%05d", i)] = noteSlot{Body: "note"}
	}

	_, err := decideSaveNote(full, event.Version(notesMaxPerOwner), SaveCallNote{BasicCommand: basicCommand(t, "note.save", "1001"), Owner: "1001", CallID: "call-new", Body: "late"})
	if err == nil || !errors.Is(err, readmodels.ErrNoteListFull) {
		t.Fatalf("a NEW call past the cap must refuse with the ErrNoteListFull contract: %v", err)
	}
	errorfamilytest.AssertFamily(t, err, errorfamily.Conflict)

	_, err = decideSaveDisposition(full, event.Version(notesMaxPerOwner), SaveCallDisposition{BasicCommand: basicCommand(t, "note.disposition_save", "1001"), Owner: "1001", CallID: "call-new", Disposition: domain.DispositionAnswered})
	if err == nil || !errors.Is(err, readmodels.ErrNoteListFull) {
		t.Fatalf("the disposition door shares the cap: %v", err)
	}

	events, err := decideSaveNote(full, event.Version(notesMaxPerOwner), SaveCallNote{BasicCommand: basicCommand(t, "note.save", "1001"), Owner: "1001", CallID: "call-00000", Body: "still open"})
	if err != nil || len(events) != 1 {
		t.Fatalf("an existing call at the cap must stay open: %v %d", err, len(events))
	}
}

// TestNoteStateSurvivesRestart pins the journal-reload fold.
func TestNoteStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s := bootNotes(t, dir)
	if err := s.Set(ctx, owner, "call-a", "across boots"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := s.SetDisposition(ctx, owner, "call-a", domain.DispositionAnswered); err != nil {
		t.Fatalf("SetDisposition: %v", err)
	}

	s2 := bootNotes(t, dir)
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := s2.Get(ctx, owner, "call-a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.Body == "across boots" && view.Disposition == domain.DispositionAnswered.String() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("note did not survive the restart: %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestNoteVerdictFirstAndClearedBody pins the slot-shape pins the
// engine store's suite carried: a verdict on a NOTELESS call creates
// the slot with an empty body, and an empty-body save CLEARS the note
// while the slot stays (a cleared note is a state, not an absence).
func TestNoteVerdictFirstAndClearedBody(t *testing.T) {
	s := bootNotes(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.SetDisposition(ctx, owner, "call-9", domain.DispositionVoicemail); err != nil {
		t.Fatalf("set disposition on noteless call: %v", err)
	}
	view := awaitNote(t, s, owner, "call-9")
	if view.Disposition != domain.DispositionVoicemail.String() || view.Body != "" {
		t.Fatalf("verdict-first slot: %+v", view)
	}
	var err error

	if err := s.Set(ctx, owner, "call-9", "rueckruf vereinbart"); err != nil {
		t.Fatalf("set note: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err = s.Get(ctx, owner, "call-9")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.Body == "rueckruf vereinbart" {
			if view.Disposition != domain.DispositionVoicemail.String() {
				t.Fatalf("note save must carry the verdict forward: %+v", view)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("note never landed: %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err := s.Set(ctx, owner, "call-9", ""); err != nil {
		t.Fatalf("clear: %v", err)
	}
	deadline = time.Now().Add(5 * time.Second)
	for {
		view, err = s.Get(ctx, owner, "call-9")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if view != nil && view.Body == "" {
			if view.Disposition != domain.DispositionVoicemail.String() {
				t.Fatalf("cleared body must keep the verdict: %+v", view)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("cleared body never landed (slot must stay): %+v", view)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
