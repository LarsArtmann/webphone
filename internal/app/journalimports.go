// The Phase-3 journal imports (SUPERB A21–A24): snippets, agent-audit
// rows, and push devices port from the legacy readmodels engine into
// their per-owner journal streams on the first boot after each swap.
// Policy mirrors importDialWindows (flag-gated, log-and-retry,
// full-success-then-flag); the append mechanics ride the ONE shared
// primitive (importutil.go).
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

// flagSnippetsJournalImported is the snippets importer's app_flags marker.
const flagSnippetsJournalImported = "snippets_journal_imported"

// importSnippets ports the legacy engine snippet rows into the journal.
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
	perOwner := map[string][]any{}
	for _, v := range views {
		if !importableOwner(log, "snippet journal import", v.Owner) {
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.SnippetSaved{
			Key:       domain.SnippetKey{ID: v.ID},
			Owner:     v.Owner,
			Body:      v.Body,
			Quick:     v.Quick,
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.SnippetsStreamType, perOwner)
	if err != nil {
		log.Error("snippet journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagSnippetsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("snippet journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("snippets imported into the journal", "imported", imported, "legacy_rows", len(views))
}

// flagAgentAuditJournalImported is the agent-audit importer's app_flags marker.
const flagAgentAuditJournalImported = "agent_audit_journal_imported"

// importAgentAudit ports the legacy engine audit rows into the journal
// (rows keep their ids, actors and seq stamps so Settings' newest-20
// window survives the upgrade). Rows import in seq order per owner —
// the stream order then agrees with the happening order the window
// sorts by.
func importAgentAudit(ctx context.Context, db *store.Database, audit *readmodels.AgentAudit, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagAgentAuditJournalImported)
	if err != nil {
		log.Error("agent-audit journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := audit.ExportAll(ctx)
	if err != nil {
		log.Error("agent-audit journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]any{}
	for _, v := range views {
		if !importableOwner(log, "agent-audit journal import", v.Owner) {
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.AgentAuditAppended{
			Key:     domain.AgentAuditKey{ID: v.ID},
			Owner:   v.Owner,
			Actor:   v.Actor,
			Event:   v.Event,
			Subject: v.Subject,
			Outcome: v.Outcome,
			At:      time.Unix(v.At, 0).UTC(),
			Seq:     v.Seq,
		})
	}
	for owner, rows := range perOwner {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.AgentAuditAppended), b.(domain.AgentAuditAppended)
			switch {
			case x.Seq < y.Seq:
				return -1
			case x.Seq > y.Seq:
				return 1
			}
			return strings.Compare(x.Key.ID, y.Key.ID)
		})
		perOwner[owner] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.AgentAuditStreamType, perOwner)
	if err != nil {
		log.Error("agent-audit journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagAgentAuditJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("agent-audit journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("agent-audit rows imported into the journal", "imported", imported, "legacy_rows", len(views))
}

// flagPushDevicesJournalImported is the push-devices importer's app_flags marker.
const flagPushDevicesJournalImported = "push_devices_journal_imported"

// importPushDevices ports the legacy engine device rows into the
// journal. Ring continuity is the whole point: without the port, every
// installed app silently stops ringing after the upgrade. Rows keep
// their tokens (the wake addresses) and registration stamps.
func importPushDevices(ctx context.Context, db *store.Database, devices *readmodels.PushDevices, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagPushDevicesJournalImported)
	if err != nil {
		log.Error("push-device journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := devices.ExportAll(ctx)
	if err != nil {
		log.Error("push-device journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]any{}
	for _, v := range views {
		if !importableOwner(log, "push-device journal import", v.Owner) {
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.PushDeviceRegistered{
			Key:       domain.PushDeviceKey{Owner: v.Owner, DeviceID: v.DeviceID},
			Transport: v.Transport,
			Token:     v.Token,
			UpdatedAt: time.Unix(0, v.UpdatedAt),
		})
	}
	for owner, rows := range perOwner {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.PushDeviceRegistered), b.(domain.PushDeviceRegistered)
			switch {
			case x.UpdatedAt.Before(y.UpdatedAt):
				return -1
			case x.UpdatedAt.After(y.UpdatedAt):
				return 1
			}
			return strings.Compare(x.Key.DeviceID, y.Key.DeviceID)
		})
		perOwner[owner] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.PushDevicesStreamType, perOwner)
	if err != nil {
		log.Error("push-device journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagPushDevicesJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("push-device journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("push devices imported into the journal", "imported", imported, "legacy_rows", len(views))
}
