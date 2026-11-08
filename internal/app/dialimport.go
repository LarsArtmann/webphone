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
// idempotent: every fact replaces its slot, and versions continue from
// what Load reports). Timestamp fidelity: the facts carry the ORIGINAL
// row stamps, never journal order — the imported windows answer
// remaining-cooldown math exactly as the old rows did.
package app

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

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

	imported := 0
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

	// Deterministic owner order: a crash mid-import leaves a prefix of
	// streams imported; the retry's tail sync continues each stream from
	// its own Load, so ordering only affects the log, never correctness.
	owners := make([]string, 0, len(perOwner))
	for owner := range perOwner {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	for _, owner := range owners {
		facts := perOwner[owner]
		// Same-stamp ordering inside a stream: windows before overrides,
		// then remote name — a stable, explainable order.
		sort.SliceStable(facts, func(i, j int) bool {
			ti, ri := importFactStamp(facts[i]), importFactRemote(facts[i])
			tj, rj := importFactStamp(facts[j]), importFactRemote(facts[j])
			if !ti.Equal(tj) {
				return ti.Before(tj)
			}
			return ri < rj
		})
		if err := appendDialFacts(ctx, svc.Sys, owner, facts); err != nil {
			log.Error("dial journal import: append failed; skipping rest (retries next boot)",
				"err", err, "owner", owner, "imported_so_far", imported)
			return
		}
		imported += len(facts)
	}

	if err := store.SetFlag(ctx, db.SQL(), flagDialJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("dial journal import: flag write failed (imports re-run idempotently next boot)",
			"err", err, "imported", imported)
		return
	}
	log.Info("dial windows imported into the journal",
		"imported", imported, "windows", len(wins), "overrides", len(overs))
}

// appendDialFacts raw-appends synthetic facts to the owner's stream,
// continuing from the stream's current tail (Load) — the crash-safety
// primitive: re-runs never duplicate versions and per-slot replaces
// converge the state.
func appendDialFacts(ctx context.Context, sys *system.System, owner string, facts []any) error {
	streamID, err := id.ParseStreamID(owner)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "cqrs.dial_import", "parse owner stream id %q", owner)
	}
	ref := id.NewStreamRef(cqrs.DialStreamType, streamID)
	existing, err := sys.EventStore().Load(ctx, ref)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_import", "load stream tail")
	}
	ver := event.Version(len(existing))
	actor, _ := id.ParseActorID("system:importer")

	events := make([]event.Event, 0, len(facts))
	for _, fact := range facts {
		ver++
		evt, err := event.New(event.Type(metaengine.EventTypeName(fact)), streamID, cqrs.DialStreamType, ver, fact,
			event.WithActor(actor))
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "cqrs.dial_import", "build synthetic fact")
		}
		events = append(events, evt)
	}
	if len(events) == 0 {
		return nil
	}
	if err := sys.EventStore().AppendBatch(ctx, ref, events); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.dial_import", "append synthetic facts")
	}
	return nil
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

// flagSnippetsJournalImported is the snippets importer's app_flags marker.
const flagSnippetsJournalImported = "snippets_journal_imported"

// importSnippets ports the legacy engine snippet rows into the journal
// (same policy and idempotence story as importDialWindows).
func importSnippets(ctx context.Context, db *store.Database, snippets *readmodels.Snippets, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagSnippetsJournalImported)
	if err != nil {
		log.Error("snippet journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := snippets.ExportAll(ctx)
	if err != nil {
		log.Error("snippet journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]domain.Snippet{}
	for _, v := range views {
		owner, err := domain.ParseExtension(v.Owner)
		if err != nil {
			log.Warn("snippet journal import: unparseable owner row skipped", "owner", v.Owner)
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.Snippet{
			ID:        domain.MustSnippetID(v.ID),
			Owner:     owner,
			Body:      v.Body,
			Quick:     v.Quick,
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	imported := 0
	for _, owner := range sortedOwners(perOwner) {
		streamID, err := id.ParseStreamID(owner)
		if err != nil {
			log.Error("snippet journal import: bad owner stream; skipping (retries next boot)", "err", err, "owner", owner)
			return
		}
		ref := id.NewStreamRef(cqrs.SnippetsStreamType, streamID)
		existing, err := svc.Sys.EventStore().Load(ctx, ref)
		if err != nil {
			log.Error("snippet journal import: load tail failed; skipping (retries next boot)", "err", err, "owner", owner)
			return
		}
		ver := event.Version(len(existing))
		actor, _ := id.ParseActorID("system:importer")
		events := make([]event.Event, 0, len(perOwner[owner]))
		for _, snip := range perOwner[owner] {
			ver++
			evt, err := event.New(event.Type(metaengine.EventTypeName(domain.SnippetSaved{})), streamID, cqrs.SnippetsStreamType, ver,
				domain.SnippetSaved{
					Key:       domain.SnippetKey{ID: snip.ID.String()},
					Owner:     owner,
					Body:      snip.Body,
					Quick:     snip.Quick,
					CreatedAt: snip.CreatedAt,
				}, event.WithActor(actor))
			if err != nil {
				log.Error("snippet journal import: build fact failed; skipping (retries next boot)", "err", err)
				return
			}
			events = append(events, evt)
		}
		if len(events) == 0 {
			continue
		}
		if err := svc.Sys.EventStore().AppendBatch(ctx, ref, events); err != nil {
			log.Error("snippet journal import: append failed; skipping (retries next boot)", "err", err, "owner", owner)
			return
		}
		imported += len(events)
	}
	if err := store.SetFlag(ctx, db.SQL(), flagSnippetsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("snippet journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("snippets imported into the journal", "imported", imported, "legacy_rows", len(views))
}

func sortedOwners(perOwner map[string][]domain.Snippet) []string {
	owners := make([]string, 0, len(perOwner))
	for owner := range perOwner {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}
