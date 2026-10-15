package readmodels

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

func mustSnippet(t *testing.T, owner, body string, quick bool, created time.Time) domain.Snippet {
	t.Helper()
	return domain.Snippet{
		ID:        domain.GenerateSnippetID(),
		Owner:     domain.MustParseExtension(owner),
		Body:      body,
		Quick:     quick,
		CreatedAt: created,
	}
}

func TestSnippetsSaveListDelete(t *testing.T) {
	engine := openTestEngine(t)
	snippets := engine.Snippets()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)
	owner := domain.MustParseExtension("1001")

	if list, err := snippets.List(ctx, owner); err != nil || len(list) != 0 {
		t.Fatalf("fresh list: %v %+v", err, list)
	}

	first := mustSnippet(t, "1001", "Thanks, on it!", true, base)
	second := mustSnippet(t, "1001", "Please call back later", false, base.Add(time.Second))
	if err := snippets.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}
	if err := snippets.Save(ctx, second); err != nil {
		t.Fatalf("save second: %v", err)
	}

	list, err := snippets.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].ID != first.ID || !list[0].Quick || list[1].Quick {
		t.Fatalf("list order/quick wrong: %+v", list)
	}

	// Owner scoping: another extension sees nothing.
	other, err := snippets.List(ctx, domain.MustParseExtension("1002"))
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("owner scoping broken: %d rows for 1002", len(other))
	}

	// Delete is owner-scoped; a miss is ErrNotFound.
	if err := snippets.Delete(ctx, domain.MustParseExtension("1002"), second.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign delete: got %v, want ErrNotFound", err)
	}
	if err := snippets.Delete(ctx, owner, second.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := snippets.Delete(ctx, owner, second.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("double delete: got %v, want ErrNotFound", err)
	}
	list, err = snippets.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(list) != 1 || list[0].ID != first.ID {
		t.Fatalf("delete left the wrong rows: %+v", list)
	}
}

// TestSnippetsReplaceKeepsCreationTime pins the upsert contract: a
// second save with the same id edits in place (body + quick flip)
// without consuming cap or changing order, and keeps the slot's
// creation time (the SQL ON CONFLICT(id) DO UPDATE SET body, quick
// semantics).
func TestSnippetsReplaceKeepsCreationTime(t *testing.T) {
	engine := openTestEngine(t)
	snippets := engine.Snippets()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)

	original := mustSnippet(t, "1001", "Thanks, on it!", true, base)
	if err := snippets.Save(ctx, original); err != nil {
		t.Fatalf("save: %v", err)
	}

	edited := original
	edited.Body = "Thanks — updated"
	edited.Quick = false
	edited.CreatedAt = base.Add(48 * time.Hour)
	if err := snippets.Save(ctx, edited); err != nil {
		t.Fatalf("replace: %v", err)
	}

	list, err := snippets.List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("replace duplicated the slot: %d rows", len(list))
	}
	if list[0].Body != "Thanks — updated" || list[0].Quick {
		t.Fatalf("replace did not take: %+v", list[0])
	}
	if list[0].ID != original.ID {
		t.Fatalf("replace minted a new id: %s, want %s", list[0].ID, original.ID)
	}
	if !list[0].CreatedAt.Equal(base) {
		t.Fatalf("replace moved creation time: %v, want %v", list[0].CreatedAt, base)
	}
}

// TestSnippetsCapRejectsNewIds pins the per-owner cap: a NEW id past
// SnippetsMaxPerExtension is refused with this package's
// ErrSnippetListFull contract; replacing an existing id stays open at
// the cap, another owner still has room, and a delete frees the slot.
func TestSnippetsCapRejectsNewIds(t *testing.T) {
	engine := openTestEngine(t)
	snippets := engine.Snippets()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	base := time.Unix(1700000000, 0)
	var firstID domain.SnippetID

	for i := range SnippetsMaxPerExtension {
		snippet := domain.Snippet{
			ID:        domain.GenerateSnippetID(),
			Owner:     owner,
			Body:      "filler",
			Quick:     i%2 == 0,
			CreatedAt: base.Add(time.Duration(i) * time.Second),
		}
		if i == 0 {
			firstID = snippet.ID
		}
		if err := snippets.Save(ctx, snippet); err != nil {
			t.Fatalf("fill to cap (%d): %v", i, err)
		}
	}

	overflow := mustSnippet(t, "1001", "one too many", false, base.Add(time.Hour))
	if err := snippets.Save(ctx, overflow); !errors.Is(err, ErrSnippetListFull) {
		t.Fatalf("save past the cap: want ErrSnippetListFull, got %v", err)
	}

	// Replace of an existing id stays open at the cap.
	existing := mustSnippet(t, "1001", "filler", false, base)
	existing.ID = firstID
	existing.Body = "edited at cap"
	if err := snippets.Save(ctx, existing); err != nil {
		t.Fatalf("replace at cap: %v", err)
	}

	// Another owner still has room (the cap is per-owner).
	fresh := mustSnippet(t, "1002", "fresh", true, base)
	if err := snippets.Save(ctx, fresh); err != nil {
		t.Fatalf("other owner at their own count: %v", err)
	}

	// A delete frees the slot: the refused snippet fits afterwards.
	if err := snippets.Delete(ctx, owner, firstID); err != nil {
		t.Fatalf("delete at cap: %v", err)
	}
	if err := snippets.Save(ctx, overflow); err != nil {
		t.Fatalf("save after freeing a slot: %v", err)
	}
	list, err := snippets.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != SnippetsMaxPerExtension {
		t.Fatalf("owner list = %d, want exactly the cap", len(list))
	}
}

// TestSnippetsSurviveRestart pins engine durability: the projected
// views live in readmodels.db, so a restart keeps every snippet.
func TestSnippetsSurviveRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	first, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	snippet := mustSnippet(t, "1001", "Durable", true, time.Unix(1700000000, 0))
	if err := first.Snippets().Save(ctx, snippet); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := first.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	second, err := Open(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = second.Shutdown() })
	list, err := second.Snippets().List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list after restart: %v", err)
	}
	if len(list) != 1 || list[0].Body != "Durable" || list[0].ID != snippet.ID || !list[0].Quick {
		t.Fatalf("snippets lost across restart: %+v", list)
	}
}
