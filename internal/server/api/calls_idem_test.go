package api

import (
	"testing"

	"github.com/larsartmann/webphone/internal/server/hooks"
)

// TestRecordCallIdemContract pins the call-journal key recorder: an
// empty key (client sent none) records nothing — the legacy
// never-dedupe shape — and a non-empty key is recorded. The handler's
// use of these semantics (record only on 204 outcomes, 502 stays
// retryable, extension-namespaced keys) is pinned end-to-end by
// TestAPICallLoggingContract in the server package's crm_test.go.
func TestRecordCallIdemContract(t *testing.T) {
	h := New(Deps{CallsIdem: hooks.NewGate(nil, hooks.CallIdempotencyTTL)})
	h.recordCallIdem("")
	if h.deps.CallsIdem.Seen("") {
		t.Fatal("an empty key must never be recorded: keyless reports never dedupe")
	}
	h.recordCallIdem("1001:uuid-1")
	if !h.deps.CallsIdem.Seen("1001:uuid-1") {
		t.Fatal("a recorded key must be seen")
	}
	if h.deps.CallsIdem.Seen("1001:uuid-2") {
		t.Fatal("a different key must not be seen")
	}
}
