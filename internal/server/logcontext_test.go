package server

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestInternalErrorLogsFamilyFields pins the LogErrorContext adoption
// (2026-09-24): the operator log line carries the op-prefixed message
// plus the library's family/code/retryable attrs, and a transient
// classification logs at Warn (the library's self-healing semantics).
func TestInternalErrorLogsFamilyFields(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	h := &handlers{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/partials/messages/t-1", nil)
	h.internalError(w, r, "load conversation", errors.New("sql: boom"))

	line := buf.String()
	if !strings.Contains(line, `msg="load conversation: sql: boom"`) {
		t.Fatalf("log line lost the op-prefixed message: %q", line)
	}
	for _, attr := range []string{"family=", "code=", "retryable="} {
		if !strings.Contains(line, attr) {
			t.Fatalf("log line lost the %s attr: %q", attr, line)
		}
	}
	if !strings.Contains(line, "level=WARN") {
		t.Fatalf("an unclassified (transient) error must log Warn: %q", line)
	}
}
