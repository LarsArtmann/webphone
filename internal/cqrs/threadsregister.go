// Thread-aggregate registration and boot (SUPERB Phase 5): the
// decider/state/commands live in threads.go; this file owns the
// system-side wiring — decider + command registration, the D8
// snapshot strategy, the shared dispatch helper, and the public
// boot used by the composition root and tests.
package cqrs

import (
	"context"
	"fmt"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/snapshot/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"
)

// D8's snapshot strategy for the thread aggregate: a phone thread is
// hundreds of events, not millions, and the state carries the full id
// lists — snapshots bound the fold cost. The A38 perf data resolved the
// numbers: 200 bounds the tail replay of long-lived threads, and
// hot-read streams snapshot sooner via ReadPressure (threshold 50) — a
// thread dispatched into daily rides BOTH edges.
const (
	snapshotEveryNEvents          = 200
	snapshotReadPressureThreshold = 50
)

// threadSnapshotStrategy is the resolved D8 strategy: EveryNEvents OR
// read-pressure, whichever fires first (miswires — non-positive N or a
// bad threshold — panic loudly at boot).
var threadSnapshotStrategy = func() snapshot.SnapshotStrategy {
	every, err := snapshot.EveryNEvents(snapshotEveryNEvents)
	if err != nil {
		panic(fmt.Sprintf("cqrs: thread snapshot strategy: %v", err))
	}
	rp, err := snapshot.NewReadPressure(snapshotReadPressureThreshold,
		snapshot.WithInnerStrategy(every))
	if err != nil {
		panic(fmt.Sprintf("cqrs: thread snapshot strategy: %v", err))
	}
	return rp
}()

// RegisterThreads wires the thread aggregate into a System (D8: the
// conservative snapshot strategy bounds the id-list fold cost of
// long-lived threads).
func RegisterThreads(sys *system.System) error {
	if err := system.RegisterDecider[threadState](sys, ThreadsStreamType, decider.Decider[threadState]{
		Initial: threadState{}.initial(),
		Apply:   decider.StrictApply(applyThreadState, threadStreamEventTypes),
	}, system.WithSnapshotStrategy(threadSnapshotStrategy)); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread decider")
	}
	if err := registerThreadCommands(sys); err != nil {
		return err
	}
	return nil
}

func registerThreadCommands(sys *system.System) error {
	if err := system.RegisterCommand[*AppendMessage, threadState](sys, "thread.append",
		func(_ context.Context, cmd *AppendMessage) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideAppendMessage(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.append")
	}
	if err := system.RegisterCommand[*AdvanceDelivery, threadState](sys, "thread.advance_delivery",
		func(_ context.Context, cmd *AdvanceDelivery) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideAdvanceDelivery(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.advance_delivery")
	}
	if err := system.RegisterCommand[*SetThreadFlag, threadState](sys, "thread.set_flag",
		func(_ context.Context, cmd *SetThreadFlag) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideSetThreadFlag(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.set_flag")
	}
	if err := system.RegisterCommand[*MarkThreadRead, threadState](sys, "thread.mark_read",
		func(_ context.Context, cmd *MarkThreadRead) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideMarkThreadRead(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.mark_read")
	}
	if err := system.RegisterCommand[*DeleteThread, threadState](sys, "thread.delete",
		func(_ context.Context, cmd *DeleteThread) system.Op[threadState] {
			return system.Execute[threadState](context.Background(), cmd.StreamID(), ThreadsStreamType,
				func(state threadState, ver event.Version) ([]event.Event, error) {
					return decideDeleteThread(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.thread_register", "register thread.delete")
	}
	return nil
}

// dispatchThreadCommand dispatches a thread command keyed by the thread's
// stream key (the dispatchOwnerCommand twin for a non-owner key).
func dispatchThreadCommand[C command.Command](ctx context.Context, sys *system.System, name, key string,
	build func(*command.BasicCommand) C,
) error {
	streamID, err := id.ParseStreamID(key)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "cqrs.thread_stream", "parse thread stream key")
	}
	base, err := command.New(command.Type(name), streamID)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.command_build", "build %s", name)
	}
	return sys.CommandDispatcher().Dispatch(ctx, build(base))
}

// NewThreadsSystem boots a minimal system over the thread aggregate at
// dsn — the batteries' boot and the seam tests' composition helper
// (production composes through internal/app's named provider).
func NewThreadsSystem(ctx context.Context, dsn string) (*system.System, error) {
	decl := ThreadsDomainDecl()
	decl.ProjectionHostOptions = append(decl.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, decl, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.threads_boot", "boot thread system at %s", dsn)
	}
	if err := sys.Start(ctx); err != nil {
		_ = sys.Close() //cqrs-lint:ignore(C023) deliberate cleanup idiom: best-effort close; the primary error/drain owns reporting
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.threads_boot", "start thread system at %s", dsn)
	}
	return sys, nil
}

// NewCoreSystem boots one system over the retention-swept core
// (threads, faxes, transcripts, notes, reviews, contacts, snippets,
// dial) — the test compositions' helper: ONE journal per composition,
// every gate over it (a second system would split-brain pushes away
// from writes).
func NewCoreSystem(ctx context.Context, dsn string) (*system.System, error) {
	decl := Domain(ThreadsDomainDecl(), FaxesDomainDecl(), TranscriptsDomainDecl(), NotesDomainDecl(), ReviewsDomainDecl(), ContactsDomainDecl(), SnippetsDomainDecl(), DialDomainDecl(), PushDevicesDomainDecl(), AgentAuditDomainDecl())
	decl.ProjectionHostOptions = append(decl.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, decl, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.core_boot", "boot core system at %s", dsn)
	}
	if err := sys.Start(ctx); err != nil {
		_ = sys.Close() //cqrs-lint:ignore(C023) deliberate cleanup idiom: best-effort close; the primary error/drain owns reporting
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.core_boot", "start core system at %s", dsn)
	}
	return sys, nil
}
