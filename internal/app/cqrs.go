// CqrsService adapts the go-cqrs-lite system composition root to the
// container's lifecycle contract (SUPERB plan Phase 1; the composition
// contract is docs/planning/2026-11-08_cqrs-composition-contract.md).
// system.New does NOT replace this package's injector — it runs INSIDE
// it as ONE named eager service, exactly like "sqlite" and "metaengine":
// the container keeps owning object lifetime, the System keeps owning
// its internal engines, and system.db stays engine-owned (the two-file
// spirit: its DDL never meets webphone's hand ladder).
package app

import (
	"context"
	"database/sql"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
	_ "modernc.org/sqlite" // registers the "sqlite" driver for the DLQ handle

	"github.com/larsartmann/webphone/internal/config"
	cqrsdomain "github.com/larsartmann/webphone/internal/cqrs"
)

// cqrsGracefulCloseBound caps GracefulClose's drain at shutdown: the
// HTTP drain has already waited out in-flight dispatches, so this bound
// only guards a wedged projection worker from stretching restarts.
const cqrsGracefulCloseBound = 10 * time.Second

// cqrsWorkerLagBound is the projection-lag ceiling the health leg
// tolerates before failing /healthz naming the worker (contract §6:
// start at 30s, tighten with evidence once real projections ride).
const cqrsWorkerLagBound = 30 * time.Second

// cqrsDLQThreshold is the per-event failure count before a poison
// event is dead-lettered and the checkpoint advances past it.
const cqrsDLQThreshold = 3

// CqrsService adapts *system.System to the container's lifecycle
// contract. Shutdown runs GracefulClose (drain, then close) and closes
// the DLQ's own database handle; both are idempotent at this seam.
type CqrsService struct {
	Sys *system.System
	// DLQ is the poison-event store wired into the projection host;
	// readable for the operator surface later phases own (A12+).
	DLQ projectionhost.DeadLetterStore

	dlqDB *sql.DB
	log   *slog.Logger

	startMu      sync.Mutex
	started      bool
	shutdown     bool
	shutdownM    sync.Mutex
	dialsOnce    sync.Once
	dials        *cqrsdomain.DialWindows
	snippetsOnce sync.Once
	snippets     *cqrsdomain.Snippets
	enrollOnce   sync.Once
	enroll       *cqrsdomain.EnrollTokens
}

// NewCqrs builds the system composition root over <dataDir>/system.db:
// ONE sqlite engine, TWO instances on it (source-of-truth + projections
// — journal, projections, dedup, checkpoints, snapshots and DLQ all in
// the one file, the spike-proven topology), WAL pragmas matching
// webphone's sqlite posture. The domain carries the dial aggregate
// since Phase 2 (SUPERB): commands, decider, and the journal
// projections the gate reads.
func NewCqrs(cfg config.Config, log *slog.Logger) (*CqrsService, error) {
	return newCqrs(cfg, log, cqrsdomain.Domain(
		cqrsdomain.DialDomainDecl(),
		cqrsdomain.SnippetsDomainDecl(),
		cqrsdomain.EnrollDomainDecl(),
	))
}

// Dials builds (once) the dispatch-backed dial-gate service — the SAME
// surface the engine-backed store exposed, now answering from the
// journal's decisions.
func (s *CqrsService) Dials() *cqrsdomain.DialWindows {
	s.dialsOnce.Do(func() { s.dials = cqrsdomain.NewDialWindows(s.Sys) })
	return s.dials
}

// SnippetsSvc builds (once) the dispatch-backed snippet service.
func (s *CqrsService) SnippetsSvc() *cqrsdomain.Snippets {
	s.snippetsOnce.Do(func() { s.snippets = cqrsdomain.NewSnippets(s.Sys) })
	return s.snippets
}

// EnrollSvc builds (once) the dispatch-backed enrollment-token store
// — the seam the userauth door consumes (SUPERB Phase 3, A22).
func (s *CqrsService) EnrollSvc() *cqrsdomain.EnrollTokens {
	s.enrollOnce.Do(func() { s.enroll = cqrsdomain.NewEnrollTokens(s.Sys) })
	return s.enroll
}

