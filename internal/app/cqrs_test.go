package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/samber/do/v2"
)

// TestCqrsServiceResolvesByName pins the composition contract §2: the
// system rides the injector as ONE named eager service, and New leaves
// system.db on disk beside webphone.db/readmodels.db (engine-owned,
// never the hand ladder's file).
func TestCqrsServiceResolvesByName(t *testing.T) {
	application, err := New(testConfig(t, false), discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		if err := application.Shutdown(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	})
	svc, err := do.InvokeNamed[*CqrsService](application.injector, cqrsProviderName)
	if err != nil {
		t.Fatalf("resolve cqrs: %v", err)
	}
	if svc.Sys == nil || svc.DLQ == nil {
		t.Fatal("resolved service must carry the System and the DLQ store")
	}
	if _, err := os.Stat(filepath.Join(application.cfg.DataDir, "system.db")); err != nil {
		t.Fatalf("system.db must exist after New: %v", err)
	}
}

// TestCqrsStartIsOnceOnlyAndShutdownIdempotent pins the lifecycle
// latches: a second Start is inert (system.Start is once-only) and a
// second Shutdown returns nil (the cascade may redial it).
func TestCqrsStartIsOnceOnlyAndShutdownIdempotent(t *testing.T) {
	svc, err := NewCqrs(testConfig(t, false), discardLogger())
	if err != nil {
		t.Fatalf("NewCqrs: %v", err)
	}
	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("second Start must be inert: %v", err)
	}
	if err := svc.Shutdown(); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	if err := svc.Shutdown(); err != nil {
		t.Fatalf("second Shutdown must be nil: %v", err)
	}
}

// TestCqrsHealthPassesOnHollowRoot pins the Phase-1 posture: no
// workers, engine green → the probe leg is green.
func TestCqrsHealthPassesOnHollowRoot(t *testing.T) {
	svc, err := NewCqrs(testConfig(t, false), discardLogger())
	if err != nil {
		t.Fatalf("NewCqrs: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })
	if err := svc.HealthCheck(context.Background()); err != nil {
		t.Fatalf("hollow-root health must pass: %v", err)
	}
}

// TestCqrsWorkerTriageNamesTheWorker pins the /healthz failure shape:
// a dead or lagging worker is named in the leg's error — the operator's
// path from a 503 body to the wedged component.
func TestCqrsWorkerTriageNamesTheWorker(t *testing.T) {
	healthy := []projectionhost.WorkerState{{Name: "projections", Status: projectionhost.WorkerLive}}
	if err := cqrsWorkerTriage(healthy); err != nil {
		t.Fatalf("healthy workers must pass: %v", err)
	}

	failed := []projectionhost.WorkerState{{Name: "projections", Status: projectionhost.WorkerFailed, LastError: "boom"}}
	err := cqrsWorkerTriage(failed)
	if err == nil || !strings.Contains(err.Error(), `"projections"`) || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("failed worker must be named with its last error, got: %v", err)
	}

	lagging := []projectionhost.WorkerState{{
		Name:   "projections",
		Status: projectionhost.WorkerLive,
		Lag:    cqrsWorkerLagBound + time.Second,
	}}
	err = cqrsWorkerTriage(lagging)
	if err == nil || !strings.Contains(err.Error(), `"projections"`) || !strings.Contains(err.Error(), "lag") {
		t.Fatalf("lagging worker must be named, got: %v", err)
	}
}

// cqrsPoisonFact is the test-only fact whose journal payload cannot
// decode into the fold's typed parameter — the deterministic poison
// path through the projection worker into the dead-letter store.
type cqrsPoisonFact struct {
	Key string
}

// TestCqrsPoisonEventDeadLetters pins the A12 contract: an event that
// repeatedly fails projection lands in the DLQ (named projection, named
// event) instead of wedging the worker forever, and the worker catches
// up afterwards — health green again.
func TestCqrsPoisonEventDeadLetters(t *testing.T) {
	const (
		streamType  = "cqrs-poison-test"
		commandName = "emit-poison"
		collection  = "cqrs_poison_views"
	)
	domain := system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := system.RegisterDecider[int](sys, streamType, decider.Decider[int]{
				Initial: 0,
				Apply:   func(state int, _ event.Event) (int, error) { return state, nil },
			}); err != nil {
				panic(err)
			}
			if err := system.RegisterCommand[*command.BasicCommand, int](sys, commandName,
				func(_ context.Context, cmd *command.BasicCommand) system.Op[int] {
					return system.Execute[int](context.Background(), cmd.StreamID(), streamType,
						func(_ int, ver event.Version) ([]event.Event, error) {
							// Raw garbage bytes as the payload: event.New
							// passes []byte through untouched, so the fold's
							// typed decode fails — identically, every retry.
							evt, err := event.New(event.Type("cqrsPoisonFact"),
								cmd.StreamID(), streamType, ver+1, []byte("<<not-cbor>>"))
							if err != nil {
								return nil, err
							}
							return []event.Event{evt}, nil
						})
				}); err != nil {
				panic(err)
			}
		},
		Projections: []system.ProjectionDeclaration{
			system.RawQuery(metaengine.Query[string, cqrsPoisonFact](collection,
				metaengine.OnRecord(cqrsPoisonFact{},
					func(_ record.Record, e cqrsPoisonFact) (string, cqrsPoisonFact) { return e.Key, e }),
			)),
		},
		// Fast backoff so the three DLQ retries fit inside test time.
		ProjectionHostOptions: []projectionhost.HostOption{
			projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond),
		},
	}
	svc, err := newCqrs(testConfig(t, false), discardLogger(), domain)
	if err != nil {
		t.Fatalf("newCqrs with poison domain: %v", err)
	}
	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = svc.Shutdown() })

	streamID, err := id.ParseStreamID("poison-stream-1")
	if err != nil {
		t.Fatal(err)
	}
	cmd, err := command.New(commandName, streamID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Sys.CommandDispatcher().Dispatch(ctx, cmd); err != nil {
		t.Fatalf("dispatch poison command: %v", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	for {
		entries, err := svc.DLQ.List(ctx, "")
		if err != nil {
			t.Fatalf("DLQ list: %v", err)
		}
		if len(entries) > 0 {
			// The worker name is the store-wide "projections" adapter; the
			// entry must name it and the poison event's type.
			if entries[0].ProjectionName == "" {
				t.Fatalf("dead-letter entry must name its projection, got %+v", entries[0])
			}
			if entries[0].EventType != "cqrsPoisonFact" {
				t.Fatalf("dead-letter entry names the wrong event type: %q", entries[0].EventType)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("poison event never reached the dead-letter store")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// After the dead-letter the checkpoint advances and the worker
	// catches back up — health green again (worker survives poison).
	healthDeadline := time.Now().Add(15 * time.Second)
	for {
		if err = svc.HealthCheck(ctx); err == nil {
			break
		}
		if time.Now().After(healthDeadline) {
			t.Fatalf("worker must survive a dead-lettered poison event: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestCqrsShutdownOrderInCascade pins the A12 ordering at the App
// level: App.Start boots the system's workers and the container
// cascade closes the service on Shutdown — cleanly, nil error.
func TestCqrsShutdownOrderInCascade(t *testing.T) {
	application, err := New(testConfig(t, false), discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	svc, err := do.InvokeNamed[*CqrsService](application.injector, cqrsProviderName)
	if err != nil {
		t.Fatalf("resolve cqrs: %v", err)
	}
	if err := application.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown with cqrs in the cascade: %v", err)
	}
	if !svc.isShutdown() {
		t.Fatal("the cascade must have run CqrsService.Shutdown")
	}
}
