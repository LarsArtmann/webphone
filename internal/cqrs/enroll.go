// The enroll-token aggregate (SUPERB Phase 3, A22): one stream per
// token hash (the SHA-256 of the URL secret — a hash is a valid stream
// id; there is no owner axis here by construction). The consume-once
// burn answers from the DECISION (unknown/already-burned is an inert
// no-op, exactly the retired engine store's contract), expiry
// classification stays at the userauth door (Lookup exposes the
// expiry), and the mint-side opportunistic sweep survives as burn
// commands on each expired token's own stream. NO importer: tokens
// live 15 minutes, so an upgrade carries no meaningful continuity —
// in-flight enrollment links simply get re-minted (the re-mint
// doctrine, already the engine store's rule).
package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	"github.com/larsartmann/go-error-family"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// EnrollStreamType keys per-token enrollment streams.
const EnrollStreamType = "enroll"

// MintEnrollToken stores one hashed token with its expiry (a re-mint
// of the same hash overwrites — the stream IS the token).
type MintEnrollToken struct {
	*command.BasicCommand
	// TokenHash rides explicitly: the stream id renders branded, and
	// the fact payload keys the row exactly as the legacy one did.
	TokenHash string
	Email     string
	UserID    string
	ExpiresAt time.Time
}

// BurnEnrollToken consumes one token; burning an unknown or
// already-burned hash decides NO events (the inert no-op).
type BurnEnrollToken struct {
	*command.BasicCommand
	TokenHash string
}

// enrollState is one token's aggregate state (the stream's whole
// history collapses to "what does this hash resolve to right now").
type enrollState struct {
	Email     string
	UserID    string
	ExpiresAt time.Time
	Minted    bool
}

func (enrollState) initial() enrollState { return enrollState{} }

// enrollStreamEventTypes is the aggregate's stream vocabulary: the
// domain decl's Events entry AND the StrictApply guard list (an unknown
// type on the stream fails replay loudly instead of silently skipping).
var enrollStreamEventTypes = []event.Type{
	event.Type(metaengine.EventTypeName(domain.EnrollTokenMinted{})),
	event.Type(metaengine.EventTypeName(domain.EnrollTokenBurned{})),
}

func applyEnrollState(s enrollState, evt event.Event) (enrollState, error) {
	switch evt.Type() {
	case event.Type(metaengine.EventTypeName(domain.EnrollTokenMinted{})):
		fact, err := event.DecodePayloadAuto[domain.EnrollTokenMinted](evt)
		if err != nil {
			return s, errorfamily.WrapCorruptionf(err, "cqrs.enroll_fold", "decode EnrollTokenMinted")
		}
		s = enrollState{Email: fact.Email, UserID: fact.UserID, ExpiresAt: fact.ExpiresAt, Minted: true}
	case event.Type(metaengine.EventTypeName(domain.EnrollTokenBurned{})):
		s = enrollState{}
	}
	return s, nil
}

func decideMintEnrollToken(_ enrollState, ver event.Version, cmd MintEnrollToken) ([]event.Event, error) {
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.EnrollTokenMinted{})), cmd.StreamID(), EnrollStreamType, ver.Increment(), domain.EnrollTokenMinted{
		Key:       domain.EnrollTokenKey{TokenHash: cmd.TokenHash},
		Email:     cmd.Email,
		UserID:    cmd.UserID,
		ExpiresAt: cmd.ExpiresAt,
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.enroll_mint", "build EnrollTokenMinted")
	}
	return []event.Event{evt}, nil
}

func decideBurnEnrollToken(s enrollState, ver event.Version, cmd BurnEnrollToken) ([]event.Event, error) {
	if !s.Minted {
		return nil, nil
	}
	evt, err := event.New(event.Type(metaengine.EventTypeName(domain.EnrollTokenBurned{})), cmd.StreamID(), EnrollStreamType, ver.Increment(), domain.EnrollTokenBurned{
		Key: domain.EnrollTokenKey{TokenHash: cmd.TokenHash},
	}, event.WithSchemaVersion(1))
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.enroll_burn", "build EnrollTokenBurned")
	}
	return []event.Event{evt}, nil
}

