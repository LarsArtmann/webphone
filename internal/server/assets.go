package server

import (
	"io/fs"
	"net/http"

	"github.com/larsartmann/webphone/internal/web/assets"
)

// assets serves the embedded static tree: the island's ES modules and
// stylesheet, the vendored sip.js bundle, the shell stylesheet and glue
// script. Everything is same-origin; Content-Types come from the file
// extensions via http.FileServer's FS-based serving.
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
	// /assets/app.css, /assets/shell.js → shell files
	mux.HandleFunc("GET /assets/app.css", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "app.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("GET /assets/shell.js", func(w http.ResponseWriter, r *http.Request) {
		serveEmbedded(w, r, "shell.js", "application/javascript; charset=utf-8")
	})
	return noStore(mux)
}

func serveEmbedded(w http.ResponseWriter, r *http.Request, name, contentType string) {
	content, err := fs.ReadFile(assets.FS(), name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(content) //nolint:erraudit // best-effort write; the response is already committed
}

// noStore keeps development honest: a stale island bundle is the worst
// kind of bug to chase. Static assets are tiny; caching buys nothing here.
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
