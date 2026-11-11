// Package httpclient is the ONE home for outbound HTTP clients that must
// not share the process-global connection pool. http.DefaultTransport's
// idle-connection pool is process-wide: a parallel test's
// httptest.Server.Close "helpfully" retires idle connections on it, which
// can kill another seam's in-flight roundtrip mid-test (the flake class
// the crm repair of 2026-11-08 killed at one seam; every seam mints its
// own pool through here). Production behavior is unchanged: same stock
// transport settings, one private pool per client.
package httpclient

import (
	"net/http"
	"time"
)

// Isolated returns an *http.Client with the given timeout and its own
// clone of the default transport (own connection pool, stock settings
// otherwise).
func Isolated(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
	}
}
