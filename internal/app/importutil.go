// The journal importers' shared primitive (SUPERB): every legacy-engine
// → journal port raw-appends synthetic facts per stream, continuing
// from each stream's own tail. The policy (flag-gated, log-and-retry,
// full-success-then-flag) lives in each importer; this file owns the
// append mechanics once.
package app

import (
	"context"
	"log/slog"
	"maps"
	"slices"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
)

// importerActor is the journal actor synthetic import events carry (one
// identity for every legacy port; the audit surface greps it).
func importerActor() id.ActorID {
	actor, _ := id.ParseActorID("system:importer")
	return actor
}

// importableOwner is the importers' shared row guard: a legacy row
// whose owner does not parse as an extension is SKIPPED with a WARN
// (one bad row must never abort a whole owner's port).
func importableOwner(log *slog.Logger, importer, owner string) bool {
	if _, err := domain.ParseExtension(owner); err != nil {
		log.Warn(importer+": unparseable owner row skipped", "owner", owner, "err", err)
		return false
	}
	return true
}

// appendJournalFacts raw-appends per-stream fact batches into the
// journal: each stream's events continue from its own Load (per-stream
// tail sync makes re-runs idempotent — versions never duplicate, and a
// crash mid-import leaves a deterministic prefix the retry extends),
// facts append in the caller's slice order (import ordering is the
// caller's timestamp-fidelity duty), and the actor is the importer.
// Returns the number of events appended.
func appendJournalFacts(ctx context.Context, sys *system.System, streamType id.StreamType, perStream map[string][]any) (int, error) {
	imported := 0
	for _, key := range slices.Sorted(maps.Keys(perStream)) {
		streamID, err := id.ParseStreamID(key)
		if err != nil {
			return imported, errorfamily.WrapRejectionf(err, "cqrs.journal_import", "parse %s stream id %q", streamType, key)
		}
		ref := id.NewStreamRef(streamType, streamID)
		existing, err := sys.EventStore().Load(ctx, ref)
		if err != nil {
			return imported, errorfamily.WrapInfrastructuref(err, "cqrs.journal_import", "load %s stream tail for %s", streamType, key)
		}
		ver := event.Version(len(existing))
		actor := importerActor()
		events := make([]event.Event, 0, len(perStream[key]))
		for _, fact := range perStream[key] {
			ver++
			evt, err := event.New(event.Type(metaengine.EventTypeName(fact)), streamID, streamType, ver, fact,
				event.WithActor(actor))
			if err != nil {
				return imported, errorfamily.WrapInfrastructuref(err, "cqrs.journal_import", "build synthetic fact for %s %s", streamType, key)
			}
			events = append(events, evt)
		}
		if len(events) == 0 {
			continue
		}
		if err := sys.EventStore().AppendBatch(ctx, ref, events); err != nil {
			return imported, errorfamily.WrapInfrastructuref(err, "cqrs.journal_import", "append %s facts for %s", streamType, key)
		}
		imported += len(events)
	}
	return imported, nil
}
