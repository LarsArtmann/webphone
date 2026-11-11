package fax_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/cqrs"
)

// newTestFaxes boots a throwaway system with the fax domain over a
// temp dir and returns its dispatch-backed fax store (the journal
// backend replaced the engine-backed readmodels store behind the same
// contract when the fax twins dissolved).
func newTestFaxes(t *testing.T) *cqrs.Faxes {
	t.Helper()
	ctx := context.Background()
	decl := cqrs.FaxesDomainDecl()
	decl.ProjectionHostOptions = append(decl.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, decl, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "system.db"), Pragmas: []string{"wal"}},
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
	return cqrs.NewFaxes(sys)
}
