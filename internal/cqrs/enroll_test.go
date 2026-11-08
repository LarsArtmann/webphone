package cqrs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// bootEnrollSystem boots a REAL system with the enroll domain over a
// dir (a dir reuse is the restart case: the journal replays).
func bootEnrollSystem(t *testing.T, dir string) *EnrollTokens {
	t.Helper()
	ctx := context.Background()
	dsn := filepath.Join(dir, "system.db")
	domainCfg := EnrollDomainDecl()
	domainCfg.ProjectionHostOptions = append(domainCfg.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, domainCfg, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dsn, Pragmas: []string{"wal"}},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: "db", Durability: system.DurabilityNormal},
			{Role: system.RoleProjections, Engine: "db", Durability: system.DurabilityNormal},
		},
	})
	if err != nil {
		t.Fatalf("system.New: %v", err)
	}
	if err := sys.Start(ctx); err != nil {
		t.Fatalf("sys.Start: %v", err)
	}
	t.Cleanup(func() { _ = sys.GracefulClose(ctx) })
	return NewEnrollTokens(sys)
}

// hashFor builds a distinct token hash per test call (the stream key
// only needs to be a stable string — the real seam hashes; here any
// unique UTF-8 id does).
var enrollSeq int

func hashFor(t *testing.T) string {
	t.Helper()
	enrollSeq++
	return "hash-" + time.Now().Format("150405.000000000") + "-" + string(rune('a'+enrollSeq%26)) + string(rune('a'+enrollSeq/26))
}

// awaitLookupFound polls the projection until the hash resolves.
func awaitLookupFound(t *testing.T, s *EnrollTokens, hash string) (email, userID string, expiresAt time.Time) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		email, userID, expiresAt, found, err := s.Lookup(context.Background(), hash)
		if err != nil {
			t.Fatalf("Lookup: %v", err)
		}
		if found {
			return email, userID, expiresAt
		}
		if time.Now().After(deadline) {
			t.Fatalf("enroll projection never surfaced token %s", hash)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestEnrollMintLookupBurn pins the consume-once round-trip: mint
// lands in the projection with the account facts and expiry, burn
// removes, and a second burn is the inert no-op (the retired store's
// contract).
func TestEnrollMintLookupBurn(t *testing.T) {
	s := bootEnrollSystem(t, t.TempDir())
	ctx := context.Background()
	hash := hashFor(t)
	want := time.Now().Add(15 * time.Minute).Truncate(time.Second)

	if err := s.Mint(ctx, hash, "ops@example.com", "user-1", want); err != nil {
		t.Fatalf("Mint: %v", err)
	}
	email, userID, expiresAt := awaitLookupFound(t, s, hash)
	if email != "ops@example.com" || userID != "user-1" {
		t.Fatalf("lookup resolved wrong account: email=%q userID=%q", email, userID)
	}
	if !expiresAt.Equal(want) {
		t.Fatalf("expiry drifted: want %v got %v", want, expiresAt)
	}

	if err := s.Burn(ctx, hash); err != nil {
		t.Fatalf("Burn: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, _, _, found, err := s.Lookup(ctx, hash)
		if err != nil {
			t.Fatalf("Lookup after burn: %v", err)
		}
		if !found {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("burned token never left the projection")
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err := s.Burn(ctx, hash); err != nil {
		t.Fatalf("second burn must be inert, got: %v", err)
	}
}

// TestEnrollBurnInertForUnknown pins that burning a never-minted
// hash answers success without emitting anything (the decision knows
// the stream is empty).
func TestEnrollBurnInertForUnknown(t *testing.T) {
	s := bootEnrollSystem(t, t.TempDir())
	if err := s.Burn(context.Background(), "never-minted"); err != nil {
		t.Fatalf("Burn unknown must be inert, got: %v", err)
	}
	_, _, _, found, err := s.Lookup(context.Background(), "never-minted")
	if err != nil || found {
		t.Fatalf("unknown hash must not resolve: found=%v err=%v", found, err)
	}
}

// TestEnrollSweepOnMint pins the mint-side opportunistic sweep: an
// expired token leaves on the NEXT mint (each sweep is a burn command
// on the expired token's own stream).
func TestEnrollSweepOnMint(t *testing.T) {
	s := bootEnrollSystem(t, t.TempDir())
	ctx := context.Background()
	stale := hashFor(t)
	fresh := hashFor(t)

	if err := s.Mint(ctx, stale, "ops@example.com", "user-1", time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("mint stale: %v", err)
	}
	awaitLookupFound(t, s, stale)

	if err := s.Mint(ctx, fresh, "ops@example.com", "user-2", time.Now().Add(15*time.Minute)); err != nil {
		t.Fatalf("mint fresh (sweep pass): %v", err)
	}
	awaitLookupFound(t, s, fresh)

	deadline := time.Now().Add(5 * time.Second)
	for {
		_, _, _, found, err := s.Lookup(ctx, stale)
		if err != nil {
			t.Fatalf("Lookup stale: %v", err)
		}
		if !found {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("expired token was never swept by the following mint")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestEnrollExpiredStillResolvesUntilSwept pins the door-side
// classification substrate: an expired token still RESOLVES (with its
// past expiry) until something sweeps or burns it — userauth owns the
// expired-vs-invalid distinction, not the store.
func TestEnrollExpiredStillResolvesUntilSwept(t *testing.T) {
	s := bootEnrollSystem(t, t.TempDir())
	ctx := context.Background()
	hash := hashFor(t)
	past := time.Now().Add(-2 * time.Minute).Truncate(time.Second)

	if err := s.Mint(ctx, hash, "ops@example.com", "user-9", past); err != nil {
		t.Fatalf("mint expired: %v", err)
	}
	_, _, expiresAt := awaitLookupFound(t, s, hash)
	if !expiresAt.Equal(past) {
		t.Fatalf("expired token must resolve with its past expiry: want %v got %v", past, expiresAt)
	}
}

// TestEnrollStateSurvivesRestart pins the journal reload: a token
// minted before the restart still resolves and still burns after it.
func TestEnrollStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	hash := hashFor(t)

	s1 := bootEnrollSystem(t, dir)
	if err := s1.Mint(ctx, hash, "ops@example.com", "user-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("mint before restart: %v", err)
	}
	awaitLookupFound(t, s1, hash)

	s2 := bootEnrollSystem(t, dir)
	_, _, _, found, err := s2.Lookup(ctx, hash)
	if err != nil || !found {
		t.Fatalf("token must reload from the journal: found=%v err=%v", found, err)
	}
	if err := s2.Burn(ctx, hash); err != nil {
		t.Fatalf("burn after restart: %v", err)
	}
}
