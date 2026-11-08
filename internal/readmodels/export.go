package readmodels

import (
	"context"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-error-family"
)

// exportAllViews reads every row of a collection through its typed
// reader — the ONE home of the journal importers' export shape
// (WithLimit rides because the planner's default ceiling is 100).
// The code names the calling seam for the error family registry.
func exportAllViews[V any](ctx context.Context, reader *metaengine.TypedReader[V], code, what string) ([]V, error) {
	views, err := reader.Scan(ctx, metaengine.WithLimit(1_000_000))
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(err, code, what)
	}
	return views, nil
}
