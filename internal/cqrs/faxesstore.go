// Fax journal backend (SUPERB Phase 6): the projections in system.db
// and the gate the server keeps calling — fax.Store's write contract
// (Create/UpdateStatus/AdvanceIdempotent) and the FaxGate read/retention
// contract, both answered by dispatch + projection scans. Idempotency
// rides sys.Dedup() (system.db's durable CAS window): the webhook
// lane's replay guard outlives the process, superseding the engine's
// readmodels.db-scoped ApplyIdempotent.
package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/store"
)

// systemFaxesCollection names the fax-jobs projection in system.db
// (readmodels' own "faxes" collection is the frozen import lane in its
// own database — same folds, different home).
const systemFaxesCollection = "faxes"

// faxDedupCollection names the durable idempotency window's keyspace.
const faxDedupCollection = "fax_replay"

// faxDedupTTL is the replay window's life (the hooks seam's 1h TTL
// twin — a replay after the window re-dispatches and the decider's
// same-status refresh keeps it convergent anyway).
const faxDedupTTL = time.Hour

// Faxes is the dispatch-backed fax store over the system.
type Faxes struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.FaxView]
	dedup  metaengine.DedupStore
	vers   streamVersions
	mu     sync.Mutex
}

var (
	_ fax.Store          = (*Faxes)(nil)
	_ readmodels.FaxGate = (*Faxes)(nil)
)

// NewFaxes builds the backend over a System (the dedup capability is
// asserted here — a future engine swap that drops it fails at boot,
// not at the first replay).
func NewFaxes(sys *system.System) *Faxes {
	dedup, err := sys.Dedup()
	if err != nil {
		panic(fmt.Sprintf("cqrs: fax backend needs the system dedup capability: %v", err))
	}
	return &Faxes{
		sys:    sys,
		reader: metaengine.NewReader[readmodels.FaxView](sys.MetaEngine(), systemFaxesCollection),
		dedup:  dedup,
		vers:   streamVersions{},
	}
}

// dispatchFax dispatches one fax command keyed by the job's stream key.
func dispatchFax[C command.Command](ctx context.Context, sys *system.System, name string, faxID domain.FaxID,
	build func(*command.BasicCommand) C,
) error {
	streamID, err := id.ParseStreamID(faxID.String())
	if err != nil {
		return errorfamily.WrapRejectionf(err, "cqrs.fax_stream", "parse fax stream key %s", faxID)
	}
	base, err := command.New(command.Type(name), streamID)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.command_build", "build %s", name)
	}
	return sys.CommandDispatcher().Dispatch(ctx, build(base))
}

func (s *Faxes) awaitOwnWrite(ctx context.Context, id domain.FaxID) {
	awaitOwnWrite(ctx, s.sys, FaxesStreamType, FaxStreamKey(id), systemFaxesCollection, s.vers)
}

// Create records a fax job's birth row (outbound queued or inbound
// received) and awaits its projection visibility — the service's
// send/receive handlers re-render in the same request.
func (s *Faxes) Create(ctx context.Context, job domain.FaxJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := dispatchFax(ctx, s.sys, "fax.record", job.ID,
		func(base *command.BasicCommand) *RecordFax {
			return &RecordFax{BasicCommand: base, Job: job}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, job.ID)
	return nil
}

// UpdateStatus moves one job's lifecycle. Illegal moves (the inbound
// no-lifecycle rule, off-lane targets) are inert zero-event dispatches
// — the settled state answers, exactly the legacy contract.
func (s *Faxes) UpdateStatus(ctx context.Context, id domain.FaxID, delivery domain.FaxDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := dispatchFax(ctx, s.sys, "fax.advance_status", id,
		func(base *command.BasicCommand) *AdvanceFaxStatus {
			return &AdvanceFaxStatus{BasicCommand: base, FaxID: id, Delivery: delivery, At: time.Now()}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, id)
	return nil
}

// AdvanceIdempotent is the webhook lane's durable replay guard: the
// idem key's first sighting records a window and dispatches; a replay
// inside the window is a no-op. A crash between dispatch and record
// re-dispatches on retry — the decider's same-status refresh keeps
// that convergent (one redundant fact, same row image).
func (s *Faxes) AdvanceIdempotent(ctx context.Context, idemKey string, id domain.FaxID, delivery domain.FaxDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen, err := s.dedup.DedupCheckAndRecord(ctx, faxDedupCollection, idemKey, faxDedupTTL, time.Now())
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.fax_idem", "fax replay window for %s", idemKey)
	}
	if seen {
		return nil
	}
	if err := dispatchFax(ctx, s.sys, "fax.advance_status", id,
		func(base *command.BasicCommand) *AdvanceFaxStatus {
			return &AdvanceFaxStatus{BasicCommand: base, FaxID: id, Delivery: delivery, At: time.Now()}
		}); err != nil {
		return err
	}
	s.awaitOwnWrite(ctx, id)
	return nil
}

// List returns the owner's fax jobs, newest first, bounded by limit.
func (s *Faxes) List(ctx context.Context, owner domain.Extension, limit int) ([]domain.FaxJob, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(limit))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_list", "list faxes for %s", owner)
	}
	out := make([]domain.FaxJob, 0, len(views))
	for _, v := range views {
		out = append(out, readmodels.FaxJobOfView(v))
	}
	return out, nil
}

