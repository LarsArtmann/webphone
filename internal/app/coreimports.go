// The core journal import (SUPERB A35): webphone.db threads, messages
// and attachments port into their per-thread journal streams on the
// first boot after the core swap. Policy mirrors the family (flag-gated,
// log-and-retry, full-success-then-flag); the append mechanics ride the
// ONE shared primitive (importutil.go). This file exists to be deleted
// when the transitional window closes.
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// importThreadsIntoJournal ports the legacy thread/message/attachment
// rows into the journal. Two fidelity decisions worth naming:
//
//   - Every MessageAppended fact carries the FINAL legacy summary image,
//     not a stepwise-derived one. Read receipts and the organization
//     flags are legacy ROW state — they are not derivable from the
//     message sequence — so the legacy row is the only honest source.
//     The threads-collection folds adopt images wholesale and the LAST
//     write wins, so the final projection row IS the legacy row; a
//     replay rebuilds the same state from the same images.
//
//   - Messages append in the legacy list's own order (created_at, rowid)
//     and keep their original stamps and ids, so journal sequence and
//     transcript order agree with what the legacy SQL served.
//
// Retry safety: appendJournalFacts tail-syncs each stream's versions and
// the folds are keyed upserts, so a crash mid-import leaves a
// deterministic prefix whose re-append converges on the same rows.
func importThreadsIntoJournal(ctx context.Context, db *store.Database, svc *CqrsService, log *slog.Logger) {
	done, err := store.FlagPresent(ctx, db.SQL(), store.FlagThreadsJournalImported)
	if err != nil {
		log.Error("thread journal import: flag read failed; skipping (retries next boot)", "err", err)
		return
	}
	if done {
		return
	}
	rows, err := store.ReadLegacyThreads(ctx, db.SQL())
	if err != nil {
		log.Error("thread journal import: thread read failed; skipping (retries next boot)", "err", err)
		return
	}
	messages, err := store.ReadLegacyMessages(ctx, db.SQL())
	if err != nil {
		log.Error("thread journal import: message read failed; skipping (retries next boot)", "err", err)
		return
	}
	attachments, err := store.ReadLegacyAttachments(ctx, db.SQL())
	if err != nil {
		log.Error("thread journal import: attachment read failed; skipping (retries next boot)", "err", err)
		return
	}

	byThread := map[string][]store.LegacyMessage{}
	skippedMessages := 0
	for _, m := range messages {
		byThread[m.ThreadID] = append(byThread[m.ThreadID], m)
	}
	byMessage := map[string][]store.LegacyAttachment{}
	for _, a := range attachments {
		byMessage[a.MessageID] = append(byMessage[a.MessageID], a)
	}

	perStream := map[string][]any{}
	threads, msgs, atts := 0, 0, 0
	for _, row := range rows {
		if !importableOwner(log, "thread journal import", row.Owner) {
			continue
		}
		thread, ok := row.ToThread()
		if !ok {
			log.Warn("thread journal import: unparseable thread row skipped",
				"thread", row.ID, "owner", row.Owner, "network", row.Network, "remote", row.Remote)
			continue
		}
		threadMsgs := byThread[thread.ID.String()]
		// The preview fields mirror the legacy list join: the last message
		// by (created_at, rowid) — ReadLegacyMessages already sorted.
		image := domain.ThreadSummaryImage{
			ThreadID:     thread.ID.String(),
			Owner:        thread.Owner.String(),
			Network:      string(thread.Remote.Network()),
			Remote:       thread.Remote.String(),
			LastActivity: thread.LastActivityAt.UnixNano(),
			Unread:       thread.Unread,
			Pinned:       thread.Pinned,
			Archived:     thread.Archived,
			Muted:        thread.Muted,
		}
		if len(threadMsgs) > 0 {
			last, ok := threadMsgs[len(threadMsgs)-1].ToMessage(thread)
			if ok {
				image.LastBody, image.LastDirection, image.LastChannel = last.Body,
					string(last.Direction), string(last.Channel)
				image.LastAttachment = len(byMessage[last.ID.String()]) > 0
			}
		}
		key := cqrs.ThreadStreamKey(thread.Owner, thread.Remote.Network(), thread.Remote)
		for _, m := range threadMsgs {
			msg, ok := m.ToMessage(thread)
			if !ok {
				skippedMessages++
				log.Warn("thread journal import: unparseable message row skipped",
					"message", m.ID, "thread", m.ThreadID)
				continue
			}
			fact := domain.MessageAppended{
				Key:           domain.ThreadRowKey{ThreadID: thread.ID.String()},
				Summary:       image,
				MessageID:     msg.ID.String(),
				ThreadID:      msg.ThreadID.String(),
				Owner:         msg.Owner.String(),
				Network:       string(msg.Remote.Network()),
				Remote:        msg.Remote.String(),
				Direction:     string(msg.Direction),
				Channel:       string(msg.Channel),
				Body:          msg.Body,
				Status:        string(msg.Status),
				ProviderRef:   msg.ProviderRef,
				FailureKind:   string(msg.FailureKind),
				FailureDetail: msg.FailureDetail,
				CreatedAt:     msg.CreatedAt,
			}
			perStream[key] = append(perStream[key], fact)
			msgs++
			for _, a := range byMessage[msg.ID.String()] {
				attachment, ok := a.ToAttachment(msg)
				if !ok {
					log.Warn("thread journal import: unparseable attachment row skipped",
						"attachment", a.ID, "message", a.MessageID)
					continue
				}
				perStream[key] = append(perStream[key], domain.AttachmentStored{
					ID:        attachment.ID.String(),
					MessageID: attachment.MessageID.String(),
					Owner:     thread.Owner.String(),
					Name:      attachment.Name,
					MimeType:  attachment.MimeType,
					SizeBytes: attachment.SizeBytes,
					Path:      attachment.Path,
				})
				atts++
			}
		}
		if len(perStream[key]) > 0 {
			threads++
		}
	}

	imported, err := appendJournalFacts(ctx, svc.Sys, cqrs.ThreadsStreamType, perStream)
	if err != nil {
		log.Error("thread journal import: append failed; skipping rest (retries next boot)",
			"err", err, "imported_so_far", imported)
		return
	}
	if err := store.SetFlag(ctx, db.SQL(), store.FlagThreadsJournalImported, time.Now().UTC().Format(time.RFC3339)); err != nil {
		log.Error("thread journal import: flag write failed (imports re-run idempotently next boot)",
			"err", err, "imported", imported)
		return
	}
	log.Info("thread journal import complete",
		"threads", threads, "messages", msgs, "attachments", atts,
		"events", imported, "skipped_messages", skippedMessages)
}
