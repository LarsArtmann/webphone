package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/web/views"
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
	notifier := NewNotifier(hubs, store.NewMessages(db), newTestEngine(t).Faxes(), nil, false, views.FaxCapability{Send: true, Receive: true})

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

// TestFaxPushCarriesTheCapability pins T20.2: the SSE fax-list push
// renders through the SAME lane capability as the full panel, so a
// pushed payload can never promise a form or an affordance the render
// withholds — the off-deployment's push carries the honest empty copy.
func TestFaxPushCarriesTheCapability(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	hubs := NewHubs()
	owner := domain.MustParseExtension("100")
	notifier := NewNotifier(hubs, store.NewMessages(db), newTestEngine(t).Faxes(), nil, false, views.FaxCapability{})

	events := hubs.get(owner).Hub().Subscribe()
	t.Cleanup(func() { hubs.get(owner).Hub().Unsubscribe(events) })

	notifier.Project(context.Background(), domain.FaxAdvanced{Owner: owner, Fax: domain.GenerateFaxID()})

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		select {
		case ev := <-events:
			if ev.Event != "fax" {
				continue
			}
			body := string(ev.Data)
			if !strings.Contains(body, "No faxes yet.") {
				t.Fatalf("off-deployment fax push lost the honest empty copy; payload:\n%s", body)
			}
			if strings.Contains(body, "send a PDF above") {
				t.Fatalf("off-deployment fax push promises a send affordance; payload:\n%s", body)
			}
			return
		case <-deadline.C:
			t.Fatal("projection never pushed the fax list")
		}
	}
}
