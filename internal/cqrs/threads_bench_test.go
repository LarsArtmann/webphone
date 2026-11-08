package cqrs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"

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

// BenchmarkThreadAppendBareDispatch is the budget's own shape (the P2
// precedent): dispatch → decider → journal append, NO read-your-writes
// await — the apply lands asynchronously. The A38 gate number.
func BenchmarkThreadAppendBareDispatch(b *testing.B) {
	t := &testing.T{}
	sys := bootDomainSystem(t, b.TempDir(), ThreadsDomainDecl())
	defer func() { _ = sys.Close() }()
	ctx := context.Background()
	owner := mustExt(t, "100")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// A fresh stream per iteration: the budget's "single event"
		// shape (the P2 dial precedent — tiny state, the decision and
		// the journal append isolated). The one-stream-to-1400-events
		// grind is BenchmarkThreadAppendDispatch's job.
		remote := domain.MustParseAddress(domain.NetworkSMS, fmt.Sprintf("+4917%07d", i))
		key := ThreadStreamKey(owner, remote.Network(), remote)
		msg := domain.NewInboundMessage(domain.GenerateMessageID(), domain.ThreadID{}, owner, remote,
			fmt.Sprintf("bench %d", i), 0, time.Now())
		msg.ThreadID = threadIDForBench(key)
		if err := dispatchThreadCommand(ctx, sys, "thread.append", key,
			func(base *command.BasicCommand) *AppendMessage {
				return &AppendMessage{BasicCommand: base, Msg: msg}
			}); err != nil {
			b.Fatalf("dispatch %d: %v", i, err)
		}
	}
}

// threadIDForBench keeps one stable thread id per stream without paying
// a projection read per iteration (the bare-dispatch bench must not
// measure the await it deliberately skips).
var benchThreadIDs = map[string]domain.ThreadID{}

func threadIDForBench(key string) domain.ThreadID {
	if id, ok := benchThreadIDs[key]; ok {
		return id
	}
	id := domain.GenerateThreadID()
	benchThreadIDs[key] = id
	return id
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
