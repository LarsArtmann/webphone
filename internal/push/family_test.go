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

// TestParseTransportVocabulary pins the Transport vocabulary door (the
// defined type's one parser): both live transports resolve, anything
// else — wrong name, wrong case, empty — is a Rejection carrying the
// push.transport_parse code (the registration API answers it 400).
func TestParseTransportVocabulary(t *testing.T) {
	for _, want := range []push.Transport{push.TransportNtfy, push.TransportFcm} {
		got, err := push.ParseTransport(string(want))
		if err != nil || got != want {
			t.Fatalf("ParseTransport(%q) = %q, %v; want %q, nil", string(want), got, err, string(want))
		}
	}
	for _, raw := range []string{"", "apns", "Ntfy", "webpush"} {
		_, err := push.ParseTransport(raw)
		if err == nil {
			t.Fatalf("ParseTransport(%q) accepted — the vocabulary is closed", raw)
		}
		errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
		if got := errorfamily.Code(err); got != "push.transport_parse" {
			t.Fatalf("ParseTransport(%q) code = %q, want push.transport_parse", raw, got)
		}
	}
}
