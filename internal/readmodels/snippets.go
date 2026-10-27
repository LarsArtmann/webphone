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
)

// SnippetsMaxPerExtension bounds one extension's snippet list (the
// contacts-cap posture: the bound is the durability story, the UI
// renders the whole list).
const SnippetsMaxPerExtension = 100

// ErrSnippetListFull is returned when a NEW snippet would exceed the
// per-owner cap; replacing an existing id stays open. The sentinel
// homed here with the read-model swap (the SQL snippets store left
// with its file, the contract stays).
var ErrSnippetListFull = errors.New("snippet list full") //nolint:erraudit // sentinel: identity, not an error family (plan guardrail #4); classified via init registration

func init() {
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrSnippetListFull: errorfamily.Rejection,
	})
}

// SnippetView is the engine-side projection of a reply snippet.
// Unix-nanos created_at keeps list ordering EXACTLY numeric in sqlite
// (the SQL store compared unix seconds with a rowid tiebreak; nanos
// make ties virtually impossible and the order stable).
//
// branching-flow COMPOSITION_mixin (baselined): the engine projection
// row mirrors the retired SQL row verbatim (parity port).
// brandid-lint:ignore(BD003) engine projection row: the json tags are the storage schema (strong-id wire ruling)
type SnippetView struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Body      string `json:"body"`
	Quick     bool   `json:"quick"`
	CreatedAt int64  `json:"created_at"`
}

// snippetsCollection names the engine collection (query name).
const snippetsCollection = "snippets"

// brandid-lint:ignore(BD003) engine key-extraction input: wire shape (strong-id wire ruling)
type snippetInput struct{ ID string }

// snippetsQuery declares the snippets read model: a keyed projection
// upserted by SnippetSaved facts, removed by SnippetDeleted facts,
// filtered by owner and sorted oldest-first (the stable order the
// picker and the chips render). The insert fold is the single
// SnippetSaved fold on purpose — the replace-keeps-creation-time
// semantics live in the store seam, which pre-reads the existing view
// under its mutex before emitting.
var snippetsQuery = metaengine.Query[snippetInput, SnippetView](snippetsCollection,
	metaengine.OnRecord(domain.SnippetSaved{},
		func(_ record.Record, e domain.SnippetSaved) (domain.SnippetKey, SnippetView) {
			return e.Key, SnippetView{
				ID:        e.Key.ID,
				Owner:     e.Owner,
				Body:      e.Body,
				Quick:     e.Quick,
				CreatedAt: e.CreatedAt.UnixNano(),
			}
		}),
	metaengine.OnRecord(domain.SnippetDeleted{}, metaengine.Remove[SnippetView]()),
	metaengine.FilterOnField[SnippetView]("owner", metaengine.FilterEq),
	metaengine.SortOnField[SnippetView]("created_at", false),
	metaengine.Volume(humanScaleVolume),
)

// Snippets is the engine-backed reply-snippet store: the same contract
// the SQL store held (Save upsert by id keeping the creation time on
// replace; owner-scoped oldest-first List; owner-scoped Delete by id;
// per-owner cap for NEW ids).
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Contacts/Faxes) — deliberate symmetry,
// not a mixin candidate.
type Snippets struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[SnippetView]
	mu     sync.Mutex
}

func newSnippets(models *metaengine.Store) *Snippets {
	return &Snippets{
		models: models,
		reader: metaengine.NewReader[SnippetView](models, snippetsCollection),
	}
}

// Save upserts a snippet by id: an existing slot is replaced keeping
// its creation time (the SQL ON CONFLICT(id) DO UPDATE SET body, quick
// semantics); a NEW id past the per-owner cap is refused with this
// package's ErrSnippetListFull contract.
func (s *Snippets) Save(ctx context.Context, snippet domain.Snippet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := domain.SnippetKey{ID: snippet.ID.String()}

	prev, found, err := s.reader.Get(ctx, key)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.snippet_save", "load snippet slot")
	}
	if found {
		// Replace: the creation time stays with the slot.
		snippet.CreatedAt = time.Unix(0, prev.CreatedAt)
	} else if err := s.ownerHasRoom(ctx, snippet.Owner); err != nil {
		return err
	}

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.SnippetSaved{}), domain.SnippetSaved{
		Key:       key,
		Owner:     snippet.Owner.String(),
		Body:      snippet.Body,
		Quick:     snippet.Quick,
		CreatedAt: snippet.CreatedAt,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.snippet_save", "apply snippet save")
	}
	return nil
}

// ownerHasRoom rejects a NEW id when the owner is at the cap.
func (s *Snippets) ownerHasRoom(ctx context.Context, owner domain.Extension) error {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return err
	}
	if len(views) >= SnippetsMaxPerExtension {
		return errorfamily.WrapRejectionf(ErrSnippetListFull, "readmodels.snippet_list_full",
			"%s at %d snippets", owner.String(), SnippetsMaxPerExtension)
	}
	return nil
}

// List returns the owner's snippets, oldest first — a stable order for
// the picker and the chips.
func (s *Snippets) List(ctx context.Context, owner domain.Extension) ([]domain.Snippet, error) {
	views, err := s.ownerViews(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Snippet, 0, len(views))
	for _, v := range views {
		out = append(out, domain.Snippet{
			ID:        domain.MustSnippetID(v.ID),
			Owner:     domain.MustParseExtension(v.Owner),
			Body:      v.Body,
			Quick:     v.Quick,
			CreatedAt: time.Unix(0, v.CreatedAt),
		})
	}
	return out, nil
}

func (s *Snippets) ownerViews(ctx context.Context, owner domain.Extension) ([]SnippetView, error) {
	// WithLimit is load-bearing: an unlimited engine scan silently
	// ceilings at the planner's default (100 rows), and both the
	// user-facing list (the whole list renders) and the cap check need
	// every row the owner owns — the per-owner cap bounds this at 100.
	views, err := s.reader.Scan(ctx,
		metaengine.WithFilter("owner", metaengine.FilterEq, owner.String()),
		metaengine.WithSort("created_at", false),
		metaengine.WithLimit(SnippetsMaxPerExtension+1),
	)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "readmodels.snippet_list", "list snippets for %s", owner.String())
	}
	return views, nil
}

// Delete removes one snippet scoped to its owner. The id-to-slot
// resolution is a direct key lookup; a miss or a wrong-owner id
// answers the not-found sentinel (the SQL DELETE's owner scoping).
func (s *Snippets) Delete(ctx context.Context, owner domain.Extension, id domain.SnippetID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, found, err := s.reader.Get(ctx, domain.SnippetKey{ID: id.String()})
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.snippet_delete", "load snippet slot")
	}
	if !found || prev.Owner != owner.String() {
		return notFound("readmodels.snippet_missing", "snippet %s", id.String())
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.SnippetDeleted{}), domain.SnippetDeleted{
		Key: domain.SnippetKey{ID: id.String()},
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.snippet_delete", "apply snippet delete")
	}
	return nil
}