// newCqrs is the construction seam tests (and, from Phase 2 on, the
// real domain registration) ride: same deployment shape, swappable
// domain config.
func newCqrs(cfg config.Config, log *slog.Logger, domain system.DomainConfig) (*CqrsService, error) {
	ctx := context.Background()
	dsn := filepath.Join(cfg.DataDir, "system.db")

	// The DLQ store needs a *sql.DB the System never exposes (engines
	// own their handles), so it rides a second modernc connection to
	// the SAME file — WAL absorbs the second writer and the table is
	// projectionhost's own DDL, not webphone SQL. Opened BEFORE
	// system.New because ProjectionHostOptions are consumed during
	// construction; modernc creates the file if the System has not yet.
	dlqDB, err := sql.Open("sqlite", dsn+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.dlq", "open dead-letter handle on %s", dsn)
	}
	dlq, err := projectionhost.NewSQLiteDeadLetterStore(ctx, dlqDB)
	if err != nil {
		_ = dlqDB.Close()
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.dlq", "bootstrap dead-letter store")
	}

	deployment := system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {
				Driver:  "sqlite",
				DSN:     dsn,
				Pragmas: []string{"wal"},
			},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	}

	domain.ProjectionHostOptions = append(domain.ProjectionHostOptions,
		projectionhost.WithDeadLetterStore(dlq, cqrsDLQThreshold),
		projectionhost.WithLogger(log),
	)

	sys, err := system.New(ctx, domain, deployment)
	if err != nil {
		_ = dlqDB.Close()
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.open", "compose system over %s", dsn)
	}
	if domain.Commands == nil && len(domain.Projections) == 0 {
		log.Warn("cqrs hollow root", "reason", "phase 1 plumbing — first aggregate lands in phase 2")
	}
	return &CqrsService{Sys: sys, DLQ: dlq, dlqDB: dlqDB, log: log}, nil
}

// Start boots the projection workers and managed timers, ONCE:
// system.Start is once-only (ErrAlreadyStarted), so the adapter latches
// it — a second App.Start must not red-tip the System.
func (s *CqrsService) Start(ctx context.Context) error {
	s.startMu.Lock()
	defer s.startMu.Unlock()
	if s.started {
		return nil
	}
	if err := s.Sys.Start(ctx); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.start", "start system workers")
	}
	s.started = true
	return nil
}

// Shutdown drains the System (GracefulClose under a 10s bound: workers
// first, then engines) and closes the DLQ handle. Idempotent at this
// seam; the container cascade owns when it runs (App.Shutdown →
// injector.Shutdown), always AFTER the HTTP drain.
func (s *CqrsService) Shutdown() error {
	s.shutdownM.Lock()
	defer s.shutdownM.Unlock()
	if s.shutdown {
		return nil
	}
	s.shutdown = true

	ctx, cancel := context.WithTimeout(context.Background(), cqrsGracefulCloseBound)
	defer cancel()
	err := s.Sys.GracefulClose(ctx)
	if closeErr := s.dlqDB.Close(); closeErr != nil {
		s.log.Warn("cqrs dead-letter handle close failed", "error", closeErr)
	}
	return err
}

// HealthCheck is the optional non-critical probe leg: engine health
// plus worker triage. A failed worker or a projection lag beyond the
// bound fails the leg NAMING the worker — the same shape the
// cqrs-htmx projections leg uses on /healthz.
func (s *CqrsService) HealthCheck(ctx context.Context) error {
	if err := s.Sys.HealthCheck(ctx); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.engine", "system engine health")
	}
	return cqrsWorkerTriage(s.Sys.WorkerStatus())
}

// cqrsWorkerTriage fails on a dead or too-far-behind projection worker,
// naming the worker in the message — the operator's way from a 503 body
// to the exact wedged component.
func cqrsWorkerTriage(workers []projectionhost.WorkerState) error {
	for _, w := range workers {
		if w.Status == projectionhost.WorkerFailed {
			return errorfamily.Newf(errorfamily.Infrastructure, "cqrs.worker_failed",
				"projection worker %q failed (last error: %s)", w.Name, w.LastError)
		}
		if w.Lag > cqrsWorkerLagBound {
			return errorfamily.Newf(errorfamily.Infrastructure, "cqrs.worker_lag",
				"projection worker %q lag %s beyond the %s bound", w.Name, w.Lag, cqrsWorkerLagBound)
		}
	}
	return nil
}

// isShutdown reports whether the shutdown latch has fired (test
// observability into the cascade's reach).
func (s *CqrsService) isShutdown() bool {
	s.shutdownM.Lock()
	defer s.shutdownM.Unlock()
	return s.shutdown
}

// cqrsProviderName is the injector name the probe discovers health by
// (contract §2: named "cqrs", eager, fail-fast).
const cqrsProviderName = "cqrs"

// registerCqrs wires the named provider; called from New so the
// registration sits beside its siblings in the composition root.
func registerCqrs(injector do.Injector, cfg config.Config, log *slog.Logger) {
	//samber-linter:allow hw-4 eagerly resolved via InvokeNamed in New
	do.ProvideNamed(injector, cqrsProviderName, func(do.Injector) (*CqrsService, error) {
		return NewCqrs(cfg, log)
	})
}
