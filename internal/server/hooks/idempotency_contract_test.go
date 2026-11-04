package hooks

import (
	"context"
	"testing"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	idempotency "github.com/larsartmann/go-idempotency"
	"github.com/larsartmann/go-idempotency/contract"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// contractWindow adapts the hooks.DedupWindow port (collection-named,
// caller-injected time) onto the go-idempotency Store contract (flat
// keys, wall clock, sentinel errors) so the library's full invariant
// suite can audit the real window. The Record shim rides CheckAndRecord's
// duplicate verdict: Record must be a no-op on an existing key, which is
// exactly what leaving a dup behind means. Test-only (T18): production
// consumes the port directly.
type contractWindow struct {
	window DedupWindow
	coll   string
}

var _ idempotency.Store = contractWindow{}

func (c contractWindow) Seen(ctx context.Context, key string) (bool, error) {
	return c.window.Seen(ctx, c.coll, key, time.Now())
}

func (c contractWindow) Record(ctx context.Context, key string, ttl time.Duration) error {
	if ttl <= 0 {
		return idempotency.ErrInvalidTTL
	}
	_, err := c.window.CheckAndRecord(ctx, c.coll, key, ttl, time.Now())
	return err
}

func (c contractWindow) CheckAndRecord(ctx context.Context, key string, ttl time.Duration) error {
	if ttl <= 0 {
		return idempotency.ErrInvalidTTL
	}
	dup, err := c.window.CheckAndRecord(ctx, c.coll, key, ttl, time.Now())
	if err != nil {
		return err
	}
	if dup {
		return idempotency.ErrDuplicate
	}
	return nil
}

// engineWindow is the test-local port shim over the engine's native
// DedupStore capability — the production twin lives in internal/app
// (engineDedupWindow); this copy exists because app imports hooks, so
// hooks cannot import app to reuse it.
type engineWindow struct {
	store metaengine.DedupStore
}

func (w engineWindow) Seen(ctx context.Context, collection, key string, now time.Time) (bool, error) {
	return w.store.DedupSeen(ctx, collection, key, now)
}

func (w engineWindow) CheckAndRecord(
	ctx context.Context, collection, key string, ttl time.Duration, now time.Time,
) (bool, error) {
	return w.store.DedupCheckAndRecord(ctx, collection, key, ttl, now)
}

func (w engineWindow) Sweep(ctx context.Context, collection string, now time.Time) (int, error) {
	return w.store.DedupSweep(ctx, collection, now)
}

// TestDedupWindowSatisfiesIdempotencyContract runs the go-idempotency
// contract suite (atomicity under concurrency, TTL expiry, duplicate
// detection, key independence) against the REAL metaengine-backed
// window — a fresh readmodels engine per subtest. A red here means the
// idem-gate seam's semantics drifted from the standard idempotency
// contract the webhook and call lanes depend on. TimingScale 3 keeps
// the wall-clock expiry subtests honest on loaded CI runners.
func TestDedupWindowSatisfiesIdempotencyContract(t *testing.T) {
	contract.RunTestsStrict(t, func(t *testing.T) idempotency.Store {
		engine, err := readmodels.Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := engine.Shutdown(); err != nil {
				t.Errorf("engine shutdown: %v", err)
			}
		})
		return contractWindow{window: engineWindow{store: engine.DedupStore()}, coll: DedupCollection}
	}, contract.Options{TimingScale: 3})
}
