// The dial-window journal import (SUPERB Phase 2, A18): the ONE-TIME
// data continuity step of the dial-gate backend swap. Deployments
// upgrading carry their cooldown slots in the readmodels engine's
// dial collections; the journal starts empty in its own system.db.
// On the first boot after the swap every legacy row is re-emitted as a
// synthetic fact into the per-owner stream (raw append, the migration
// primitive — the decider's deny path must not gate its own history),
// then the app_flags marker stamps completion.
//
// Failure policy mirrors the legacy imports: an error LOGS and leaves
// the boot standing — the flag is only written after full success, so
// the next boot retries (per-stream tail sync makes the retry
// idempotent). Timestamp fidelity: the facts carry the ORIGINAL row
// stamps, never journal order — the imported windows answer
// remaining-cooldown math exactly as the old rows did.
package app

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// flagDialJournalImported is the app_flags marker (webphone.db, per
// the D2 ruling: app_flags stay in the legacy ladder until its phase).
const flagDialJournalImported = "dial_journal_imported"

// importDialWindows ports the legacy engine rows into the journal.
func importDialWindows(ctx context.Context, db *store.Database, dials *readmodels.DialWindows, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagDialJournalImported)
	if err != nil {
		log.Error("dial journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}

	wins, overs, err := dials.ExportAll(ctx)
	if err != nil {
		log.Error("dial journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}

	perOwner := map[string][]any{}
	for _, w := range wins {
		perOwner[w.Owner] = append(perOwner[w.Owner], domain.DialAttempted{
			Key:    w.Key,
			Owner:  w.Owner,
			Remote: w.Remote,
			At:     time.Unix(w.At, 0).UTC(),
		})
	}
	for _, o := range overs {
		perOwner[o.Owner] = append(perOwner[o.Owner], domain.DialOverrideSaved{
			Key:      o.Key,
			Owner:    o.Owner,
			Remote:   o.Remote,
			Consumed: o.Consumed,
			At:       time.Unix(o.At, 0).UTC(),
		})
	}

	// Same-stamp ordering inside a stream: windows before overrides,
	// then remote name — a stable, explainable order.
	for owner, facts := range perOwner {
		slices.SortStableFunc(facts, func(a, b any) int {
			ta, tb := importFactStamp(a), importFactStamp(b)
			if !ta.Equal(tb) {
				if ta.Before(tb) {
					return -1
				}
				return 1
			}
			return strings.Compare(importFactRemote(a), importFactRemote(b))
		})
		perOwner[owner] = facts
	}

	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.DialStreamType, perOwner)
	if err != nil {
		log.Error("dial journal import: append failed; skipping rest (retries next boot)",
			"err", err, "imported_so_far", imported)
		return
	}

	if err := store.SetFlag(ctx, db.SQL(), flagDialJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("dial journal import: flag write failed (imports re-run idempotently next boot)",
			"err", err, "imported", imported)
		return
	}
	log.Info("dial windows imported into the journal",
		"imported", imported, "windows", len(wins), "overrides", len(overs))
}

func importFactStamp(fact any) time.Time {
	switch f := fact.(type) {
	case domain.DialAttempted:
		return f.At
	case domain.DialOverrideSaved:
		return f.At
	}
	return time.Time{}
}

func importFactRemote(fact any) string {
	switch f := fact.(type) {
	case domain.DialAttempted:
		return f.Remote
	case domain.DialOverrideSaved:
		return f.Remote
	}
	return ""
}
