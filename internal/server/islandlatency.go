package server

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
)

// islandLatencyLegs is the CLOSED leg vocabulary of the island's
// latency beacons (the measure-first latency design,
// docs/planning/2026-11-08_latency-measure-first-design.md): per
// interval, the named legs a total splits into. The set is a server
// decision, not wire input — every beacon leg outside it is dropped,
// so a buggy or hostile client cannot grow the label cardinality of
// /metrics (AGGREGATES ONLY, the same posture as the request
// histogram: no extension, no remote, no free-form label).
var islandLatencyLegs = map[string]map[string]bool{
	// dial-press → first remote audio: where the user's wait actually
	// went. audio_gap is the honest leg — ringback vs instant-audio
	// deployments differ exactly there.
	"dial_to_audio": {"island_overhead": true, "signaling": true, "media_attach": true, "audio_gap": true},
	// spoken segment → text on screen: capture flush, the provider
	// round trip split at first byte, and the island's parse+insert.
	"asr_round_trip": {"capture_to_post": true, "post_to_first_byte": true, "first_byte_to_complete": true, "parse_to_insert": true},
}

// islandMaxLegMS bounds one leg's value: anything beyond a day is
// clock skew, not latency — the beacon is dropped rather than
// poisoning a bucket.
const islandMaxLegMS = 86_400_000

// islandASRBytesBuckets bound the segment-size histogram in bytes.
// The 25 MiB transcribe cap is the ceiling; the low end covers the
// ~4 s phone chunk the live capture produces.
var islandASRBytesBuckets = [...]float64{512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608, 16777216, 33554432}

// islandBucket is one labeled histogram series: non-cumulative bucket
// counts (the exposition accumulates), a sum and a total.
type islandBucket struct {
	bounds []float64
	counts []uint64
	sum    float64
	total  uint64
}

func newIslandBucket(bounds []float64) *islandBucket {
	return &islandBucket{
		bounds: bounds,
		counts: make([]uint64, len(bounds)+1),
	}
}

func (b *islandBucket) observe(v float64) {
	idx := len(b.bounds)
	for i, bound := range b.bounds {
		if v <= bound {
			idx = i
			break
		}
	}
	b.counts[idx]++ //nolint:branching-flow // counts is sized len(bounds)+1 and idx defaults to the overflow bucket, so the index is always in range
	b.sum += v
	b.total++
}

// write emits one cumulative histogram series under name. extra
// carries this series' fixed labels WITHOUT braces ("" for none, e.g.
// `interval="dial_to_audio",leg="signaling"`); the le= bound is
// appended here so the label list stays valid Prometheus text.
func (b *islandBucket) write(sb *strings.Builder, name, extra string) {
	bucketLabels := func() string {
		if extra == "" {
			return "{le="
		}
		return "{" + extra + ",le="
	}
	seriesLabels := func() string {
		if extra == "" {
			return ""
		}
		return "{" + extra + "}"
	}
	var cumulative uint64
	for i, bound := range b.bounds {
		cumulative += b.counts[i] //nolint:branching-flow // i ranges over bounds, which sizes counts
		fmt.Fprintf(sb, "%s_bucket%s\"%g\"} %d\n", name, bucketLabels(), bound, cumulative)
	}
	cumulative += b.counts[len(b.bounds)] //nolint:branching-flow // the overflow bucket exists: counts is sized len(bounds)+1
	fmt.Fprintf(sb, "%s_bucket%s\"+Inf\"} %d\n", name, bucketLabels(), cumulative)
	fmt.Fprintf(sb, "%s_sum%s %g\n", name, seriesLabels(), b.sum)
	fmt.Fprintf(sb, "%s_count%s %d\n", name, seriesLabels(), b.total)
}

type islandSeriesKey struct {
	interval string
	leg      string
}

// islandLatency folds the island's latency beacons into Prometheus
// histograms for /metrics. AGGREGATES ONLY: the labels are the fixed
// interval/leg vocabulary above — never an extension, a remote or any
// wire-supplied free-form value. The family renders only series with
// at least one observation (the CRM-counter posture: a deploy where
// nothing flowed publishes nothing that reads as broken).
//
// Safe for concurrent use.
type islandLatency struct {
	mu       sync.Mutex
	series   map[islandSeriesKey]*islandBucket
	asrBytes *islandBucket
}

func newIslandLatency() *islandLatency {
	return &islandLatency{series: map[islandSeriesKey]*islandBucket{}}
}

// RecordBeacon folds one beacon. It IS the error handling: an unknown
// interval, an unknown leg, a negative or absurd value, a negative
// size — any malformed beacon is dropped SILENTLY (instrumentation
// never alerts on itself, the design note's failure posture), so the
// method has no error return to wrap and no error family to register.
func (m *islandLatency) RecordBeacon(interval string, legsMS map[string]int64, segmentBytes *int64) {
	allowed := islandLatencyLegs[interval]
	if allowed == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, leg := range slices.Sorted(maps.Keys(legsMS)) {
		if !allowed[leg] {
			continue
		}
		ms := legsMS[leg]
		if ms < 0 || ms > islandMaxLegMS {
			continue
		}
		key := islandSeriesKey{interval: interval, leg: leg}
		bucket := m.series[key]
		if bucket == nil {
			bucket = newIslandBucket(latencyBuckets[:])
			m.series[key] = bucket
		}
		bucket.observe(float64(ms) / 1000)
	}
	if segmentBytes != nil && *segmentBytes >= 0 {
		if m.asrBytes == nil {
			m.asrBytes = newIslandBucket(islandASRBytesBuckets[:])
		}
		m.asrBytes.observe(float64(*segmentBytes))
	}
}

// writePrometheus appends the island latency histograms (and the ASR
// segment-size histogram) to the exposition. Series render in sorted
// (interval, leg) order so a scrape diff stays stable.
func (m *islandLatency) writePrometheus(sb *strings.Builder) {
	m.mu.Lock()
	series := make(map[islandSeriesKey]*islandBucket, len(m.series))
	maps.Copy(series, m.series)
	asrBytes := m.asrBytes
	m.mu.Unlock()

	if len(series) == 0 && asrBytes == nil {
		return
	}
	fmt.Fprintf(sb, "# HELP webphone_island_latency_seconds Island-measured latency legs: dial-press to first remote audio, and an ASR segment to text on screen (aggregates only; the leg vocabulary is server-fixed).\n")
	fmt.Fprintf(sb, "# TYPE webphone_island_latency_seconds histogram\n")
	keys := make([]islandSeriesKey, 0, len(series))
	for k := range series {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b islandSeriesKey) int {
		if a.interval != b.interval {
			return strings.Compare(a.interval, b.interval)
		}
		return strings.Compare(a.leg, b.leg)
	})
	for _, k := range keys {
		series[k].write(sb, "webphone_island_latency_seconds",
			fmt.Sprintf("interval=%q,leg=%q", k.interval, k.leg))
	}
	if asrBytes != nil {
		fmt.Fprintf(sb, "# HELP webphone_island_asr_segment_bytes Live ASR segment sizes posted to the transcribe seam (the size context every round-trip percentile needs).\n")
		fmt.Fprintf(sb, "# TYPE webphone_island_asr_segment_bytes histogram\n")
		asrBytes.write(sb, "webphone_island_asr_segment_bytes", "")
	}
}
