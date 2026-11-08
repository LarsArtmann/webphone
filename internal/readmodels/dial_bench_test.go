package readmodels_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// BenchmarkDialAttemptEngine is the PRE-SWAP baseline (the engine-backed
// gate's check+record transaction): recorded beside the journal-backed
// numbers in the SUPERB plan errata (A19).
func BenchmarkDialAttemptEngine(b *testing.B) {
	engine, err := readmodels.Open(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = engine.Shutdown() })
	ctx := context.Background()
	owner := mustExtension(&testing.T{}, "1001")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		remote := fmt.Sprintf("+1606%06d", i)
		if allowed, _, err := engine.DialWindows().Attempt(ctx, owner, remote, time.Hour); err != nil || !allowed {
			b.Fatalf("attempt %d: allowed=%v err=%v", i, allowed, err)
		}
	}
}
