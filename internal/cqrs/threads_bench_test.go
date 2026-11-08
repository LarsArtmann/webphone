package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// BenchmarkThreadAppendDispatch measures the core send lane's full
// contract: dispatch → decider → journal append → read-your-writes
// await (the handler re-renders list + transcript in the SAME request,
// so the await is part of the honest number, unlike the dial verdict
// that answers from the decision alone). Recorded in the SUPERB plan
// errata against the P2 baseline; the A38 budget is p95 < 2 ms.
func BenchmarkThreadAppendDispatch(b *testing.B) {
	t := &testing.T{}
	sys := bootDomainSystem(t, b.TempDir(), ThreadsDomainDecl())
	defer func() { _ = sys.Close() }()
	backend := NewThreads(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+4917000001")
	threadID, err := backend.FindThread(ctx, owner, remote, time.Now())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := domain.NewInboundMessage(domain.GenerateMessageID(), threadID, owner, remote,
			fmt.Sprintf("bench %d", i), 0, time.Now())
		if err := backend.AppendMessage(ctx, msg); err != nil {
			b.Fatalf("append %d: %v", i, err)
		}
	}
}

// BenchmarkThreadListRead measures the panel's list read (the owner's
// thread summaries) over a realistically dense owner — 50 threads, the
// newest each carrying a message. The legacy SQL twin lives in
// internal/store; the A38 budget is within 2× of that number.
func BenchmarkThreadListRead(b *testing.B) {
	t := &testing.T{}
	sys := bootDomainSystem(t, b.TempDir(), ThreadsDomainDecl())
	defer func() { _ = sys.Close() }()
	backend := NewThreads(sys)
	ctx := context.Background()
	owner := mustExt(t, "100")
	for i := range 50 {
		remote := domain.MustParseAddress(domain.NetworkSMS, fmt.Sprintf("+49170%06d", i))
		threadID, err := backend.FindThread(ctx, owner, remote, time.Now())
		if err != nil {
			b.Fatal(err)
		}
		msg := domain.NewInboundMessage(domain.GenerateMessageID(), threadID, owner, remote,
			fmt.Sprintf("seed %d", i), 0, time.Now())
		if err := backend.AppendMessage(ctx, msg); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := backend.ListThreads(ctx, owner); err != nil {
			b.Fatalf("list: %v", err)
		}
	}
}
