package readmodels

import (
	"context"
	"sync"
	"time"

	metaengine "github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
)

// EnrollTokenView is the engine-side projection of one enrollment
// token row (the hash is the key; the raw secret never exists here).
type EnrollTokenView struct {
	TokenHash string `json:"token_hash"`
	Email     string `json:"email"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// enrollTokensCollection names the engine collection (query name).
const enrollTokensCollection = "enroll_tokens" //cqrs-lint:ignore(S001) a storage-collection NAME, not a credential — trips only the "token" substring heuristic

type enrollInput struct{ TokenHash string }

// enrollTokensQuery declares the enrollment-token read model: a keyed
// insert/burn projection — lookup and burn are direct key reads/writes,
// no owner scoping and no list shape exist (a token store is never
// enumerated; the sweep scans by expiry through the keyed folds).
var enrollTokensQuery = metaengine.Query[enrollInput, EnrollTokenView](enrollTokensCollection,
	metaengine.OnRecord(domain.EnrollTokenMinted{},
		func(_ record.Record, e domain.EnrollTokenMinted) (domain.EnrollTokenKey, EnrollTokenView) {
			return e.Key, EnrollTokenView{
				TokenHash: e.Key.TokenHash,
				Email:     e.Email,
				UserID:    e.UserID,
				ExpiresAt: e.ExpiresAt.Unix(),
			}
		}),
	metaengine.OnRecord(domain.EnrollTokenBurned{}, metaengine.Remove[EnrollTokenView]()),
)

// EnrollTokens is the engine-backed enrollment-token store: the same
// contract the SQL table held (hash-keyed mint/lookup/burn, expired
// rows swept opportunistically on every mint). Token classification
// (invalid vs expired vs unknown) stays in the userauth service — the
// seam only ever reports found/not-found.
//
// branching-flow COMPOSITION_mixin (baselined): the shared engine-seam
// shape (models+reader+mu, see Faxes/Contacts/Snippets) — deliberate
// symmetry, not a mixin candidate.
//
// NO legacy import: tokens live 15 minutes, so an upgrade carries no
// meaningful continuity — in-flight enrollment links at the moment of
// an upgrade simply get re-minted (the old wp_enroll_tokens table in
// usermgmt.db goes dead storage).
type EnrollTokens struct {
	models *metaengine.Store
	reader *metaengine.TypedReader[EnrollTokenView]
	mu     sync.Mutex
}

func newEnrollTokens(models *metaengine.Store) *EnrollTokens {
	return &EnrollTokens{
		models: models,
		reader: metaengine.NewReader[EnrollTokenView](models, enrollTokensCollection),
	}
}

// Mint stores one hashed token with its expiry and opportunistically
// sweeps the expired ones first (the SQL mint's preamble).
func (s *EnrollTokens) Mint(ctx context.Context, tokenHash, email, userID string, expiresAt time.Time) error { //nolint:branching-flow // verbatim port of the retired SQL store's Mint signature (parity rule); classification lives in userauth, not here
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sweepExpired(ctx, time.Now()); err != nil {
		return err
	}
	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.EnrollTokenMinted{}), domain.EnrollTokenMinted{
		Key:       domain.EnrollTokenKey{TokenHash: tokenHash},
		Email:     email,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.enroll_mint", "store enrollment token")
	}
	return nil
}

// Lookup resolves a token hash to its account and expiry. found=false
// covers unknown, burned, and swept rows alike — the caller owns the
// user-facing classification (including the expired-vs-invalid
// operator-log distinction, which needs the expiry).
func (s *EnrollTokens) Lookup(ctx context.Context, tokenHash string) (email, userID string, expiresAt time.Time, found bool, err error) {
	view, found, err := s.reader.Get(ctx, domain.EnrollTokenKey{TokenHash: tokenHash})
	if err != nil {
		return "", "", time.Time{}, false, errorfamily.WrapInfrastructuref(err, "readmodels.enroll_lookup", "look up enrollment token")
	}
	if !found {
		return "", "", time.Time{}, false, nil
	}
	return view.Email, view.UserID, time.Unix(view.ExpiresAt, 0), true, nil
}

// Burn removes one token (the consume-once arm). Removing an
// already-burned or unknown hash is inert — the consume outcome is the
// same.
func (s *EnrollTokens) Burn(ctx context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.EnrollTokenBurned{}), domain.EnrollTokenBurned{
		Key: domain.EnrollTokenKey{TokenHash: tokenHash},
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "readmodels.enroll_burn", "burn enrollment token")
	}
	return nil
}

// sweepPageSize bounds one expiry scan/expunge batch; the loop
// converges because every batch removes what it scanned.
const enrollSweepPageSize = 1000

func (s *EnrollTokens) sweepExpired(ctx context.Context, now time.Time) error {
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithLimit(enrollSweepPageSize),
		)
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "readmodels.enroll_sweep", "scan enrollment tokens")
		}
		if len(views) == 0 {
			return nil
		}
		removed := 0
		for _, v := range views {
			if time.Unix(v.ExpiresAt, 0).Before(now) {
				if err := s.models.Apply(ctx, metaengine.EventTypeName(domain.EnrollTokenBurned{}), domain.EnrollTokenBurned{
					Key: domain.EnrollTokenKey{TokenHash: v.TokenHash},
				}); err != nil {
					return errorfamily.WrapInfrastructuref(err, "readmodels.enroll_sweep", "sweep expired enrollment token")
				}
				removed++
			}
		}
		if removed == 0 {
			return nil
		}
	}
}
