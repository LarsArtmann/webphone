package server

import (
	"testing"

	"github.com/larsartmann/httputil/httpspec"
)

// TestHTTPSpectChainConformance runs httputil's 19-spec behavioral
// suite against the full New() chain (enrichment → request log →
// security headers → recovery → routes). Deliberate divergences are
// skip-listed HERE with their reason — each skip is a documented
// product decision, not an unknown failure.
func TestHTTPSpectChainConformance(t *testing.T) {
	server := newTestServer(t)
	httpspec.Run(t, server.handler)
}