// Get returns one job scoped to the owner.
func (s *Faxes) Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error) {
	job, err := s.GetAny(ctx, id)
	if err != nil {
		return domain.FaxJob{}, err
	}
	if job.Owner != owner {
		return domain.FaxJob{}, store.ErrNotFound
	}
	return job, nil
}

// ListAll returns every owner's jobs, newest first, bounded by limit
// (the admin inventory).
func (s *Faxes) ListAll(ctx context.Context, limit int) ([]domain.FaxJob, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(limit))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_list_all", "list all faxes")
	}
	out := make([]domain.FaxJob, 0, len(views))
	for _, v := range views {
		out = append(out, readmodels.FaxJobOfView(v))
	}
	return out, nil
}

// GetAny fetches one job by id across owners (the admin seam).
func (s *Faxes) GetAny(ctx context.Context, id domain.FaxID) (domain.FaxJob, error) {
	view, found, err := s.reader.Get(ctx, domain.FaxKey{ID: id.String()})
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "cqrs.fax_get", "get fax %s", id)
	}
	if !found {
		return domain.FaxJob{}, store.ErrNotFound
	}
	return readmodels.FaxJobOfView(view), nil
}

// ByProviderRef resolves one job by its gateway correlation id (the
// webhook lane's lookup).
func (s *Faxes) ByProviderRef(ctx context.Context, ref string) (domain.FaxJob, error) {
	if ref == "" {
		return domain.FaxJob{}, store.ErrNotFound
	}
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("provider_ref", metaengine.FilterEq, ref),
		metaengine.WithLimit(2))
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "cqrs.fax_by_ref", "resolve fax ref %s", ref)
	}
	if len(views) == 0 {
		return domain.FaxJob{}, store.ErrNotFound
	}
	return readmodels.FaxJobOfView(views[0]), nil
}

// Count reports the total stored fax jobs via the engine's native
// COUNT (the metrics surface's webphone_faxes_total).
func (s *Faxes) Count(ctx context.Context) (int64, error) {
	n, err := engineCount(ctx, s.sys, systemFaxesCollection)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "cqrs.fax_count", "count faxes")
	}
	return n, nil
}

// SweepBefore is the retention leg: every job older than the cutoff
// tombstones out of the views and its journal stream trims to zero
// bytes; the document blob paths return for the caller to unlink.
func (s *Faxes) SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error) {
	views, err := s.reader.Scan(ctx,
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(0))
	if err != nil {
		return nil, 0, errorfamily.WrapInfrastructuref(err, "cqrs.fax_sweep_scan", "sweep scan faxes")
	}
	cutoffNanos := cutoff.UnixNano()
	var paths []string
	var removed int64
	for _, v := range views {
		if v.CreatedAt >= cutoffNanos {
			continue
		}
		id := domain.MustFaxID(v.ID)
		if err := dispatchFax(ctx, s.sys, "fax.expunge", id,
			func(base *command.BasicCommand) *ExpungeFax {
				return &ExpungeFax{BasicCommand: base, FaxID: id}
			}); err != nil {
			return paths, removed, err
		}
		s.awaitOwnWrite(ctx, id)
		ref, err := streamRef(FaxesStreamType, FaxStreamKey(id))
		if err != nil {
			return paths, removed, err
		}
		if _, err := s.sys.TrimStream(ctx, ref); err != nil {
			return paths, removed, errorfamily.WrapInfrastructuref(err, "cqrs.fax_trim", "trim journal of fax %s", id)
		}
		if v.DocumentPath != "" {
			paths = append(paths, v.DocumentPath)
		}
		removed++
	}
	return paths, removed, nil
}

// FaxesDomainDecl is the fax aggregate's share of the system's
// DomainConfig.
func FaxesDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterFaxes(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register fax aggregate: %v", err)) //cqrs-lint:ignore(C009) boot-time fail-fast on a composition miswire (DO-1; the decl comment owns the why)
			}
		},
		Projections: []system.ProjectionDeclaration{
			system.RawQuery(readmodels.FaxesProjection(systemFaxesCollection)),
		},
		Events: faxStreamEventTypes,
	}
}

// NewFaxesSystem boots a minimal system over the fax aggregate at dsn —
// the battery's boot helper (production composes through internal/app's
// named provider).
func NewFaxesSystem(ctx context.Context, dsn string) (*system.System, error) {
	decl := FaxesDomainDecl()
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
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_boot", "boot fax system at %s", dsn)
	}
	if err := sys.Start(ctx); err != nil {
		_ = sys.Close() //cqrs-lint:ignore(C023) deliberate cleanup idiom: best-effort close; the primary error/drain owns reporting
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.fax_boot", "start fax system at %s", dsn)
	}
	return sys, nil
}
