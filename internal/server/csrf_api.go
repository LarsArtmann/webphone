package server

import (
	"encoding/json/v2"
	"net/http"

	"github.com/larsartmann/httputil"
)

// refreshCSRF hands the island the fresh masked CSRF token after the login
// rotation deleted the old cookie. Riding inside the CSRF middleware, this
// GET makes nosurf regenerate the cookie (missing cookie ⇒ new token) and
// exposes its masked token; the island swaps it into the meta tag and the
// body's hx-headers (session.js) so every later POST validates again.
// The answer is a random token only: no session required, and cross-origin
// JavaScript cannot read it (same-origin policy), so GET-open is safe.
func (h *handlers) refreshCSRF(w http.ResponseWriter, r *http.Request) {
	token := httputil.CSRFTokenFromRequest(r)
	if token == "" {
		http.Error(w, "no CSRF token available", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Cache-safety: the answer is bound to THIS client's csrf_token cookie
	// (nosurf's masked pairing), so a shared cache storing it could hand
	// one browser's token to another. `no-store` forbids caching outright;
	// `Vary: Cookie` additionally tells any intermediary that violates
	// no-store to key on the cookie. The consuming stack's vhost carries
	// no proxy_cache, so in practice nothing caches this route today;
	// the headers make the contract explicit rather than incidental.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Vary", "Cookie")
	_ = json.MarshalWrite(w, map[string]string{"token": token}) //nolint:erraudit // best-effort write; the paired cookie is already set
}
