package api

import (
	"errors"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
)

// TestSummaryWrapFamilyPin locks the summarize seam's own wrap: the
// transport failure the api handler wraps around an unreachable agent
// classifies Transient under the agent.summarize code. Agent VERDICTS
// pass through unwrapped (they carry textagent.http + their own
// status), pinned in internal/textagent/family_test.go. The draft
// fan-out that follows is fire-and-forget (`_ = Push.Fan`): its
// failures surface as the Fanout's greppable warns under
// push.registry / push.transport + the transports' push.ntfy_send /
// push.fcm_send — pinned in internal/push, not here (the historical
// push.draft_devices / push.draft_send codes named no production
// constructor and were deleted with this file's ghost pin).
func TestSummaryWrapFamilyPin(t *testing.T) {
	wrapped := errorfamily.WrapTransientf(errors.New("dial refused"), "agent.summarize", "summarize call c1")
	errorfamilytest.AssertFamily(t, wrapped, errorfamily.Transient)
	errorfamilytest.AssertCode(t, wrapped, "agent.summarize")
}
