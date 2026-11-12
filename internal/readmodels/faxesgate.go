package readmodels

import (
	"context"
	"time"

	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"

	"github.com/larsartmann/webphone/internal/domain"
)

// humanScaleVolume is the Volume hint every query declaration carries:
// per-owner row counts stay in the hundreds/thousands (one human's
// lifetime of threads/faxes), never the millions system tuning
// assumes.
const humanScaleVolume = 1

// FaxView is the fax projection row both backends fold into (the
// journal-backed system.db projection reads/writes this exact shape;
// Unix-nanos timestamps keep list ordering EXACTLY numeric — the
// retired SQL store compared unix seconds with a rowid tiebreak, and
// nanos make ties virtually impossible and the order stable).
//
// branching-flow COMPOSITION_mixin (baselined): the projection row
// mirrors the retired SQL row verbatim (parity port).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
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

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type faxInput struct{ ID string }

// FaxGate is the fax seam (SUPERB Phase 6): every Faxes method the
// server, the notifier and the metrics surface call, answered by the
// journal-backed backend. ErrNotFound keeps its store meaning.
type FaxGate interface {
	List(ctx context.Context, owner domain.Extension, limit int) ([]domain.FaxJob, error)
	Get(ctx context.Context, owner domain.Extension, id domain.FaxID) (domain.FaxJob, error)
	ListAll(ctx context.Context, limit int) ([]domain.FaxJob, error)
	GetAny(ctx context.Context, id domain.FaxID) (domain.FaxJob, error)
	ByProviderRef(ctx context.Context, ref string) (domain.FaxJob, error)
	// Count reports the total stored fax jobs across all owners (the
	// metrics surface's webphone_faxes_total).
	Count(ctx context.Context) (int64, error)
	// SweepBefore is the retention leg: expired jobs tombstone out of
	// the views AND their journal streams trim to zero bytes; the
	// document blob paths return for the caller to unlink.
	SweepBefore(ctx context.Context, cutoff time.Time) ([]string, int64, error)
}

// FaxJobOfView converts one projection row into the domain job (the
// ONE view→domain home).
func FaxJobOfView(v FaxView) domain.FaxJob {
	return domain.FaxJob{
		ID:           domain.MustFaxID(v.ID),
		Owner:        domain.MustParseExtension(v.Owner),
		Remote:       domain.MustParsePhone(v.Remote),
		Direction:    v.Direction,
		Status:       v.Status,
		Pages:        v.Pages,
		DocumentPath: domain.RelPathFromRow(v.DocumentPath),
		ProviderRef:  v.ProviderRef,
		Error:        v.Error,
		CreatedAt:    time.Unix(0, v.CreatedAt),
		UpdatedAt:    time.Unix(0, v.UpdatedAt),
	}
}

// FaxesProjection builds the fax-jobs read model over the GIVEN
// collection — the folds' one home, consumed verbatim by the cqrs
// system's journal projection (system.db). Inserted by FaxReceipted
// facts, merged by FaxStatusAdvanced facts (the provider-ref
// keep-when-empty rule lives here, once), removed by retention's
// FaxExpunged facts.
func FaxesProjection(collection string) metaengine.QueryDecl[faxInput, FaxView] {
	return metaengine.Query[faxInput, FaxView](collection,
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
		metaengine.Volume(humanScaleVolume),
	)
}
