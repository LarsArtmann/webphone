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
