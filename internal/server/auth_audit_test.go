package server

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestAuthAuditNeverLogsSecrets pins the audit contract: the semantic
// auth events reach the journal (login/logout greppable) and the log
// carries IDENTIFIERS only — the submitted password, the session token,
// or any credential never appears.
func TestAuthAuditNeverLogsSecrets(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	server, svc := newPasskeyServer(t)
	enrollCredential(t, svc, pkEmail)
	c := clientFor(t, server)
	c.login("1000", pkPassword)
	if resp, body := c.do(http.MethodDelete, "/api/session", nil, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: %d %s", resp.StatusCode, body)
	}

	out := buf.String()
	for _, want := range []string{"auth login", "auth logout"} {
		if !strings.Contains(out, want) {
			t.Errorf("journal missing the %q audit event:\n%s", want, out)
		}
	}
	if strings.Contains(out, pkPassword) {
		t.Errorf("journal leaked the submitted password:\n%s", out)
	}
}

// TestAgentAuditRidesThePanelActions pins the M18 audit trail: one
// greppable "agent <event>" line per drafts-panel action with the actor
// extension and the draft id — success AND failure — and never the
// draft's body text.
func TestAgentAuditRidesThePanelActions(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	stub := newStubReviewAPI(t)
	c := clientWithAgent(t, stub)
	c.login("1001", "pw")

	if resp, _ := c.do(http.MethodPost, "/messages/drafts/d-7/approve", nil, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("approve: %d", resp.StatusCode)
	}
	out := buf.String()
	if !strings.Contains(out, "agent approve") || !strings.Contains(out, `draft=d-7`) || !strings.Contains(out, `extension=1001`) {
		t.Errorf("journal missing the approve audit line:\n%s", out)
	}
	if strings.Contains(out, "Guten Morgen") {
		t.Errorf("the journal must never carry draft body text:\n%s", out)
	}

	// A failed action audits its failure honestly.
	stub.status.Store(http.StatusServiceUnavailable)
	if resp, _ := c.do(http.MethodPost, "/messages/drafts/d-7/discard", nil, ""); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("discard against a fail-closed agent: %d", resp.StatusCode)
	}
	if !strings.Contains(buf.String(), `agent discard`) || !strings.Contains(buf.String(), `outcome=failed`) {
		t.Errorf("journal missing the failed-discard audit line:\n%s", buf.String())
	}
}

// TestAgentAuditNamesTheHumanActor pins the multi-user substrate's line
// shapes: a session with a login email audits under BOTH attrs (actor =
// the human, extension = the line they acted on), while an extension
// login keeps the historical shape byte-identical (no actor attr — the
// extension already IS the actor).
func TestAgentAuditNamesTheHumanActor(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	h := &handlers{deps: Deps{}}
	h.agentAudit("approve", "lars@example.com", "1000", "d-1", "ok")
	line := buf.String()
	for _, want := range []string{"agent approve", "extension=1000", "actor=lars@example.com", "draft=d-1", "outcome=ok"} {
		if !strings.Contains(line, want) {
			t.Errorf("email-actor line missing %q:\n%s", want, line)
		}
	}

	buf.Reset()
	h.agentAudit("discard", "", "1001", "d-2", "failed")
	plain := buf.String()
	for _, want := range []string{"agent discard", "extension=1001", "draft=d-2"} {
		if !strings.Contains(plain, want) {
			t.Errorf("extension-actor line missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "actor=") {
		t.Errorf("extension login must keep the historical shape (no actor attr):\n%s", plain)
	}
}

// TestAgentAuditDualWritesTheDurableRow pins the D5 read model: with
// the store wired, one agentAudit call lands BOTH the journal line and
// a queryable row (actor, event, subject, outcome); a nil store keeps
// the line alone (byte-parity for test compositions); a foreign
// extension shape never reaches the store but still logs.
func TestAgentAuditDualWritesTheDurableRow(t *testing.T) {
	engine := newTestEngine(t)
	h := &handlers{deps: Deps{AgentAudit: engine.AgentAudit()}}

	h.agentAudit("approve", "lars@example.com", "1000", "d-9", "ok")
	h.agentAudit("mode", "", "1000", "", "pinned")

	rows, err := engine.AgentAudit().Recent(context.Background(), domain.MustParseExtension("1000"), 10)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 durable rows, got %d", len(rows))
	}
	// Newest first: the mode pin landed after the approve.
	if rows[0].Event != "mode" || rows[0].Actor != "" || rows[0].Outcome != "pinned" {
		t.Errorf("newest row = %+v", rows[0])
	}
	if rows[1].Event != "approve" || rows[1].Actor != "lars@example.com" || rows[1].Subject != "d-9" || rows[1].Outcome != "ok" {
		t.Errorf("oldest row = %+v", rows[1])
	}

	// Nil store: the line still writes (no panic, no row).
	bare := &handlers{deps: Deps{}}
	bare.agentAudit("discard", "", "1000", "d-1", "ok")
	// Undialable extension shape: logged, never stored (letters alone
	// ARE dialable — they would store under their sanitized owner).
	h.agentAudit("approve", "", "---", "d-2", "ok")
	rows, err = engine.AgentAudit().Recent(context.Background(), domain.MustParseExtension("1000"), 10)
	if err != nil || len(rows) != 2 {
		t.Errorf("foreign shape must not store (rows=%d err=%v)", len(rows), err)
	}
}
