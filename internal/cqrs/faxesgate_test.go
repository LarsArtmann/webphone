package cqrs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// TestFaxListAllAndRefGuards pins the admin inventory and the webhook
// lookup's guards (the readmodels fax twins' surviving coverage): ListAll
// is newest-first across owners and the limit bounds it; ByProviderRef
// answers ErrNotFound for an empty ref, an unknown ref, and still
// resolves a known one.
func TestFaxListAllAndRefGuards(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	a := mustExt(t, "100")
	b := mustExt(t, "101")
	base := time.Now().Add(-time.Hour)
	newer := faxJob(t, b, "+491710", domain.GenerateFaxID(), base.Add(time.Minute))
	older := faxJob(t, a, "+491711", domain.GenerateFaxID(), base)
	for _, job := range []domain.FaxJob{older, newer} {
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	inventory, err := faxes.ListAll(ctx, 10)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(inventory) != 2 || inventory[0].ID != newer.ID || inventory[1].ID != older.ID {
		t.Fatalf("ListAll = [%s %s], want newest-first across owners", inventory[0].ID, inventory[1].ID)
	}
	if page, err := faxes.ListAll(ctx, 1); err != nil || len(page) != 1 || page[0].ID != newer.ID {
		t.Fatalf("limit must bound the inventory: (%+v, %v)", page, err)
	}
	if _, err := faxes.ByProviderRef(ctx, ""); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("empty ref must be not-found, got %v", err)
	}
	if _, err := faxes.ByProviderRef(ctx, "no-such-ref"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown ref must be not-found, got %v", err)
	}
	if _, err := faxes.GetAny(ctx, domain.GenerateFaxID()); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown id must be not-found, got %v", err)
	}
}

// TestFaxMovesLegalityMirror drives every row of the faxMoves table
// through the seam (the decider consumes the machine table verbatim —
// the ONE legality home — and the projection stores what it sends) and
// pins the resulting statuses.
func TestFaxMovesLegalityMirror(t *testing.T) {
	_, faxes := faxBattery(t)
	ctx := context.Background()
	owner := mustExt(t, "100")

	moves := []struct {
		from domain.FaxStatus
		to   domain.FaxStatus
	}{
		{domain.FaxQueued, domain.FaxSending},          // accepted
		{domain.FaxQueued, domain.FaxFailed},           // submitFailed
		{domain.FaxSending, domain.FaxTransmitted},     // settled
		{domain.FaxQueued, domain.FaxTransmitted},      // verdict
		{domain.FaxSending, domain.FaxTransmitted},     // verdict
		{domain.FaxFailed, domain.FaxTransmitted},      // verdict (late reversal)
		{domain.FaxTransmitted, domain.FaxTransmitted}, // same-status refresh
		{domain.FaxQueued, domain.FaxFailed},           // verdict
		{domain.FaxSending, domain.FaxFailed},          // verdict
		{domain.FaxTransmitted, domain.FaxFailed},      // verdict (late reversal)
		{domain.FaxFailed, domain.FaxFailed},           // same-status refresh
	}
	for i, m := range moves {
		job := domain.FaxJob{
			ID:        domain.GenerateFaxID(),
			Owner:     owner,
			Remote:    domain.MustParsePhone("+491712"),
			Direction: domain.FaxOutbound,
			Status:    m.from,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			UpdatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := faxes.Create(ctx, job); err != nil {
			t.Fatalf("move %d create at %s: %v", i, m.from, err)
		}
		if err := faxes.UpdateStatus(ctx, job.ID, domain.FaxDelivery{Status: m.to}); err != nil {
			t.Fatalf("move %d (%s -> %s): %v", i, m.from, m.to, err)
		}
		got, err := faxes.Get(ctx, owner, job.ID)
		if err != nil {
			t.Fatalf("move %d get: %v", i, err)
		}
		if got.Status != m.to {
			t.Fatalf("move %d: status %s, want %s", i, got.Status, m.to)
		}
	}
}
