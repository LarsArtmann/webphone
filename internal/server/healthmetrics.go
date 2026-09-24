package server

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/larsartmann/go-health"
)

// healthOutcomes counts go-health probe evaluations per check and outcome
// — the M18 telemetry seam, mirroring the CRM lookup counters. Only
// go-health evaluations flow here: /startupz evaluates on every request
// until its latch (so the family captures boot-time check failures, the
// operator-interesting events), while /healthz's continuous readiness is
// cqrs-htmx's own handler and never rides the probe. /livez is
// fetch-free by design and evaluates nothing.
type healthOutcomes struct {
	mu     sync.Mutex
	counts map[string]map[health.Status]int64
}

func newHealthOutcomes() *healthOutcomes {
	return &healthOutcomes{counts: make(map[string]map[health.Status]int64)}
}

// record is the health.WithEvaluationHook callback: fast (one lock, no
// I/O) per the library's contract — it runs on the evaluation path.
func (o *healthOutcomes) record(resp health.Response) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for name, check := range resp.Checks {
		if o.counts[name] == nil {
			o.counts[name] = make(map[health.Status]int64)
		}
		o.counts[name][check.Status]++
	}
}

// render appends the metric family, check- then outcome-sorted for
// stable output. Nothing is emitted before the first evaluation, so a
// fresh process publishes no zero-lines.
func (o *healthOutcomes) render(b *strings.Builder) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.counts) == 0 {
		return
	}
	fmt.Fprint(b, "# HELP webphone_health_checks_total go-health probe evaluations by check and outcome (startup evaluations until the latch; /healthz readiness does not ride the probe).\n# TYPE webphone_health_checks_total counter\n")
	for _, name := range slices.Sorted(maps.Keys(o.counts)) {
		for _, outcome := range slices.Sorted(maps.Keys(o.counts[name])) {
			fmt.Fprintf(b, "webphone_health_checks_total{check=%q,outcome=%q} %d\n", name, outcome, o.counts[name][outcome])
		}
	}
}
