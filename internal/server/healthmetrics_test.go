// M18 pins: the go-health evaluation-hook counters render into /metrics
// as aggregates (check name + outcome only), appear only after the first
// probe evaluation, and count every check of every evaluation.
package server

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-health"
)

func TestHealthOutcomesRenderCountsByCheckAndOutcome(t *testing.T) {
	outcomes := newHealthOutcomes()
	outcomes.record(health.Response{Checks: map[string]health.Check{
		"sqlite":   {Status: health.StatusPass},
		"blob-dir": {Status: health.StatusFail, Error: "write probe"},
	}})
	outcomes.record(health.Response{Checks: map[string]health.Check{
		"sqlite": {Status: health.StatusPass},
	}})

	var b strings.Builder
	outcomes.render(&b)
	got := b.String()
	for _, want := range []string{
		"# TYPE webphone_health_checks_total counter",
		`webphone_health_checks_total{check="blob-dir",outcome="fail"} 1`,
		`webphone_health_checks_total{check="sqlite",outcome="pass"} 2`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render missing %q in:\n%s", want, got)
		}
	}

	var empty strings.Builder
	fresh := newHealthOutcomes()
	fresh.render(&empty)
	if empty.Len() != 0 {
		t.Errorf("fresh outcomes must render nothing, got:\n%s", empty.String())
	}
}

// TestStartupzFeedsTheMetricsCounters pins the end-to-end seam: a
// /startupz hit evaluates the probe (until its latch), the hook counts
// every check outcome, and /metrics publishes the family.
func TestStartupzFeedsTheMetricsCounters(t *testing.T) {
	server := newTestServer(t)

	for _, path := range []string{"/startupz", "/metrics"} {
		req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d %s", path, resp.StatusCode, body)
		}
		if path == "/metrics" {
			for _, want := range []string{
				`webphone_health_checks_total{check="sqlite",outcome="pass"}`,
				`webphone_health_checks_total{check="blob-dir",outcome="pass"}`,
			} {
				if !strings.Contains(string(body), want) {
					t.Errorf("metrics missing %q", want)
				}
			}
		}
	}
}
