// Push seam classification pins: the disabled sentinel is
// Infrastructure (nil deps), a rejected credential is a Rejection (the
// operator's to fix), a dead device token is Infrastructure (prune and
// move on) — and the wraps callers add survive.
package push_test

import (
	"fmt"
	"testing"

	"github.com/larsartmann/go-error-family"
	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/larsartmann/webphone/internal/push"
)

func TestSentinelsClassify(t *testing.T) {
	errorfamilytest.AssertFamily(t, push.ErrDisabled, errorfamily.Infrastructure)
	errorfamilytest.AssertFamily(t, push.ErrUnauthorized, errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, push.ErrUnregistered, errorfamily.Infrastructure)
	errorfamilytest.AssertFamily(t, fmt.Errorf("ring: %w", push.ErrUnauthorized), errorfamily.Rejection)
	errorfamilytest.AssertFamily(t, fmt.Errorf("ring: %w", push.ErrUnregistered), errorfamily.Infrastructure)
}
