package cqrs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// bootPushDevicesSystem boots a REAL system with the push-devices
// domain over a dir (a dir reuse is the restart case).
func bootPushDevicesSystem(t *testing.T, dir string) *PushDevices {
	t.Helper()
	ctx := context.Background()
	dsn := filepath.Join(dir, "system.db")
	domainCfg := PushDevicesDomainDecl()
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
	return NewPushDevices(sys)
}

// awaitDeviceCount polls the projection until the owner holds want
// devices.
func awaitDeviceCount(t *testing.T, s *PushDevices, owner string, want int) []readmodels.PushDeviceView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		views, err := s.List(context.Background(), mustExt(t, owner))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(views) == want {
			return views
		}
		if time.Now().After(deadline) {
			t.Fatalf("device projection never reached %d rows: got %d", want, len(views))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestPushDevicesRegisterListRemove pins the registry round-trip:
// register lands, list is newest-first, remove drops, and a second
// remove is inert.
func TestPushDevicesRegisterListRemove(t *testing.T) {
	s := bootPushDevicesSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Upsert(ctx, owner, "phone-1", "ntfy", "topic-a"); err != nil {
		t.Fatalf("register 1: %v", err)
	}
	if err := s.Upsert(ctx, owner, "tablet-2", "fcm", "fcm-token-2"); err != nil {
		t.Fatalf("register 2: %v", err)
	}
	views := awaitDeviceCount(t, s, "1001", 2)
	if views[0].DeviceID != "tablet-2" || views[1].DeviceID != "phone-1" {
		t.Fatalf("List must answer newest first: %+v", views)
	}
	if views[0].Transport != "fcm" || views[0].Token != "fcm-token-2" {
		t.Fatalf("row facts drifted: %+v", views[0])
	}

	if err := s.Remove(ctx, owner, "phone-1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	views = awaitDeviceCount(t, s, "1001", 1)
	if views[0].DeviceID != "tablet-2" {
		t.Fatalf("wrong device survived: %+v", views)
	}
	if err := s.Remove(ctx, owner, "phone-1"); err != nil {
		t.Fatalf("second remove must be inert: %v", err)
	}
}

// TestPushDevicesRotationReplaces pins the rotation semantics: a
// known device id replaces its token (cap-free at the limit).
func TestPushDevicesRotationReplaces(t *testing.T) {
	s := bootPushDevicesSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Upsert(ctx, owner, "phone-1", "ntfy", "topic-old"); err != nil {
		t.Fatalf("register: %v", err)
	}
	awaitDeviceCount(t, s, "1001", 1)
	if err := s.Upsert(ctx, owner, "phone-1", "ntfy", "topic-new"); err != nil {
		t.Fatalf("rotate: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		views, err := s.List(ctx, owner)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(views) == 1 && views[0].Token == "topic-new" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("rotation never replaced the token: %+v", views)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestPushDevicesCapRejectsNewDevice pins the 8-per-owner cap: a NEW
// device id past the cap answers ErrListFull while rotation at the
// cap still succeeds.
func TestPushDevicesCapRejectsNewDevice(t *testing.T) {
	s := bootPushDevicesSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for i := 0; i < readmodels.PushDevicesMaxPerExtension; i++ {
		if err := s.Upsert(ctx, owner, string(rune('a'+i)), "ntfy", "topic"); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}
	awaitDeviceCount(t, s, "1001", readmodels.PushDevicesMaxPerExtension)

	err := s.Upsert(ctx, owner, "one-too-many", "ntfy", "topic")
	if !errors.Is(err, readmodels.ErrListFull) {
		t.Fatalf("a NEW device past the cap must answer ErrListFull, got: %v", err)
	}
	// Rotation at the cap stays free (the device keeps its slot).
	if err := s.Upsert(ctx, owner, "a", "ntfy", "topic-rotated"); err != nil {
		t.Fatalf("rotation at the cap must succeed: %v", err)
	}
}

// TestPushDevicesStateSurvivesRestart pins the journal reload: the
// registry (and the cap) survive a restart.
func TestPushDevicesStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s1 := bootPushDevicesSystem(t, dir)
	if err := s1.Upsert(ctx, owner, "phone-1", "ntfy", "topic-a"); err != nil {
		t.Fatalf("register before restart: %v", err)
	}
	awaitDeviceCount(t, s1, "1001", 1)

	s2 := bootPushDevicesSystem(t, dir)
	views := awaitDeviceCount(t, s2, "1001", 1)
	if views[0].Token != "topic-a" {
		t.Fatalf("device drifted across restart: %+v", views[0])
	}
	if err := s2.Remove(ctx, owner, "phone-1"); err != nil {
		t.Fatalf("remove after restart: %v", err)
	}
}
