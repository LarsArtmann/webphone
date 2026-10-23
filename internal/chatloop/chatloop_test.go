package chatloop

import (
	"context"
	"testing"
	"time"
)

func TestNextBackoffDoublesAndCaps(t *testing.T) {
	cases := []struct {
		in, max, want time.Duration
	}{
		{time.Second, 30 * time.Second, 2 * time.Second},
		{20 * time.Second, 30 * time.Second, 30 * time.Second},
		{time.Minute, 30 * time.Second, 30 * time.Second},
	}
	for _, tc := range cases {
		if got := NextBackoff(tc.in, tc.max); got != tc.want {
			t.Fatalf("NextBackoff(%s, %s) = %s, want %s", tc.in, tc.max, got, tc.want)
		}
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
