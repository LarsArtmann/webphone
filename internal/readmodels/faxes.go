package readmodels

import (
	"context"
	"errors"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/store"
)

// FaxView is the engine-side projection of a fax job row. Unix-nanos
// timestamps keep list ordering EXACTLY numeric in sqlite (the SQL
// store compared unix seconds with a rowid tiebreak; nanos make ties
// virtually impossible and the order stable).
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the retired SQL row verbatim (parity port).
//brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type FaxView struct {
	ID           string              `json:"id"`
	Owner        string              `json:"owner"`
	Remote       string              `json:"remote"`
	Direction    domain.FaxDirection `json:"direction"`
	Status       domain.FaxStatus    `json:"status"`
	Pages        int                 `json:"pages"`
	DocumentPath string              `json:"document_path"`
	ProviderRef  string              `json:"provider_ref"`
	Error        string              `json:"error"`
	CreatedAt    int64               `json:"created_at"`
	UpdatedAt    int64               `json:"updated_at"`
}

// faxesCollection names the engine collection (query name).
const faxesCollection = "faxes"

//brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type faxInput struct{ ID string }

// faxesQuery declares the faxes read model: a keyed projection
// inserted by FaxReceipted facts, merged by FaxStatusAdvanced facts
// (the update fold receives the previous view, so the provider-ref
// keep-when-empty rule lives in ONE place), and removed by retention's
// FaxExpunged facts. Filtered by owner, provider ref (the webhook's
// lookup path), and created-at (retention), sorted newest-first.
var faxesQuery = metaengine.Query[faxInput, FaxView](faxesCollection,
	metaengine.OnRecord(domain.FaxReceipted{},
		func(_ record.Record, e domain.FaxReceipted) (domain.FaxKey, FaxView) {
			return e.Key, FaxView{
				ID:           e.Key.ID,
				Owner:        e.Owner,
				Remote:       e.Remote,
				Direction:    e.Direction,
				Status:       e.Status,
				Pages:        e.Pages,
				DocumentPath: e.DocumentPath,
				ProviderRef:  e.ProviderRef,
				Error:        e.Error,
				CreatedAt:    e.CreatedAt.UnixNano(),
				UpdatedAt:    e.UpdatedAt.UnixNano(),
			}
		}),
	metaengine.OnRecord(domain.FaxStatusAdvanced{},
		func(_ record.Record, e domain.FaxStatusAdvanced, prev FaxView) FaxView {
			prev.Status = e.Status
			if e.ProviderRef != "" {
				prev.ProviderRef = e.ProviderRef
			}
			prev.Error = e.Error
			prev.Pages = e.Pages
			prev.UpdatedAt = e.AdvancedAt.UnixNano()
			return prev
		}),
	metaengine.OnRecord(domain.FaxExpunged{}, metaengine.Remove[FaxView]()),
	metaengine.FilterOnField[FaxView]("owner", metaengine.FilterEq),
	metaengine.FilterOnField[FaxView]("provider_ref", metaengine.FilterEq),
	metaengine.FilterOnField[FaxView]("created_at", metaengine.FilterLt),
	metaengine.SortOnField[FaxView]("created_at", true),
)

// Faxes is the engine-backed fax store: the same contract the SQL
// store held (Create inserts; UpdateStatus advances with the
// provider-ref keep-when-empty rule and answers ErrNotFound on a
// missing row; List is owner-scoped newest-first with an explicit
// limit; Get is owner-scoped; ByProviderRef is the webhook's global
// lookup). Fax rows are never deleted.
//
// mu serializes the pre-read/apply sequences the same way the Contacts
// and Snippets seams do: the binary is single-process by doctrine.
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+agg+mu, see Contacts/Snippets) — deliberate
// symmetry, not a mixin candidate.
type Faxes struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[FaxView]
	agg    metaengine.AggregateReader
	mu     sync.Mutex
}

func newFaxes(models *metaengine.Store, agg metaengine.AggregateReader) *Faxes {
	return &Faxes{
		models: models,
		reader: metaengine.NewReader[FaxView](models, faxesCollection),
		agg:    agg,
	}
}

