package matrix

import (
	"testing"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// TestFamilyPins locks the seam's error taxonomy: auth and permission
// verdicts are Rejections (retrying changes nothing), transport
// trouble and rate limits are Transient (the send-failure arm offers
// retry), foreign networks and malformed mxc URIs are Rejections at
// the door.
func TestFamilyPins(t *testing.T) {
	dead := "http://127.0.0.1:1"
	c := New(dead, "@w:example.org", "tok")

	_, transportErr := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkMatrix, "@peer:example.org"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, transportErr, errorfamily.Transient)
	errorfamilytest.AssertCode(t, transportErr, "matrix.request")

	_, networkErr := c.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkSMS, "+441632960961"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, networkErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, networkErr, "matrix.network")

	_, _, mxcErr := c.Download(t.Context(), "https://example.org/not-mxc")
	errorfamilytest.AssertFamily(t, mxcErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, mxcErr, "matrix.mxc")
}
