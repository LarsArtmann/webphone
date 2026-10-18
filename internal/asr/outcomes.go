package asr

import (
	"context"
	"sync"
	"time"
)

// latencyRing is the rolling window behind AvgLatency: recent round
// trips only, so one marathon voicemail clip cannot skew the average
// the operator reads on the health surface.
const latencyRing = 32

// OutcomeSnapshot is the provider's recent-outcome state: the health
// leg's entire truth. It is deliberately LOCAL — no probe traffic is
// ever generated; the numbers describe the transcriptions that already
// happened.
type OutcomeSnapshot struct {
	Calls       uint64
	Failures    uint64
	AvgLatency  time.Duration
	LastSuccess time.Time
	LastError   string
	LastErrorAt time.Time
}

// outcomes tracks every Transcribe round trip for both provider wires
// (OpenAI-compatible and Google): count, failures, a rolling latency
// average, and the timestamps of the last success and last failure.
// Failing-since-last-success is the health verdict's only signal —
// see healthCheck.
type outcomes struct {
	mu         sync.Mutex
	calls      uint64
	failures   uint64
	ring       [latencyRing]time.Duration
	ringFilled int
	ringNext   int
	lastOK     time.Time
	lastErr    string
	lastErrAt  time.Time
}

// record observes one round trip. err nil = success.
func (o *outcomes) record(latency time.Duration, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.calls++
	o.ring[o.ringNext] = latency
	o.ringNext = (o.ringNext + 1) % latencyRing
	if o.ringFilled < latencyRing {
		o.ringFilled++
	}
	if err != nil {
		o.failures++
		o.lastErr = err.Error()
		o.lastErrAt = time.Now()
		return
	}
	o.lastOK = time.Now()
}

// snapshot copies the current state.
func (o *outcomes) snapshot() OutcomeSnapshot {
	o.mu.Lock()
	defer o.mu.Unlock()
	var total time.Duration
	for i := range o.ringFilled {
		total += o.ring[i]
	}
	avg := time.Duration(0)
	if o.ringFilled > 0 {
		avg = total / time.Duration(o.ringFilled)
	}
	return OutcomeSnapshot{
		Calls:       o.calls,
		Failures:    o.failures,
		AvgLatency:  avg,
		LastSuccess: o.lastOK,
		LastError:   o.lastErr,
		LastErrorAt: o.lastErrAt,
	}
}

// HealthCheck is the OPTIONAL non-critical health leg (#22): it never
// probes the provider — it reports whether recent REAL transcriptions
// succeeded. Healthy verdicts: no traffic yet (nothing observed), or a
// success after the last failure (recovered). Degraded: failures with
// no success since — the error names the last cause and its age so the
// dashboard card answers "why is transcription down" without a
// round trip.
func (o *outcomes) HealthCheck(_ context.Context) error {
	s := o.snapshot()
	if s.Calls == 0 || s.Failures == 0 {
		return nil
	}
	if s.LastSuccess.After(s.LastErrorAt) {
		return nil
	}
	age := time.Since(s.LastErrorAt).Round(time.Second)
	return asrDegradedError{snapshot: s, age: age}
}

// asrDegradedError renders the degraded verdict; a distinct type so
// tests (and any future triage) can classify it without string
// matching.
type asrDegradedError struct {
	snapshot OutcomeSnapshot
	age      time.Duration
}

func (e asrDegradedError) Error() string {
	return "transcriptions failing since the last success: last error " + e.age.String() + " ago: " + e.snapshot.LastError
}
