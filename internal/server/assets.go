package server

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"strings"

	"github.com/larsartmann/webphone/internal/web/assets"
)

// assets serves the embedded static tree: the island's ES modules and
// stylesheet, the vendored sip.js bundle, the shell stylesheet, glue
// script, and theme preload. Everything is same-origin; Content-Types
// come from the file extensions via http.FileServer's FS-based serving.
func (h *handlers) assets() http.Handler {
	sub, err := fs.Sub(assets.FS(), "island")
	if err != nil {
		panic("assets: island subtree missing: " + err.Error())
	}
	fileServer := http.FileServerFS(sub)
	vendorServer := http.FileServerFS(assets.FS())

	mux := http.NewServeMux()
	// /assets/island/... → the island tree (app/*.js, style.css)
	mux.Handle("/assets/island/", http.StripPrefix("/assets/island/", fileServer))
	// /assets/vendor/... → vendored sip.min.js + license notice
	mux.Handle("/assets/vendor/", http.StripPrefix("/assets/", vendorServer))
	// /assets/app.css, /assets/shell.js, /assets/theme-preload.js → shell files
	mux.HandleFunc("GET /assets/app.css", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "app.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("GET /assets/shell.js", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "shell.js", "application/javascript; charset=utf-8")
	})
	mux.HandleFunc("GET /assets/theme-preload.js", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "theme-preload.js", "application/javascript; charset=utf-8")
	})
	// The templ-components Tailwind build (v4, layered — coexistence
	// verdict 2026-09-24): regenerated from the ADOPTED components'
	// sources whenever the library version bumps or a wave adopts more
	// components (recipe: docs/planning/2026-09-24_16-38_*.md).
	mux.HandleFunc("GET /assets/tw.css", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "tw.css", "text/css; charset=utf-8")
	})
	// The dashboard's own scoped Tailwind build: same generator, but
	// sourced from the go-health-dashboard page + the templ-components
	// subtrees IT renders (layout/display/feedback/utils/datastar). Kept
	// separate from tw.css so the app surface's build stays the proven
	// adopted-components set — neither page loads the other's classes.
	// Regenerate together with tw.css on templ-components bumps (recipe
	// in AGENTS.md, assets section).
	mux.HandleFunc("GET /assets/health.css", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "health.css", "text/css; charset=utf-8")
	})
	// Revalidate-always + content ETag + scoped gzip: a repeat visit skips
	// the payload transfer (304) while a redeploy is still picked up on the
	// next request. gzip wraps ONLY this static subtree — /events must stay
	// uncompressed (its SSE frames are flushed incrementally).
	return noStore(assetHeaders(gzipAssets(mux)))
}

// assetHeaders gives every embedded asset a strong content ETag so repeat
// visits revalidate with 304 instead of re-downloading. Cache-Control
// stays no-cache (see noStore): the browser may cache but MUST revalidate,
// so a redeploy is picked up immediately while the body transfer is
// skipped. The ETag is the sha256 of the served bytes, so it changes
// exactly when the content does.
func assetHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := strings.CutPrefix(r.URL.Path, "/assets/")
		if !ok || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		content, err := fs.ReadFile(assets.FS(), name)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		sum := sha256.Sum256(content)
		etag := `"` + hex.EncodeToString(sum[:]) + `"`
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("ETag", etag)
		if ifNoneMatch(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ifNoneMatch reports whether an If-None-Match header value selects etag.
// Handles the wildcard and comma-separated lists (weak W/ prefixes are
// compared as strong, which is safe for our stable content hashes).
func ifNoneMatch(header, etag string) bool {
	if header == "" {
		return false
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}

// gzipAssets compresses the static subtree when the client accepts gzip.
// It never sees /events (that handler is mounted outside this subtree),
// and it runs BELOW assetHeaders, so a 304 short-circuits before any
// Content-Encoding is set.
func gzipAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// gzipResponseWriter drops the Content-Length the file server would set
// (the compressed length differs) and streams the body through gzip.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (g *gzipResponseWriter) WriteHeader(status int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(status)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

func serveEmbedded(w http.ResponseWriter, r *http.Request, name, contentType string) {
	content, err := fs.ReadFile(assets.FS(), name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(content) //nolint:erraudit // best-effort write; the response is already committed
}

// noStore keeps development honest: a stale island bundle is the worst
// kind of bug to chase. Assets are served no-cache (must-revalidate) —
// the content ETag (assetHeaders) turns the revalidation into a bodyless
// 304, so a redeploy is picked up immediately without re-downloading
// unchanged bytes.
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

// favicon serves the embedded island favicon at the site root.
func (h *handlers) favicon(w http.ResponseWriter, r *http.Request) {
	serveEmbedded(w, r, "island/favicon.svg", "image/svg+xml")
}
