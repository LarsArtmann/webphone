// The concurrent first-sight probe (TODO_LIST full-code-review row): two
// simultaneous Receive lanes for the same unseen (owner, remote) both
// miss FindThread's projection scan and answer different fresh thread
// ids. The probe pins what actually happens next: both appends target
// the SAME journal stream (ThreadStreamKey is owner|network|remote), the
// first adopter wins the thread identity, and the decider's
// cqrs.thread_identity rejection refuses the loser — one thread, never
// two. The loser's message is recoverable by a re-delivery (the webhook
// lane's provider retry / the receiver loop's retry), which this probe
// also pins: the retried append resolves the winner's thread and lands.
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

	threads, err := svc.Threads(ctx, owner)
	if err != nil {
		t.Fatalf("Threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("thread split: %d threads exist for one (owner, remote), want 1 (losers: %v)", len(threads), losers)
	}

	for _, err := range losers {
		errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	}

	// The loser's message is recoverable by re-delivery: the retried
	// Receive resolves the winner's thread and both messages share it.
	retried, err := svc.Receive(ctx, domain.InboundMessage{
		Owner: owner, From: remote, Body: "retried re-delivery", ReceivedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("re-delivery after the race: %v", err)
	}
	if retried.ThreadID != winners[0].ThreadID {
		t.Fatalf("re-delivery landed on thread %s, winner holds %s", retried.ThreadID, winners[0].ThreadID)
	}

	threads, err = svc.Threads(ctx, owner)
	if err != nil {
		t.Fatalf("Threads after re-delivery: %v", err)
	}
	if len(threads) != 1 || threads[0].Thread.Unread != 2 {
		t.Fatalf("after re-delivery: %d threads, unread %d — want 1 thread, both messages counted", len(threads), threads[0].Unread)
	}
}
