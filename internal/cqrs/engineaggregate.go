package cqrs

import (
	"context"

	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"
)

// engineCount reads the engine-native COUNT aggregate over the System's
// engine (the metrics surfaces' totals — no row materialization, the
// same posture the readmodels engine held). The System exposes its
// engine only through TimerEngine (the deployment's sole engine; the
// name is the upstream seam, the use is "the engine"); the sqlite
// engine implements AggregateReader, and a swap that dropped the
// capability fails the call loudly instead of answering zero.
func engineCount(ctx context.Context, sys *system.System, code, collection string) (int64, error) {
	eng := sys.TimerEngine()
	if eng == nil {
		return 0, errorfamily.Newf(errorfamily.Infrastructure, code, "the system exposes no engine for %s", collection)
	}
	reader, ok := eng.(metaengine.AggregateReader)
	if !ok {
		return 0, errorfamily.Newf(errorfamily.Infrastructure, code, "engine %T lacks the aggregate capability", eng)
	}
	n, err := reader.Aggregate(ctx, collection, metaengine.AggregateCount, "", nil)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, code, "count %s", collection)
	}
	return int64(n), nil
}
