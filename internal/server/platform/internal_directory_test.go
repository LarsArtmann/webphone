package platform

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMachineSurfaceAcceptsIPv6Loopback pins the ::1 half of the
// loopback posture: loopbackPeer's net.IP.IsLoopback handles the IPv6
// loopback, but no test ever proved it (every existing case rides an
// implicit 127.0.0.1 RemoteAddr). The stack's localhost forwarding can
// present [::1], and the machine-surface prologue must treat it as the
// same host, not a misdeployment.
func TestMachineSurfaceAcceptsIPv6Loopback(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:51000", "[::1]:51001"} {
		r := httptest.NewRequest(http.MethodGet, "/internal/directory/1001", nil)
		r.RemoteAddr = addr
		if !loopbackPeer(r) {
			t.Errorf("%s must read as a loopback peer", addr)
		}
	}
	r := httptest.NewRequest(http.MethodGet, "/internal/directory/1001", nil)
	r.RemoteAddr = "192.0.2.10:51002"
	if loopbackPeer(r) {
		t.Error("a remote peer must not read as loopback")
	}
}