// Create inserts a new fax job row.
func (s *Faxes) Create(ctx context.Context, job domain.FaxJob) error {
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.FaxReceipted{}), domain.FaxReceipted{
		Key:          domain.FaxKey{ID: job.ID.String()},
		Owner:        job.Owner.String(),
		Remote:       job.Remote.String(),
		Direction:    job.Direction,
		Status:       job.Status,
		Pages:        job.Pages,
		DocumentPath: job.DocumentPath,
		ProviderRef:  job.ProviderRef,
		Error:        job.Error,
		CreatedAt:    job.CreatedAt,
		UpdatedAt:    job.UpdatedAt,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.fax_insert", "insert fax job")
	}
	return nil
}

// UpdateStatus advances a job's status; providerRef updates the
// correlation id when non-empty. errorText is stored verbatim on
// failure. A missing row answers the not-found sentinel (the SQL
// UPDATE's updatedOrNotFound arm).
//
// branching-flow provider correlation id + verbatim refusal text; a ProviderRef type would ripple into webhook DTOs and golden-pinned rows (triage 2026-10-11)
func (s *Faxes) UpdateStatus(
	ctx context.Context, id domain.FaxID, status domain.FaxStatus, providerRef, errorText string, pages int, //nolint:branching-flow // swap-safe verdict strings; a ProviderRef brand ripples into webhook DTOs + golden-pinned rows (triage 2026-10-11)
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found, err := s.reader.Get(ctx, domain.FaxKey{ID: id.String()}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.fax_update", "load fax job")
	} else if !found {
		return errors.Join( //nolint:erraudit // Join keeps the sentinel errors.Is-matchable; family via ErrNotFound registration (same shape the SQL store used)
			store.ErrNotFound,
			errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "fax %s", id.String()),
		)
	}

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.FaxStatusAdvanced{}), domain.FaxStatusAdvanced{
		Key:         domain.FaxKey{ID: id.String()},
		Status:      status,
		ProviderRef: providerRef,
		Error:       errorText,
		Pages:       pages,
		AdvancedAt:  time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.fax_update", "apply fax status advance")
	}
	return nil
}

// AdvanceIdempotent is UpdateStatus riding the engine's in-process
// idempotent apply: a repeated idemKey is a no-op. The REPLAY lane of
// the status webhook uses it so a double dispatch inside this process
// cannot double-write; DURABLE replay protection stays where it has
// been since the dedup swap — the hooks seam's engine-backed dedup
// window (ApplyIdempotent's dedup is process-local by design).
func (s *Faxes) AdvanceIdempotent(
	ctx context.Context, idemKey string, id domain.FaxID, status domain.FaxStatus, providerRef, errorText string, pages int, //nolint:branching-flow // swap-safe verdict strings; a ProviderRef brand ripples into webhook DTOs + golden-pinned rows (triage 2026-10-11)
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found, err := s.reader.Get(ctx, domain.FaxKey{ID: id.String()}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.fax_update", "load fax job")
	} else if !found {
		return errors.Join( //nolint:erraudit // Join keeps the sentinel errors.Is-matchable; family via ErrNotFound registration (same shape the SQL store used)
			store.ErrNotFound,
			errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "fax %s", id.String()),
		)
	}

	if err := s.models.ApplyIdempotent(ctx, idemKey, metaengine.EventTypeName(domain.FaxStatusAdvanced{}), domain.FaxStatusAdvanced{
		Key:         domain.FaxKey{ID: id.String()},
		Status:      status,
		ProviderRef: providerRef,
		Error:       errorText,
		Pages:       pages,
		AdvancedAt:  time.Now(),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.fax_update", "apply fax status advance")
	}
	return nil
}

// List returns the owner's fax jobs, newest first, bounded by limit
// (the fax list's page size and the export's larger window both ride
// this seam).
func (s *Faxes) List(ctx context.Context, owner domain.Extension, limit int) ([]domain.FaxJob, error) {
	// WithLimit is load-bearing: an unlimited engine scan silently
	// ceilings at the planner's default (100 rows); the caller's limit
	// is the page shape the SQL store kept.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", true),
		metaengine.WithLimit(limit),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.fax_list", "list faxes for %s", owner.String())
	}
	out := make([]domain.FaxJob, 0, len(views))
	for _, v := range views {
		out = append(out, v.toJob())
	}
	return out, nil
}

