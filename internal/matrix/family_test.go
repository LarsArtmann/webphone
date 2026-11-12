package matrix

import (
	"strings"
	"net/http/httptest"
	"net/http"
	"io"
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
	c := New(ClientConfig{Homeserver: dead, UserID: "@w:example.org", Token: "tok"})

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

// TestOversizedAnswerIsItsOwnTransient (T18 F49): a homeserver answer
// beyond the read cap is its own Transient verdict (matrix.too_large),
// never a truncation-induced Infrastructure decode error — the
// receiver's fresh-sync reset keys on it.
func TestOversizedAnswerIsItsOwnTransient(t *testing.T) {
	filler := strings.Repeat(" ", answerCap+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, filler)
	}))
	t.Cleanup(server.Close)

	c := New(ClientConfig{Homeserver: server.URL, UserID: "@w:example.org", Token: "tok"})
	_, err := c.Sync(t.Context(), "s1", 0)
	errorfamilytest.AssertFamily(t, err, errorfamily.Transient)
	errorfamilytest.AssertCode(t, err, "matrix.too_large")
}
