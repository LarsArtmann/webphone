// Throwaway Tailwind-coexistence spike (plan M7, 2026-09-24). The
// route renders templ-components beside real webphone surfaces with the
// library CSS toggled by ?tw= so an A/B computed-style probe can rule
// on coexistence. Deleted with spike.templ once the M7 verdict lands.
package server

import (
	"net/http"
	"time"

	"github.com/larsartmann/webphone/internal/web/views"
)

func (h *handlers) spikeTailwind(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireSession(w, r); !ok {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	withTailwind := r.URL.Query().Get("tw") != "0"
	now := time.Now()
	if err := views.SpikeTailwind(withTailwind, now).Render(r.Context(), w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
