package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/samber/do/v2"

	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/store"
)

func testConfig(t testing.TB, dashboard bool) config.Config {
	t.Helper()
	return config.Config{
		Addr:          ":0",
		DataDir:       t.TempDir(),
		WebsocketPath: "/sip",
		SessionTTL:    time.Hour,
		Gateway:       config.Gateway{Mode: config.GatewayLoopback, WebhookSecret: "test-secret"},
		Dashboard:     config.Dashboard{Enable: dashboard, Title: "webphone test"},
	}
}

// startApp builds the app, starts its background loops under a
// canceled-on-cleanup context, and serves its handler over an
// httptest server. Shutdown always runs — the DO-2 rule.
func startApp(t testing.TB, dashboard bool) *App {
	t.Helper()
	application, err := New(testConfig(t, dashboard), discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(cancel)
	t.Cleanup(func() {
		if err := application.Shutdown(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	})
	return application
}

// TestAppServesProbeTriple pins the container-built probe's wire
// behavior: /healthz stays the cqrshtmx readiness JSON with both named
// checks, /livez is fetch-free pass, /startupz latches on the critical
// pair served from the injector's named services.
func TestAppServesProbeTriple(t *testing.T) {
	application := startApp(t, false)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	get := func(path string) (int, string) {
		t.Helper()
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}

	if code, body := get("/healthz"); code != http.StatusOK || !strings.Contains(body, `"sqlite"`) || !strings.Contains(body, `"blob-dir"`) {
		t.Fatalf("healthz: code %d body %s", code, body)
	}
	if code, body := get("/livez"); code != http.StatusOK || !strings.Contains(body, `"pass"`) {
		t.Fatalf("livez: code %d body %s", code, body)
	}
	if code, body := get("/startupz"); code != http.StatusOK {
		t.Fatalf("startupz: code %d body %s — the critical pair is healthy, the latch must pass", code, body)
	}
}

// TestAppDashboardDisabledByDefault pins the opt-in contract: without
// dashboard.enable the /health subtree does not exist — the styled 404
// answers, status 404.
func TestAppDashboardDisabledByDefault(t *testing.T) {
	application := startApp(t, false)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("/health with dashboard disabled: code %d, want 404", resp.StatusCode)
	}
}

// TestAppDashboardMountsHealthSubtree pins the mounted surface: HTML
// page with per-request nonce'd inline scripts under the dashboard's
// verified CSP, the embedded Datastar SDK served same-origin, the
// scoped Tailwind build as the stylesheet, and the namespaced probe
// aliases answering from the same probe as the root triple.
func TestAppDashboardMountsHealthSubtree(t *testing.T) {
	application := startApp(t, true)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	get := func(path, accept string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	resp := get("/health", "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/health: code %d", resp.StatusCode)
	}
	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "unsafe-eval") || !strings.Contains(csp, "nonce-") {
		t.Fatalf("/health CSP is not the dashboard policy: %q", csp)
	}
	body, _ := io.ReadAll(resp.Body)
	page := string(body)
	if !strings.Contains(page, "/assets/health.css") {
		t.Error("dashboard page does not reference its scoped Tailwind build /assets/health.css")
	}
	if !strings.Contains(page, "/health/datastar.js") {
		t.Error("dashboard page does not reference the embedded Datastar SDK")
	}
	if !strings.Contains(page, "nonce=") {
		t.Error("dashboard page has no nonce'd inline scripts — the CSP would silence it")
	}

	// Per-request nonces: two loads must not share one.
	second := get("/health", "")
	if second.Header.Get("Content-Security-Policy") == csp {
		_ = second.Body.Close()
		t.Error("CSP nonce repeated across requests — nonces must be per-request")
	}
	_ = second.Body.Close()

	// Content negotiation: Accept: application/json answers go-health's
	// response shape with the container's check names.
	jsonResp := get("/health", "application/json")
	defer func() { _ = jsonResp.Body.Close() }()
	jsonBody, _ := io.ReadAll(jsonResp.Body)
	for _, check := range []string{"sqlite", "blob-dir", "dashboard"} {
		if !strings.Contains(string(jsonBody), `"`+check+`"`) {
			t.Errorf("/health JSON missing check %q: %s", check, jsonBody)
		}
	}

	// The embedded SDK serves JavaScript same-origin.
	sdk := get("/health/datastar.js", "")
	defer func() { _ = sdk.Body.Close() }()
	if sdk.StatusCode != http.StatusOK || !strings.Contains(sdk.Header.Get("Content-Type"), "javascript") {
		t.Fatalf("datastar.js: code %d type %s", sdk.StatusCode, sdk.Header.Get("Content-Type"))
	}

	// Namespaced probe aliases answer from the same probe as /livez.
	alias := get("/health/livez", "")
	defer func() { _ = alias.Body.Close() }()
	if alias.StatusCode != http.StatusOK {
		t.Fatalf("/health/livez alias: code %d", alias.StatusCode)
	}
}

// TestAppDashboardSSEStreams pins the SSE mount: the stream answers
// text/event-stream with the Datastar patch event type. The read is
// bounded by a client timeout — the stream never ends on its own.
func TestAppDashboardSSEStreams(t *testing.T) {
	application := startApp(t, true)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(server.URL + "/health/sse")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/health/sse: code %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("/health/sse content-type: %q", ct)
	}
	buf := make([]byte, 512)
	n, _ := resp.Body.Read(buf)
	if !strings.Contains(string(buf[:n]), "event:") {
		t.Fatalf("SSE stream carried no event in first %d bytes: %q", n, buf[:n])
	}
}

// TestAppShutdownClosesDatabaseAndIsIdempotent pins the DO-2 cascade:
// Shutdown closes the SQLite handle exactly once (a closed handle
// fails HealthCheck), and a second Shutdown is inert.
func TestAppShutdownClosesDatabaseAndIsIdempotent(t *testing.T) {
	cfg := testConfig(t, false)
	application, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	db, err := do.InvokeNamed[*store.Database](application.injector, "sqlite")
	if err != nil {
		t.Fatalf("resolve sqlite: %v", err)
	}
	if err := db.HealthCheck(context.Background()); err != nil {
		t.Fatalf("live database before shutdown: %v", err)
	}

	if err := application.Shutdown(); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	if err := application.Shutdown(); err != nil {
		t.Fatalf("second Shutdown must be inert, got: %v", err)
	}
	if err := db.HealthCheck(context.Background()); err == nil {
		t.Fatal("database still answers HealthCheck after Shutdown — the cascade did not close it")
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestAppServesMainPageUnderStrictCSP guards the boundary: the shell
// page keeps webphone's strict app-wide CSP (no unsafe-eval, no nonce)
// — the relaxed dashboard policy applies to the /health subtree ONLY.
func TestAppServesMainPageUnderStrictCSP(t *testing.T) {
	application := startApp(t, true)
	server := httptest.NewServer(application.Handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	csp := resp.Header.Get("Content-Security-Policy")
	if strings.Contains(csp, "unsafe-eval") || strings.Contains(csp, "nonce-") {
		t.Fatalf("main page CSP was relaxed by the dashboard wiring: %q", csp)
	}
}
