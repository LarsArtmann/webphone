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
// status), pinned in internal/textagent/family_test.go.
func TestSummaryWrapFamilyPin(t *testing.T) {
	wrapped := errorfamily.WrapTransientf(errors.New("dial refused"), "agent.summarize", "summarize call c1")
	errorfamilytest.AssertFamily(t, wrapped, errorfamily.Transient)
	errorfamilytest.AssertCode(t, wrapped, "agent.summarize")
}

// TestPushDraftWrapCodePins locks the draft fan-out's two wrap codes
// (the fire-and-forget goroutine's only greppable trail): the device
// list failure and the per-device send failure wrap Infrastructure
// under push.draft_devices / push.draft_send (summary_api.go).
func TestPushDraftWrapCodePins(t *testing.T) {
	devices := errorfamily.WrapInfrastructuref(errors.New("db down"), "push.draft_devices", "list draft push targets")
	errorfamilytest.AssertFamily(t, devices, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, devices, "push.draft_devices")

	send := errorfamily.WrapInfrastructuref(errors.New("ntfy 500"), "push.draft_send", "ship draft-ready to dev-1")
	errorfamilytest.AssertFamily(t, send, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, send, "push.draft_send")
}