// Get fetches one job scoped to its owner.
func (s *Faxes) Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error) {
	view, found, err := s.reader.Get(ctx, domain.FaxKey{ID: id.String()})
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "readmodels.fax_get", "load fax job")
	}
	if !found || view.Owner != owner.String() {
		return domain.FaxJob{}, errors.Join( //nolint:erraudit // Join keeps the sentinel errors.Is-matchable; family via ErrNotFound registration
			store.ErrNotFound,
			errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "fax %s", id.String()),
		)
	}
	return view.toJob(), nil
}

// ByProviderRef resolves a job by its gateway correlation id — the
// status webhook's lookup path. An empty ref answers not-found (the
// SQL store's guard).
func (s *Faxes) ByProviderRef(ctx context.Context, ref string) (domain.FaxJob, error) {
	if ref == "" {
		return domain.FaxJob{}, errors.Join( //nolint:erraudit // Join keeps the sentinel errors.Is-matchable; family via ErrNotFound registration
			store.ErrNotFound,
			errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "empty provider ref"),
		)
	}
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("provider_ref", metaengine.FilterEq, ref),
		metaengine.WithLimit(2),
	)
	if err != nil {
		return domain.FaxJob{}, errorfamily.WrapInfrastructuref(err, "readmodels.fax_by_ref", "resolve fax by provider ref")
	}
	if len(views) == 0 {
		return domain.FaxJob{}, errors.Join( //nolint:erraudit // Join keeps the sentinel errors.Is-matchable; family via ErrNotFound registration
			store.ErrNotFound,
			errorfamily.Newf(errorfamily.Rejection, "readmodels.fax_missing", "provider ref %s", ref),
		)
	}
	return views[0].toJob(), nil
}

func (v FaxView) toJob() domain.FaxJob {
	return domain.FaxJob{
		ID:           domain.MustFaxID(v.ID),
		Owner:        domain.MustParseExtension(v.Owner),
		Remote:       domain.MustParsePhone(v.Remote),
		Direction:    v.Direction,
		Status:       v.Status,
		Pages:        v.Pages,
		DocumentPath: v.DocumentPath,
		ProviderRef:  v.ProviderRef,
		Error:        v.Error,
		CreatedAt:    time.Unix(0, v.CreatedAt),
		UpdatedAt:    time.Unix(0, v.UpdatedAt),
	}
}

// Count reports the total stored fax jobs across all owners (the
// metrics surface's webphone_faxes_total). Served by the engine's
// native COUNT aggregate — no row materialization (the SQL fax_jobs
// table is retired storage after the import window).
func (s *Faxes) Count(ctx context.Context) (int64, error) {
	n, err := s.agg.Aggregate(ctx, faxesCollection, metaengine.AggregateCount, "", nil)
	if err != nil {
		return 0, errorfamily.WrapInfrastructuref(err, "readmodels.fax_count", "count faxes")
	}
	return int64(n), nil
}

// sweepPageSize bounds one retention scan/expunge batch; the loop
// converges because every batch removes what it scanned.
const sweepPageSize = 1000

// SweepBefore deletes fax jobs created before the cutoff (the
// retention pass's engine arm) and returns their spooled document
// paths for the CALLER to unlink — the same contract the SQL sweep
// kept (paths collected before the rows go; a crash between expunge
// and unlink leaves an orphaned blob, bounded and documented).
func (s *Faxes) SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var paths []string
	var removed int64
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithFilter("created_at", metaengine.FilterLt, cutoff.UnixNano()),
			metaengine.WithSort("created_at", false),
			metaengine.WithLimit(sweepPageSize),
		)
		if err != nil {
			return paths, removed, errorfamily.WrapInfrastructuref(err, "readmodels.fax_sweep", "scan expirable faxes")
		}
		if len(views) == 0 {
			return paths, removed, nil
		}
		for _, v := range views {
			if v.DocumentPath != "" {
				paths = append(paths, v.DocumentPath)
			}
			if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.FaxExpunged{}), domain.FaxExpunged{
				Key: domain.FaxKey{ID: v.ID},
			}); err != nil {
				return paths, removed, errorfamily.WrapInfrastructuref(err, "readmodels.fax_sweep", "expunge fax %s", v.ID)
			}
			removed++
		}
	}
}
