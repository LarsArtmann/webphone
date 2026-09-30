package app

import (
	"net/http"
	"time"

	"github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
	"github.com/larsartmann/httputil"
	"github.com/samber/do/v2"

	"github.com/larsartmann/webphone/internal/config"
)

// dashboardRateLimit fences the dashboard-owned routes (HTML, SSE,
// trend, export): one shared token bucket of 60 requests per minute —
// generous for a human operator, bounded for a runaway client. The
// probe aliases are never limited (kubelet contract).
var dashboardRateLimit = struct {
	requests int
	window   time.Duration
}{60, time.Minute}

// dashboardCSP sets the subtree's Content-Security-Policy: the
// dashboard library's verified policy — everything self-hosted,
// nonce'd inline scripts, 'unsafe-eval' ONLY because the Datastar SDK
// compiles its data-* expressions with the Function constructor. This
// OVERRIDES the app-wide strict CSP (which has no nonce and no
// unsafe-eval) for the /health subtree alone; every other surface
// keeps the stricter policy.
func dashboardCSP(nonce string) string { return dashboard.RecommendedCSP(nonce) }

// newDashboard mounts the go-health-dashboard at /health: registers
// the dashboard in the container (shutdown cascade + a non-critical
// "dashboard" health check reporting pusher staleness) and returns the
// subtree handler with per-request nonce generation and the CSP
// override wrapped around it.
//
// Route namespacing: the dashboard's conventional kubelet probe paths
// (/healthz, /readyz, /startupz) are overridden to /health/livez,
// /health/readyz, /health/startupz — webphone owns the conventional
// probe triple at the root already, served from the SAME probe
// instance, so the aliases are plain duplicates, never a second truth.
// Favicon stays off ("" disables): webphone serves its own.
func newDashboard(injector *do.RootScope, probe *health.Probe, cfg config.Config) http.Handler {
	opts := []dashboard.Option{
		dashboard.WithTitle(cfg.Dashboard.Title),
		// The dashboard's own scoped Tailwind build (/assets/health.css,
		// sourced from the dashboard page + the templ-components subtrees
		// it renders). Without a CSSPath the library falls back to the
		// Tailwind Play CDN, which the CSP blocks.
		dashboard.WithCSSPath("/assets/health.css"),
		// Serve the pinned Datastar SDK from the embedded bundle — a
		// same-origin script, the only kind script-src 'self' allows.
		dashboard.WithEmbeddedDatastarSDK(),
		dashboard.WithTrend(trendSamples),
		dashboard.WithRateLimit(dashboardRateLimit.requests, dashboardRateLimit.window),
		dashboard.WithRoutes(dashboard.Routes{
			Dashboard:  "/health",
			SSE:        "/health/sse",
			Favicon:    "",
			Liveness:   "/health/livez",
			Readiness:  "/health/readyz",
			Startup:    "/health/startupz",
			Healthz:    "",
			Metrics:    "",
			Trend:      "/health/trend",
			Export:     "/health/export",
			Introspect: "",
			DatastarJS: "/health/datastar.js",
		}),
		dashboard.WithNonceExtractor(httputil.NonceFromRequest),
	}

	dash := dashboard.New(probe, opts...)
	// Register under a plain name (not the type name): the probe's
	// check list shows "dashboard", not "*dashboard.Dashboard". The
	// cascade shuts the pusher down regardless of naming.
	do.ProvideNamedValue(injector, "dashboard", dash)

	mux := http.NewServeMux()
	dash.RegisterRoutes(mux)

	// Per-request nonce → CSP override → dashboard routes. The nonce
	// middleware carries no CSPBuilder of its own: this subtree's
	// policy is the dashboard's verified one, not httputil's generic
	// recommendation.
	nonce := httputil.Nonce(httputil.NonceConfig{})
	return nonce(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", dashboardCSP(httputil.NonceFromRequest(r)))
		mux.ServeHTTP(w, r)
	}))
}
