package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/larsartmann/webphone/internal/store"
)

// processStart is the metrics' clock zero; set when the server
// package loads (process start within padding).
var processStart = time.Now()

// metrics renders the Prometheus text exposition. AGGREGATES ONLY by
// design (plan T26a): counts and process facts, never per-extension
// data — a scraped metrics endpoint must not become a data leak, so
// the pinning test fails on any extension-like string in the body.
func (h *handlers) metrics(w http.ResponseWriter, r *http.Request) {
	counts, err := store.ReadCounts(r.Context(), h.deps.DB)
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	var b strings.Builder
	writeMetric := func(name, help, kind string, value any) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s %s\n%s %v\n", name, help, name, kind, name, value)
	}
	writeMetric("webphone_build_info", "Build version as a labeled constant.", "gauge",
		fmt.Sprintf(`{version="%s"}`, DisplayVersion()))
	writeMetric("webphone_uptime_seconds", "Seconds since process start.", "gauge",
		int64(time.Since(processStart).Seconds()))
	writeMetric("webphone_threads_total", "Stored conversation threads across all extensions.", "gauge", counts.Threads)
	writeMetric("webphone_messages_total", "Stored messages (all directions) across all extensions.", "gauge", counts.Messages)
	writeMetric("webphone_faxes_total", "Stored fax jobs across all extensions.", "gauge", counts.Faxes)
	writeMetric("webphone_contacts_total", "Stored personal contacts across all extensions.", "gauge", counts.Contacts)
	writeMetric("webphone_sessions_stored", "Session rows in the store (expired rows are swept lazily).", "gauge", counts.Sessions)
	// CRM integration observability: the resolver's upstream-outcome
	// counters (aggregates by construction — no numbers, no contact
	// names). The family is absent entirely when the integration is off,
	// so a disabled deploy never publishes zero-lines that read as
	// "CRM broken".
	if h.deps.CRM.Enabled() {
		hit, miss, failure := h.deps.CRM.LookupCounters()
		fmt.Fprintf(&b, "# HELP webphone_crm_lookups_total CRM number-resolution lookups by outcome (since process start).\n# TYPE webphone_crm_lookups_total counter\n")
		fmt.Fprintf(&b, "webphone_crm_lookups_total{outcome=\"hit\"} %d\n", hit)
		fmt.Fprintf(&b, "webphone_crm_lookups_total{outcome=\"miss\"} %d\n", miss)
		fmt.Fprintf(&b, "webphone_crm_lookups_total{outcome=\"failure\"} %d\n", failure)
	}
	// go-health probe outcomes (M18): aggregate counters by construction
	// (check names and statuses only). Absent until the first evaluation.
	h.health.render(&b)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(b.String())) //nolint:erraudit // best-effort body write; nothing left to do on failure
}
