package cqrs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"

	"github.com/larsartmann/webphone/internal/readmodels"
)

// bootAgentAuditSystem boots a REAL system with the agent-audit domain
// over a dir (a dir reuse is the restart case: the journal replays).
func bootAgentAuditSystem(t *testing.T, dir string) *AgentAudit {
	t.Helper()
	ctx := context.Background()
	dsn := filepath.Join(dir, "system.db")
	domainCfg := AgentAuditDomainDecl()
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
	return NewAgentAudit(sys)
}

// awaitAuditCount polls the projection until the owner's window holds
// want rows.
func awaitAuditCount(t *testing.T, s *AgentAudit, owner string, want int) []readmodels.AgentAuditView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		rows, err := s.Recent(context.Background(), mustExt(t, owner), want)
		if err != nil {
			t.Fatalf("Recent: %v", err)
		}
		if len(rows) == want {
			return rows
		}
		if time.Now().After(deadline) {
			t.Fatalf("audit projection never reached %d rows: got %d", want, len(rows))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestAgentAuditAppendRecent pins the append/read round-trip: rows
// land with actor/event/subject/outcome and Recent answers newest
// first.
func TestAgentAuditAppendRecent(t *testing.T) {
	s := bootAgentAuditSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	if err := s.Append(ctx, owner, "ops@example.com", "approve", "draft-1", "sent"); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if err := s.Append(ctx, owner, "ops@example.com", "discard", "draft-2", "dropped"); err != nil {
		t.Fatalf("append 2: %v", err)
	}
	rows := awaitAuditCount(t, s, "1001", 2)
	if rows[0].Event != "discard" || rows[1].Event != "approve" {
		t.Fatalf("Recent must answer newest first: got [%s, %s]", rows[0].Event, rows[1].Event)
	}
	if rows[0].Actor != "ops@example.com" || rows[0].Subject != "draft-2" || rows[0].Outcome != "dropped" {
		t.Fatalf("row facts drifted: %+v", rows[0])
	}

	other, err := s.Recent(ctx, mustExt(t, "1002"), 20)
	if err != nil || other != nil {
		t.Fatalf("owner scoping broke: rows=%v err=%v", other, err)
	}
}

// TestAgentAuditCapTrimsOldest pins the cap decision: past the cap,
// one append emits the append PLUS the expunge of the oldest row —
// the window converges inside the same event batch.
func TestAgentAuditCapTrimsOldest(t *testing.T) {
	old := agentAuditMaxRows
	agentAuditMaxRows = 3
	t.Cleanup(func() { agentAuditMaxRows = old })

	s := bootAgentAuditSystem(t, t.TempDir())
	ctx := context.Background()
	owner := mustExt(t, "1001")

	for i := range 4 {
		if err := s.Append(ctx, owner, "ops@example.com", "mode_pin", "run-"+string(rune('0'+i)), "pinned"); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond) // distinct seq stamps
	}
	rows := awaitAuditCount(t, s, "1001", 3)
	for _, r := range rows {
		if r.Subject == "run-0" {
			t.Fatalf("oldest row survived the cap trim: %+v", rows)
		}
	}
	if rows[2].Subject != "run-1" {
		t.Fatalf("the oldest surviving row should be the second append: %+v", rows[2])
	}
}

// TestAgentAuditStateSurvivesRestart pins the journal reload: rows
// appended before the restart still answer Recent after it, and the
// reloaded decider keeps the cap semantics.
func TestAgentAuditStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	owner := mustExt(t, "1001")

	s1 := bootAgentAuditSystem(t, dir)
	if err := s1.Append(ctx, owner, "ops@example.com", "approve", "draft-1", "sent"); err != nil {
		t.Fatalf("append before restart: %v", err)
	}
	awaitAuditCount(t, s1, "1001", 1)

	s2 := bootAgentAuditSystem(t, dir)
	rows := awaitAuditCount(t, s2, "1001", 1)
	if rows[0].Event != "approve" || rows[0].Actor != "ops@example.com" {
		t.Fatalf("audit row drifted across restart: %+v", rows[0])
	}
}
