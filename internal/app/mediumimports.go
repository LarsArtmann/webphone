// The Phase-4 journal imports (SUPERB A29): contacts, transcript
// segments, call notes, and call reviews port from the legacy
// readmodels engine into their journal streams on the first boot after
// each swap. Policy mirrors importSnippets (flag-gated, log-and-retry,
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

// flagContactsJournalImported is the contacts importer's app_flags marker.
const flagContactsJournalImported = "contacts_journal_imported"

// importContactsIntoJournal ports the legacy engine contact rows into
// the per-owner contacts streams. Rows keep their ids and creation
// stamps (the rename-keeps-identity semantics re-derive from the
// imported facts), sorted by creation then id so the list order
// survives the upgrade byte-identically.
func importContactsIntoJournal(ctx context.Context, db *store.Database, contacts *readmodels.Contacts, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagContactsJournalImported)
	if err != nil {
		log.Error("contacts journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := contacts.ExportAll(ctx)
	if err != nil {
		log.Error("contacts journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]any{}
	for _, v := range views {
		if _, err := domain.ParseExtension(v.Owner); err != nil {
			log.Warn("contacts journal import: unparseable owner row skipped", "owner", v.Owner)
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.ContactSaved{
			Key:       domain.ContactKey{Owner: v.Owner, Phone: v.Phone},
			ID:        v.ID,
			Name:      v.Name,
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	for owner, rows := range perOwner {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.ContactSaved), b.(domain.ContactSaved)
			switch {
			case x.CreatedAt.Before(y.CreatedAt):
				return -1
			case x.CreatedAt.After(y.CreatedAt):
				return 1
			}
			return strings.Compare(x.ID, y.ID)
		})
		perOwner[owner] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.ContactsStreamType, perOwner)
	if err != nil {
		log.Error("contacts journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagContactsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("contacts journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("contacts imported into the journal", "imported", imported, "legacy_rows", len(views))
}

// flagTranscriptsJournalImported is the transcripts importer's app_flags marker.
const flagTranscriptsJournalImported = "transcripts_journal_imported"

// importTranscriptsIntoJournal ports the legacy engine transcript
// segments into the per-(owner, call) transcript streams, grouped per
// call stream and sorted by Seq (the original append order — the
// stream order then agrees with the spoken order the reassembly
// renders; never the export's arbitrary row order). The per-owner
// projection cap is not enforced during import: a replay re-grows rows
// until the next append's enforcement converges (the documented
// durability posture).
func importTranscriptsIntoJournal(ctx context.Context, db *store.Database, transcripts *readmodels.Transcripts, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagTranscriptsJournalImported)
	if err != nil {
		log.Error("transcripts journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := transcripts.ExportAll(ctx)
	if err != nil {
		log.Error("transcripts journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perCall := map[string][]any{}
	for _, v := range views {
		if _, err := domain.ParseExtension(v.Owner); err != nil {
			log.Warn("transcripts journal import: unparseable owner row skipped", "owner", v.Owner)
			continue
		}
		key := cqrs.TranscriptStreamKey(v.Owner, v.CallID)
		perCall[key] = append(perCall[key], domain.TranscriptAppended{
			Key:       domain.TranscriptSegmentKey{ID: v.ID},
			Owner:     v.Owner,
			CallID:    v.CallID,
			Direction: v.Direction,
			Remote:    v.Remote,
			StartedAt: time.Unix(v.StartedAt, 0),
			Seq:       v.Seq,
			Text:      v.Text,
		})
	}
	for key, rows := range perCall {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.TranscriptAppended), b.(domain.TranscriptAppended)
			switch {
			case x.Seq < y.Seq:
				return -1
			case x.Seq > y.Seq:
				return 1
			}
			return strings.Compare(x.Key.ID, y.Key.ID)
		})
		perCall[key] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.TranscriptsStreamType, perCall)
	if err != nil {
		log.Error("transcripts journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagTranscriptsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("transcripts journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("transcript segments imported into the journal", "imported", imported, "legacy_rows", len(views))
}

// flagCallNotesJournalImported is the call-notes importer's app_flags marker.
const flagCallNotesJournalImported = "call_notes_journal_imported"

// importCallNotesIntoJournal ports the legacy engine note slots into
// the per-owner call-notes streams. The view row IS the whole slot
// snapshot (body + disposition together), so one fact per slot
// reproduces it exactly; slots sort by update time then call id.
func importCallNotesIntoJournal(ctx context.Context, db *store.Database, notes *readmodels.Notes, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagCallNotesJournalImported)
	if err != nil {
		log.Error("call-notes journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := notes.ExportAll(ctx)
	if err != nil {
		log.Error("call-notes journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]any{}
	for _, v := range views {
		if _, err := domain.ParseExtension(v.Owner); err != nil {
			log.Warn("call-notes journal import: unparseable owner row skipped", "owner", v.Owner)
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.CallNoteSaved{
			Key:         domain.CallNoteKey{Owner: v.Owner, CallID: v.CallID},
			Owner:       v.Owner,
			CallID:      v.CallID,
			Body:        v.Body,
			Disposition: domain.Disposition(v.Disposition),
			UpdatedAt:   time.Unix(v.UpdatedAt, 0),
		})
	}
	for owner, rows := range perOwner {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.CallNoteSaved), b.(domain.CallNoteSaved)
			switch {
			case x.UpdatedAt.Before(y.UpdatedAt):
				return -1
			case x.UpdatedAt.After(y.UpdatedAt):
				return 1
			}
			return strings.Compare(x.CallID, y.CallID)
		})
		perOwner[owner] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.NotesStreamType, perOwner)
	if err != nil {
		log.Error("call-notes journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagCallNotesJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("call-notes journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("call notes imported into the journal", "imported", imported, "legacy_rows", len(views))
}

// flagCallReviewsJournalImported is the call-reviews importer's app_flags marker.
const flagCallReviewsJournalImported = "call_reviews_journal_imported"

// importCallReviewsIntoJournal ports the legacy engine review slots
// into the per-owner call-reviews streams (one wholesale snapshot fact
// per slot, sorted by update time then call id).
func importCallReviewsIntoJournal(ctx context.Context, db *store.Database, reviews *readmodels.Reviews, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), flagCallReviewsJournalImported)
	if err != nil {
		log.Error("call-reviews journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	views, err := reviews.ExportAll(ctx)
	if err != nil {
		log.Error("call-reviews journal import: export failed; skipping (retries next boot)", "err", err)
		return
	}
	perOwner := map[string][]any{}
	for _, v := range views {
		if _, err := domain.ParseExtension(v.Owner); err != nil {
			log.Warn("call-reviews journal import: unparseable owner row skipped", "owner", v.Owner)
			continue
		}
		perOwner[v.Owner] = append(perOwner[v.Owner], domain.CallReviewSaved{
			Key:        domain.CallReviewKey{Owner: v.Owner, CallID: v.CallID},
			Owner:      v.Owner,
			CallID:     v.CallID,
			State:      domain.ReviewState(v.State),
			DraftID:    v.DraftID,
			Text:       v.Text,
			Language:   v.Language,
			Confidence: v.Confidence,
			UpdatedAt:  time.Unix(v.UpdatedAt, 0),
		})
	}
	for owner, rows := range perOwner {
		slices.SortStableFunc(rows, func(a, b any) int {
			x, y := a.(domain.CallReviewSaved), b.(domain.CallReviewSaved)
			switch {
			case x.UpdatedAt.Before(y.UpdatedAt):
				return -1
			case x.UpdatedAt.After(y.UpdatedAt):
				return 1
			}
			return strings.Compare(x.CallID, y.CallID)
		})
		perOwner[owner] = rows
	}
	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.ReviewsStreamType, perOwner)
	if err != nil {
		log.Error("call-reviews journal import: append failed; skipping rest (retries next boot)", "err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), flagCallReviewsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("call-reviews journal import: flag write failed (imports re-run idempotently next boot)", "err", err, "imported", imported)
		return
	}
	log.Info("call reviews imported into the journal", "imported", imported, "legacy_rows", len(views))
}
