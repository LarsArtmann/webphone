package cqrs

import (
	"context"
	"testing"
	"time"

)

func TestDiagProjectionCatchup(t *testing.T) {
	dials := bootDialSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")
	allowed, at, err := dials.Attempt(ctx, owner, "+15550001111", time.Hour)
	t.Logf("attempt: allowed=%v at=%v err=%v", allowed, at, err)
	for _, w := range dials.sys.WorkerStatus() {
		t.Logf("worker: %+v", w)
	}
	t.Logf("lag: %+v", dials.sys.LagPerProjection())
	time.Sleep(200 * time.Millisecond)
	for _, w := range dials.sys.WorkerStatus() {
		t.Logf("worker after: %+v", w)
	}
	last, found, err := dials.Last(ctx, owner, "+15550001111")
	t.Logf("last: found=%v last=%v err=%v", found, last, err)
	slots, err := dials.Active(ctx, owner, time.Hour)
	t.Logf("active: %+v err=%v", slots, err)
}
