package chatloop

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// TestNextDelayJittersWithinTheLadder pins the go-retry contract the
// receivers inherit: attempt n waits the exponential base (1s·2^(n−1),
// capped) plus up to 50% additive jitter — never below the base, never
// above min(1.5·base, max). Attempt 6's base exceeds the 30s cap, so
// the whole window collapses onto the cap exactly.
func TestNextDelayJittersWithinTheLadder(t *testing.T) {
	cases := []struct {
		attempt int
		max     time.Duration
		lo, hi  time.Duration
	}{
		{1, 30 * time.Second, time.Second, 1500 * time.Millisecond},
		{2, 30 * time.Second, 2 * time.Second, 3 * time.Second},
		{3, 30 * time.Second, 4 * time.Second, 6 * time.Second},
		{5, 30 * time.Second, 16 * time.Second, 24 * time.Second},
		{6, 30 * time.Second, 30 * time.Second, 30 * time.Second},
	}
	for _, tc := range cases {
		for range 50 {
			got := nextDelay(tc.attempt, tc.max)
			if got < tc.lo || got > tc.hi {
				t.Fatalf("nextDelay(%d) = %s, want within [%s, %s]", tc.attempt, got, tc.lo, tc.hi)
			}
		}
	}
}

// TestBackoffStopsOnRejection pins the loud stop: a Rejection-classified
// failure ends the loop immediately (alive=false, no attempt consumed,
// no sleep) with an Error line naming the permanence.
func TestBackoffStopsOnRejection(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	err := errorfamily.NewRejection("matrix.token", "matrix: token revoked")

	attempt, alive := Backoff(t.Context(), log, "matrix: sync", err, 3, 30*time.Second)

	if alive {
		t.Fatal("a Rejection must stop the loop")
	}
	if attempt != 3 {
		t.Fatalf("attempt = %d, want 3 (a permanent failure consumes nothing)", attempt)
	}
	for _, want := range []string{"ERROR", "matrix: sync", "permanent"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log missing %q: %s", want, buf.String())
		}
	}
}

// TestBackoffArmsNextAttemptOnTransient pins the retry path: a
// Transient-classified failure warns, waits the attempt's jittered
// delay (shrunk here), and returns the next attempt number; a canceled
// context during that wait ends the loop.
func TestBackoffArmsNextAttemptOnTransient(t *testing.T) {
	oldWait := firstWait
	firstWait = time.Millisecond
	t.Cleanup(func() { firstWait = oldWait })

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	err := errorfamily.WrapTransient(context.Canceled, "signal.connect", "signal: dial")

	attempt, alive := Backoff(t.Context(), log, "signal: daemon unreachable", err, 1, 30*time.Second)
	if !alive || attempt != 2 {
		t.Fatalf("alive=%v attempt=%d, want true/2 after a transient failure", alive, attempt)
	}
	if !bytes.Contains(buf.Bytes(), []byte("WARN")) {
		t.Errorf("a retryable failure must warn: %s", buf.String())
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	attempt, alive = Backoff(ctx, log, "signal: daemon unreachable", err, attempt, 30*time.Second)
	if alive {
		t.Fatal("a canceled context must end the loop")
	}
	if attempt != 2 {
		t.Fatalf("attempt = %d, want 2 (a canceled wait consumes nothing)", attempt)
	}
}

func TestSleepEndsOnContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if Sleep(ctx, time.Minute) {
		t.Fatal("a canceled context must end the sleep immediately")
	}
	if !Sleep(t.Context(), time.Millisecond) {
		t.Fatal("a live context must let the nap finish")
	}
}

func TestDefaultResolvesNil(t *testing.T) {
	if Default(nil) == nil {
		t.Fatal("nil must resolve to the slog default")
	}
	log := Default(nil)
	if Default(log) != log {
		t.Fatal("a set logger must pass through unchanged")
	}
}
