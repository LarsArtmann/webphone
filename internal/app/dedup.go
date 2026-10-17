package app

import (
	"context"
	"log/slog"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"

	"github.com/larsartmann/webphone/internal/server/hooks"
)

// dedupSweepInterval reclaims expired engine dedup rows. Expiry is
// honored on read regardless, so the sweep is space reclamation, not
// correctness: a lazy cadence is the honest choice for a single-writer
// sqlite the hot webhook lane also writes to.
const dedupSweepInterval = 15 * time.Minute

// engineDedupWindow adapts the metaengine DedupStore capability into
// the hooks.DedupWindow port (method names differ; the adapter lives
// here because readmodels must stay server-free and the hooks lane
// stays metaengine-free — the composition root is the one place that
// may see both).
type engineDedupWindow struct {
	store metaengine.DedupStore
}

var _ hooks.DedupWindow = engineDedupWindow{}

func (w engineDedupWindow) Seen(ctx context.Context, collection, key string, now time.Time) (bool, error) { //nolint:branching-flow // verbatim port of metaengine.DedupStore's parameter order — the engine API IS the contract
	return w.store.DedupSeen(ctx, collection, key, now)
}

func (w engineDedupWindow) CheckAndRecord(
	ctx context.Context, collection, key string, ttl time.Duration, now time.Time, //nolint:branching-flow // verbatim port of metaengine.DedupStore's parameter order — the engine API IS the contract
) (bool, error) {
	return w.store.DedupCheckAndRecord(ctx, collection, key, ttl, now)
}

func (w engineDedupWindow) Sweep(ctx context.Context, collection string, now time.Time) (int, error) {
	return w.store.DedupSweep(ctx, collection, now)
}

// startDedupSweeper runs the expired-row reclamation loop until ctx
// ends. Started from App.Start alongside the retention sweep — the
// same lifecycle home for every background loop.
func startDedupSweeper(ctx context.Context, log *slog.Logger, window hooks.DedupWindow) {
	go func() {
		ticker := time.NewTicker(dedupSweepInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				removed, err := window.Sweep(ctx, hooks.DedupCollection, now)
				if err != nil {
					log.Warn("dedup sweep failed", "error", err, "collection", hooks.DedupCollection)
					continue
				}
				if removed > 0 {
					log.Info("dedup sweep", "collection", hooks.DedupCollection, "expired_removed", removed)
				}
			}
		}
	}()
}
