package domain

import (
	"testing"
	"time"
)

// TestOrClockPinsTheZeroFallback pins the inbound-timestamp fallback:
// a provider-stamped moment passes through WITHOUT consulting the
// clock (lazy: real timestamps never pay for a time.Now), the zero
// time means "happened when we saw it" and takes the clock's answer.
func TestOrClockPinsTheZeroFallback(t *testing.T) {
	stamped := time.Date(2026, 9, 23, 4, 29, 0, 0, time.UTC)
	called := false
	clock := func() time.Time {
		called = true
		return stamped.Add(time.Hour)
	}
	if got := OrClock(stamped, clock); got != stamped {
		t.Fatalf("stamped moment must pass through: %v", got)
	}
	if called {
		t.Fatal("the clock must stay lazy when a timestamp is present")
	}
	zero := time.Time{}
	if got := OrClock(zero, clock); got != stamped.Add(time.Hour) {
		t.Fatalf("zero timestamp must take the clock's answer: %v", got)
	}
	if !called {
		t.Fatal("the clock must be consulted for the zero timestamp")
	}
}
