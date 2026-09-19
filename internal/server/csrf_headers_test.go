package server

import (
	"net/http"
	"strings"
	"testing"
)

// TestCSRFRefreshCacheSafety pins the cache contract on GET /api/csrf:
// the masked token pairs with one client's cookie, so the response must
// be uncachable (no-store) and, for intermediaries that ignore that,
// keyed per cookie.
func TestCSRFRefreshCacheSafety(t *testing.T) {
	server := newTestServer(t)
	resp, err := server.Client().Get(server.URL + "/api/csrf")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/csrf: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if vary := resp.Header.Get("Vary"); !strings.Contains(vary, "Cookie") {
		t.Errorf("Vary = %q, want it to name Cookie", vary)
	}
}
