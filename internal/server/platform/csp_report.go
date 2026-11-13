package platform

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
)

// reportingEndpoints names the reporting endpoint group referenced by the
// CSP's report-to directive. Both ride every response through the
// reportingEndpoints middleware: a violation seen by ANY surface's browser
// (the strict shell CSP, the dashboard's unsafe-eval lane, the login nonce
// lane) flows to the one journal.
const reportingEndpoints = `csp-endpoint="/csp-report"`

// cspReportMaxBody caps a violation report: real reports are a few hundred
// bytes, so 4 KiB bounds a hostile flood without clipping honest ones.
const cspReportMaxBody = 4 << 10

// cspReportPayload mirrors the browser's violation report shape (the
// Reporting API's "csp-report" object). Only the fields the journal line
// needs are decoded; unknown fields stay legal so a browser adding keys
// never turns a report into a 400.
type cspReportPayload struct {
	Report struct {
		EffectiveDirective string `json:"effective-directive"`
		ViolatedDirective  string `json:"violated-directive"`
		BlockedURI         string `json:"blocked-uri"`
		SourceFile         string `json:"source-file"`
		LineNumber         int    `json:"line-number"`
		DocumentURI        string `json:"document-uri"`
	} `json:"csp-report"`
}

// cspReport receives CSP violation reports from browsers. Session-free by
// design — the browser delivers reports for the page it rendered, without
// credentials, and demanding a session or CSRF token would blind the very
// lane that surfaces a real XSS-class event. The fences are structural
// instead: the cspLimiter bucket (per peer host, like every flood budget),
// the 4 KiB body cap, and a journal that only ever grows one greppable line
// per accepted report. The answer is 204 for anything that parses or fails
// to parse — report delivery is best-effort by the Reporting API contract,
// and the browser discards the response either way.
func (p *Hub) CSPReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := http.MaxBytesReader(w, r.Body, cspReportMaxBody)
	var payload cspReportPayload
	if err := json.UnmarshalRead(body, &payload); err != nil {
		// Oversize dies inside Decode with MaxBytesError and gets the
		// honest 413; malformed JSON stays best-effort 204 — a broken
		// report must not turn the journal into a 400-log.
		if _, tooBig := errors.AsType[*http.MaxBytesError](err); tooBig {
			http.Error(w, "report too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	report := payload.Report
	directive := report.EffectiveDirective
	if directive == "" {
		directive = report.ViolatedDirective
	}
	slog.Info("csp_report",
		"directive", directive,
		"blocked", report.BlockedURI,
		"source_file", report.SourceFile,
		"line", report.LineNumber,
		"document", report.DocumentURI,
	)
	w.WriteHeader(http.StatusNoContent)
}

// reportingEndpointsMiddleware sets the Reporting-Endpoints header on every
// response, pairing the CSP's report-to directive with its target group.
// It rides the middleware chain next to SecurityHeaders so the strict shell
// CSP, the dashboard subtree and the login lane all carry it.
func ReportingEndpointsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Reporting-Endpoints", reportingEndpoints)
		next.ServeHTTP(w, r)
	})
}
