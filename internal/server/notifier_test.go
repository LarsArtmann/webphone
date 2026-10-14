package server

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// TestProjectNotificationPolicy pins the event → SSE projection (train
// E1): message events owe the thread list AND the transcript, a flag
// change owes the list ONLY (the event carries no thread id on purpose
// — an empty transcript would wipe the open conversation), and a fax
// advance owes the fax list. Payload bytes are pinned separately by
// the SSE tests; this is the POLICY table.
func TestProjectNotificationPolicy(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	hubs := NewHubs()
	owner := domain.MustParseExtension("100")
	notifier := NewNotifier(hubs, store.NewMessages(db), store.NewFaxes(db), nil, false)

	events := hubs.get(owner).Hub().Subscribe()
	t.Cleanup(func() { hubs.get(owner).Hub().Unsubscribe(events) })

	ctx := context.Background()

	notifier.Project(ctx, domain.MessageReceived{Owner: owner, Thread: domain.GenerateThreadID()})
	expectNamed(t, events, "threads")
	expectNamed(t, events, "thread")

	notifier.Project(ctx, domain.MessageDeliveryAdvanced{Owner: owner, Thread: domain.GenerateThreadID()})
	expectNamed(t, events, "threads")
	expectNamed(t, events, "thread")

	notifier.Project(ctx, domain.ThreadFlagChanged{Owner: owner})
	expectNamed(t, events, "threads")
	expectQuiet(t, events, "flag change must not push a transcript")

	notifier.Project(ctx, domain.FaxAdvanced{Owner: owner, Fax: domain.GenerateFaxID()})
	expectNamed(t, events, "fax")
	expectQuiet(t, events, "fax advance owes exactly one push")
}

func expectNamed(t *testing.T, events <-chan sse.Event, name string) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		select {
		case ev := <-events:
			if ev.Event == name {
				return
			}
		case <-deadline.C:
			t.Fatalf("projection never pushed %q", name)
		}
	}
}

func expectQuiet(t *testing.T, events <-chan sse.Event, because string) {
	t.Helper()
	deadline := time.NewTimer(150 * time.Millisecond)
	defer deadline.Stop()
	select {
	case ev := <-events:
		t.Fatalf("unexpected %q push: %s", ev.Event, because)
	case <-deadline.C:
	}
}
