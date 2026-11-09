package server

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCspHeaderCarriesTheReportingLane pins the reporting lane: the strict
// CSP carries both the legacy report-uri and the report-to group, and the
// Reporting-Endpoints header names the group's target. A violation the
// browser sees must have somewhere to go, or the lane is decorative.
func TestCspHeaderCarriesTheReportingLane(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "report-uri /csp-report") {
		t.Fatalf("CSP lost the report-uri directive: %s", csp)
	}
	if !strings.Contains(csp, "report-to csp-endpoint") {
		t.Fatalf("CSP lost the report-to directive: %s", csp)
	}
	if got := resp.Header.Get("Reporting-Endpoints"); got != `csp-endpoint="/csp-report"` {
		t.Fatalf("Reporting-Endpoints = %q, want the csp-endpoint group", got)
	}
}

// TestCspReportJournalsTheViolation is the planted-violation probe: a
// nonce-less inline script report POSTed to the intake must land as one
// greppable journal line naming the directive and the source.
func TestCspReportJournalsTheViolation(t *testing.T) {
	prev := slog.Default()
	buf := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	ts := newTestServer(t)
	body := `{"csp-report":{"effective-directive":"script-src","blocked-uri":"inline","source-file":"https://home.lan/assets/evil.js","line-number":7,"document-uri":"https://home.lan/"}}`
	resp, err := http.Post(ts.URL+"/csp-report", "application/csp-report", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("csp-report status = %d, want 204", resp.StatusCode)
	}
	line := buf.String()
	for _, want := range []string{"csp_report", "script-src", "inline", "evil.js", "line=7"} {
		if !strings.Contains(line, want) {
			t.Fatalf("journal line %q lost %q", line, want)
		}
	}
}

// TestCspReportFences pins the intake's fences: oversize bodies die 413,
// malformed JSON stays best-effort 204 (report delivery is best-effort by
// contract; a broken report must not become a 400-log), GET is a 405, and
// flooding past the burst answers 429.
func TestCspReportFences(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Post(ts.URL+"/csp-report", "application/csp-report",
		strings.NewReader(fmt.Sprintf(`{"csp-report":{"blocked-uri":"%s"}}`, strings.Repeat("x", cspReportMaxBody))))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize report status = %d, want 413", resp.StatusCode)
	}

	resp, err = http.Post(ts.URL+"/csp-report", "application/csp-report", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("malformed report status = %d, want 204", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/csp-report")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET /csp-report status = %d, want 405", resp.StatusCode)
	}

	lastStatus := 0
	for range cspBurst + 5 {
		resp, err := http.Post(ts.URL+"/csp-report", "application/csp-report",
			strings.NewReader(`{"csp-report":{"effective-directive":"img-src"}}`))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		lastStatus = resp.StatusCode
		resp.Body.Close()
	}
	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("flood last status = %d, want 429 after burst %d", lastStatus, cspBurst)
	}
}

// TestCspReportDecodeKnowsTheBrowserShape guards the wire: the payload
// decodes the Reporting API's csp-report object and keeps unknown fields
// legal, so a browser adding keys never turns a report into a 400.
func TestCspReportDecodeKnowsTheBrowserShape(t *testing.T) {
	var payload cspReportPayload
	raw := `{"csp-report":{"document-uri":"https://home.lan/","violated-directive":"style-src","effective-directive":"style-src","blocked-uri":"inline","status-code":200,"unknown-future-field":true}}`
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("browser-shaped report failed to decode: %v", err)
	}
	if payload.Report.EffectiveDirective != "style-src" || payload.Report.BlockedURI != "inline" {
		t.Fatalf("decoded %+v", payload.Report)
	}
}

// TestReportingEndpointsMiddlewareSetsTheHeader pins the middleware arm
// directly (a planted header reader) so a chain reorder cannot silently
// drop the header while the CSP keeps referencing the group.
func TestReportingEndpointsMiddlewareSetsTheHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	reportingEndpointsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Reporting-Endpoints"); got != `csp-endpoint="/csp-report"` {
		t.Fatalf("Reporting-Endpoints = %q", got)
	}
}
