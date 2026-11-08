// The fax journal import (SUPERB Phase 6, A40): readmodels.db fax rows
// port into their per-job journal streams on the first boot after the
// fax swap. Policy mirrors the family (flag-gated, log-and-retry,
// full-success-then-flag); each job ports as ONE FaxReceipted fact
// carrying its CURRENT row image — the row after every move — so the
// fold rebuilds the settled state without replaying the lifecycle.
// This file exists to be deleted when the transitional window closes.
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// flagFaxesJournalImported is the faxes importer's app_flags marker.
const flagFaxesJournalImported = "faxes_journal_imported"

// importFaxesIntoJournal ports the legacy engine fax rows into the
// journal. Retry safety: appendJournalFacts tail-syncs each stream's
// versions and the FaxReceipted fold is a keyed upsert of the full row
// image, so a crash mid-import leaves a deterministic prefix whose
// re-append converges on the same rows.
func importFaxesIntoJournal(ctx context.Context, db *store.Database, faxes *readmodels.Faxes, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagFaxesJournalImported)
	if err != nil {
		log.Error("fax journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	jobs, err := faxes.ListAll(ctx, 0)
	if err != nil {
		log.Error("fax journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perStream := map[string][]any{}
	count := 0
	for _, job := range jobs {
		if !importableOwner(log, "fax journal import", job.Owner.String()) {
			continue
		}
		perStream[cqrs.FaxStreamKey(job.ID)] = append(perStream[cqrs.FaxStreamKey(job.ID)], domain.FaxReceipted{
			Key:          domain.FaxKey{ID: job.ID.String()},
			Owner:        job.Owner.String(),
			Remote:       job.Remote.String(),
			Direction:    job.Direction,
			Status:       job.Status,
			Pages:        job.Pages,
			DocumentPath: job.DocumentPath,
			ProviderRef:  job.ProviderRef,
			Error:        job.Error,
			CreatedAt:    job.CreatedAt,
			UpdatedAt:    job.UpdatedAt,
		})
		count++
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.FaxesStreamType, perStream)
	if err != nil {
		log.Error("fax journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagFaxesJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("fax journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("fax journal import complete", "faxes", count, "events", imported)
}
