package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// importLegacyContacts is the ONE-TIME data continuity step of the
// contacts read-model swap (2026-10-15): deployments upgrading from the
// SQL contacts store carry their contacts in webphone.db, and the
// engine store starts empty in its own readmodels.db. On the first boot
// after the swap the legacy rows are re-saved into the engine store
// (Save's upsert-keeps-id semantics make a crash-and-retry idempotent),
// then the app_flags marker stamps completion.
//
// Failure policy: an import error LOGS and leaves the boot standing —
// the flag is only written after full success, so the next boot retries
// the whole import (Save converges on already-imported slots). A failed
// import never takes the phone system down; a completed one never runs
// again.
func importLegacyContacts(ctx context.Context, db *store.Database, engine *readmodels.Engine, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), store.FlagLegacyContactsImported)
	if err != nil {
		log.Error("legacy contacts import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}

	legacy, err := store.ReadLegacyContacts(ctx, db.SQL())
	if err != nil {
		log.Error("legacy contacts import: read failed; skipping (retries next boot)", "err", err)
		return
	}

	contacts := engine.Contacts()
	imported, skipped := 0, 0
	for _, row := range legacy {
		contact, ok := row.ToContact()
		if !ok {
			skipped++
			continue
		}
		if err := contacts.Save(ctx, contact); err != nil {
			log.Error("legacy contacts import: save failed; skipping rest (retries next boot)",
				"err", err, "imported_so_far", imported)
			return
		}
		imported++
	}

	if err := store.SetFlag(ctx, db.SQL(), store.FlagLegacyContactsImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("legacy contacts import: flag write failed (imports re-run idempotently next boot)",
			"err", err, "imported", imported)
		return
	}
	log.Info("legacy contacts imported into read models",
		"imported", imported, "skipped", skipped, "legacy_rows", len(legacy))
}

// importLegacySnippets is the ONE-TIME data continuity step of the
// snippets read-model swap (2026-10-15): same shape, policy, and
// idempotence story as importLegacyContacts above.
func importLegacySnippets(ctx context.Context, db *store.Database, engine *readmodels.Engine, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), store.FlagLegacySnippetsImported)
	if err != nil {
		log.Error("legacy snippets import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}

	legacy, err := store.ReadLegacySnippets(ctx, db.SQL())
	if err != nil {
		log.Error("legacy snippets import: read failed; skipping (retries next boot)", "err", err)
		return
	}

	snippets := engine.Snippets()
	imported, skipped := 0, 0
	for _, row := range legacy {
		snippet, ok := row.ToSnippet()
		if !ok {
			skipped++
			continue
		}
		if err := snippets.Save(ctx, snippet); err != nil {
			log.Error("legacy snippets import: save failed; skipping rest (retries next boot)",
				"err", err, "imported_so_far", imported)
			return
		}
		imported++
	}

	if err := store.SetFlag(ctx, db.SQL(), store.FlagLegacySnippetsImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("legacy snippets import: flag write failed (imports re-run idempotently next boot)",
			"err", err, "imported", imported)
		return
	}
	log.Info("legacy snippets imported into read models",
		"imported", imported, "skipped", skipped, "legacy_rows", len(legacy))
}
