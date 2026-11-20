package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
)

// loggingDeadLetterStore decorates the projection host's dead-letter
// store: every poison landing WARNs once, greppably. The library's
// sendToDLQ is silent (projectionhost worker.go), and without this
// decorator the only trace of a poisoned projection would be a row in
// system.db nobody reads — the ES core's one unobservable failure
// mode. Only Store is intercepted; the raw store stays reachable for
// the admin surface via CqrsService.DLQ.
type loggingDeadLetterStore struct {
	projectionhost.DeadLetterStore
	log *slog.Logger
}

// Store records the entry and WARNs. The underlying error is returned
// unwrapped: the library's call site already wraps it as
// "store dead-letter entry", and a second wrap would double-report.
func (s loggingDeadLetterStore) Store(ctx context.Context, entry projectionhost.DeadLetterEntry) error {
	if err := s.DeadLetterStore.Store(ctx, entry); err != nil {
		return err
	}
	s.log.Warn("projection dead-lettered an event",
		"projection", entry.ProjectionName,
		"event_type", entry.EventType,
		"event_id", entry.EventID,
		"error", entry.Error,
		"code", entry.ErrorCode,
		"hint", "fix the handler, then replay; the entry persists in system.db (projection_dead_letters)",
	)
	return nil
}

// warnDLQDepth surfaces poison events that survived a restart: the
// DLQ is durable, so entries that landed in a previous boot must
// re-WARN in this one. Reads only when the store implements the
// admin surface; a read failure WARNs and never blocks the boot —
// an observability probe must not take the product down with it.
func warnDLQDepth(ctx context.Context, store projectionhost.DeadLetterStore, log *slog.Logger) {
	admin, ok := store.(projectionhost.DeadLetterStoreAdmin)
	if !ok {
		return
	}
	count, err := admin.Count(ctx)
	if err != nil {
		log.Warn("projection dead-letter queue unreadable", "error", err.Error())
		return
	}
	if count == 0 {
		return
	}
	attrs := []any{"count", count}
	if entries, listErr := admin.ListPaged(ctx, "", 0, 1); listErr == nil && len(entries) > 0 {
		// ListPaged orders by failed_at ascending: the first entry IS
		// the oldest poison still waiting.
		attrs = append(attrs, "oldest_age", time.Since(entries[0].FailedAt).Round(time.Second).String())
	}
	log.Warn("projection dead-letter queue carries poison events from a previous run", attrs...)
}
