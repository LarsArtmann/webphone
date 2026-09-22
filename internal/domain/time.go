package domain

import "time"

// OrClock reports stamped, or the clock's answer when stamped is the zero
// time: an inbound event without a provider timestamp happened when we
// saw it. The clock stays lazy so a provider-supplied moment never pays
// for a time.Now call.
func OrClock(stamped time.Time, clock func() time.Time) time.Time {
	if stamped.IsZero() {
		return clock()
	}
	return stamped
}
