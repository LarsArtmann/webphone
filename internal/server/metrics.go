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
	// Threads and messages live in the journal-backed projections since
	// the A34 swap, but pre-import history still counts from the legacy
	// tables: until the A35 importer stamps its marker the two stores hold
	// DISJOINT row sets, so the honest total is the SUM. Once the marker
	// flips the journal holds every row and the legacy tables would
	// double-count — their contribution drops to zero by reading the same
	// flag the importer sets.
	imported, err := store.FlagPresent(r.Context(), h.deps.DB, store.FlagThreadsJournalImported)
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	legacyThreads, legacyMessages := counts.Threads, counts.Messages
	if imported {
		legacyThreads, legacyMessages = 0, 0
	}
	threadCount, err := h.deps.Messages.CountThreads(r.Context())
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	messageCount, err := h.deps.Messages.CountMessages(r.Context())
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	// Contacts and faxes live in the metaengine read models since the
	// T07/T11 swaps — the engine's native COUNT aggregate is their
	// count source now (a dead SQL table would freeze or 500 the
	// whole metrics surface).
	contactCount, err := h.deps.Contacts.Count(r.Context())
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	faxCount, err := h.deps.Faxes.Count(r.Context())
	if err != nil {
		http.Error(w, "counts unavailable", http.StatusInternalServerError)
		return
	}
	transcriptCount, err := h.deps.Transcripts.Count(r.Context())
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
	writeMetric("webphone_threads_total", "Stored conversation threads across all extensions.", "gauge", legacyThreads+threadCount)
	writeMetric("webphone_messages_total", "Stored messages (all directions) across all extensions.", "gauge", legacyMessages+messageCount)
	writeMetric("webphone_faxes_total", "Stored fax jobs across all extensions.", "gauge", faxCount)
	writeMetric("webphone_contacts_total", "Stored personal contacts across all extensions.", "gauge", contactCount)
	writeMetric("webphone_transcripts_total", "Stored live-call transcript segments across all extensions.", "gauge", transcriptCount)
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

	// Request-duration aggregates last: they describe the traffic that
	// produced the counts above (aggregates only, no path labels).
	h.latency.writePrometheus(&b)
	// The island beacon family renders only when the seam is on AND a
	// beacon has flowed (the recorder itself omits empty exposition).
	if h.island != nil {
		h.island.writePrometheus(&b)
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(b.String())) //nolint:erraudit // best-effort body write; nothing left to do on failure
}
