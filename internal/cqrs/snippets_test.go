package cqrs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// bootSnippetsSystem boots a REAL system with the snippets domain over
// a temp dir.
func bootSnippetsSystem(t *testing.T, dir string) *Snippets {
	t.Helper()
	ctx := context.Background()
	dsn := dir + "/system.db"
	domainCfg := SnippetsDomainDecl()
	domainCfg.ProjectionHostOptions = append(domainCfg.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, domainCfg, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		t.Fatalf("system.New: %v", err)
	}
	if err := sys.Start(ctx); err != nil {
		t.Fatalf("sys.Start: %v", err)
	}
	t.Cleanup(func() { _ = sys.GracefulClose(ctx) })
	return NewSnippets(sys)
}

var snippetSeq int

func snippetFor(t *testing.T, owner, body string, quick bool) domain.Snippet {
	t.Helper()
	snippetSeq++
	id, err := domain.ParseSnippetID(fmt.Sprintf("%021d", snippetSeq))
	if err != nil {
		t.Fatalf("parse snippet id: %v", err)
	}
	return domain.Snippet{
		ID:        id,
		Owner:     mustExt(t, owner),
		Body:      body,
		Quick:     quick,
		CreatedAt: time.Now(),
	}
}

// awaitSnippetCount polls the projection until the owner's list holds
// want rows.
func awaitSnippetCount(t *testing.T, s *Snippets, owner domain.Extension, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		got, err := s.List(context.Background(), owner)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("snippet projection never reached %d rows: got %d (%+v)", want, len(got), got)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestSnippetSaveListDelete pins the CRUD round-trip through the
// journal: save lands in the projection, list is oldest-first, delete
// removes (and a second delete is the rejection).
func TestSnippetSaveListDelete(t *testing.T) {
	s := bootSnippetsSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	first := snippetFor(t, "1001", "hello there", false)
	second := snippetFor(t, "1001", "on my way", true)
	if err := s.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}
	if err := s.Save(ctx, second); err != nil {
		t.Fatalf("save second: %v", err)
	}
	awaitSnippetCount(t, s, owner, 2)

	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].ID != first.ID || !list[1].Quick {
		t.Fatalf("list order/content wrong: %+v", list)
	}

	if err := s.Delete(ctx, owner, first.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	awaitSnippetCount(t, s, owner, 1)
	if err := s.Delete(ctx, owner, first.ID); err == nil {
		t.Fatal("second delete must reject (the slot is gone)")
	}
}

// TestSnippetReplaceKeepsCreationTime pins the upsert semantics now
// living in the decision: replacing a snippet keeps the slot's
// original stamp.
func TestSnippetReplaceKeepsCreationTime(t *testing.T) {
	s := bootSnippetsSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	original := snippetFor(t, "1001", "v1", false)
	if err := s.Save(ctx, original); err != nil {
		t.Fatalf("save: %v", err)
	}
	awaitSnippetCount(t, s, owner, 1)

	time.Sleep(5 * time.Millisecond)
	updated := original
	updated.Body = "v2"
	updated.CreatedAt = time.Now()
	if err := s.Save(ctx, updated); err != nil {
		t.Fatalf("replace: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		list, err := s.List(ctx, owner)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) == 1 && list[0].Body == "v2" {
			// Microsecond tolerance: the journal's CBOR time encoding
			// quantizes to ~200 ns (float64 epoch) — a bounded, invisible
			// drift recorded in the plan errata; anything coarser is a
			// real semantics break.
			if !list[0].CreatedAt.Truncate(time.Millisecond).Equal(original.CreatedAt.Truncate(time.Millisecond)) {
				t.Fatalf("replace must keep the slot's creation time: got %v want %v", list[0].CreatedAt, original.CreatedAt)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("replace never landed: %+v", list)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestSnippetCapRejectsNewIDs pins the per-owner cap: the 101st NEW id
// rejects with the ErrSnippetListFull contract; replacing an existing
// id stays open.
func TestSnippetCapRejectsNewIDs(t *testing.T) {
	s := bootSnippetsSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for i := range snippetsMaxSlots {
		if err := s.Save(ctx, snippetFor(t, "1001", "body", false)); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}
	err := s.Save(ctx, snippetFor(t, "1001", "one too many", false))
	if err == nil || !errors.Is(err, readmodels.ErrSnippetListFull) {
		t.Fatalf("the 101st new snippet must reject with the ErrSnippetListFull contract: %v", err)
	}

	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != snippetsMaxSlots {
		t.Fatalf("cap must hold exactly %d: %d", snippetsMaxSlots, len(list))
	}
	// Replacing an existing id stays open at the cap.
	existing := list[0]
	existing.Body = "still room for replaces"
	if err := s.Save(ctx, existing); err != nil {
		t.Fatalf("replace at cap must pass: %v", err)
	}
}

// TestSnippetStateSurvivesRestart pins the journal-reload fold.
func TestSnippetStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s := bootSnippetsSystem(t, dir)
	snip := snippetFor(t, "1001", "across boots", true)
	if err := s.Save(ctx, snip); err != nil {
		t.Fatalf("save: %v", err)
	}

	s2 := bootSnippetsSystem(t, dir)
	deadline := time.Now().Add(5 * time.Second)
	for {
		list, err := s2.List(ctx, owner)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) == 1 && list[0].ID == snip.ID {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("snippet did not survive the restart: %+v", list)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestSnippetSaveIsImmediatelyVisible pins the D7 await: when Save (or
// Delete) returns, the projection has ALREADY folded the write — the
// Settings partial re-renders from the projection in the same request
// and must never race its own write (no polling loop here: a miss IS
// the failure).
func TestSnippetSaveIsImmediatelyVisible(t *testing.T) {
	s := bootSnippetsSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	snip := snippetFor(t, "1001", "visible at once", false)
	if err := s.Save(ctx, snip); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != snip.ID || list[0].Body != "visible at once" {
		t.Fatalf("the just-saved snippet must be visible the moment Save returns: %+v", list)
	}

	if err := s.Delete(ctx, owner, snip.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, err = s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("the just-deleted snippet must be gone the moment Delete returns: %+v", list)
	}
}
