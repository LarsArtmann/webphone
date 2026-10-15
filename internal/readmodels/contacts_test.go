package readmodels

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

func mustContact(t *testing.T, owner, name, phone string, created time.Time) domain.Contact {
	t.Helper()
	return domain.Contact{
		ID:        domain.GenerateContactID(),
		Owner:     domain.MustParseExtension(owner),
		Name:      name,
		Phone:     domain.MustParsePhone(phone),
		CreatedAt: created,
	}
}

func openTestEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	return engine
}

func TestContactsSaveListDelete(t *testing.T) {
	engine := openTestEngine(t)
	contacts := engine.Contacts()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)

	older := mustContact(t, "1001", "Older", "+15550000001", base)
	newer := mustContact(t, "1001", "Newer", "+15550000002", base.Add(time.Hour))
	if err := contacts.Save(ctx, older); err != nil {
		t.Fatalf("save older: %v", err)
	}
	if err := contacts.Save(ctx, newer); err != nil {
		t.Fatalf("save newer: %v", err)
	}

	list, err := contacts.List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list length %d, want 2", len(list))
	}
	if list[0].Name != "Newer" || list[1].Name != "Older" {
		t.Fatalf("list order: [%s, %s], want newest first", list[0].Name, list[1].Name)
	}

	// Owner scoping: another extension sees nothing.
	other, err := contacts.List(ctx, domain.MustParseExtension("1002"))
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("owner scoping broken: %d rows for 1002", len(other))
	}

	// Delete is owner-scoped by id.
	if err := contacts.Delete(ctx, domain.MustParseExtension("1001"), older.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, err = contacts.List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Newer" {
		t.Fatalf("delete left the wrong rows: %+v", list)
	}

	// Deleting an unknown id answers the not-found sentinel.
	err = contacts.Delete(ctx, domain.MustParseExtension("1001"), domain.GenerateContactID())
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown id delete: %v, want ErrNotFound", err)
	}
	// Deleting under the WRONG owner is a not-found too (scoping).
	err = contacts.Delete(ctx, domain.MustParseExtension("1002"), newer.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("cross-owner delete: %v, want ErrNotFound", err)
	}
}

// TestContactsRenameKeepsIdentity pins the upsert contract: a second
// save with the same number renames the slot, keeping its ID and
// creation time (the SQL ON CONFLICT(owner, phone) DO UPDATE SET name
// semantics).
func TestContactsRenameKeepsIdentity(t *testing.T) {
	engine := openTestEngine(t)
	contacts := engine.Contacts()
	ctx := context.Background()
	base := time.Unix(1700000000, 0)

	original := mustContact(t, "1001", "Alice", "+15550000003", base)
	if err := contacts.Save(ctx, original); err != nil {
		t.Fatalf("save: %v", err)
	}

	rename := mustContact(t, "1001", "Alicia", "+15550000003", base.Add(48*time.Hour))
	if err := contacts.Save(ctx, rename); err != nil {
		t.Fatalf("rename: %v", err)
	}

	list, err := contacts.List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("rename duplicated the slot: %d rows", len(list))
	}
	if list[0].Name != "Alicia" {
		t.Fatalf("rename did not take: %q", list[0].Name)
	}
	if list[0].ID != original.ID {
		t.Fatalf("rename replaced the id: %s, want %s", list[0].ID, original.ID)
	}
	if !list[0].CreatedAt.Equal(base) {
		t.Fatalf("rename moved creation time: %v, want %v", list[0].CreatedAt, base)
	}
}

// TestContactsCapRejectsNewNumbers pins the per-owner cap: a NEW
// number past ContactsMaxPerExtension is refused with this package's
// ErrListFull contract; renames of EXISTING numbers stay open, and a
// delete frees the slot again.
func TestContactsCapRejectsNewNumbers(t *testing.T) {
	engine := openTestEngine(t)
	contacts := engine.Contacts()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	base := time.Unix(1700000000, 0)
	var firstID domain.ContactID

	for i := range ContactsMaxPerExtension {
		contact := domain.Contact{
			ID:        domain.GenerateContactID(),
			Owner:     owner,
			Name:      "c",
			Phone:     domain.MustParsePhone("+15551" + digitsOf(i, 7)),
			CreatedAt: base.Add(time.Duration(i) * time.Second),
		}
		if i == 0 {
			firstID = contact.ID
		}
		if err := contacts.Save(ctx, contact); err != nil {
			t.Fatalf("fill to cap (%d): %v", i, err)
		}
	}

	over := mustContact(t, "1001", "Over", "+15559999999", base.Add(time.Hour))
	if err := contacts.Save(ctx, over); !errors.Is(err, ErrListFull) {
		t.Fatalf("save past the cap: want ErrListFull, got %v", err)
	}

	// Rename of an existing number stays open at the cap.
	existing := domain.Contact{
		ID:        domain.GenerateContactID(),
		Owner:     owner,
		Name:      "Renamed",
		Phone:     domain.MustParsePhone("+15551" + digitsOf(0, 7)),
		CreatedAt: base.Add(2 * time.Hour),
	}
	if err := contacts.Save(ctx, existing); err != nil {
		t.Fatalf("rename at cap: %v", err)
	}

	// Another owner still has room (the cap is per-owner).
	fresh := mustContact(t, "1002", "Fresh", "+15551000000", base)
	if err := contacts.Save(ctx, fresh); err != nil {
		t.Fatalf("other owner at their own count: %v", err)
	}

	// A delete frees the slot: the refused number fits afterwards.
	if err := contacts.Delete(ctx, owner, firstID); err != nil {
		t.Fatalf("delete at cap: %v", err)
	}
	if err := contacts.Save(ctx, over); err != nil {
		t.Fatalf("save after freeing a slot: %v", err)
	}
	list, err := contacts.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != ContactsMaxPerExtension {
		t.Fatalf("owner list = %d, want exactly the cap", len(list))
	}
}

// TestContactsSurviveRestart pins engine durability: the projected
// views live in readmodels.db, so a restart keeps every contact.
func TestContactsSurviveRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	first, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	contact := mustContact(t, "1001", "Durable", "+15550000004", time.Unix(1700000000, 0))
	if err := first.Contacts().Save(ctx, contact); err != nil {
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
	list, err := second.Contacts().List(ctx, domain.MustParseExtension("1001"))
	if err != nil {
		t.Fatalf("list after restart: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Durable" || list[0].ID != contact.ID {
		t.Fatalf("contacts lost across restart: %+v", list)
	}
}

// digitsOf renders n as a zero-padded digit string of the given width.
func digitsOf(n, width int) string {
	out := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		out[i] = byte('0' + n%10)
		n /= 10
	}
	return string(out)
}
