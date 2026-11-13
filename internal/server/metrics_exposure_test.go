package server

import (
	"net/http"
	"strings"
	"testing"
)

// The exposition is AGGREGATES ONLY, like every other family on /metrics: a
// scrape must never reveal which paths were requested.
// The exposition is AGGREGATES ONLY over the live mux: a scrape must
// never reveal which paths were requested (the recorder-level pin moved
// with the platform carve; this is the mux-level twin).
func TestLatencyMetricsCarryNoPathsOrIdentifiers(t *testing.T) {
	c := signIn(t, newTestServer(t))

	if resp, _ := c.do(http.MethodGet, "/nope", nil, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("probe: status %d, want 404", resp.StatusCode)
	}
	resp, body := c.do(http.MethodGet, "/metrics", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics: %d", resp.StatusCode)
	}
	page := string(body)
	for _, want := range []string{
		"webphone_http_request_duration_seconds_count",
		"webphone_http_request_duration_seconds_bucket{le=\"+Inf\"}",
		"webphone_http_requests_total{status_class=\"4xx\"}",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("metrics missing %s:\n%s", want, page)
		}
	}
	for _, leak := range []string{"/nope", "1001", "method="} {
		if strings.Contains(page, leak) {
			t.Errorf("latency metrics leak %q — aggregates only:\n%s", leak, page)
		}
	}
}
