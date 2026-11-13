package api

import (
	"encoding/json/v2"
	"net/http"
)

// BeaconBodyLimit is the beacon route's own cap (mounted as the
// route's httputil.MaxBodySize in the server's route table): a beacon
// is two leg names and a handful of numbers — 2 KiB is generous. The
// island endpoint contract stays small-JSON; the transcript-sized
// limits deliberately do not apply here.
const BeaconBodyLimit = 2048

// IslandBeaconSink is the beacon route's seam: the server-owned
// histogram recorder (internal/server's islandLatency). The interface
// keeps internal/server/api free of the server package (the same
// narrow-seam pattern as Server), and a test stub folds beacons
// without a real /metrics exposition.
//
// The sink does the vocabulary validation (the fixed interval/leg
// sets) and IS its own error handling: a malformed beacon is dropped
// SILENTLY — instrumentation never alerts on itself, so the interface
// has no error return and no error family rides this route.
type IslandBeaconSink interface {
	RecordBeacon(interval string, legsMS map[string]int64, segmentBytes *int64)
}

// LatencyBeaconRequest is the POST /api/metrics/latency wire shape.
// Named (not inline) so the machine-API contract derives its schema
// from this type (internal/server/apidocs.go).
type LatencyBeaconRequest struct {
	// Interval names the measured interval ("dial_to_audio" |
	// "asr_round_trip"); unknown names are dropped server-side.
	Interval string `description:"The measured interval: dial_to_audio (dial-press to first remote audio) or asr_round_trip (spoken segment to text on screen)." json:"interval"`
	// LegsMs carries the interval's named legs in milliseconds; the
	// server folds each leg into its own labeled histogram.
	LegsMs map[string]int64 `description:"The interval's named legs in milliseconds (server-fixed vocabulary, e.g. island_overhead/signaling/media_attach/audio_gap for dial_to_audio)." json:"legs_ms"`
	// Bytes is the ASR segment's audio size in bytes — the size
	// context every round-trip percentile needs (a big segment is
	// supposed to be slower). Optional; only the ASR interval sends it.
	Bytes *int64 `description:"The segment's audio size in bytes (ASR beacons only)." json:"bytes,omitempty"`
}

// LatencyBeacon accepts one fire-and-forget island latency beacon and
// answers 204 unconditionally: the endpoint is instrumentation, not a
// control surface. A malformed body, an unknown interval, a junk leg —
// every failure is a silent 204 (the design note's failure posture:
// alerting on the instrument would be the alarm-cries-wolf class), so
// there is deliberately no 4xx arm and no error family here. The
// route is session-gated like every island endpoint, because an
// unauthenticated metrics sink would let a stranger shape the
// operator's latency distributions.
func (h *Handlers) LatencyBeacon(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.deps.Server.RequireSession(w, r); !ok {
		return
	}
	if h.deps.IslandLatency == nil {
		http.NotFound(w, r)
		return
	}
	var body LatencyBeaconRequest
	if err := json.UnmarshalRead(r.Body, &body); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.deps.IslandLatency.RecordBeacon(body.Interval, body.LegsMs, body.Bytes)
	w.WriteHeader(http.StatusNoContent)
}
