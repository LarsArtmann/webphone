package cqrs

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// basicCommand mints the BasicCommand a dispatcher would hand the
// decider — the decision-level cap tests need a real stream id.
func basicCommand(t *testing.T, name, stream string) *command.BasicCommand {
	t.Helper()
	streamID, err := id.ParseStreamID(stream)
	if err != nil {
		t.Fatalf("parse stream id: %v", err)
	}
	base, err := command.New(command.Type(name), streamID)
	if err != nil {
		t.Fatalf("build command: %v", err)
	}
	return base
}

// bootDomainSystem boots a REAL system with the given domain decl over
// dir — the Phase-4 batteries' shared harness (the snippets boot shape;
// fast projection backoff so awaits converge inside the budget).
func bootDomainSystem(t *testing.T, dir string, decl system.DomainConfig) *system.System {
	t.Helper()
	ctx := context.Background()
	decl.ProjectionHostOptions = append(decl.ProjectionHostOptions,
		projectionhost.WithBackoff(time.Millisecond, 5*time.Millisecond))
	sys, err := system.New(ctx, decl, system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			"db": {Driver: "sqlite", DSN: dir + "/system.db", Pragmas: []string{"wal"}},
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
	return sys
}
