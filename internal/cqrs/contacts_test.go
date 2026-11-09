package cqrs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

func bootContacts(t *testing.T, dir string) *Contacts {
	t.Helper()
	return NewContacts(bootDomainSystem(t, dir, ContactsDomainDecl()))
}

var contactSeq int

// contactFor mints a contact with a fresh 21-char nanoid-slot id (the
// id's provenance is the client; the decision keeps whatever slot id
// the stream already holds on rename).
func contactFor(t *testing.T, owner, phone, name string) domain.Contact {
	t.Helper()
	contactSeq++
	return domain.Contact{
		ID:        domain.MustContactID(fmt.Sprintf("%021d", contactSeq)),
		Owner:     mustExt(t, owner),
		Name:      name,
		Phone:     domain.MustParsePhone(phone),
		CreatedAt: time.Now(),
	}
}

// TestContactSaveListRenameKeepsIdentity pins the upsert semantics now
// living in the decision: the list is newest-first, and a rename (same
// phone, new name, fresh id and stamp from the client) keeps the
// slot's ORIGINAL id and creation time.
func TestContactSaveListRenameKeepsIdentity(t *testing.T) {
	s := bootContacts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	first := contactFor(t, "1001", "+15550001", "Ada")
	if err := s.Save(ctx, first); err != nil {
		t.Fatalf("save first: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	second := contactFor(t, "1001", "+15550002", "Grace")
	if err := s.Save(ctx, second); err != nil {
		t.Fatalf("save second: %v", err)
	}

	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 contacts: %+v", list)
	}
	// Newest first (created_at descending).
	if list[0].Phone != second.Phone || list[1].Phone != first.Phone {
		t.Fatalf("list must be newest-first: %+v", list)
	}

	time.Sleep(5 * time.Millisecond)
	rename := contactFor(t, "1001", "+15550001", "Ada Lovelace")
	if err := s.Save(ctx, rename); err != nil {
		t.Fatalf("rename: %v", err)
	}
	list, err = s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after rename: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("rename must not grow the list: %+v", list)
	}
	for _, c := range list {
		if c.Phone != first.Phone {
			continue
		}
		if c.ID != first.ID {
			t.Fatalf("rename must keep the slot's id: got %s want %s", c.ID, first.ID)
		}
		if !c.CreatedAt.Truncate(time.Millisecond).Equal(first.CreatedAt.Truncate(time.Millisecond)) {
			t.Fatalf("rename must keep the slot's creation time: got %v want %v", c.CreatedAt, first.CreatedAt)
		}
		if c.Name != "Ada Lovelace" {
			t.Fatalf("rename must land the new name: %q", c.Name)
		}
		return
	}
	t.Fatalf("renamed slot missing: %+v", list)
}

// TestContactOwnerScopingKeepsListsApart pins the owner axis: a second
// owner's list never sees the first owner's rows.
func TestContactOwnerScopingKeepsListsApart(t *testing.T) {
	s := bootContacts(t, t.TempDir())
	ctx := context.Background()

	if err := s.Save(ctx, contactFor(t, "1001", "+15550001", "Ada")); err != nil {
		t.Fatalf("save: %v", err)
	}
	other, err := s.List(ctx, mustExt(t, "1002"))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("owner 1002 must not see 1001's contacts: %+v", other)
	}
}

// TestContactCapRejectsNewNumbers pins the 500-per-owner cap for NEW
// numbers (the full dispatch path, the snippets-cap precedent);
// replacing an existing number at the cap stays open.
func TestContactCapRejectsNewNumbers(t *testing.T) {
	if testing.Short() {
		t.Skip("500 dispatches")
	}
	s := bootContacts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for i := range contactsMaxPerOwner {
		if err := s.Save(ctx, contactFor(t, "1001", fmt.Sprintf("+1555%05d", i), "cap")); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}
	err := s.Save(ctx, contactFor(t, "1001", "+155599999", "one too many"))
	if err == nil || !errors.Is(err, readmodels.ErrListFull) {
		t.Fatalf("the 501st NEW number must reject with the ErrListFull contract: %v", err)
	}

	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != contactsMaxPerOwner {
		t.Fatalf("cap must hold exactly %d: %d", contactsMaxPerOwner, len(list))
	}
	existing := list[0]
	existing.Name = "replace at cap"
	if err := s.Save(ctx, existing); err != nil {
		t.Fatalf("replace at cap must pass: %v", err)
	}
}

// TestContactDeleteUnknownRejects pins the missing rejection (the
// handler's 404 mapping depends on the family).
func TestContactDeleteUnknownRejects(t *testing.T) {
	s := bootContacts(t, t.TempDir())
	ctx := context.Background()

	err := s.Delete(ctx, mustExt(t, "1001"), domain.MustContactID(fmt.Sprintf("%021d", 424242)))
	if err == nil {
		t.Fatal("deleting an unknown id must reject")
	}
}

// TestContactSaveIsImmediatelyVisible pins the D7 await: when Save (or
// Delete) returns, the projection has ALREADY folded the write — the
// Contacts tab partial re-renders from the projection in the same
// request (no polling loop here: a miss IS the failure).
func TestContactSaveIsImmediatelyVisible(t *testing.T) {
	s := bootContacts(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	ada := contactFor(t, "1001", "+15550001", "Ada")
	if err := s.Save(ctx, ada); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Ada" {
		t.Fatalf("the just-saved contact must be visible the moment Save returns: %+v", list)
	}

	if err := s.Delete(ctx, owner, ada.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, err = s.List(ctx, owner)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("the just-deleted contact must be gone the moment Delete returns: %+v", list)
	}
}

// TestContactStateSurvivesRestart pins the journal-reload fold.
func TestContactStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s := bootContacts(t, dir)
	ada := contactFor(t, "1001", "+15550001", "Ada")
	if err := s.Save(ctx, ada); err != nil {
		t.Fatalf("save: %v", err)
	}

	s2 := bootContacts(t, dir)
	deadline := time.Now().Add(5 * time.Second)
	for {
		list, err := s2.List(ctx, owner)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) == 1 && list[0].ID == ada.ID {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("contact did not survive the restart: %+v", list)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