// RegisterEnroll wires the enroll-token aggregate into a System.
func RegisterEnroll(sys *system.System) error {
	if err := system.RegisterDecider[enrollState](sys, EnrollStreamType, decider.Decider[enrollState]{
		Initial: enrollState{}.initial(),
		Apply:   decider.StrictApply(applyEnrollState, enrollStreamEventTypes),
	}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_register", "register enroll decider")
	}
	if err := system.RegisterCommand[*MintEnrollToken, enrollState](sys, "enroll.mint",
		func(_ context.Context, cmd *MintEnrollToken) system.Op[enrollState] {
			return system.Execute[enrollState](context.Background(), cmd.StreamID(), EnrollStreamType,
				func(state enrollState, ver event.Version) ([]event.Event, error) {
					return decideMintEnrollToken(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_register", "register enroll.mint")
	}
	if err := system.RegisterCommand[*BurnEnrollToken, enrollState](sys, "enroll.burn",
		func(_ context.Context, cmd *BurnEnrollToken) system.Op[enrollState] {
			return system.Execute[enrollState](context.Background(), cmd.StreamID(), EnrollStreamType,
				func(state enrollState, ver event.Version) ([]event.Event, error) {
					return decideBurnEnrollToken(state, ver, *cmd)
				})
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_register", "register enroll.burn")
	}
	return nil
}

// enrollProjection folds the token facts into the SAME view shape the
// userauth door reads today (readmodels.EnrollTokenView).
func enrollProjection() metaengine.QueryDecl[enrollKeyInput, readmodels.EnrollTokenView] {
	return metaengine.Query[enrollKeyInput, readmodels.EnrollTokenView](enrollCollection,
		metaengine.OnRecord(domain.EnrollTokenMinted{},
			func(_ record.Record, e domain.EnrollTokenMinted) (domain.EnrollTokenKey, readmodels.EnrollTokenView) {
				return e.Key, readmodels.EnrollTokenView{
					TokenHash: e.Key.TokenHash,
					Email:     e.Email,
					UserID:    e.UserID,
					ExpiresAt: e.ExpiresAt.Unix(),
				}
			}),
		metaengine.OnRecord(domain.EnrollTokenBurned{}, metaengine.Remove[readmodels.EnrollTokenView]()),
		metaengine.Volume(humanScaleVolume),
	)
}

// enrollCollection names the system engine's projection (the same
// name the readmodels engine declares — different databases, one
// shape).
const enrollCollection = "enroll_tokens"

// enrollKeyInput is the projection's key-extraction input (the same
// shape the readmodels engine declares).
type enrollKeyInput struct{ TokenHash string }

// EnrollDomainDecl is the enroll aggregate's share of the system's
// DomainConfig (a registration failure is a boot-time miswire and
// panics loudly).
func EnrollDomainDecl() system.DomainConfig {
	return system.DomainConfig{
		Commands: func(sys *system.System) {
			if err := RegisterEnroll(sys); err != nil {
				panic(fmt.Sprintf("cqrs: register enroll aggregate: %v", err)) //cqrs-lint:ignore(C009) boot-time fail-fast on a composition miswire (DO-1; the decl comment owns the why)
			}
		},
		Projections: []system.ProjectionDeclaration{system.RawQuery(enrollProjection())},
		Events:      enrollStreamEventTypes,
	}
}

// ─── the backend the userauth door keeps calling ───────────────────

// NewEnrollSystem boots the minimal enroll-only system over dsn — the
// operator CLI's one-shot companion (cmd/webphone -enroll-passkey).
// The CLI is a SEPARATE process from the server: it must append to the
// SAME journal (<dataDir>/system.db) the server verifies from, with
// its own projection host running so the mint is folded into the
// shared planned table before the process exits (the cross-process
// read-your-writes contract the stack's enroll-verify VM check pins).
func NewEnrollSystem(ctx context.Context, dsn string) (*system.System, error) {
	sys, err := system.New(ctx, Domain(EnrollDomainDecl()), system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.enroll_boot", "boot enroll system at %s", dsn)
	}
	if err := sys.Start(ctx); err != nil {
		_ = sys.Close() //cqrs-lint:ignore(C023) deliberate cleanup idiom: best-effort close; the primary error/drain owns reporting
		return nil, errorfamily.WrapInfrastructuref(err, "cqrs.enroll_boot", "start enroll system at %s", dsn)
	}
	return sys, nil
}

// EnrollTokens is the dispatch-backed enrollment-token store: the
// same contract the engine-backed store exposed (hash-keyed
// mint/lookup/burn, expired rows swept opportunistically on every
// mint). Token classification (invalid vs expired vs unknown) stays
// in userauth — the seam only ever reports found/not-found plus the
// expiry.
type EnrollTokens struct {
	sys    *system.System
	reader *metaengine.TypedReader[readmodels.EnrollTokenView]
	vers   streamVersions
	mu     sync.Mutex
}

var _ readmodels.EnrollGate = (*EnrollTokens)(nil)

// NewEnrollTokens builds the backend over a System.
func NewEnrollTokens(sys *system.System) *EnrollTokens {
	return &EnrollTokens{sys: sys, reader: metaengine.NewReader[readmodels.EnrollTokenView](sys.MetaEngine(), enrollCollection), vers: streamVersions{}}
}

// Mint stores one hashed token with its expiry and opportunistically
// sweeps the expired ones first (the engine mint's preamble — each
// swept token is a burn command on its own stream).
//
// brandid-lint:ignore(BD003) store-seam transport: identitymodel.UserID parses strictly at the userauth door (strong-id wire ruling)
func (s *EnrollTokens) Mint(ctx context.Context, tokenHash, email, userID string, expiresAt time.Time) error { //nolint:branching-flow // verbatim port of the retired engine store's Mint signature (parity rule); classification lives in userauth, not here
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sweepExpired(ctx, time.Now()); err != nil {
		return err
	}
	if err := dispatchOwnerCommand(ctx, s.sys, "enroll.mint", tokenHash,
		func(base *command.BasicCommand) *MintEnrollToken {
			return &MintEnrollToken{BasicCommand: base, TokenHash: tokenHash, Email: email, UserID: userID, ExpiresAt: expiresAt}
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_mint", "dispatch enroll.mint")
	}
	// The mint's read-back is CROSS-PROCESS: the operator CLI mints
	// while the server verifies. The await folds the token into the
	// shared system.db projection before the CLI exits, so the
	// server's Lookup sees it without racing the server's own
	// projection cadence (D7 mechanism; best-effort, bounded 250 ms).
	awaitOwnWrite(ctx, s.sys, EnrollStreamType, tokenHash, enrollCollection, s.vers)
	return nil
}

// Lookup resolves a token hash to its account and expiry. found=false
// covers unknown, burned, and swept rows alike — the caller owns the
// user-facing classification (including the expired-vs-invalid
// operator-log distinction, which needs the expiry).
func (s *EnrollTokens) Lookup(ctx context.Context, tokenHash string) (email, userID string, expiresAt time.Time, found bool, err error) {
	view, found, err := s.reader.Get(ctx, domain.EnrollTokenKey{TokenHash: tokenHash})
	if err != nil {
		return "", "", time.Time{}, false, errorfamily.WrapInfrastructuref(err, "cqrs.enroll_lookup", "look up enrollment token")
	}
	if !found {
		return "", "", time.Time{}, false, nil
	}
	return view.Email, view.UserID, time.Unix(view.ExpiresAt, 0), true, nil
}

// Burn removes one token (the consume-once arm). Burning an unknown
// or already-burned hash decides no events — the consume outcome is
// the same.
func (s *EnrollTokens) Burn(ctx context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.dispatchBurn(ctx, tokenHash)
}

func (s *EnrollTokens) dispatchBurn(ctx context.Context, tokenHash string) error {
	if err := dispatchOwnerCommand(ctx, s.sys, "enroll.burn", tokenHash,
		func(base *command.BasicCommand) *BurnEnrollToken {
			return &BurnEnrollToken{BasicCommand: base, TokenHash: tokenHash}
		}); err != nil {
		return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_burn", "dispatch enroll.burn")
	}
	return nil
}

// enrollSweepPageSize bounds one expiry scan/burn batch (the retired
// engine store's page size).
const enrollSweepPageSize = 1000

// sweepExpired burns every expired token the projection currently
// shows. The burned-set keeps the loop from re-dispatching against
// projection lag (each burn is inert at the decision anyway); it
// terminates when a scan holds nothing NEW to burn — the retired
// store's `removed == 0` early-out, lag-tolerated.
func (s *EnrollTokens) sweepExpired(ctx context.Context, now time.Time) error {
	burned := make(map[string]struct{})
	for {
		views, err := s.reader.Scan(ctx,
			metaengine.WithLimit(enrollSweepPageSize),
		)
		if err != nil {
			return errorfamily.WrapInfrastructuref(err, "cqrs.enroll_sweep", "scan enrollment tokens")
		}
		if len(views) == 0 {
			return nil
		}
		removed := 0
		for _, v := range views {
			if _, done := burned[v.TokenHash]; done {
				continue
			}
			if time.Unix(v.ExpiresAt, 0).Before(now) {
				if err := s.dispatchBurn(ctx, v.TokenHash); err != nil {
					return err
				}
				burned[v.TokenHash] = struct{}{}
				removed++
			}
		}
		if removed == 0 {
			return nil
		}
	}
}
