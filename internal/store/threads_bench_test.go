package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// BenchmarkLegacyThreadListRead is the legacy SQL twin of the journal
// gate's BenchmarkThreadListRead (cqrs package): same shape, same
// density — 50 threads, one message each — so the A38 perf record
// compares like with like (the budget: the journal read stays within
// 2x of this number).
func BenchmarkLegacyThreadListRead(b *testing.B) {
	db, err := Open(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	messages := NewMessages(db)
	ctx := context.Background()
	owner := domain.MustParseExtension("100")
	for i := range 50 {
		remote := domain.MustParseAddress(domain.NetworkSMS, fmt.Sprintf("+49170%06d", i))
		threadID, err := messages.FindThread(ctx, owner, remote, time.Now())
		if err != nil {
			b.Fatal(err)
		}
		msg := domain.NewInboundMessage(domain.GenerateMessageID(), threadID, owner, remote,
			fmt.Sprintf("seed %d", i), 0, time.Now())
		if err := messages.AppendMessage(ctx, msg); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := messages.ListThreads(ctx, owner); err != nil {
			b.Fatalf("list: %v", err)
		}
	}
}
