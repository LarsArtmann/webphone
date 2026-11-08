package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkDialAttemptDispatch measures the journal-backed gate's
// check+record transaction: dispatch → decider → journal append →
// publish (the projection apply lands ~1-2ms later and is deliberately
// NOT awaited — the verdict answers from the decision). Recorded in
// the SUPERB plan errata against the engine-backed baseline.
func BenchmarkDialAttemptDispatch(b *testing.B) {
	t := &testing.T{}
	dials := bootDialSystem(t, b.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		remote := fmt.Sprintf("+1606%06d", i)
		if allowed, _, err := dials.Attempt(ctx, owner, remote, time.Hour); err != nil || !allowed {
			b.Fatalf("attempt %d: allowed=%v err=%v", i, allowed, err)
		}
	}
}

// BenchmarkDialAttemptDeny measures the deny path (decision-only, no
// journal write): the hot path a retry loop hammers.
func BenchmarkDialAttemptDeny(b *testing.B) {
	t := &testing.T{}
	dials := bootDialSystem(t, b.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")
	if _, _, err := dials.Attempt(ctx, owner, "+1606000000", time.Hour); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if allowed, _, err := dials.Attempt(ctx, owner, "+1606000000", time.Hour); err != nil || allowed {
			b.Fatalf("deny: allowed=%v err=%v", allowed, err)
		}
	}
}

// BenchmarkDialWindowRead measures the projection read the typeahead
// hits (owner-filtered scan of active slots).
func BenchmarkDialWindowRead(b *testing.B) {
	t := &testing.T{}
	dials := bootDialSystem(t, b.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")
	for i := 0; i < 50; i++ {
		remote := fmt.Sprintf("+1607%05d", i)
		if _, _, err := dials.Attempt(ctx, owner, remote, time.Hour); err != nil {
			b.Fatal(err)
		}
	}
	time.Sleep(200 * time.Millisecond) // let the projection drain

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := dials.Active(ctx, owner, time.Hour); err != nil {
			b.Fatal(err)
		}
	}
}
