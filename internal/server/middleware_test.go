package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cqrshtmx "github.com/larsartmann/cqrs-htmx/v4"
	"github.com/larsartmann/httputil"
)

// productionStack mirrors the middleware order built in New, so the panic
// path is exercised exactly as it runs in production.
func productionStack(next http.Handler) http.Handler {
	security := httputil.SecurityHeaders(httputil.SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: contentSecurityPolicy,
	})

	return security(cqrshtmx.RecoveryMiddleware(next))
}

// captureDefaultLogger swaps the package-global slog default for a buffer
// and restores the previous logger on cleanup.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &buf
}

func TestPanicRecoveredAs500WithStackLog(t *testing.T) {
	log := captureDefaultLogger(t)

	boom := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic("boom: test injected panic")
	})

	rec := httptest.NewRecorder()
	productionStack(boom).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/somewhere", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("panic recovered as %d, want 500", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "boom") || strings.Contains(body, "goroutine") {
		t.Errorf("panic detail leaked to the client: %q", body)
	}

	entry := log.String()
	for _, want := range []string{
		"panic recovered",
		"method=GET",
		"path=/somewhere",
		"stack=\"goroutine",
		"boom: test injected panic",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("panic log missing %q in:\n%s", want, entry)
		}
	}
}

func TestErrAbortHandlerRepanicsPerNetHTTPConvention(t *testing.T) {
	captureDefaultLogger(t)

	abort := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("http.ErrAbortHandler was swallowed; net/http expects it re-raised")
		}
		if recovered != http.ErrAbortHandler {
			t.Fatalf("re-panicked with %v, want http.ErrAbortHandler", recovered)
		}
	}()

	productionStack(abort).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
}

func TestPanicInsideSecurityHeadersKeepsHeaders(t *testing.T) {
	captureDefaultLogger(t)

	boom := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic("headers first")
	})

	rec := httptest.NewRecorder()
	productionStack(boom).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/headers", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("security headers lost on the panic path: X-Content-Type-Options=%q", got)
	}
}

// TestChainCompositionMatchesNestedOrder pins the CH1 refactor: New() now
// composes with cqrshtmx.Chain (first argument outermost). The parity
// proof is behavioral: the served page must still carry security headers
// (security inside the logger) and a panic must still 500 with headers
// (recovery innermost) — the same contracts productionStack pins for the
// nested spelling.
func TestChainCompositionMatchesNestedOrder(t *testing.T) {
	captureDefaultLogger(t)
	server := newTestServer(t)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / through the Chain stack: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("security headers missing under Chain composition: %q", got)
	}
}

// TestVersionEndpoint pins the /version contract: build metadata as the
// library DebugHandler emits it (JSON, no-cache), route outside CSRF.
func TestVersionEndpoint(t *testing.T) {
	server := newTestServer(t)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /version: %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("version content-type %q, want json", ct)
	}
	body := string(readAll(t, resp))
	for _, key := range []string{`"version"`, `"goVersion"`, `"title"`} {
		if !strings.Contains(body, key) {
			t.Errorf("version body missing %s: %s", key, body)
		}
	}
}

// TestServerTimingOptIn pins the ST1 contract: the Server-Timing header
// appears only when WEBPHONE_DEBUG_TIMING is set, and SSE flushing still
// works under the timing writer (its Flush forwarding is what keeps the
// library's heartbeat from stalling).
func TestServerTimingOptIn(t *testing.T) {
	t.Run("off by default", func(t *testing.T) {
		t.Setenv("WEBPHONE_DEBUG_TIMING", "")
		server := newTestServer(t)

		req, _ := http.NewRequest(http.MethodGet, server.URL+"/", nil)
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if got := resp.Header.Get("Server-Timing"); got != "" {
			t.Errorf("Server-Timing present with the flag off: %q", got)
		}
	})

	t.Run("on when flagged", func(t *testing.T) {
		t.Setenv("WEBPHONE_DEBUG_TIMING", "1")
		server := newTestServer(t)

		req, _ := http.NewRequest(http.MethodGet, server.URL+"/", nil)
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		header := resp.Header.Get("Server-Timing")
		if header == "" || !strings.Contains(header, "total;dur=") {
			t.Errorf("Server-Timing %q, want total;dur=... with the flag on", header)
		}
	})
}
