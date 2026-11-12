package platform

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// The histogram's buckets are cumulative in the exposition even though the
// recorder stores non-cumulative counts: a metrics backend reparses these
// lines, so an off-by-one here silently misreports every latency panel.
func TestLatencyHistogramExpositionIsCumulative(t *testing.T) {
	l := NewLatencyRecorder()
	l.Record(http.MethodGet, "/", http.StatusOK, 7*time.Millisecond)     // le=0.01
	l.Record(http.MethodGet, "/", http.StatusNotFound, 2*time.Second)    // le=2.5
	l.Record(http.MethodGet, "/metrics", http.StatusOK, 100*time.Second) // excluded: self-scrape

	var b strings.Builder
	l.writePrometheus(&b)
	body := b.String()

	for _, want := range []string{
		"webphone_http_request_duration_seconds_bucket{le=\"0.005\"} 0",
		"webphone_http_request_duration_seconds_bucket{le=\"0.01\"} 1",
		"webphone_http_request_duration_seconds_bucket{le=\"0.025\"} 1",
		"webphone_http_request_duration_seconds_bucket{le=\"1\"} 1",
		"webphone_http_request_duration_seconds_bucket{le=\"2.5\"} 2",
		"webphone_http_request_duration_seconds_bucket{le=\"+Inf\"} 2",
		"webphone_http_request_duration_seconds_count 2",
		"webphone_http_requests_total{status_class=\"2xx\"} 1",
		"webphone_http_requests_total{status_class=\"4xx\"} 1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
	// The 100 s self-scrape is the difference between the sum of the two
	// recorded observations (2.007 s) and anything larger.
	if !strings.Contains(body, "webphone_http_request_duration_seconds_sum 2.007") {
		t.Errorf("the 100 s self-scrape must not be recorded:\n%s", body)
	}
}
