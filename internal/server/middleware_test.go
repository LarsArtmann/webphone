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
