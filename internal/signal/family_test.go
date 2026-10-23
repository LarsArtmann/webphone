package signal

import (
	"path/filepath"
	"testing"

	errorfamilytest "github.com/larsartmann/go-error-family/errorfamilytest"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/gateway"
)

// TestFamilyPins locks the seam's error taxonomy: an unreachable
// daemon is Transient (the socket may come back), daemon RPC errors
// are Rejections (the daemon's own verdicts), foreign networks and
// unreadable attachments refuse at the door.
func TestFamilyPins(t *testing.T) {
	dead := New(filepath.Join(t.TempDir(), "absent.sock"), "+4915112345678", nil)
	_, connectErr := dead.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkSignal, "+491599887766"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, connectErr, errorfamily.Transient)
	errorfamilytest.AssertCode(t, connectErr, "signal.connect")

	_, networkErr := dead.SendMessage(t.Context(), gateway.OutboundMessage{
		To: domain.MustParseAddress(domain.NetworkSMS, "+441632960961"), Body: "x",
	})
	errorfamilytest.AssertFamily(t, networkErr, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, networkErr, "signal.network")

	_, readErr := dead.SendMessage(t.Context(), gateway.OutboundMessage{
		To:          domain.MustParseAddress(domain.NetworkSignal, "+491599887766"),
		Attachments: []gateway.OutboundAttachment{{Name: "gone.png", MimeType: "image/png", Path: filepath.Join(t.TempDir(), "gone.png")}},
	})
	errorfamilytest.AssertFamily(t, readErr, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, readErr, "signal.attachment_read")
}
