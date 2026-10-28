package readmodels

import (
	"context"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestAgentAuditAppendRecentAndCap pins the D5 read model: rows append
// immutably, Recent answers the owner's newest window first, the cap
// trims the OLDEST (audit rows are system facts), and owners never see
// each other's rows.
func TestAgentAuditAppendRecentAndCap(t *testing.T) {
	engine, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = engine.Shutdown() })
	audit := engine.AgentAudit()
	ctx := context.Background()
	owner := domain.MustParseExtension("1001")
	other := domain.MustParseExtension("1002")

	if rows, err := audit.Recent(ctx, owner, 10); err != nil || rows != nil {
		t.Fatalf("empty recent = %v %v; want nil, nil", rows, err)
	}

	for i := 0; i < 3; i++ {
		if err := audit.Append(ctx, owner, "lars@example.com", "approve", "d-1", "ok"); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	if err := audit.Append(ctx, other, "", "mode", "", "pinned"); err != nil {
		t.Fatalf("other append: %v", err)
	}

	rows, err := audit.Recent(ctx, owner, 2)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("limit: got %d rows, want 2", len(rows))
	}
	// Newest first.
	if rows[0].Seq < rows[1].Seq {
		t.Errorf("recent must answer newest first: %+v", rows)
	}
	// Owner scoping: 1001 never sees 1002's row.
	for _, row := range rows {
		if row.Owner != "1001" || row.Actor != "lars@example.com" {
			t.Errorf("foreign row leaked: %+v", row)
		}
	}

	// The cap trims the OLDEST rows (system facts, transcripts posture).
	for i := 0; i < AgentAuditMaxPerExtension+5; i++ {
		if err := audit.Append(ctx, owner, "", "approve", "d", "ok"); err != nil {
			t.Fatalf("flood append %d: %v", i, err)
		}
	}
	rows, err = audit.Recent(ctx, owner, AgentAuditMaxPerExtension+10)
	if err != nil {
		t.Fatalf("recent after flood: %v", err)
	}
	if len(rows) > AgentAuditMaxPerExtension {
		t.Errorf("cap exceeded: %d rows", len(rows))
	}
}
