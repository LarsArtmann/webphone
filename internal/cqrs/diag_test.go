package cqrs

import (
	"context"
	"testing"
	"time"
)

func TestDiagCapExpunge(t *testing.T) {
	_ = t.TempDir()
	dials := bootDialSystem(t, "/tmp/diagdb")
	ctx := context.Background()
	owner := mustExt(t, "1001")

	backdated := time.Now().Add(-time.Hour)
	timeNow = func() time.Time { return backdated }
	if _, _, err := dials.Attempt(ctx, owner, "+15550001", time.Hour); err != nil {
		t.Fatalf("oldest attempt: %v", err)
	}
	timeNow = time.Now
	awaitLast(t, dials, owner, "+15550001", backdated)
	for i := 0; i < dialWindowMaxPerOwner; i++ {
		if _, _, err := dials.Attempt(ctx, owner, remoteOf(i), time.Hour); err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	time.Sleep(300 * time.Millisecond)
	_, found, _ := dials.Last(ctx, owner, "+15550001")
	t.Logf("oldest found after cap: %v", found)
	for _, w := range dials.sys.WorkerStatus() {
		t.Logf("worker: %+v", w)
	}
}
