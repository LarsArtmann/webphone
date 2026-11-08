// The bounded read-your-writes await (SUPERB D7 ruling): dispatch
// returns before the projection worker folds the event (~1–2 ms, the
// A06 spike), and seams that re-read state in the SAME request (the
// contacts panel re-render, the settings partial) must not race their
// own write. The mechanism: read the committed tail of the stream the
// backend just wrote (tail sync from a cached version), then poll the
// projection host's LIVE checkpoints until the projection has
// processed at or past our last event ID — bounded at 250 ms, WARN and
// proceed on expiry (the availability posture: staleness, never data
// loss).
package cqrs

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"
)

// visibilityBudget bounds the read-your-writes await (D7: 250 ms hard
// cap, ~2 ms typical).
const visibilityBudget = 250 * time.Millisecond

// Visibility polls back off from 200 µs to 5 ms.
const (
	visibilityPollStart = 200 * time.Microsecond
	visibilityPollMax   = 5 * time.Millisecond
)

// streamVersions caches each stream's committed length. Guarded by the
// owning backend's mutex: every write to a stream dispatches through
// its backend, so the cache stays exact (the boot importer runs before
// serving, restart resets it lazily).
type streamVersions map[string]event.Version

// streamRef builds the journal ref for a stream key (any non-empty
// UTF-8 key is a legal stream id; the key is never parsed back).
func streamRef(streamType id.StreamType, key string) (id.StreamRef, error) {
	streamID, err := id.ParseStreamID(key)
	if err != nil {
		return id.StreamRef{}, errorfamily.WrapRejectionf(err, "cqrs.stream_ref", "parse %s stream key %q", streamType, key)
	}
	return id.NewStreamRef(streamType, streamID), nil
}

// committedTail returns the events committed to the stream after the
// cached version (the whole stream on first sight), advancing the
// cache to the new committed length. Backends call it AFTER their
// dispatch: the tail is then exactly the just-committed batch.
func committedTail(ctx context.Context, sys *system.System, streamType id.StreamType, key string, vers streamVersions) ([]event.Event, error) {
	ref, err := streamRef(streamType, key)
	if err != nil {
		return nil, err
	}
	from, cached := vers[key]
	if !cached {
		from = 0
	}
	tail, err := sys.EventStore().LoadFromVersion(ctx, ref, from)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.stream_tail", "load %s tail for %s", streamType, key)
	}
	vers[key] = from + event.Version(len(tail))
	return tail, nil
}

// visibilityWorkerName is the System's projection worker name (the
// default wiring registers exactly one worker, "projections", that
// folds every declared projection — system/projection_wiring.go). The
// await targets it by name; its live checkpoint names the last folded
// event across the whole host.
const visibilityWorkerName = "projections"

// awaitVisibility blocks until the projection worker's live checkpoint
// has processed at or past afterID (event IDs are monotonic ULIDs:
// lexical order IS processing order), bounded by visibilityBudget.
// Expiry WARNs and returns — the write IS committed; a lagging fold is
// staleness, not loss. A missing worker keeps waiting (a stopped
// worker's checkpoint never advances; the budget resolves it).
func awaitVisibility(ctx context.Context, sys *system.System, projection, afterID string) {
	if afterID == "" {
		return
	}
	host := sys.ProjectionHost()
	if host == nil {
		return
	}
	deadline := time.Now().Add(visibilityBudget)
	delay := visibilityPollStart
	for {
		for _, state := range host.Status() {
			if state.Name == visibilityWorkerName && strings.Compare(state.Checkpoint, afterID) >= 0 {
				return
			}
		}
		if time.Now().After(deadline) {
			slog.Warn("cqrs projection visibility budget expired (write committed; read may be briefly stale)",
				"projection", projection, "after_event", afterID, "budget", visibilityBudget.String())
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, visibilityPollMax)
	}
}

// warnVisibilityTailFailed is the shared best-effort tail-read WARN:
// the write IS committed, only the await's tail read failed.
func warnVisibilityTailFailed(err error, aggregate, owner string) {
	slog.Warn("write committed but the visibility tail read failed", "err", err, "aggregate", aggregate, "owner", owner)
}
