package hooks

import (
	"testing"
	"time"
)

// TestIdemStoreSeenRecord pins the dedupe primitives: a key is a replay
// only after Record() (failed attempts are never recorded), and an
// expired entry is no longer a replay.
func TestIdemStoreFirstTime(t *testing.T) {
	s := newIdemStore(50 * time.Millisecond)
	if s.Seen("k") {
		t.Fatal("unrecorded key reported as replay")
	}
	s.Record("k")
	if !s.Seen("k") {
		t.Fatal("recorded key not reported as replay")
	}
	time.Sleep(60 * time.Millisecond)
	if s.Seen("k") {
		t.Fatal("expired entry still reported as replay")
	}
}
