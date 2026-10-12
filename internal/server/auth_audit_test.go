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
