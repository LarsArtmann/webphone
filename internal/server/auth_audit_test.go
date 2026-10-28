package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"
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

	agentAudit("approve", "lars@example.com", "1000", "d-1", "ok")
	line := buf.String()
	for _, want := range []string{"agent approve", "extension=1000", "actor=lars@example.com", "draft=d-1", "outcome=ok"} {
		if !strings.Contains(line, want) {
			t.Errorf("email-actor line missing %q:\n%s", want, line)
		}
	}

	buf.Reset()
	agentAudit("discard", "", "1001", "d-2", "failed")
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
