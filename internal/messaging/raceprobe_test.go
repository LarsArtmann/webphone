// The concurrent first-sight probe (TODO_LIST full-code-review row): two
// simultaneous Receive lanes for the same unseen (owner, remote) race
// FindThread's projection scan. BOTH outcomes are legal, and the probe
// accepts either: (a) the fold lands between the two scans — the second
// lane RESOLVES the first lane's thread and appends to it, zero
// refusals; (b) both lanes miss — each mints a fresh thread identity,
// both appends target the SAME journal stream (ThreadStreamKey is
// owner|network|remote), the first adopter wins the thread identity,
// and the decider's cqrs.thread_identity rejection refuses the loser.
// The invariant is interleaving-independent: ONE thread, never two,
// and the refused message stays recoverable by a re-delivery (the
// webhook lane's provider retry / the receiver loop's retry), which
// this probe also pins: the retried append resolves the winner's
// thread and lands. The original probe demanded outcome (b) exactly
// and flapped under -shuffle and suite parallelism when the scheduler
// produced (a).
package messaging_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/cqrs"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/messaging"
)

// awaitThreadFold polls the projection until the (owner, remote) thread
// has settled to exactly one thread carrying want messages — the same
// convergence pattern as the transcript tests' awaitTranscriptCount.
// Receive returns when the DECIDER has its verdict, but the projection
// fold that Threads reads is async to that; the probe's exact-count
// read depended on fold timing and flapped under -count>1 and suite
// parallelism (#249). The poll asserts the production invariant
// (read-your-writes convergence, the D7 checkpoint budget) instead of
// a race with the fold worker.
func awaitThreadFold(t *testing.T, svc *messaging.Service, owner domain.Extension, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		threads, err := svc.Threads(context.Background(), owner)
		if err != nil {
			t.Fatalf("Threads: %v", err)
		}
		if len(threads) == 1 && threads[0].Thread.Unread == want {
			return
		}
		if time.Now().After(deadline) {
			unread := -1
			if len(threads) == 1 {
				unread = threads[0].Thread.Unread
			}
			t.Fatalf("thread projection never settled to 1 thread with %d messages: got %d threads, unread %d", want, len(threads), unread)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestConcurrentFirstSightYieldsOneThread(t *testing.T) {
	blobs, err := blob.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sys, err := cqrs.NewThreadsSystem(context.Background(), "file:"+filepath.Join(t.TempDir(), "system.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sys.Close() })
	svc := messaging.New(cqrs.NewThreads(sys), blobs, nil, nil, nil)

	ctx := context.Background()
	owner := domain.MustParseExtension("100")
	remote := domain.MustParseAddress(domain.NetworkSMS, "+491709")

	const racers = 2
	start := make(chan struct{})
	results := make(chan error, racers)
	ids := make(chan domain.Message, racers)
	var wg sync.WaitGroup
	for i := range racers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			inbound := domain.InboundMessage{
				Owner: owner, From: remote,
				Body: "first-sight racer", ReceivedAt: time.Now(),
			}
			<-start
			msg, err := svc.Receive(ctx, inbound)
			if err != nil {
				results <- err
				return
			}
			ids <- msg
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	close(ids)

	var winners []domain.Message
	var losers []error
	for id := range ids {
		winners = append(winners, id)
	}
	for err := range results {
		losers = append(losers, err)
	}
	if len(winners)+len(losers) != racers {
		t.Fatalf("races settled without a verdict: %d winners, %d losers", len(winners), len(losers))
	}

	// The settle assert counts LANDED messages (len(winners)), not
	// racers: under outcome (a) both lanes landed, under (b) exactly
	// one did. len(threads) == 1 is the no-split invariant either way.
	awaitThreadFold(t, svc, owner, len(winners))

	for _, err := range losers {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	// The loser's message is recoverable by re-delivery: the retried
	// Receive resolves the winner's thread and both messages share it.
	if len(winners) == 0 {
		t.Fatalf("every racer was refused; nothing adopted the thread identity")
	}
	retried, err := svc.Receive(ctx, domain.InboundMessage{
		Owner: owner, From: remote, Body: "retried re-delivery", ReceivedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("re-delivery after the race: %v", err)
	}
	if retried.ThreadID != winners[0].ThreadID {
		t.Fatalf("re-delivery landed on thread %s, winner holds %s", retried.ThreadID, winners[0].ThreadID)
	}

	awaitThreadFold(t, svc, owner, len(winners)+1)
}
